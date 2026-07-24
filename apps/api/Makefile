.PHONY: all build run test test-quick test-race clean docker-build docker-run docker-stop lint lint-go lint-js fmt vet deps tidy check lint-fix lint-go-fix lint-js-fix wire migrate seed dev profile race build-prod build-production

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet
# Version locked via .golangci-version file
GOLANGCILINT_VERSION=$(shell cat .golangci-version)
GOLANGCILINT_INSTALL_DIR=/workspace/go/bin
GOLANGCILINT=$(GOLANGCILINT_INSTALL_DIR)/golangci-lint

# Binary name
BINARY_NAME=vyzorix-update-server
BINARY_UNIX=$(BINARY_NAME)-unix

# Go build flags
LDFLAGS=-ldflags="-s -w"

# Production build flags (optimized, stripped)
PROD_LDFLAGS=-ldflags="-s -w -extldflags '-static'"

# Directory paths
DATA_DIR=./data
BIN_DIR=./bin
PUBLIC_DIR=./public

all: deps tidy build

deps:
	$(GOMOD) download

tidy:
	$(GOMOD) tidy

build:
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/api

# Production build - optimized binary to bin/ directory
build-prod build-production:
	@echo "Building production binary..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=1 $(GOBUILD) $(PROD_LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/api
	@echo "Production binary built: $(BIN_DIR)/$(BINARY_NAME)"

build-go:
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/api

frontend:
	@if command -v npm >/dev/null 2>&1; then npm run build; else echo "npm not found"; fi

build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) ./cmd/api

build-linux-arm64:
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/api

# Run production binary from bin/
run-prod:
	@echo "Starting production server..."
	@./$(BIN_DIR)/$(BINARY_NAME)

run:
	$(GOCMD) run ./cmd/api

run-local: export DATA_DIR=$(DATA_DIR)
run-local: export BIN_DIR=$(BIN_DIR)
run-local: export PUBLIC_DIR=$(PUBLIC_DIR)
run-local:
	$(GOCMD) run .

test:
	$(GOTEST) -v -race -cover ./...

test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

bench:
	$(GOTEST) -bench=. -benchmem -memprofile=mem.out ./...
	$(GOCMD) tool pprof -http=:8080 mem.out

# =============================================
# WIRE DEPENDENCY INJECTION
# =============================================

wire:
	@echo "Generating wire dependencies..."
	@cd internal/api/wire && wire

# =============================================
# TEST VARIANTS
# =============================================

test-quick:
	$(GOTEST) ./...

test-race:
	$(GOTEST) -v -race ./...

race:
	$(GOTEST) -race ./...

# =============================================
# PROFILING
# =============================================

profile:
	$(GOTEST) -cpuprofile=cpu.out -memprofile=mem.out ./...
	@echo "Profiles saved to cpu.out and mem.out"
	@echo "View with: go tool pprof -http=:8080 cpu.out"

# =============================================
# DATABASE
# =============================================

migrate:
	@echo "Running migrations (automatic on startup)..."
	@$(GOCMD) run ./cmd/api -migrate 2>/dev/null || echo "Migrations run automatically on server start"

seed:
	@echo "Seeding database..."
	@$(GOCMD) run ./cmd/api -seed 2>/dev/null || echo "Seed command not configured"

# =============================================
# DEVELOPMENT
# =============================================

dev:
	@echo "Starting development server..."
	@$(GOCMD) run ./cmd/api

# =============================================
# STRICT LINTING - CATCHES ALL THE BUGS
# =============================================

fmt:
	$(GOFMT) -s -w .

vet:
	$(GOVET) -composites=false ./...

lint-go:
	@echo "Installing golangci-lint $(GOLANGCILINT_VERSION)..."
	@REQUIRED_VERSION=$(shell cat .golangci-version) && \
	VERSION_NO_V=$$(echo "$$REQUIRED_VERSION" | sed 's/^v//') && \
	INSTALLED_VERSION=$$($(GOLANGCILINT) --version 2>/dev/null | grep -o 'version [0-9.]*' | cut -d' ' -f2) && \
	if [ ! -f "$(GOLANGCILINT)" ] || [ "$$INSTALLED_VERSION" != "$$VERSION_NO_V" ]; then \
		mkdir -p $(GOLANGCILINT_INSTALL_DIR); \
		curl -sSfL https://github.com/golangci/golangci-lint/releases/download/$$REQUIRED_VERSION/golangci-lint-$$VERSION_NO_V-linux-amd64.tar.gz -o /tmp/golangci-lint.tar.gz && \
		tar -xzf /tmp/golangci-lint.tar.gz -C /tmp && \
		cp /tmp/golangci-lint-$$VERSION_NO_V-linux-amd64/golangci-lint $(GOLANGCILINT) && \
		chmod +x $(GOLANGCILINT) && \
		rm -rf /tmp/golangci-lint* /tmp/golangci-lint-$$VERSION_NO_V-linux-amd64 && \
		echo "Installed $$REQUIRED_VERSION"; \
	else \
		echo "golangci-lint $$VERSION_NO_V already installed (version locked via .golangci-version)"; \
	fi
	@echo "Running Go linting..."
	$(GOLANGCILINT) run ./...

lint-go-fix:
	$(GOLANGCILINT) run --fix ./...

lint-js:
	@echo "Running STRICT ESLint..."
	cd src && npm run lint

lint-js-fix:
	cd src && npm exec -- eslint . --fix

lint: lint-go lint-js

lint-fix: lint-go-fix lint-js-fix

check: fmt lint vet test

ci: fmt vet test

clean:
	rm -f $(BINARY_NAME) $(BINARY_UNIX) coverage.out coverage.html *.prof
	rm -rf dist/

docker-build:
	docker build -t vyzorix-update-server:latest .

docker-build-no-cache:
	docker build --no-cache -t vyzorix-update-server:latest .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-rebuild: docker-down docker-build docker-up

docker-logs:
	docker-compose logs -f

docker-shell:
	docker-compose exec vyzorix-server sh

init-data:
	@mkdir -p $(DATA_DIR) $(BIN_DIR) $(PUBLIC_DIR)
	@cp -n $(DATA_DIR)/version.json 2>/dev/null || cp api/v1/version.json $(DATA_DIR)/version.json
	@cp -n $(DATA_DIR)/changelog.json 2>/dev/null || cp api/v1/changelog.json $(DATA_DIR)/changelog.json
	@touch $(DATA_DIR)/vyzorix.db
	@echo "Created $(DATA_DIR)"

init-env: init-data
	@if [ ! -f .env ]; then cp .env.example .env; echo "Created .env"; else echo ".env exists"; fi

gen-secret:
	@openssl rand -hex 32

validate-firebase:
	@if [ -z "$$FIREBASE_CREDENTIALS" ]; then echo "FIREBASE_CREDENTIALS not set"; exit 1; fi
	@echo "$$FIREBASE_CREDENTIALS" | jq . > /dev/null && echo "Firebase JSON valid"

secrets-check:
	@if command -v trufflehog >/dev/null 2>&1; then trufflehog filesystem . --no-update; else echo "trufflehog not installed"; fi
