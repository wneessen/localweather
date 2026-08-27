package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-co-op/gocron/v2"

	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/geocode"
	"github.com/wneessen/localweather/internal/log"
	"github.com/wneessen/localweather/internal/weather"
)

// Server represents the main application server, managing HTTP services, cron jobs, metrics, and database interactions.
type Server struct {
	conf        *config.Config
	geobus      *geobus.Service
	geocoder    geocode.Geocoder
	geobusUnsub func()
	log         *log.Logger
	httpserv    *http.Server
	mux         chi.Router
	cron        gocron.Scheduler
	weatherLock sync.RWMutex
	weather     weather.Provider
}

type Params struct {
	Log  *log.Logger
	Cron gocron.Scheduler
	/*
		Queries *model.Queries
		PgxPool *pgxpool.Pool

	*/
}

func New(params Params, conf *config.Config) *Server {
	if params.Log == nil {
		params.Log = log.New(conf)
	}

	server := &Server{
		conf: conf,
		log:  params.Log,
		mux:  chi.NewMux(),
		cron: params.Cron,
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

	s.log.Info("starting cron scheduler")
	if err := s.cronjobs(ctx); err != nil {
		return fmt.Errorf("failed to set up cron jobs: %w", err)
	}

	s.log.Info("selecting geocoder provider")
	if err := s.initGeocoder(); err != nil {
		return fmt.Errorf("failed to select geocoding provider: %w", err)
	}

	s.log.Info("selecting weather provider")
	if err := s.initWeather(); err != nil {
		return fmt.Errorf("failed to select weather provider: %w", err)
	}

	s.log.Info("starting geobus service")
	if err := s.startGeobus(ctx); err != nil {
		return fmt.Errorf("failed to start geobus provider: %w", err)
	}

	s.log.Info("starting http backend", "listen_addr", s.conf.ListenAddr())
	s.httpRoutes(ctx)
	if err := s.httpserv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start http server: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the server by stopping HTTP services, unregistering metrics, and halting the scheduler.
func (s *Server) Stop() error {
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
