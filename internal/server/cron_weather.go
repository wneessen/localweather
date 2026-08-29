package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/types"
)

func (s *Server) cronjobWeatherdataUpdate(ctx context.Context) {
	now := time.Now()
	const action = "weatherdata_update"

	s.weatherLock.Lock()
	defer s.weatherLock.Unlock()

	location, err := s.queries.CurrentAddress(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logJobCompletionWithError(action, now,
			fmt.Errorf("failed to retrieve current address from database: %w", err))
		return
	}
	if location.ID == 0 {
		s.logJobCompletionWithError(action, now, apperror.ErrLocationNotSet)
		return
	}

	data, err := s.weather.Fetch(ctx,
		types.Coordinate{Latitude: location.Latitude, Longitude: location.Longitude})
	if err != nil {
		s.logJobCompletionWithError(action, now, fmt.Errorf("failed to retrieve weather data: %w", err))
		return
	}
	if data.Current.Temperature.IsSet() {
		s.log.Debug("current temperature", slog.Float64("temperature", data.Current.Temperature.Value()),
			slog.String("provider", s.weather.Name()))
	}

	s.logJobCompletion(action, now, false)
}
