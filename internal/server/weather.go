// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/wneessen/localweather/internal/http"
	"github.com/wneessen/localweather/internal/weather"
	openmeteo "github.com/wneessen/localweather/internal/weather/provider/open-meteo"
)

// initWeather initializes the weather provider for the server and logs the selected provider.
func (s *Server) initWeather() error {
	weatherProvider, err := s.selectWeatherProvider()
	if err != nil {
		return err
	}
	s.weather = weatherProvider
	s.log.Debug("weather provider initialized", slog.String("provider", s.geocoder.Name()))
	return nil
}

// selectWeatherProvider determines the appropriate weather provider based on the configuration and initializes it.
// It returns the selected weather provider or an error if the configuration is invalid or initialization fails.
func (s *Server) selectWeatherProvider() (provider weather.Provider, err error) {
	switch strings.ToLower(s.conf.Weather.Provider) {
	case "open-meteo":
		provider, err = openmeteo.New(http.New(s.log), s.log, s.conf.Units)
		if err != nil {
			return provider, fmt.Errorf("failed to create Open-Meteo weather provider: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported weather provider: %s", s.conf.Weather.Provider)
	}
	return provider, nil
}

// forceWeatherUpdate triggers an immediate execution of the scheduled weather update job if it exists.
func (s *Server) forceWeatherUpdate(ctx context.Context) error {
	for _, job := range s.cron.Jobs() {
		if job.ID() == s.weatherJobID {
			if err := job.RunNow(); err != nil {
				return err
			}
		}
	}
	return nil
}
