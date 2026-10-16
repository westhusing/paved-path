# Paved path: one entry point for everything a developer does in this repo.
# The CI pipelines run the same commands as these targets (directly, not via
# make), so a gate you pass locally is the one CI checks. The trivy image scan
# has no target here on purpose — scanning a built image is a pipeline concern.
#
# Multi-service: targets act on all services by default, or one via SERVICE=.
#   make test                      # everything
#   make image SERVICE=billing     # just billing

SERVICES ?= gateway billing worker
SERVICE  ?= gateway
IMG      ?= registry.example.com/paved-path
# Support both the docker CLI plugin and standalone docker-compose.
COMPOSE  := $(shell docker compose version >/dev/null 2>&1 && echo "docker compose" || echo "docker-compose")

.PHONY: help bootstrap test race lint build scan ci-lint image sign release release-all fleet-metrics setup-time up down

help: ## list targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-14s %s\n", $$1, $$2}'

# govulncheck is pinned, not @latest: an unpinned tool version in a
# "reproducible" repo is the same class of bug as an unpinned base image.
# v1.9.0 requires Go >= 1.26, which is why go.mod declares 1.27.
GOVULNCHECK_VERSION ?= v1.9.0

bootstrap: ## install dev tools (Go via brew; govulncheck optional)
	@command -v go >/dev/null || { echo "install Go: brew install go"; exit 1; }
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) || true

test: ## run all tests (all services)
	go test ./...

race: ## race-detector pass (needs cgo; CI runs this on every commit)
	CGO_ENABLED=1 go test -race ./...

lint: ## static checks (golangci-lint if present, else go vet)
	@if command -v golangci-lint >/dev/null; then \
	  golangci-lint run; \
	else \
	  go vet ./...; \
	fi

build: ## build every service (reproducible flags)
	@for s in $(SERVICES); do \
	  echo "==> $$s"; \
	  CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/$$s ./services/$$s || exit 1; \
	done

scan: ## vulnerability scan (govulncheck if present, else go vet)
	@if command -v govulncheck >/dev/null; then \
	  govulncheck ./...; \
	else \
	  go vet ./...; \
	fi

# Lint the CI config itself, locally. The workflow files are code; this runs
# the same linters the actionlint job does (actionlint + shellcheck), so a
# local pass predicts CI. The GitLab templates have no .gitlab-ci.yml to feed
# `glab ci lint` (this repo IS the template), so it validates they parse as
# YAML instead — the class of error that lint would catch. Comments live above
# the target, not in the recipe: BSD make echoes in-recipe `#` lines.
ci-lint:
	@if command -v actionlint >/dev/null; then \
	  echo "==> actionlint"; actionlint .github/workflows/; \
	else \
	  echo "actionlint not installed (CI runs it); skipping"; \
	fi
	@if command -v shellcheck >/dev/null; then \
	  echo "==> shellcheck"; shellcheck scripts/*.sh; \
	else \
	  echo "shellcheck not installed (CI runs it); skipping"; \
	fi
	@if command -v python3 >/dev/null && python3 -c 'import yaml' 2>/dev/null; then \
	  echo "==> gitlab yaml"; python3 -c 'import glob,yaml; [yaml.safe_load(open(f)) for f in sorted(glob.glob(".gitlab/ci/*.yml"))]; print("ok")'; \
	elif command -v ruby >/dev/null; then \
	  echo "==> gitlab yaml"; ruby -ryaml -e 'Dir[".gitlab/ci/*.yml"].each{|f| YAML.load_file(f)}; puts "ok"'; \
	else \
	  echo "no yaml parser (python3+yaml or ruby) found; skipping GitLab template check"; \
	fi

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
