package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"github.com/wneessen/localweather/internal/database/migrations"
	"github.com/wneessen/localweather/internal/log"
)

func Migrate(ctx context.Context, db *sql.DB, logger *log.Logger) (*goose.Provider, error) {
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, migrations.FS)
	if err != nil {
		return provider, fmt.Errorf("failed to create migration provider: %w", err)
	}

	current, err := provider.GetDBVersion(ctx)
	if err != nil {
		return provider, fmt.Errorf("failed to determine current schema version: %w", err)
	}
	logger.Info("checking database schema", "current_version", current)

	results, err := provider.Up(ctx)
	if err != nil {
		return provider, fmt.Errorf("failed to apply migrations: %w", err)
	}
	for _, r := range results {
		logger.Info("applied migration", "version", r.Source.Version,
			"name", r.Source.Path, "duration", r.Duration)
	}
	return provider, nil
}
