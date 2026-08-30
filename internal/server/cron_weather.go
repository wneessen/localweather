package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/database/model"
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
		s.logJobCompletionWithError(action, now, apperror.ErrCurrentLocationNotSet)
		return
	}

	data, err := s.weather.Fetch(ctx,
		types.Coordinate{Latitude: location.Latitude, Longitude: location.Longitude})
	if err != nil {
		s.logJobCompletionWithError(action, now, fmt.Errorf("failed to retrieve weather data: %w", err))
		return
	}

	currentDB := model.UpdateCurrentWeatherParams{
		AddressID:           location.ID,
		Timestamp:           data.Current.InstantTime.UnixMicro(),
		Temperature:         data.Current.Temperature.Value(),
		ApparentTemperature: data.Current.ApparentTemperature.Value(),
		WeatherCode:         int64(data.Current.WeatherCode.Value()),
		WindSpeed:           data.Current.WindSpeed.Value(),
		WindGusts:           data.Current.WindGusts.Value(),
		WindDirection:       data.Current.WindDirection.Value(),
		RelativeHumidity:    data.Current.RelativeHumidity.Value(),
		PressureMsl:         data.Current.PressureMSL.Value(),
		IsDay:               data.Current.IsDay.Value(),
		UpdatedAt:           time.Now().UnixMicro(),
	}
	if err = s.queries.UpdateCurrentWeather(ctx, currentDB); err != nil {
		s.logJobCompletionWithError(action, now, fmt.Errorf("failed to update current weather in database: %w", err))
	}

	s.logJobCompletion(action, now, false)
}
