.PHONY: all build run test clean docker-build docker-run docker-stop lint lint-go lint-js fmt vet deps tidy check lint-fix lint-go-fix lint-js-fix

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet
GOLANGCILINT=/workspace/go/bin/golangci-lint

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

deps:
	$(GOMOD) download

tidy:
	$(GOMOD) tidy

build:
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

build-go:
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

frontend:
	@if command -v npm >/dev/null 2>&1; then npm run build; else echo "npm not found"; fi

build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) .

run:
	$(GOCMD) run .

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
# STRICT LINTING - CATCHES ALL THE BUGS
# =============================================

fmt:
	$(GOFMT) -s -w .

vet:
	$(GOVET) -composites=false ./...

lint-go:
	@echo "Running STRICT Go linting..."
	@if [ ! -f "$(GOLANGCILINT)" ]; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /workspace/go/bin v1.64.5; fi
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