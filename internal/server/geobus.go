package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/geobus/provider/cityname_file"
	"github.com/wneessen/localweather/internal/geobus/provider/coordinates_file"
	"github.com/wneessen/localweather/internal/geobus/provider/geoapi"
	"github.com/wneessen/localweather/internal/geobus/provider/geoip"
	"github.com/wneessen/localweather/internal/geobus/provider/gpsd"
	"github.com/wneessen/localweather/internal/geobus/provider/ichnaea"
	"github.com/wneessen/localweather/internal/http"
	"github.com/wneessen/localweather/internal/log"
)

const (
	SubID       = "geobus-update"
	burstWindow = 5 * time.Second
)

// startGeobus initializes the geobus service, sets up providers, subscribes to updates, and starts processing them.
func (s *Server) startGeobus(ctx context.Context) error {
	service, err := geobus.New(s.log)
	if err != nil {
		return fmt.Errorf("failed to create geobus service: %w", err)
	}
	s.geobus = service

	list, err := s.geobusProviderList()
	if err != nil {
		return fmt.Errorf("failed to create geobus provider list: %w", err)
	}

	geobus.TrackProviders(ctx, s.geobus, SubID, list...)
	sub, unsub := s.geobus.Subscribe(SubID, 1)
	go s.processGeobusUpdate(ctx, sub)
	s.geobusUnsub = unsub

	return nil
}

// geobusProviderList aggregates and initializes active geolocation providers based on configuration, returning a list.
func (s *Server) geobusProviderList() ([]geobus.Provider, error) {
	httpClient := http.New(s.log)
	var provider []geobus.Provider

	if !s.conf.Geobus.DisableCoordinatesFile {
		provider = append(provider, coordinates_file.NewCoordinatesFileProvider(s.conf.Geobus.CoordinatesFile, s.log))
	}

	if !s.conf.Geobus.DisableCitynameFile {
		cnf, err := cityname_file.NewCitynameFileProvider(s.conf.Geobus.CitynameFile, s.geocoder, s.log)
		if err != nil {
			return nil, fmt.Errorf("failed to create cityname file provider: %w", err)
		}
		provider = append(provider, cnf)
	}

	if !s.conf.Geobus.DisableGeoIP {
		gip, err := geoip.NewGeoIPProvider(httpClient, s.log)
		if err != nil {
			return nil, fmt.Errorf("failed to create GeoIP provider: %w", err)
		}
		provider = append(provider, gip)
	}

	if !s.conf.Geobus.DisableGeoAPI {
		gap, err := geoapi.NewGeoAPIProvider(httpClient, s.log)
		if err != nil {
			return nil, fmt.Errorf("failed to create GeoAPI provider: %w", err)
		}
		provider = append(provider, gap)
	}

	if !s.conf.Geobus.DisableICHNAEA {
		mls, err := ichnaea.NewICHNAEAProvider(httpClient, s.log)
		if err != nil {
			return nil, fmt.Errorf("failed to create GeoAPI provider: %w", err)
		}
		provider = append(provider, mls)
	}

	if !s.conf.Geobus.DisableGPSD {
		provider = append(provider, gpsd.NewGPSdProvider(s.log))
	}

	if len(provider) == 0 {
		return nil, apperror.ErrNoGeobusProvider
	}

	return provider, nil
}

// processGeobusUpdate subscribes to geolocation updates, processes location data, and updates the
// service state accordingly.
/*
func (s *Server) processGeobusUpdate(ctx context.Context, sub <-chan geobus.Result) {
	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-sub:
			if !ok {
				return
			}
			s.log.Debug("received geolocation update",
				slog.Float64("latitude", r.Coordinates.Latitude),
				slog.Float64("longitude", r.Coordinates.Longitude),
				slog.Float64("altitude", r.Coordinates.Altitude),
				slog.String("accuracy", r.Coordinates.Accuracy.String()),
				slog.String("provider", r.Provider))
			if err := s.updateCurrentLocation(ctx, r.Coordinates, r.Provider); err != nil {
				s.log.Error("failed to update current location", log.ErrAttr(err))
			}
		}
	}
}
*/
// processGeobusUpdate subscribes to geolocation updates, processes location data, and updates the
// service state accordingly. Updates arriving within geobusWindow are collected and only the one
// with the best accuracy is processed.
func (s *Server) processGeobusUpdate(ctx context.Context, sub <-chan geobus.Result) {
	var (
		best geobus.Result
		open bool
	)

	timer := time.NewTimer(burstWindow)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()

	process := func(r geobus.Result) {
		s.log.Debug("processing geolocation update",
			slog.Float64("latitude", r.Coordinates.Latitude),
			slog.Float64("longitude", r.Coordinates.Longitude),
			slog.Float64("altitude", r.Coordinates.Altitude),
			slog.String("accuracy", r.Coordinates.Accuracy.String()),
			slog.String("provider", r.Provider))

		if err := s.updateCurrentLocation(ctx, r.Coordinates, r.Provider); err != nil {
			s.log.Error("failed to update current location", log.ErrAttr(err))
		}

		// Run weather data update job based on the new location
		for _, job := range s.cron.Jobs() {
			if job.ID() == s.weatherJobID {
				if err := job.RunNow(); err != nil {
					s.log.Error("failed to run weather data update job", log.ErrAttr(err))
				}
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-sub:
			if !ok {
				if open {
					process(best)
				}
				return
			}

			s.log.Debug("received geolocation update",
				slog.Float64("latitude", r.Coordinates.Latitude),
				slog.Float64("longitude", r.Coordinates.Longitude),
				slog.Float64("altitude", r.Coordinates.Altitude),
				slog.String("accuracy", r.Coordinates.Accuracy.String()),
				slog.String("provider", r.Provider))
			if !open {
				best, open = r, true
				timer.Reset(burstWindow)
				continue
			}
			if r.Coordinates.Accuracy.Float64() < best.Coordinates.Accuracy.Float64() {
				best = r
			}
		case <-timer.C:
			open = false
			process(best)
			best = geobus.Result{} // drop reference
		}
	}
}
