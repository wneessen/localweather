package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nathan-osman/go-sunrise"
	"github.com/wneessen/go-moonphase"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/database/model"
	"github.com/wneessen/localweather/internal/formatter"
	"github.com/wneessen/localweather/internal/types"
	"github.com/wneessen/localweather/internal/weather"
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

	sunriseTimeUTC, sunsetTimeUTC := sunrise.SunriseSunset(location.Latitude, location.Longitude, now.Year(),
		now.Month(), now.Day())
	currentDB := model.UpdateCurrentWeatherParams{
		AddressID:    location.ID,
		Timestamp:    sql.NullInt64{Int64: data.Current.InstantTime.UnixMicro(), Valid: true},
		Timezone:     data.Timezone,
		TimezoneAbbr: sql.NullString{String: data.TimezoneAbbreviation, Valid: data.TimezoneAbbreviation != ""},
		BaseUnit:     s.conf.Units,
		UpdatedAt:    now.UnixMicro(),
		SunriseUtc:   sunriseTimeUTC.UTC().UnixMicro(),
		SunsetUtc:    sunsetTimeUTC.UTC().UnixMicro(),
		Temperature: sql.NullFloat64{
			Float64: data.Current.Temperature.Value(),
			Valid:   data.Current.Temperature.IsSet(),
		},
		ApparentTemperature: sql.NullFloat64{
			Float64: data.Current.ApparentTemperature.Value(),
			Valid:   data.Current.ApparentTemperature.IsSet(),
		},
		WindSpeed: sql.NullFloat64{
			Float64: data.Current.WindSpeed.Value(),
			Valid:   data.Current.WindSpeed.IsSet(),
		},
		WindGusts: sql.NullFloat64{
			Float64: data.Current.WindGusts.Value(),
			Valid:   data.Current.WindGusts.IsSet(),
		},
		RelativeHumidity: sql.NullFloat64{
			Float64: data.Current.RelativeHumidity.Value(),
			Valid:   data.Current.RelativeHumidity.IsSet(),
		},
		PressureMsl: sql.NullFloat64{
			Float64: data.Current.PressureMSL.Value(),
			Valid:   data.Current.PressureMSL.IsSet(),
		},
		IsDay: sql.NullBool{
			Bool:  data.Current.IsDay.Value(),
			Valid: data.Current.IsDay.IsSet(),
		},
		TempUnit:      data.Current.Units.Temperature,
		WinddirUnit:   data.Current.Units.WindDirection,
		HumidityUnit:  data.Current.Units.Humidity,
		PressureUnit:  data.Current.Units.Pressure,
		WindspeedUnit: data.Current.Units.WindSpeed,
	}
	if data.Current.WeatherCode.IsSet() {
		val := data.Current.WeatherCode.Value()
		currentDB.WeatherCode = sql.NullInt64{
			Int64: int64(val),
			Valid: true,
		}
		currentDB.Condition = s.t.Get(s.fmt.WeatherCondition(val))
		currentDB.Category = s.fmt.WeatherCategory(val)
		currentDB.Icon = s.fmt.WeatherSymbol(val, data.Current.IsDay.Value())

	}
	if data.Current.WindDirection.IsSet() {
		currentDB.WindDirection = sql.NullFloat64{
			Float64: data.Current.WindDirection.Value(),
			Valid:   true,
		}
		currentDB.WinddirIcon = s.fmt.WindDirectionSymbol(data.Current.WindDirection)
		currentDB.WinddirText = s.fmt.DegToString(data.Current.WindDirection)
	}

	moon := moonphase.New(now.In(time.Local))
	phase := moon.PhaseName()
	if phase != "" {
		currentDB.Moonphase = sql.NullString{String: phase, Valid: true}
		icon, ok := formatter.MoonPhaseIcon[phase]
		if ok {
			currentDB.MoonphaseIcon = sql.NullString{String: icon, Valid: true}
		}
		if url := s.fmt.MoonphaseIconURL(phase); url != "" {
			currentDB.MoonphaseIconUrl = sql.NullString{String: url, Valid: true}
		}
	}

	today := weather.NewDay(time.Now())
	if val, ok := data.Daily[today]; ok {
		if val.TemperatureMin.IsSet() {
			currentDB.TempDayMin = sql.NullFloat64{
				Float64: val.TemperatureMin.Value(),
				Valid:   true,
			}
		}
		if val.TemperatureMax.IsSet() {
			currentDB.TempDayMax = sql.NullFloat64{
				Float64: val.TemperatureMax.Value(),
				Valid:   true,
			}
		}
		if val.UVIndex.IsSet() {
			currentDB.UvIndex = sql.NullFloat64{
				Float64: val.UVIndex.Value(),
				Valid:   true,
			}
		}
	}

	if err = s.queries.UpdateCurrentWeather(ctx, currentDB); err != nil {
		s.logJobCompletionWithError(action, now, fmt.Errorf("failed to update current weather in database: %w", err))
	}

	s.logJobCompletion(action, now, false)
}
