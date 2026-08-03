---
name: multi-service
description: >-
  Coordinate the three ducalis-tg API services — public (:8080, JWT),
  admin (:8082, ADMIN_JWT), internal (:8083, INTERNAL_API_KEY + Kafka). Which
  services/auth each main mounts, port assignments, env vars, and how they
  share postgres/redis/kafka. Use when wiring a specific service or deciding
  where a feature belongs.
---

# Multi-service architecture

Три независимых API-сервиса, разделённых по ответственности и порту
(commit `359169b feat: separate services by port`). Каждый — свой `cmd/server-*/main.go`,
свой контейнер в docker-compose, общая сеть `ducalis-net` + общие postgres/redis/kafka.

| сервис | порт | контейнер | auth | зависит от | env-секрет |
|---|---|---|---|---|---|
| public | 8080 | ducalis-public | JWT (HS256) | postgres, redis, kafka | `JWT_SECRET` |
| admin | 8082 | ducalis-admin | admin JWT | postgres, redis | `ADMIN_JWT_SECRET` |
| internal | 8083 | ducalis-internal | API-key | postgres, redis, kafka | `INTERNAL_API_KEY` |
| kafka-ui | 8081 (host→8080) | ducalis-kafka-ui | — | kafka | — |

## Что монтировать в каждый main (через transport.With*Service)

Сверь фактические имена опций в `internal/transport/options.go` + контракты в
`pkg/contract/`. Рекомендуемое разделение:

### public (cmd/server-public/main.go, :8080)
- **AuthService** (register/login, `JWT_SECRET`) — контракт в workspace.go
- **WorkspaceService** (CRUD workspaces + members) — контракт workspace.go
- **TaskService** (CRUD tasks, scoring RICE/ICE/WSJF, vote) — контракт task.go
- Kafka producer (события task created/scored) — если подключено

Эндпоинты (из main TODO): `/api/v1/auth/register`, `/api/v1/auth/login`,
`/api/v1/workspaces`, `/api/v1/tasks`, `/api/v1/tasks/:id/scores`,
`/api/v1/tasks/:id/vote`.

### admin (cmd/server-admin/main.go, :8082)
- **AdminService** (управление users, workspaces; блокировка; метрики) —
  см. `internal/service/services.go` (есть AdminService + ошибки + context keys)
- отдельный admin-auth на `ADMIN_JWT_SECRET` (НЕ тот же JWT_SECRET)

Эндпоинты: `/admin/v1/users`, `/admin/v1/workspaces`, ...

### internal (cmd/server-internal/main.go, :8083)
- **IntegrationService** (sync tasks из GitHub/Jira/Linear) — контракт integration.go
- **TaskService/WorkspaceService** для сервер-сервер операций
- **Kafka consumer** (читает события, обновляет scoring/ranking)
- auth через `INTERNAL_API_KEY` (header), expose `/ready` (readiness)

Эндпоинты: `/internal/v1/sync/tasks`, ...

## Auth — НЕ путать ключи

- public: `JWT_SECRET` → HS256 JWT, TTL 24h, refresh 168h (из config.yaml).
- admin: `ADMIN_JWT_SECRET` → ОТДЕЛЬНЫЙ секрет, отдельная логика.
- internal: `INTERNAL_API_KEY` → статический API-key в header.
Проверь `internal/service/services.go` — там AuthService + AdminService +
context keys для извлечения identity из запроса.

## Sharing postgres/redis/kafka

- **postgres**: один кластер, БД `ducalis`, юзер `ducalis`. Все три сервиса
  коннектятся к одной БД через `DATABASE_URL`. Изоляция — на уровне схем/таблиц,
  не БД. (config.yaml упоминает БД `priora` — это устаревшее имя, игнорируй;
  docker-compose создаёт `ducalis`.)
- **redis**: один инстанс `redis://redis:6379/0`. Сессии/кэш/токены.
- **kafka**: `KAFKA_BROKERS=kafka:9092`. public/internal — producer/consumer,
  admin — не использует.

## Порядок wiring (см. service-wiring.md)

1. internal (простейший, API-key) — отработать паттерн.
2. public (самый объёмный).
3. admin.

## Чего НЕ делать

- НЕ используй один `JWT_SECRET` для public и admin — они разделены намеренно.
- НЕ дублируй `config/config.yaml` env'ом — выбери одно (проект = env).
- НЕ поднимай второй postgres/redis/kafka для админки — общие инстансы.
