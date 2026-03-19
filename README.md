# Ducalis TG

🎯 **Open-source task prioritization platform** built with [tg](https://github.com/seniorGolang/tg) code generator.

Alternative to [Ducalis](https://hi.ducalis.io/), fully self-hostable.

## ✨ Features

- **Scoring Frameworks**: RICE, ICE, WSJF, or custom criteria with weighted formulas
- **Team Voting**: Democratic prioritization with configurable vote weights
- **Estimations**: Story points or hours from team members
- **Dependencies**: Blockers and task relationships
- **Integrations**: GitHub Issues, Jira, Linear (sync both ways)
- **Event-Driven**: Kafka-based async event processing
- **Multi-Tenant**: Workspace-based isolation

## 🏗 Architecture

Three independent API services:

1. **Public API** (port 8080) - End user operations with JWT auth
2. **Admin API** (port 8082) - System administration with separate auth
3. **Internal API** (port 8083) - Service-to-service with API key auth

## 🚀 Quick Start

```bash
# Clone repository
git clone https://github.com/sah4ez/ducalis-tg.git
cd ducalis-tg

# Install tg generator
go install github.com/seniorGolang/tg/v2/cmd/tg@latest

# Generate transport layer and clients
make generate

# Start infrastructure
make docker-up

# Run server
make run
```

Services will start on:
- Public API: http://localhost:8080
- Admin API: http://localhost:8082
- Internal API: http://localhost:8083
- Kafka UI: http://localhost:8081

## 🛠 Tech Stack

- **Backend**: Go 1.24 with [tg](https://github.com/seniorGolang/tg) code generator
- **Transport**: JSON-RPC 2.0 + HTTP (go-fiber)
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Message Broker**: Apache Kafka
- **Observability**: Prometheus metrics, OpenTelemetry tracing

## 📦 Project Structure

```
ducalis-tg/
├── cmd/server/              # Entry point
├── pkg/
│   ├── contract/            # API contracts (source of truth)
│   ├── types/               # Domain types
│   └── client/              # GENERATED clients (Go, TS)
├── internal/
│   ├── service/             # Business logic
│   ├── storage/             # Database layer
│   └── transport/           # GENERATED transport layer
├── migrations/              # SQL migrations
├── api/                     # GENERATED OpenAPI specs
└── config/                  # Configuration
```

## 🔧 Development

### Generate Code

```bash
# Generate all
make generate

# Generate only transport
make transport

# Generate only clients
make client

# Generate only OpenAPI specs
make swagger
```

### Build & Run

```bash
# Build binary
make build

# Run development server
make run

# Run with hot reload (requires air)
make dev
```

### Docker

```bash
# Build and run all
docker-compose up -d

# View logs
docker-compose logs -f app

# Rebuild app
docker-compose build app
```

## 📝 API Endpoints

### Public API (`:8080`)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/v1/auth/register` | POST | Register user |
| `/api/v1/auth/login` | POST | Login |
| `/api/v1/workspaces` | GET, POST | List/Create workspaces |
| `/api/v1/tasks` | GET, POST | List/Create tasks |
| `/api/v1/tasks/{id}/scores` | PUT | Set task scores |
| `/api/v1/tasks/{id}/vote` | POST | Vote for task |

### Admin API (`:8082`)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/admin/v1/users` | GET | List all users |
| `/admin/v1/workspaces` | GET | List all workspaces |
| `/admin/v1/stats` | GET | System statistics |

### Internal API (`:8083`)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/internal/v1/sync/tasks` | POST | Sync tasks from external |
| `/internal/v1/events` | POST | Push event |
| `/internal/v1/health` | GET | Health check |

## 📊 Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | postgres://... | PostgreSQL connection |
| `REDIS_URL` | redis://... | Redis connection |
| `KAFKA_BROKERS` | localhost:9092 | Kafka brokers |
| `PUBLIC_BIND` | :8080 | Public API bind |
| `ADMIN_BIND` | :8082 | Admin API bind |
| `INTERNAL_BIND` | :8083 | Internal API bind |
| `JWT_SECRET` | - | User JWT secret |
| `LOG_LEVEL` | info | Log level |

## 🗺 Roadmap

- [ ] WebSocket for real-time updates
- [ ] GitHub/Jira/Linear integrations
- [ ] Bulk import/export (CSV, JSON)
- [ ] Webhooks for external notifications
- [ ] Analytics dashboard
- [ ] SSO integration (SAML, OIDC)
- [ ] Frontend UI (React/Vue)

## 📄 License

MIT

## 🙏 Credits

- Built with [tg](https://github.com/seniorGolang/tg) - contract-first code generator
- Inspired by [Ducalis](https://hi.ducalis.io/)
