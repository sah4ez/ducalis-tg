.PHONY: all install-tg generate transport client client-go client-ts swagger build build-public build-admin build-internal run clean docker-up docker-down test lint help

# Directories
CONTRACT_DIR := ./pkg/contract
TRANSPORT_DIR := ./internal/transport
CLIENT_DIR := ./pkg/client
API_DIR := ./api

# Install tg generator
install-tg:
	go install github.com/seniorGolang/tg/v3/cmd/tg@latest

# Generate all (transport, clients, swagger)
generate: transport client swagger
	goimports -l -w $(TRANSPORT_DIR) 2>/dev/null || true
	goimports -l -w $(CLIENT_DIR) 2>/dev/null || true

# Generate transport layer for all services
transport:
	@echo "Generating transport layer..."
	tg server --contracts-dir $(CONTRACT_DIR) -o $(TRANSPORT_DIR)

# Generate all clients
client: client-go client-ts

# Generate Go client (public API only)
client-go:
	@echo "Generating Go client..."
	tg client-go --contracts-dir $(CONTRACT_DIR) -o $(CLIENT_DIR)/go
	goimports -l -w $(CLIENT_DIR) 2>/dev/null || true

# Generate TypeScript client (public API only)
client-ts:
	@echo "Generating TypeScript client..."
	tg client-ts --contracts-dir $(CONTRACT_DIR) -o $(CLIENT_DIR)/ts

# Generate OpenAPI specs for all services
swagger:
	@echo "Generating OpenAPI spec..."
	mkdir -p $(API_DIR)
	tg server --contracts-dir $(CONTRACT_DIR) -o $(API_DIR)

# Build all binaries
build: build-public build-admin build-internal
	@echo "✅ All services built"

# Build public API server
build-public:
	CGO_ENABLED=0 go build -ldflags="-w -s -X main.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')" -o bin/server-public ./cmd/server-public

# Build admin API server
build-admin:
	CGO_ENABLED=0 go build -ldflags="-w -s -X main.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')" -o bin/server-admin ./cmd/server-admin

# Build internal API server
build-internal:
	CGO_ENABLED=0 go build -ldflags="-w -s -X main.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')" -o bin/server-internal ./cmd/server-internal

# Run development servers
run-public:
	go run ./cmd/server-public

run-admin:
	go run ./cmd/server-admin

run-internal:
	go run ./cmd/server-internal

run: run-public run-admin run-internal

# Run with hot reload (requires air)
dev-public:
	air --config .air-public.toml

dev-admin:
	air --config .air-admin.toml

dev-internal:
	air --config .air-internal.toml

# Docker operations
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-build:
	docker-compose build

docker-restart:
	docker-compose restart

# Database migrations
migrate-up:
	@echo "Running migrations..."
	psql $(DATABASE_URL) -f migrations/init.sql

migrate-reset:
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres
	@sleep 3
	$(MAKE) migrate-up

# Tests
test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Linting
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/
	find $(TRANSPORT_DIR) -maxdepth 1 -name '*.go' -delete

# Download dependencies
deps:
	go mod download
	go mod tidy

# Help
help:
	@echo "Ducalis - Task Prioritization Platform"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  install-tg    Install tg code generator"
	@echo "  generate      Generate all (transport, clients, swagger)"
	@echo "  transport     Generate transport layer"
	@echo "  client        Generate Go and TS clients"
	@echo "  client-go     Generate Go client"
	@echo "  client-ts     Generate TypeScript client"
	@echo "  swagger       Generate OpenAPI specs"
	@echo ""
	@echo "  build         Build all server binaries"
	@echo "  build-public  Build public API server"
	@echo "  build-admin   Build admin API server"
	@echo "  build-internal Build internal API server"
	@echo ""
	@echo "  run-public    Run public API server"
	@echo "  run-admin     Run admin API server"
	@echo "  run-internal  Run internal API server"
	@echo ""
	@echo "  docker-up     Start Docker containers"
	@echo "  docker-down   Stop Docker containers"
	@echo "  test          Run tests"
	@echo "  lint          Run linter"
	@echo "  clean         Clean build artifacts"
