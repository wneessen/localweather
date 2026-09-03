package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"

	"github.com/wneessen/localweather/internal/log"
)

func (s *Server) httpRoutes(_ context.Context) {
	logFormat := httplog.SchemaECS
	logSkipPath := []string{"/metrics", "/healthz"}
	logger := log.New(s.conf).With(slog.String("service", "http"))
	logHandler := httplog.RequestLogger(
		logger,
		&httplog.Options{
			Level: s.conf.Log.Level.SLog(),
			Skip: func(req *http.Request, code int) bool {
				for _, skip := range logSkipPath {
					if strings.HasPrefix(req.URL.Path, skip) && code == 200 {
						return true
					}
				}
				return false
			},
			Schema:        logFormat,
			RecoverPanics: true,
		},
	)

	// Global middlewares
	s.mux.Use(middleware.StripSlashes)
	s.mux.Use(middleware.URLFormat)
	s.mux.Use(logHandler)

	// Routes
	s.mux.Get("/healthz", s.handlerHealthzGet)
	s.mux.Route("/location", func(r chi.Router) {
		r.Get("/current", s.handlerLocationCurrentGet)
	})
	s.mux.Route("/weather", func(r chi.Router) {
		r.Get("/current", s.handlerWeatherCurrentGet)
		r.Get("/update", s.handlerWeatherUpdateGet)
	})
}
