// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/vorlif/spreak"

	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/database/model"
	"github.com/wneessen/localweather/internal/formatter"
	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/geocode"
	"github.com/wneessen/localweather/internal/log"
	"github.com/wneessen/localweather/internal/static"
	"github.com/wneessen/localweather/internal/weather"
)

// Server represents the main application server, managing HTTP services, cron jobs, metrics, and database interactions.
type Server struct {
	conf         *config.Config
	cron         gocron.Scheduler
	db           *sql.DB
	fmt          *formatter.Formatter
	geobus       *geobus.Service
	geocoder     geocode.Geocoder
	geobusUnsub  func()
	httpserv     *http.Server
	log          *log.Logger
	mux          chi.Router
	queries      *model.Queries
	weather      weather.Provider
	weatherJobID uuid.UUID
	weatherLock  sync.RWMutex
	t            *spreak.Localizer
}

type Params struct {
	Cron      gocron.Scheduler
	DB        *sql.DB
	Log       *log.Logger
	Queries   *model.Queries
	Localizer *spreak.Localizer
}

func New(params Params, conf *config.Config) *Server {
	if params.Log == nil {
		params.Log = log.New(conf)
	}

	server := &Server{
		conf:    conf,
		cron:    params.Cron,
		db:      params.DB,
		log:     params.Log,
		mux:     chi.NewMux(),
		queries: params.Queries,
		t:       params.Localizer,
	}

	server.httpserv = &http.Server{
		Addr:              conf.ListenAddr(),
		Handler:           server.mux,
		ReadTimeout:       conf.Server.Timeout,
		ReadHeaderTimeout: conf.Server.Timeout,
		WriteTimeout:      conf.Server.Timeout,
		IdleTimeout:       conf.Server.Timeout,
	}
	return server
}

// Start initializes and starts the server, including cron jobs, metrics registration, and HTTP server setup.
func (s *Server) Start(ctx context.Context) error {
	s.log.Info("starting localweather service")

	s.log.Debug("starting cron scheduler")
	if err := s.cronjobs(ctx); err != nil {
		return fmt.Errorf("failed to set up cron jobs: %w", err)
	}

	s.log.Debug("selecting geocoder provider")
	if err := s.initGeocoder(); err != nil {
		return fmt.Errorf("failed to select geocoding provider: %w", err)
	}

	s.log.Debug("selecting weather provider")
	if err := s.initWeather(); err != nil {
		return fmt.Errorf("failed to select weather provider: %w", err)
	}

	s.log.Debug("starting geobus service")
	if err := s.startGeobus(ctx); err != nil {
		return fmt.Errorf("failed to start geobus provider: %w", err)
	}

	s.log.Debug("creating formatter")
	form, err := formatter.New(s.conf, s.t)
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}
	s.fmt = form

	s.log.Debug("initializing sleep/resume detection")
	go s.monitorSleepResume(ctx)

	s.log.Info("starting http backend", "listen_addr", s.conf.ListenAddr())
	s.httpRoutes(ctx)
	s.fileServer(s.mux, "/static", http.FS(static.FS))
	if err := s.httpserv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start http server: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the server by stopping HTTP services, unregistering metrics, and halting the scheduler.
func (s *Server) Stop(_ context.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.log.Info("stopping http backend")
	if err := s.httpserv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop http server: %w", err)
	}

	s.log.Info("unsubscribing from geobus updates")
	if s.geobusUnsub != nil {
		s.geobusUnsub()
	}

	s.log.Info("clearing current location")
	if err := s.queries.ClearCurrentAddress(ctx); err != nil {
		return fmt.Errorf("failed to clear current address: %w", err)
	}

	s.log.Info("stopping scheduler")
	if err := s.cron.StopJobs(); err != nil {
		return fmt.Errorf("failed to stop scheduler jobs: %w", err)
	}
	if err := s.cron.Shutdown(); err != nil {
		return fmt.Errorf("failed to shut down scheduler: %w", err)
	}

	s.log.Info("localweather service gracefully stopped")
	return nil
}
