package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/geobus"
	"github.com/wneessen/localweather/internal/geobus/provider/coordinates_file"
	"github.com/wneessen/localweather/internal/geobus/provider/geoip"
	"github.com/wneessen/localweather/internal/http"
)

const (
	SubID = "geobus-update"
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
		provider = append(provider, coordinates_file.NewLocationFileProvider(s.conf.Geobus.CoordinatesFile))
	}

	if !s.conf.Geobus.DisableGeoIP {
		gip, err := geoip.NewGeoIPProvider(httpClient)
		if err != nil {
			return nil, fmt.Errorf("failed to create GeoIP provider: %w", err)
		}
		provider = append(provider, gip)
	}

	/*
		if !s.config.GeoLocation.DisableCitynameFile {
			cnf, err := cityname_file.NewCitynameFileProvider(s.config.GeoLocation.CitynameFile, s.geocoder)
			if err != nil {
				return nil, fmt.Errorf("failed to create cityname file provider: %w", err)
			}
			provider = append(provider, cnf)
		}

		if !s.config.GeoLocation.DisableGPSD {
			provider = append(provider, gpsd.NewGeolocationGPSDProvider())
		}

		if !s.config.GeoLocation.DisableGeoIP {
			gip, err := geoip.NewGeolocationGeoIPProvider(httpClient)
			if err != nil {
				return nil, fmt.Errorf("failed to create GeoIP provider: %w", err)
			}
			provider = append(provider, gip)
		}

		if !s.config.GeoLocation.DisableGeoAPI {
			gap, err := geoapi.NewGeolocationGeoAPIProvider(httpClient)
			if err != nil {
				return nil, fmt.Errorf("failed to create GeoAPI provider: %w", err)
			}
			provider = append(provider, gap)
		}

		if !s.config.GeoLocation.DisableICHNAEA {
			mls, err := ichnaea.NewGeolocationICHNAEAProvider(httpClient)
			if err != nil {
				s.logger.Error("failed to create ICHNAEA provider", logger.Err(err))
			} else {
				provider = append(provider, mls)
			}
		}

	*/
	if len(provider) == 0 {
		return nil, apperror.ErrNoGeobusProvider
	}

	return provider, nil
}

// processGeobusUpdate subscribes to geolocation updates, processes location data, and updates the
// service state accordingly.
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
				slog.Float64("latitude", r.Latitude),
				slog.Float64("longitude", r.Longitude),
				slog.Float64("altitude", r.Altitude),
				slog.Float64("accuracy", r.Accuracy.Float64()),
				slog.String("provider", r.Provider))
		}
	}
}
