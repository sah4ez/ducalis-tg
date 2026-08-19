# Frontend build stage
FROM node:20-alpine AS webbuilder
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install
COPY web/ .
RUN npm run build

# Build stage
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make bash

WORKDIR /build

# Copy go mod files first for caching
COPY go.mod go.sum* ./
RUN go mod download || go mod tidy

# Copy source code
COPY . .

# Build all binaries
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o /server-public ./cmd/server-public

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o /server-admin ./cmd/server-admin

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o /server-internal ./cmd/server-internal

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /server-public /app/server-public
COPY --from=builder /server-admin /app/server-admin
COPY --from=builder /server-internal /app/server-internal
COPY --from=webbuilder /web/dist /app/web/dist

# Copy migrations (for manual migration runs)
COPY migrations/ /app/migrations/

# Create non-root user
RUN addgroup -g 1000 ducalis && \
    adduser -u 1000 -G ducalis -s /bin/sh -D ducalis

# Set ownership
RUN chown -R ducalis:ducalis /app

USER ducalis

# No default EXPOSE - each service uses different port
# Public: 8080
# Admin: 8082
# Internal: 8083

# No default ENTRYPOINT - specify which service to run via command
# Example:
#   docker run ducalis ./server-public
#   docker run ducalis ./server-admin
#   docker run ducalis ./server-internal
