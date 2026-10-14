# Paved path: one entry point for everything a developer does in this repo.
# Every target here is also what CI runs, so "works locally" == "passes in CI".
#
# Multi-service: targets act on all services by default, or one via SERVICE=.
#   make test                      # everything
#   make image SERVICE=billing     # just billing

SERVICES ?= gateway billing worker
SERVICE  ?= gateway
IMG      ?= registry.example.com/paved-path
# Support both the docker CLI plugin and standalone docker-compose.
COMPOSE  := $(shell docker compose version >/dev/null 2>&1 && echo "docker compose" || echo "docker-compose")

.PHONY: help bootstrap test lint build scan image sign release release-all fleet-metrics setup-time up down

help: ## list targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-14s %s\n", $$1, $$2}'

bootstrap: ## install dev tools (Go via brew; govulncheck optional)
	@command -v go >/dev/null || { echo "install Go: brew install go"; exit 1; }
	go install golang.org/x/vuln/cmd/govulncheck@latest || true

test: ## run all tests (all services)
	go test ./...

lint: ## static checks (golangci-lint if present, else go vet)
	@command -v golangci-lint >/dev/null && golangci-lint run || go vet ./...

build: ## build every service (reproducible flags)
	@for s in $(SERVICES); do \
	  echo "==> $$s"; \
	  CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/$$s ./services/$$s || exit 1; \
	done

scan: ## vulnerability scan (govulncheck if present, else go vet)
	@command -v govulncheck >/dev/null && govulncheck ./... || go vet ./...

image: ## build one service image (make image SERVICE=billing)
	docker build --build-arg SERVICE=$(SERVICE) -t $(IMG)/$(SERVICE):dev .

sign: image ## sign the dev image keylessly (Sigstore)
	cosign sign --yes $(IMG)/$(SERVICE):dev

release: ## release one service (make release SERVICE=gateway VERSION=1.2.3)
	bash scripts/release.sh $(SERVICE) $(or $(VERSION),dev)

release-all: ## release every service + fleet manifest
	bash scripts/release.sh --all $(or $(VERSION),dev)

fleet-metrics: ## per-service setup time + fleet rollup (JSON report)
	bash scripts/fleet-metrics.sh

setup-time: ## measure fresh-clone -> first passing test (whole repo)
	bash scripts/setup-time.sh

up: ## run the gateway locally via docker compose
	$(COMPOSE) up --build -d

down: ## stop the local service
	$(COMPOSE) down
