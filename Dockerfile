# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate transport layer
RUN go install github.com/seniorGolang/tg/cmd/tg@v2
RUN tg transport --services ./pkg/contract --out ./internal/transport

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /priora ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary
COPY --from=builder /priora /app/priora

# Copy migrations
COPY migrations/ /app/migrations/

# Create non-root user
RUN addgroup -g 1000 priora && \
    adduser -u 1000 -G priora -s /bin/sh -D priora

USER priora

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9091/health || exit 1

ENTRYPOINT ["/app/priora"]
