root := justfile_directory()
dev_dir := root / ".dev"
db_path := dev_dir / "storage.db"
cityname_path := dev_dir / "cityname"
coordinates_path := dev_dir / "coordinates"


## Some exports for convenice
export APP_URL := "http://127.0.0.1:10002"
export GOOSE_DRIVER := "sqlite3"
export GOOSE_DBSTRING := db_path
export GOOSE_MIGRATION_DIR := root / "internal/database/migrations"

## App-specific development environment
export LOCALWEATHER_LOG_LEVEL := "debug"
export LOCALWEATHER_LOCALE := "de"
export LOCALWEATHER_SERVER_PORT := "10002"
export LOCALWEATHER_GEOCODER_PROVIDER := "opencage"
export LOCALWEATHER_DATABASE_PATH := db_path
export LOCALWEATHER_GEOBUS_COORDINATES_FILE := coordinates_path
export LOCALWEATHER_GEOBUS_CITYNAME_FILE := cityname_path
export LOCALWEATHER_GEOBUS_DISABLE_GEOAPI := "true"
export LOCALWEATHER_GEOBUS_DISABLE_GEOIP := "true"
#export LOCALWEATHER_GEOCODER_PROVIDER := "geocode-earth"
[private]
default:
    @just --list --unsorted

[group('run')]
shell:
    secretspec run -- nix develop -c zsh

[group('run')]
i18n:
  nix develop --command ~/go/bin/xspreak -D ./ -p ./internal/i18n/locale/ --copyright-holder 'Winni Neessen <wn@neessen.dev>' --package-name 'github.com/wneessen/localweather'

[group('run')]
lint:
    nix develop -c golangci-lint run

[group('run')]
logs:
    tail -f {{ dev_dir }}/service.log | jq

[group('run')]
hashes:
    #!/usr/bin/env bash
    set -euo pipefail
    ver=$(nix eval --raw .#default.version)
    base="https://github.com/wneessen/localweather/releases/download/v${ver}"
    curl -fsSL "${base}/localweather_${ver}_checksums.txt" \
      | grep -E '_(linux|darwin)_(amd64|arm64)\.tar\.gz$' \
      | while read -r sum file; do
          printf '%-45s %s\n' "$file" \
            "$(nix hash convert --hash-algo sha256 --to sri "$sum")"
        done

[group('service')]
server:
    nix develop -c secretspec run -- air | tee {{ dev_dir }}/service.log

[group('setup')]
bootstrap:
  @test -f ~/go/bin/xspreak || nix develop --command go install github.com/go-modulus/xspreak@latest
  @test -d {{ dev_dir }} || mkdir -p {{ dev_dir }}
  @test -f .air.toml || nix develop -c 'air init'
