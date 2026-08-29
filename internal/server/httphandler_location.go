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

func (s *Server) handlerLocationCurrentGet(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Accuracy    float64   `json:"accuracy"`
		Latitude    float64   `json:"latitude"`
		Longitude   float64   `json:"longitude"`
		DisplayName string    `json:"display_name"`
		Provider    string    `json:"location_provider"`
		LastSeen    time.Time `json:"last_seen"`
	}

	address, err := s.queries.CurrentAddress(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			s.renderErr(w, r, http.StatusNotFound, apperror.ErrLocationNotSet)
		default:
			s.log.Error("failed to get current address from database", log.ErrAttr(err))
			s.renderErr(w, r, http.StatusInternalServerError, apperror.ErrUnexpected)
		}
		return
	}
	data := response{
		Accuracy:    address.Accuracy,
		Latitude:    address.Latitude,
		Longitude:   address.Longitude,
		DisplayName: address.DisplayName,
		LastSeen:    time.UnixMicro(address.UpdatedAt),
		Provider:    address.Provider,
	}
	resp := NewResponse(http.StatusOK, "current location", data)
	if err := render.Render(w, r, resp); err != nil {
		s.log.Error("failed to render healthz JSON", log.ErrAttr(err))
	}
}
