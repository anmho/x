.DEFAULT_GOAL := help

.PHONY: help setup clean clean-full test verify build build-cli build-web \
	build-access-api build-omnichannel-backend build-x-stream-bot \
	stack-up stack-down stack-status stack-logs watch-up watch-status \
	platform-install deploy preflight materialize-configs project-registry

PROJECT ?=
ENV ?= dev
SERVICE ?=
TARGET ?= platform
INSTALL_DIR ?=
DRY_RUN ?=
WATCH_GO ?= 0
GOCACHE_DIR ?= /tmp/go-cache

help:
	@echo "Project X Make Targets"
	@echo ""
	@echo "Bootstrap:"
	@echo "  make setup                  # run environment checks"
	@echo "  make clean                  # remove common local artifacts"
	@echo "  make clean-full             # remove common + large frontend artifacts"
	@echo ""
	@echo "Quality:"
	@echo "  make test                   # run full verify suite"
	@echo "  make verify                 # alias of test"
	@echo ""
	@echo "Build:"
	@echo "  make build                  # build all active targets"
	@echo "  make build-cli              # build bin/platform"
	@echo "  make build-web              # build cloud console frontend"
	@echo "  make build-access-api       # build services/access-api"
	@echo "  make build-omnichannel-backend # build omnichannel api + worker"
	@echo "  make build-x-stream-bot     # build rust stream bot (if cargo is installed)"
	@echo ""
	@echo "Stack:"
	@echo "  make stack-up               # start local stack"
	@echo "  make watch-up               # start local stack with Go hot reload (wgo)"
	@echo "  make stack-down             # stop local stack"
	@echo "  make stack-status           # show stack status"
	@echo "  make watch-status           # show Go watch-mode settings"
	@echo "  make stack-logs SERVICE=... # stream service logs"
	@echo ""
	@echo "Platform CLI:"
	@echo "  make platform-install                     # install to ~/.local/bin"
	@echo "  make platform-install INSTALL_DIR=/path   # install to custom path"
	@echo ""
	@echo "Deploy:"
	@echo "  make deploy PROJECT=<name> [ENV=dev] [DRY_RUN=1]"
	@echo "  make preflight [TARGET=platform]"
	@echo ""
	@echo "Config + Registry:"
	@echo "  make materialize-configs    # materialize declarative Python config into JSON files"
	@echo "  make project-registry       # generate project + MCP registry artifacts"

setup:
	./scripts/doctor

clean:
	./scripts/clean

clean-full:
	./scripts/clean --full

test:
	./scripts/verify all

verify: test

build: build-cli build-web build-access-api build-omnichannel-backend build-x-stream-bot

build-cli:
	./platform build

build-web:
	npm --prefix apps/cloud-console run build

build-access-api:
	cd services/access-api && GOCACHE="$(GOCACHE_DIR)" go build ./cmd/api

build-omnichannel-backend:
	cd services/omnichannel/backend && GOCACHE="$(GOCACHE_DIR)" go build ./cmd/api
	cd services/omnichannel/backend && GOCACHE="$(GOCACHE_DIR)" go build ./cmd/worker

build-x-stream-bot:
	@if [ -f services/x_stream_bot/Cargo.toml ]; then \
		if command -v cargo >/dev/null 2>&1; then \
			cargo build --manifest-path services/x_stream_bot/Cargo.toml --release; \
		else \
			echo "cargo not found; skipping services/x_stream_bot build"; \
		fi; \
	fi

stack-up:
	WATCH_GO_SERVICES="$(WATCH_GO)" ./platform start

watch-up:
	WATCH_GO_SERVICES=1 ./platform start

stack-down:
	./platform stop

stack-status:
	./platform status

watch-status:
	@echo "WATCH_GO_SERVICES=$(WATCH_GO)"
	@echo "Set WATCH_GO=1 for make stack-up, or use make watch-up."

stack-logs:
	@test -n "$(SERVICE)" || (echo "usage: make stack-logs SERVICE=<service>"; exit 1)
	./platform logs "$(SERVICE)"

platform-install:
	@if [ -n "$(INSTALL_DIR)" ]; then \
		./platform install "$(INSTALL_DIR)"; \
	else \
		./platform install; \
	fi

deploy:
	@test -n "$(PROJECT)" || (echo "usage: make deploy PROJECT=<name> [ENV=dev] [DRY_RUN=1]"; exit 1)
	@if [ "$(DRY_RUN)" = "1" ]; then \
		ENV="$(ENV)" ./platform deploy --project "$(PROJECT)" --dry-run; \
	else \
		ENV="$(ENV)" ./platform deploy --project "$(PROJECT)"; \
	fi

preflight:
	./scripts/deploy-preflight "$(TARGET)"

materialize-configs:
	python3 scripts/ci/materialize_platform_configs.py

project-registry:
	python3 scripts/ci/materialize_platform_configs.py
	node scripts/ci/generate-project-registry.mjs
