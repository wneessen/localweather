root      := justfile_directory()
dev_dir   := root / ".dev"
db_path   := dev_dir / "storage.db"

## Some exports for convenice
export APP_URL := "http://127.0.0.1:10001"
export GOOSE_DRIVER := "sqlite3"
export GOOSE_DBSTRING := db_path
export GOOSE_MIGRATION_DIR := root / "internal/database/migrations"

## App-specific development environment
export LOCALWEATHER_LOG_LEVEL := "debug"
export LOCALWEATHER_GEOCODER_PROVIDER := "opencage"
export LOCALWEATHER_DATABASE_PATH := db_path
export LOCALWEATHER_GEOBUS_DISABLE_COORDINATES_FILE := "true"
#export LOCALWEATHER_GEOCODER_PROVIDER := "geocode-earth"

[private]
default:
    @just --list --unsorted

[group('run')]
shell:
    secretspec run -- nix develop -c zsh

[group('run')]
lint:
    nix develop -c golangci-lint run

[group('run')]
logs:
    tail -f {{ dev_dir }}/service.log | jq

[group('service')]
server:
    nix develop -c secretspec run -- air | tee {{ dev_dir }}/service.log

[group('setup')]
bootstrap:
    mkdir -p {{ dev_dir }}
    test -f .air.toml || nix develop -c 'air init'
