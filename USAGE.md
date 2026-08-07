# Ducalis TG - Usage

## Prerequisites

- Go 1.26+
- Docker and Docker Compose (for infrastructure)
- tg code generator (optional, for regenerating transport)

## Install

```bash
git clone https://github.com/sah4ez/ducalis-tg.git
cd ducalis-tg
go mod download
```

## Start

### Option 1: Run services directly (requires PostgreSQL running)

```bash
# Start infrastructure
docker-compose up -d postgres redis

# Run each service in a separate terminal
BIND=:8080  JWT_SECRET=my-secret go run ./cmd/server-public
BIND=:8082  ADMIN_JWT_SECRET=admin-secret go run ./cmd/server-admin
BIND=:8083  INTERNAL_API_KEY=internal-key go run ./cmd/server-internal
```

### Option 2: Build and run binaries

```bash
make build
./bin/server-public   # :8080
./bin/server-admin    # :8082
./bin/server-internal # :8083
```

### Option 3: Docker Compose (full stack)

```bash
docker-compose up --build
```

## Verify

```bash
# Health checks (should return 200 with JSON)
curl -s http://localhost:8080/health
curl -s http://localhost:8082/health
curl -s http://localhost:8083/health

# Register a user (public API)
curl -s -X POST http://localhost:8080/api/v1/auth/authService \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"AuthService.Register","params":[{"name":"Test","email":"test@example.com","password":"secret"}],"id":"1"}'

# Admin stats
curl -s http://localhost:8082/admin/v1/stats

# Readiness check (internal API)
curl -s http://localhost:8083/ready
```

## Stop

```bash
# Stop locally running binaries: Ctrl+C in each terminal

# Stop Docker services
docker-compose down

# Force stop on specific port
lsof -ti:8080 | xargs kill -9
lsof -ti:8082 | xargs kill -9
lsof -ti:8083 | xargs kill -9
```

## Environment Variables

See `.env.example` for the full list. Key variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | postgres://ducalis:ducalis123@localhost:5432/ducalis?sslmode=disable | PostgreSQL connection |
| `JWT_SECRET` | (required) | JWT signing secret for public API |
| `ADMIN_JWT_SECRET` | (required) | JWT signing secret for admin API |
| `INTERNAL_API_KEY` | (optional) | API key for internal API auth |
| `KAFKA_BROKERS` | localhost:9092 | Kafka broker addresses |
| `REDIS_URL` | redis://localhost:6379/0 | Redis connection |
| `BIND` | :8080 | Service bind address |
| `LOG_LEVEL` | info | Log level (debug, info, warn, error) |
