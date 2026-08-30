package server

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/log"
)

func (s *Server) handlerWeatherCurrentGet(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Temperature         float64 `json:"temperature,omitempty"`
		ApparentTemperature float64 `json:"apparent_temperature,omitempty"`
		WeatherCode         int     `json:"wmo_weather_code,omitempty"`
		WindSpeed           float64 `json:"wind_speed,omitempty"`
		WindGusts           float64 `json:"wind_gusts,omitempty"`
		WindDirection       float64 `json:"wind_direction,omitempty"`
		RelativeHumidity    float64 `json:"relative_humidity,omitempty"`
		PressureMSL         float64 `json:"pressure_msl,omitempty"`
		Condition           string  `json:"condition,omitempty"`
		Category            string  `json:"category,omitempty"`
		Icon                string  `json:"icon,omitempty"`
		IconURL             string  `json:"icon_url,omitempty"`
		WinddirIcon         string  `json:"winddir_icon"`
		WinddirText         string  `json:"winddir_text"`
		IsDay               bool    `json:"is_day,omitempty"`
		TempUnit            string  `json:"temperature_unit"`
		WindDirUnit         string  `json:"wind_direction_unit"`
		HumidityUnit        string  `json:"humidity_unit"`
		PressureUnit        string  `json:"pressure_unit"`
		WindspeedUnit       string  `json:"wind_speed_unit"`
		Timestamp           int64   `json:"timestamp_unix,omitempty"`
		TimestampString     string  `json:"timestamp_local,omitempty"`
		TimestampStringUTC  string  `json:"timestamp_utc,omitempty"`
		UpdatedAtString     string  `json:"updated_at_local"`
		UpdatedAtStringUTC  string  `json:"updated_at_utc"`
		DisplayName         string  `json:"display_name"`
		Latitude            float64 `json:"latitude"`
		Longitude           float64 `json:"longitude"`
		City                string  `json:"city,omitempty"`
		Country             string  `json:"country,omitempty"`
		LocationProvider    string  `json:"location_provider"`
		WeatherProvider     string  `json:"weather_provider"`
	}

	address, err := s.queries.CurrentAddress(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			s.renderErr(w, r, http.StatusNotFound, apperror.ErrCurrentLocationNotSet)
		default:
			s.log.Error("failed to fetch current address from database", log.ErrAttr(err))
			s.renderErr(w, r, http.StatusInternalServerError, apperror.ErrUnexpected)
		}
		return
	}
	data, err := s.queries.CurrentWeatherByAddressID(r.Context(), address.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			s.renderErr(w, r, http.StatusNotFound, apperror.ErrCurrentWeatherdataNotFound)
		default:
			s.log.Error("failed to fetch current weather from database", log.ErrAttr(err))
			s.renderErr(w, r, http.StatusInternalServerError, apperror.ErrUnexpected)
		}
		return
	}

	params := &response{
		UpdatedAtString:    time.UnixMicro(data.UpdatedAt).Format(time.RFC3339),
		UpdatedAtStringUTC: time.UnixMicro(data.UpdatedAt).UTC().Format(time.RFC3339),
		Latitude:           address.Latitude,
		Longitude:          address.Longitude,
		DisplayName:        address.DisplayName,
		WeatherProvider:    s.weather.Name(),
		LocationProvider:   address.Provider,
	}
	if data.Temperature.Valid {
		params.Temperature = data.Temperature.Float64
		params.TempUnit = data.TempUnit
	}
	if data.ApparentTemperature.Valid {
		params.ApparentTemperature = data.ApparentTemperature.Float64
		params.TempUnit = data.TempUnit
	}
	if data.WeatherCode.Valid {
		params.WeatherCode = int(data.WeatherCode.Int64)
		params.Condition = data.Condition
		params.Icon = data.Icon
		params.Category = data.Category
		params.IconURL = s.fmt.WeatherSymbolURL(params.WeatherCode, data.IsDay.Valid && data.IsDay.Bool)
	}
	if data.WindSpeed.Valid {
		params.WindSpeed = data.WindSpeed.Float64
		params.WindspeedUnit = data.WindspeedUnit
	}
	if data.WindGusts.Valid {
		params.WindGusts = data.WindGusts.Float64
		params.WindspeedUnit = data.WindspeedUnit
	}
	if data.WindDirection.Valid {
		params.WindDirection = data.WindDirection.Float64
		params.WinddirIcon = data.WinddirIcon
		params.WinddirText = data.WinddirText
		params.WindDirUnit = data.WinddirUnit
	}
	if data.RelativeHumidity.Valid {
		params.RelativeHumidity = data.RelativeHumidity.Float64
		params.HumidityUnit = data.HumidityUnit
	}
	if data.PressureMsl.Valid {
		params.PressureMSL = data.PressureMsl.Float64
		params.PressureUnit = data.PressureUnit
	}
	if data.IsDay.Valid {
		params.IsDay = data.IsDay.Bool
	}
	if data.Timestamp.Valid {
		params.Timestamp = data.Timestamp.Int64
		params.TimestampString = time.UnixMicro(data.Timestamp.Int64).Format(time.RFC3339)
		params.TimestampStringUTC = time.UnixMicro(data.Timestamp.Int64).UTC().Format(time.RFC3339)
	}
	if address.City.Valid {
		params.City = address.City.String
	}
	if address.Country.Valid {
		params.Country = address.Country.String
	}

	resp := NewResponse(http.StatusOK, "current location", params)
	if err = render.Render(w, r, resp); err != nil {
		s.log.Error("failed to render current weather response", log.ErrAttr(err))
	}
}
