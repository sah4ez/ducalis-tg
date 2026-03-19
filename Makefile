.PHONY: all generate transport client swagger build run clean docker-up docker-down migrate

# tg binary path
TG := go run github.com/seniorGolang/tg/cmd/tg@v2

# Directories
CONTRACT_DIR := ./pkg/contract
TRANSPORT_DIR := ./internal/transport
CLIENT_DIR := ./pkg/client
API_DIR := ./api

# Generate transport layer from contracts
generate: transport client swagger
	goimports -l -w $(TRANSPORT_DIR)
	goimports -l -w $(CLIENT_DIR)

# Generate transport server
transport:
	$(TG) transport --services $(CONTRACT_DIR) --out $(TRANSPORT_DIR)

# Generate Go client
client:
	$(TG) client -go --services $(CONTRACT_DIR) --outPath $(CLIENT_DIR)/go

# Generate JS/TS client
client-js:
	$(TG) client -js --services $(CONTRACT_DIR) --outPath $(CLIENT_DIR)/js

# Generate OpenAPI spec
swagger:
	$(TG) swagger --services $(CONTRACT_DIR) --outFile $(API_DIR)/swagger.yaml

# Build server binary
build:
	go build -o bin/server ./cmd/server

# Run development server
run:
	go run ./cmd/server

# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air

# Docker operations
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Database migrations
migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

# Tests
test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Linting (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf $(TRANSPORT_DIR)/*
	rm -rf $(CLIENT_DIR)/*
