package server

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/database/model"
	"github.com/wneessen/localweather/internal/log"
)

type weatherResponse struct {
	WeatherDataRAW       responseData    `json:"weather_data_raw"`
	WeatherDataLocalized responseData    `json:"weather_data_localized"`
	WeatherCode          int             `json:"wmo_weather_code,omitempty"`
	Condition            string          `json:"condition,omitempty"`
	Category             string          `json:"category,omitempty"`
	Icon                 string          `json:"icon,omitempty"`
	IconURL              string          `json:"icon_url,omitempty"`
	WinddirIcon          string          `json:"winddir_icon"`
	WinddirText          string          `json:"winddir_text"`
	IsDay                bool            `json:"is_day,omitempty"`
	Timestamp            int64           `json:"timestamp_unix,omitempty"`
	Units                responseUnits   `json:"units"`
	Location             responseAddress `json:"location"`
	LocationProvider     string          `json:"location_provider"`
	WeatherProvider      string          `json:"weather_provider"`
}

type responseData struct {
	Temperature         string `json:"temperature,omitempty"`
	ApparentTemperature string `json:"apparent_temperature,omitempty"`
	WindSpeed           string `json:"wind_speed,omitempty"`
	WindGusts           string `json:"wind_gusts,omitempty"`
	WindDirection       string `json:"wind_direction,omitempty"`
	RelativeHumidity    string `json:"relative_humidity,omitempty"`
	PressureMSL         string `json:"pressure_msl,omitempty"`
	TempDayMin          string `json:"temp_day_min,omitempty"`
	TempDayMax          string `json:"temp_day_max,omitempty"`
	UVIndex             string `json:"uv_index,omitempty"`
	Sunrise             string `json:"sunrise"`
	Sunset              string `json:"sunset"`
	SunriseUTC          string `json:"sunrise_utc"`
	SunsetUTC           string `json:"sunset_utc"`
	TimestampString     string `json:"timestamp_local,omitempty"`
	TimestampStringUTC  string `json:"timestamp_utc,omitempty"`
	UpdatedAtString     string `json:"updated_at_local"`
	UpdatedAtStringUTC  string `json:"updated_at_utc"`
}

type responseUnits struct {
	TempUnit      string `json:"temperature"`
	WindDirUnit   string `json:"wind_direction"`
	HumidityUnit  string `json:"humidity"`
	PressureUnit  string `json:"pressure"`
	WindspeedUnit string `json:"wind_speed"`
}

