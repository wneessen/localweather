package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	"golang.org/x/sync/errgroup"

	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/database"
	"github.com/wneessen/localweather/internal/database/model"
	"github.com/wneessen/localweather/internal/log"
	"github.com/wneessen/localweather/internal/server"
)

func main() {
	if err := start(); err != nil {
		panic(fmt.Sprintf("failed to start application: %s", err))
	}
	os.Exit(0)
}

func start() error {
	// Read the config
	confPath := "etc/app.toml"
	confPathEnv := os.Getenv("APP_CONFIG_PATH")
	if confPathEnv != "" {
		confPath = confPathEnv
	}
	path := filepath.Dir(confPath)
	file := filepath.Base(confPath)
	conf, err := config.New(path, file)
	if err != nil {
		return fmt.Errorf("failed to read/parse config: %w", err)
	}
	if err = conf.Validate(); err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	// Initialize logger
	logger := log.New(conf)

	// Catch signals to gracefully shut down the app
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	// Connect to the database
	db, err := database.Open(ctx, database.Options{
		Path:           conf.Database.Path,
		BusyTimeout:    conf.Database.BusyTimeout,
		MaxConnections: conf.Database.MaxConnections,
	})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		logger.Info("closing database")
		if cerr := db.Close(); cerr != nil {
			logger.Error("failed to close database", log.ErrAttr(cerr))
		}
	}()

	// DB migrations
	if !conf.Database.DisableAutoMigrate {
		if err = database.Migrate(ctx, db, logger); err != nil {
			return fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	// DB model / queries
	queries := model.New(db)

	// Cron task scheduler
	cron, err := gocron.NewScheduler()
	if err != nil {
		return fmt.Errorf("failed to initialize cron scheduler: %w", err)
	}

	// Create a new http.Server instance
	s := server.New(server.Params{
		Cron:    cron,
		DB:      db,
		Log:     logger,
		Queries: queries,
	}, conf)

	// Use an errgroup to wait for separate goroutines which can error
	eg, egctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return s.Start(egctx) })
	eg.Go(func() error {
		<-egctx.Done()
		logger.Info("gracefully shutting down localweather")
		return s.Stop()
	})
	if err = eg.Wait(); err != nil {
		return err
	}

	logger.Info("localweather successfully shut down")
	return nil
}
