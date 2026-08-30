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

func (s *Server) handlerWeatherCurrentGet(w http.ResponseWriter, r *http.Request) {
	type response struct {
		model.CurrentWeather
		TimestampString    string  `json:"timestamp_local"`
		TimestampStringUTC string  `json:"timestamp_utc"`
		UpdatedAtString    string  `json:"updated_at_local"`
		UpdatedAtStringUTC string  `json:"updated_at_utc"`
		Latitude           float64 `json:"latitude"`
		Longitude          float64 `json:"longitude"`
		City               string  `json:"city,omitempty"`
		Country            string  `json:"country"`
		DisplayName        string  `json:"display_name"`
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
		CurrentWeather:     data,
		TimestampString:    time.UnixMicro(data.Timestamp).Format(time.RFC3339),
		TimestampStringUTC: time.UnixMicro(data.Timestamp).UTC().Format(time.RFC3339),
		UpdatedAtString:    time.UnixMicro(data.UpdatedAt).Format(time.RFC3339),
		UpdatedAtStringUTC: time.UnixMicro(data.UpdatedAt).UTC().Format(time.RFC3339),
		Latitude:           address.Latitude,
		Longitude:          address.Longitude,
		DisplayName:        address.DisplayName,
	}
	if address.City.Valid {
		params.City = address.City.String
	}
	if address.Country.Valid {
		params.Country = address.Country.String
	}
	resp := NewResponse(http.StatusOK, "current location", params)
	if err := render.Render(w, r, resp); err != nil {
		s.log.Error("failed to render current weather response", log.ErrAttr(err))
	}
}