type responseAddress struct {
	DisplayName string  `json:"display_name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	City        string  `json:"city,omitempty"`
	Country     string  `json:"country,omitempty"`
	Timezone    string  `json:"timezone"`
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

	query := model.CurrentWeatherByAddressIDAndBaseUnitParams{
		AddressID: address.ID,
		BaseUnit:  s.conf.Units,
	}
	data, err := s.queries.CurrentWeatherByAddressIDAndBaseUnit(r.Context(), query)
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
		Condition:        data.Condition,
		Category:         data.Category,
		Icon:             data.Icon,
		LocationProvider: address.Provider,
		Timestamp:        data.Timestamp.Int64,
		WeatherProvider:  s.weather.Name(),
	}

	if data.WeatherCode.Valid {
		resp.WeatherCode = int(data.WeatherCode.Int64)
		resp.IconURL = s.fmt.WeatherSymbolURL(resp.WeatherCode, data.IsDay.Valid && data.IsDay.Bool)
	}
	if data.WindDirection.Valid {
		resp.WinddirIcon = data.WinddirIcon
		resp.WinddirText = data.WinddirText
	}
	if data.IsDay.Valid {
		resp.IsDay = data.IsDay.Bool
	}

	s.buildWeatherResponseUnits(data, resp)
	s.buildWeatherRaw(data, resp)
	s.buildWeatherLocation(data, address, resp)

	return resp
}

func (s *Server) buildWeatherRaw(data model.CurrentWeather, weather *weatherResponse) {
	resp := responseData{
		SunriseUTC:         time.UnixMicro(data.SunriseUtc).UTC().Format(time.RFC3339),
		SunsetUTC:          time.UnixMicro(data.SunsetUtc).UTC().Format(time.RFC3339),
		Sunrise:            time.UnixMicro(data.SunriseUtc).In(time.Local).Format(time.RFC3339),
		Sunset:             time.UnixMicro(data.SunsetUtc).In(time.Local).Format(time.RFC3339),
		UpdatedAtString:    time.UnixMicro(data.UpdatedAt).Format(time.RFC3339),
		UpdatedAtStringUTC: time.UnixMicro(data.UpdatedAt).UTC().Format(time.RFC3339),
	}
	loc := responseData{
		SunriseUTC:         s.fmt.LocalizeTime(time.UnixMicro(data.SunriseUtc).UTC()),
		SunsetUTC:          s.fmt.LocalizeTime(time.UnixMicro(data.SunsetUtc).UTC()),
		Sunrise:            s.fmt.LocalizeTime(time.UnixMicro(data.SunriseUtc).In(time.Local)),
		Sunset:             s.fmt.LocalizeTime(time.UnixMicro(data.SunsetUtc).In(time.Local)),
		UpdatedAtString:    s.fmt.LocalizeTime(time.UnixMicro(data.UpdatedAt)),
		UpdatedAtStringUTC: s.fmt.LocalizeTime(time.UnixMicro(data.UpdatedAt).UTC()),
	}
	if data.Timestamp.Valid {
		resp.TimestampString = time.UnixMicro(data.Timestamp.Int64).Format(time.RFC3339)
		resp.TimestampStringUTC = time.UnixMicro(data.Timestamp.Int64).UTC().Format(time.RFC3339)
		loc.TimestampString = s.fmt.LocalizeTime(time.UnixMicro(data.Timestamp.Int64))
		loc.TimestampStringUTC = s.fmt.LocalizeTime(time.UnixMicro(data.Timestamp.Int64).UTC())
	}

	floats := []struct {
		src    sql.NullFloat64
		dst    *string
		locdst *string
	}{
		{data.Temperature, &resp.Temperature, &loc.Temperature},
		{data.ApparentTemperature, &resp.ApparentTemperature, &loc.ApparentTemperature},
		{data.TempDayMin, &resp.TempDayMin, &loc.TempDayMin},
		{data.TempDayMax, &resp.TempDayMax, &loc.TempDayMax},
		{data.UvIndex, &resp.UVIndex, &loc.UVIndex},
		{data.WindDirection, &resp.WindDirection, &loc.WindDirection},
		{data.WindSpeed, &resp.WindSpeed, &loc.WindSpeed},
		{data.WindGusts, &resp.WindGusts, &loc.WindGusts},
		{data.RelativeHumidity, &resp.RelativeHumidity, &loc.RelativeHumidity},
		{data.PressureMsl, &resp.PressureMSL, &loc.PressureMSL},
	}
	for _, entry := range floats {
		if entry.src.Valid {
			*entry.dst = fmt.Sprintf("%.1f", entry.src.Float64)
			*entry.locdst = s.fmt.Humanize(entry.src.Float64)
		}
	}

	weather.WeatherDataRAW = resp
	weather.WeatherDataLocalized = loc
}

func (s *Server) buildWeatherResponseUnits(data model.CurrentWeather, weather *weatherResponse) {
	units := responseUnits{
		TempUnit:      data.TempUnit,
		WindspeedUnit: data.WindspeedUnit,
		HumidityUnit:  data.HumidityUnit,
		PressureUnit:  data.PressureUnit,
		WindDirUnit:   data.WinddirUnit,
	}

	weather.Units = units
}

func (s *Server) buildWeatherLocation(data model.CurrentWeather, address model.CurrentAddressRow, weather *weatherResponse) {
	location := responseAddress{
		DisplayName: address.DisplayName,
		Latitude:    address.Latitude,
		Longitude:   address.Longitude,
		Timezone:    data.Timezone,
	}
	if address.City.Valid {
		location.City = address.City.String
	}
	if address.Country.Valid {
		location.Country = address.Country.String
	}
	if data.TimezoneAbbr.Valid {
		location.Timezone = location.Timezone + " (" + data.TimezoneAbbr.String + ")"
	}

	weather.Location = location
}
