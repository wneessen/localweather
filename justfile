root      := justfile_directory()
log_dir   := root / ".dev"

## Some exports for convenice
export APP_URL := "http://127.0.0.1:10001"

## App-specific development environment
export LOCALWEATHER_LOG_LEVEL := "debug"
export LOCALWEATHER_GEOCODER_PROVIDER := "opencage"
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
    tail -f {{ log_dir }}/service.log | jq

[group('service')]
server:
    nix develop -c secretspec run -- air | tee {{ log_dir }}/service.log

[group('setup')]
bootstrap:
    mkdir -p {{ log_dir }}
    test -f .air.toml || nix develop -c 'air init'
