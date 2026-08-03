---
name: postgres-repo
description: >-
  Work with the ducalis-tg PostgreSQL persistence layer — internal/storage/postgres
  (pgxpool), migrations/init.sql, and the service-layer repository interfaces.
  Use when writing repository methods, SQL, dealing with JSONB columns, NotFoundError,
  the ranked_tasks view, or applying migrations.
---

# PostgreSQL persistence — ducalis-tg

## Stack

- **PostgreSQL 16** (postgres:16-alpine в docker-compose).
- **driver:** `github.com/jackc/pgx/v5` через `pgxpool` (pool-обёртка в
  `internal/storage/postgres/db.go`: max 25 / min 5 conns, 1h lifetime, healthcheck).
  Также `database/sql` для null-handling в паре мест.
- **migrations:** один файл `migrations/init.sql`, idempotent
  (`CREATE TABLE IF NOT EXISTS`). Применяется через docker entrypoint
  (`/docker-entrypoint-initdb.d/init.sql`) или `make migrate-up` (raw `psql`).
  **Нет** migration-tool (golang-migrate/goose) — просто SQL.

## Таблицы (migrations/init.sql)

`users`, `workspaces`, `members`, `tasks`, `task_dependencies`, `votes`,
`estimations`, `integrations`. Плюс триггеры `updated_at` и **view `ranked_tasks`**
(window-функции для rank/percentile — основа приоритизации).

## Репозитории (internal/storage/postgres/repositories.go)

`UserRepository`, `WorkspaceRepository`, `MemberRepository`, `TaskRepository`,
`VoteRepository`, `EstimationRepository`, `IntegrationRepository` — все на pgxpool,
hand-written SQL. **Сверь фактические имена конструкторов** (`New*Repository`) —
они должны удовлетворять интерфейсам из `internal/service/*.go`.

Ключевые паттерны:

### JSONB-колонки
`scoring_config`, `scores`, `metadata`, `config` — маршалятся через
`encoding/json` (struct → []byte → JSONB). При чтении — unmarshal в типы из
`pkg/types/`.

### NotFoundError
Кастомный тип `NotFoundError` — возвращается, когда `pgx.ErrNoRows`. Сервисы
проверяют `errors.Is(err, postgres.NotFoundError)` → 404. НЕ возвращай голый
`pgx.ErrNoRows` наверх — оборачивай.

### Service-интерфейсы vs postgres-реализации
Интерфейсы (`TaskRepository`, `WorkspaceRepository`, ...) определены в
`internal/service/*.go`. postgres-структуры их удовлетворяют. Если при wiring
(`service-wiring.md`) компилятор ругается на несоответствие сигнатур — это
первый фикс (адаптер или правка одной из сторон).

## ranked_tasks view

View `ranked_tasks` считает rank/percentile через window-функции над `tasks`
(с учётом scoring config из workspace). Для отчётов/дашбордов — селектить из
view, не пересчитывать в Go. Сверь точные колонки в init.sql.

## Миграции

Нет tooling — чистый SQL. Новая миграция = дописать блок в `migrations/init.sql`
(idempotent, `IF NOT EXISTS`). Применение:

```bash
make migrate-up        # psql $DATABASE_URL -f migrations/init.sql
make migrate-reset     # снести volyumes и заново (destructive!)
```

Внутри VM (см. infra-in-vm.md): `docker-compose up -d postgres` применит init.sql
автоматически (entrypoint). `postgresql-client` стоит в snapshot для ручных запросов:
`psql "$DATABASE_URL" -c 'SELECT * FROM ranked_tasks LIMIT 5;'`.

## Чего НЕ делать

- НЕ пиши миграции отдельными файлами без tooling — проект использует один init.sql.
- НЕ используй `database/sql` там, где репозитории хотят `pgxpool.Pool`.
- НЕ возвращай `pgx.ErrNoRows` голым — оборачивай в `NotFoundError`.
- НЕ дублируй логику rank/percentile в Go — есть view `ranked_tasks`.
