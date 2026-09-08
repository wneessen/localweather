// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package server

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/wneessen/localweather/internal/log"
)

type HealthzResponse struct {
	ServiceIsHealthy bool `json:"service_healthy"`
}

// handlerHealthzGet responds with a JSON payload indicating service health and the current UTC time.
func (s *Server) handlerHealthzGet(w http.ResponseWriter, r *http.Request) {
	resp := NewResponse(http.StatusOK, "service is up and running", HealthzResponse{ServiceIsHealthy: true})
	if err := render.Render(w, r, resp); err != nil {
		s.log.Error("failed to render healthz JSON", log.ErrAttr(err))
	}
}
