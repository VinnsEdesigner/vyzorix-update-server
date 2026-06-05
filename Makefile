.PHONY: all build run test clean docker-build docker-run docker-stop lint fmt deps tidy vet

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Binary name
BINARY_NAME=vyzorix-update-server
BINARY_UNIX=$(BINARY_NAME)-unix

# Go build flags
LDFLAGS=-ldflags="-s -w"

# Directory paths
DATA_DIR=./data
BIN_DIR=./bin
PUBLIC_DIR=./public

all: deps tidy build frontend

# Download dependencies
deps:
	$(GOMOD) download

# Tidy go.mod
tidy:
	$(GOMOD) tidy

# Build the Go binary
build:
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# Build only Go (skip frontend)
build-go:
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# Build the React frontend (TanStack Start outputs to public/)
frontend:
	@if command -v npm >/dev/null 2>&1; then \
		npm run build; \
	else \
		echo "npm not found — skipping frontend build. Install Node 22+ to build the React app."; \
	fi

# Build for Linux
build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) .

# Run locally (requires .env file and React built)
run:
	$(GOCMD) run .

# Run with environment
run-local: export DATA_DIR=$(DATA_DIR)
run-local: export BIN_DIR=$(BIN_DIR)
run-local: export PUBLIC_DIR=$(PUBLIC_DIR)
run-local:
	$(GOCMD) run .

# Run tests
test:
	$(GOTEST) -v -race -cover ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	$(GOTEST) -bench=. -benchmem -memprofile=mem.out ./...
	$(GOCMD) tool pprof -http=:8080 mem.out

# Format code
fmt:
	$(GOFMT) -s -w .

# Lint code
lint:
	$(GOVET) ./...

# Full vet
vet:
	$(GOVET) -composites=false ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_UNIX)
	rm -f coverage.out coverage.html
	rm -f *.prof
	rm -rf dist/

# Docker build
docker-build:
	docker build -t vyzorix-update-server:latest .

# Docker build with no cache
docker-build-no-cache:
	docker build --no-cache -t vyzorix-update-server:latest .

# Run via docker-compose
docker-up:
	docker-compose up -d

# Stop docker-compose
docker-down:
	docker-compose down

# Rebuild and restart
docker-rebuild: docker-down docker-build docker-up

# View logs
docker-logs:
	docker-compose logs -f

# Shell into container
docker-shell:
	docker-compose exec vyzorix-server sh

# Initialize data directory with version manifests
# These are copied from api/v1/ (source of truth) to data/ (runtime target)
init-data:
	@mkdir -p $(DATA_DIR) $(BIN_DIR) $(PUBLIC_DIR)
	@cp -n $(DATA_DIR)/version.json 2>/dev/null || cp api/v1/version.json $(DATA_DIR)/version.json
	@cp -n $(DATA_DIR)/changelog.json 2>/dev/null || cp api/v1/changelog.json $(DATA_DIR)/changelog.json
	@touch $(DATA_DIR)/vyzorix.db
	@echo "Created $(DATA_DIR) with version.json, changelog.json, and database placeholder"

# Copy env file for local development
init-env: init-data
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env from .env.example"; \
	else \
		echo ".env already exists"; \
	fi

# Generate HMAC secret for development
gen-secret:
	@openssl rand -hex 32

# Validate Firebase credentials JSON
validate-firebase:
	@if [ -z "$$FIREBASE_CREDENTIALS" ]; then \
		echo "FIREBASE_CREDENTIALS not set"; \
		exit 1; \
	fi
	@echo "$$FIREBASE_CREDENTIALS" | jq . > /dev/null && echo "Firebase credentials JSON is valid"

# Security: check for secrets in git
secrets-check:
	@if command -v trufflehog >/dev/null 2>&1; then \
		trufflehog filesystem . --no-update; \
	else \
		echo "trufflehog not installed, skipping secrets check"; \
	fi

# Run all quality checks
check: fmt lint vet test

# Format and check in one step
ci: fmt vet test
