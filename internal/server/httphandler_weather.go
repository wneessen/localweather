package server

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/database/model"
	"github.com/wneessen/localweather/internal/log"
)

type weatherResponse struct {
	Temperature         float64 `json:"temperature,omitempty"`
	ApparentTemperature float64 `json:"apparent_temperature,omitempty"`
	WeatherCode         int     `json:"wmo_weather_code,omitempty"`
	WindSpeed           float64 `json:"wind_speed,omitempty"`
	WindGusts           float64 `json:"wind_gusts,omitempty"`
	WindDirection       float64 `json:"wind_direction,omitempty"`
	RelativeHumidity    float64 `json:"relative_humidity,omitempty"`
	PressureMSL         float64 `json:"pressure_msl,omitempty"`
	TempDayMin          float64 `json:"temp_day_min,omitempty"`
	TempDayMax          float64 `json:"temp_day_max,omitempty"`
	UVIndex             float64 `json:"uv_index,omitempty"`
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
	Timezone            string  `json:"timezone"`
	LocationProvider    string  `json:"location_provider"`
	WeatherProvider     string  `json:"weather_provider"`
}

func (s *Server) handlerWeatherCurrentGet(w http.ResponseWriter, r *http.Request) {
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

	params := s.buildWeatherResponse(data, address)
	resp := NewResponse(http.StatusOK, "current location", params)
	if err = render.Render(w, r, resp); err != nil {
		s.log.Error("failed to render current weather response", log.ErrAttr(err))
	}
}

func (s *Server) buildWeatherResponse(data model.CurrentWeather, address model.CurrentAddressRow) *weatherResponse {
	resp := &weatherResponse{
		UpdatedAtString:    time.UnixMicro(data.UpdatedAt).Format(time.RFC3339),
		UpdatedAtStringUTC: time.UnixMicro(data.UpdatedAt).UTC().Format(time.RFC3339),
		Latitude:           address.Latitude,
		Longitude:          address.Longitude,
		DisplayName:        address.DisplayName,
		WeatherProvider:    s.weather.Name(),
		Timezone:           data.Timezone,
		LocationProvider:   address.Provider,
	}
	if data.TimezoneAbbr.Valid {
		resp.Timezone = resp.Timezone + " (" + data.TimezoneAbbr.String + ")"
	}
	if data.Temperature.Valid {
		resp.Temperature = data.Temperature.Float64
		resp.TempUnit = data.TempUnit
	}
	if data.ApparentTemperature.Valid {
		resp.ApparentTemperature = data.ApparentTemperature.Float64
		resp.TempUnit = data.TempUnit
	}
	if data.TempDayMin.Valid {
		resp.TempDayMin = data.TempDayMin.Float64
		resp.TempUnit = data.TempUnit
	}
	if data.TempDayMax.Valid {
		resp.TempDayMax = data.TempDayMax.Float64
		resp.TempUnit = data.TempUnit
	}
	if data.UvIndex.Valid {
		resp.UVIndex = data.UvIndex.Float64
	}
	if data.WeatherCode.Valid {
		resp.WeatherCode = int(data.WeatherCode.Int64)
		resp.Condition = data.Condition
		resp.Icon = data.Icon
		resp.Category = data.Category
		resp.IconURL = s.fmt.WeatherSymbolURL(resp.WeatherCode, data.IsDay.Valid && data.IsDay.Bool)
	}
	if data.WindSpeed.Valid {
		resp.WindSpeed = data.WindSpeed.Float64
		resp.WindspeedUnit = data.WindspeedUnit
	}
	if data.WindGusts.Valid {
		resp.WindGusts = data.WindGusts.Float64
		resp.WindspeedUnit = data.WindspeedUnit
	}
	if data.WindDirection.Valid {
		resp.WindDirection = data.WindDirection.Float64
		resp.WinddirIcon = data.WinddirIcon
		resp.WinddirText = data.WinddirText
		resp.WindDirUnit = data.WinddirUnit
	}
	if data.RelativeHumidity.Valid {
		resp.RelativeHumidity = data.RelativeHumidity.Float64
		resp.HumidityUnit = data.HumidityUnit
	}
	if data.PressureMsl.Valid {
		resp.PressureMSL = data.PressureMsl.Float64
		resp.PressureUnit = data.PressureUnit
	}
	if data.IsDay.Valid {
		resp.IsDay = data.IsDay.Bool
	}
	if data.Timestamp.Valid {
		resp.Timestamp = data.Timestamp.Int64
		resp.TimestampString = time.UnixMicro(data.Timestamp.Int64).Format(time.RFC3339)
		resp.TimestampStringUTC = time.UnixMicro(data.Timestamp.Int64).UTC().Format(time.RFC3339)
	}
	if address.City.Valid {
		resp.City = address.City.String
	}
	if address.Country.Valid {
		resp.Country = address.Country.String
	}
	return resp
}
