package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/log"
)

// Server represents the main application server, managing HTTP services, cron jobs, metrics, and database interactions.
type Server struct {
	conf     *config.Config
	log      *log.Logger
	httpserv *http.Server
	mux      chi.Router
	/*queries  *model.Queries
	pool     *pgxpool.Pool
	cron     gocron.Scheduler
	*/
}

type Params struct {
	Log *log.Logger
	/*
		Cron    gocron.Scheduler
		Queries *model.Queries
		PgxPool *pgxpool.Pool

	*/
}

func New(params Params, conf *config.Config) *Server {
	if params.Log == nil {
		params.Log = log.New(conf)
	}

	server := &Server{
		conf: conf,
		log:  params.Log,
		mux:  chi.NewMux(),
	}
	server.httpserv = &http.Server{
		Addr:              conf.ListenAddr(),
		Handler:           server.mux,
		ReadTimeout:       conf.Server.Timeout,
		ReadHeaderTimeout: conf.Server.Timeout,
		WriteTimeout:      conf.Server.Timeout,
		IdleTimeout:       conf.Server.Timeout,
	}
	return server
}
