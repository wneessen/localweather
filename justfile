root      := justfile_directory()
log_dir   := root / ".dev"

## Some exports for convenice
export APP_URL := "http://127.0.0.1:10001"

[private]
default:
    @just --list --unsorted

[group('run')]
shell:
    nix develop -c zsh

[group('run')]
lint:
    nix develop -c golangci-lint run

[group('run')]
logs:
    tail -f {{ log_dir }}/service.log | jq

[group('service')]
server:
    nix develop -c air | tee {{ log_dir }}/service.log

[group('setup')]
bootstrap:
    mkdir -p {{ log_dir }}
    test -f .air.toml || nix develop -c 'air init'
