---
name: service-wiring
description: >-
  Migrate ducalis-tg from tg v2 to v3 AND wire the generated transport,
  business services, and postgres repositories into cmd/server-{public,admin,
  internal}/main.go. The old v2 transport was DELETED; v3 transport must be
  regenerated via `tg server`. go.mod still on go 1.24 (needs 1.26). connectDB
  returns nil. This module covers both tasks: migration first, wiring second.
  Use when connecting any service layer to the entry points.
---

# Service wiring — миграция v3 + подключение

Две задачи, по порядку: **(1) МИГРАЦИЯ v2→v3**, **(2) WIRING сервисов в main.go.**

## Текущее состояние

- `pkg/contract/*.go` — контракты сервисов ✓ (v2-формат `// @tg`, в основном
  совместим с v3 — проверь `tg plugin doc astg` после регенерации)
- `internal/transport/` — **СТАРЫЙ v2-транспорт УДАЛЁН**. Остались только
  hand-written поддиректории: `tracer/` (4), `context/` (1), `viewer/` (6).
  Top-level сгенерированные файлы отсутствуют → **нужно регенерировать `tg server`**.
- `internal/service/*.go` — бизнес-логика ✓ (AuthService, TaskService,
  WorkspaceService + конструкторы `New*`)
- `internal/storage/postgres/` — репозитории ✓ (db.go: `postgres.New(url)`,
  repositories.go: UserRepo/WorkspaceRepo/TaskRepo/VoteRepo/EstimationRepo/...)
- `migrations/init.sql` — DDL ✓
- `go.mod` — **`go 1.24`** (tg v3 требует `go 1.26`); tg НЕ подключён как
  зависимость (это CLI кодогенерации, но v3 требует `go get tg/v3@latest`).
- `Makefile` — v2-команды (`tg transport`, `tg client -go`, `tg swagger`),
  `install-tg` → `/v2/`. Ссылается на `pkg/contract/public|admin|internal/`,
  которых НЕТ (контракт плоский).
- `cmd/server-*/main.go` — **STUB**: `connectDB()` возвращает `nil, nil`,
  fiber-апп только с `/health`.

## ЗАДАЧА 1: Миграция v2→v3 (ПЕРВИЧНАЯ — до любого wiring)

### 1.1. go.mod
```bash
# bump go directive 1.24 → 1.26 (tg v3 требует 1.26+)
# подключить фреймворк (/v3 суффикс обязателен):
go get github.com/seniorGolang/tg/v3@latest
# tidy под прокси (доказано в test-todo2):
GONOSUMDB=off GONOSUMCHECK=* GOPROXY=https://proxy.golang.org,direct go mod tidy
```

### 1.2. Makefile
- `install-tg`: `go install github.com/seniorGolang/tg/v2/cmd/tg` → `/v3/cmd/tg`.
- `transport`: `tg transport --services X --out Y` → `tg server -o internal/transport`
  (v3-синтаксис; проверь `tg plugin doc server` для точных флагов).
- `client-go`/`client-ts`/`swagger`: v3-синтаксис (`tg client-go -o pkg/client/go`).
- `*_CONTRACT` переменные: Makefile ждёт `pkg/contract/public|admin|internal/`
  (НЕ существуют) — приведи к реальному плоскому `pkg/contract/`.

### 1.3. Регенерировать транспорт (ПЕРВЫЙ шаг после go.mod/Makefile)
```bash
tg server -o internal/transport    # Fiber server (ОБЯЗАТЕЛЬНО)
tg pkg list                        # проверка: astg + server установлены
```
После этого в `internal/transport/` появятся top-level сгенерированные файлы
(server.go, options.go, *service-*.go и т.д.) — v3-формата. **НЕ редактируй
их руками.**

См. `tg-v3-contract.md` для всех v3-правил и never-списка.

## ЗАДАЧА 2: Wireить v3-транспорт + сервисы в main.go

Три main'а идентичны по структуре, различаются набором сервисов и env
(см. `multi-service.md` — какой сервис куда). Шаги:

### 2.1. Заменить stub connectDB() на реальный postgres.New()

`connectDB(url string) (*sql.DB, error)` сейчас возвращает `nil, nil` и
использует `database/sql`. Но репозитории в `internal/storage/postgres/`
работают с **pgxpool** (`pgxpool.Pool`), не `*sql.DB`. Нужно:

```go
import (
    "github.com/sah4ez/ducalis-tg/internal/storage/postgres"
    // НЕ database/sql — репозитории хотят *pgxpool.Pool
)

// в main():
pool, err := postgres.New(dbURL)   // см. internal/storage/postgres/db.go
if err != nil {
    logger.Fatal().Err(err).Msg("failed to connect to database")
}
defer pool.Close()
```

Проверь сигнатуру `postgres.New` в `internal/storage/postgres/db.go` — обёртка
над `pgxpool.New` (max 25 / min 5 conns, 1h lifetime, healthcheck). Если main
продолжает возвращать `*sql.DB` из сигнатуры — поменяй сигнатуру, т.к.
репозиториям нужен pool.

### 2. Создать репозитории из pool

```go
userRepo        := postgres.NewUserRepository(pool)         // имена — СВЕРЬ с repositories.go
workspaceRepo   := postgres.NewWorkspaceRepository(pool)
taskRepo        := postgres.NewTaskRepository(pool)
voteRepo        := postgres.NewVoteRepository(pool)
estimationRepo  := postgres.NewEstimationRepository(pool)
integrationRepo := postgres.NewIntegrationRepository(pool)
memberRepo      := postgres.NewMemberRepository(pool)
```

**СВЕРЬ фактические имена конструкторов** в `internal/storage/postgres/repositories.go`
— они должны удовлетворять интерфейсам из `internal/service/*.go`
(`UserRepository`, `WorkspaceRepository`, `TaskRepository`, ...). Если сигнатуры
не совпадают — это первый фикс (адаптер или правка).

### 3. Создать сервисы (бизнес-логика)

```go
authSvc      := service.NewAuthService(userRepo, logger, jwtSecret)
workspaceSvc := service.NewWorkspaceService(workspaceRepo, memberRepo, userRepo, logger)
taskSvc      := service.NewTaskService(taskRepo, logger)
// integrationSvc, adminSvc — по необходимости (см. multi-service.md)
```

**СВЕРЬ сигнатуры `New*`** в `internal/service/*.go` — конструкторы принимают
репозитории + logger (+ опционально secret). Если сигнатура изменилась, адаптируй.

### 4. Создать transport-сервер и зарегистрировать в fiber-app

**ВНИМАНИЕ: v3 API отличается от v2.** Точные имена конструктора `transport.New`,
опций (`With*Service` vs `New*Service` vs иное), и метода регистрации
(`RegisterHandlers` vs `Serve` vs `Fiber()`) появятся ТОЛЬКО после регенерации
транспорта через `tg server` (задача 1.3). СВЕРЬ фактический API в
`internal/transport/options.go` + `server.go` после генерации — не полагайся на
v2-имена. Форма (приблизительная):

```go
import "github.com/sah4ez/ducalis-tg/internal/transport"

// v3: имена опций СВЕРЬ в regenerated internal/transport/options.go.
// Возможные варианты: transport.WithAuthService(...) / transport.AuthService(...)
srv := transport.New(logger,
    transport.WithAuthService(authSvc),       // ИМЯ ПРОВЕРЬ после tg server
    transport.WithWorkspaceService(workspaceSvc),
    transport.WithTaskService(taskSvc),
)
srv.RegisterHandlers(app)   // ИЛИ srv.Serve(app) / srv.Fiber() — СВЕРЬ в server.go
```

Сгенерированный v3 транспорт даёт (структура та же, имена файлов могут отличать):
- JSON-RPC 2.0 батч-эндпоинт `POST /`
- REST-роуты (HTTP-server)
- middleware (log/metrics/trace — если включены в `// @tg` на интерфейсе)

### 5. Env-переменные (НЕ config.yaml)

`config/config.yaml` в репо **вестигиальный** — его НИКТО не загружает (нет
viper/кой-кой-кой). Живой конфиг — env через `getEnv()` в каждом main. Сверь
с docker-compose.yaml:

| env | default | сервис |
|---|---|---|
| `DATABASE_URL` | postgres://ducalis:ducalis123@localhost:5432/ducalis?sslmode=disable | все |
| `REDIS_URL` | redis://localhost:6379/0 | все |
| `KAFKA_BROKERS` | localhost:9092 | public, internal |
| `BIND` | :8080 (public) / :8082 (admin) / :8083 (internal) | каждый свой |
| `LOG_LEVEL` | debug | все |
| `JWT_SECRET` | change-me-in-production | public |
| `ADMIN_JWT_SECRET` | change-admin-secret-in-production | admin |
| `INTERNAL_API_KEY` | change-internal-api-key-in-production | internal |

## Порядок работы (рекомендация)

1. **Сначала internal** (:8083) — простейший (API-key auth, Kafka consumer).
   На нём отработай паттерн wiring, проверь что `make build-internal` +
   `curl :8083/health` + запрос к контрактному эндпоинту работает.
2. **Потом public** (:8080) — самый объёмный (auth + workspace + task + vote).
3. **Последним admin** (:8082) — admin-операции над users/workspaces.

После каждого: `make build-<svc>` + запуск + `curl` health/контрактный эндпоинт.
Тесты — обязательно (см. `tdd-rules.md`).

## Definition of done (для одного main)

- [ ] `connectDB()` заменён на `postgres.New(url)`, возвращает `*pgxpool.Pool`.
- [ ] Репозитории созданы из pool, удовлетворяют service-интерфейсам.
- [ ] Сервисы созданы, переданы в `transport.New(logger, With*Service(...))`.
- [ ] `srv.RegisterHandlers(app)` (или эквивалент) вызван в main.
- [ ] `make build-<svc>` собирается без ошибок.
- [ ] `make test` зелёный (хотя бы smoke-тест на запущенном сервисе).
- [ ] `curl http://localhost:<port>/health` → 200.
- [ ] Запрос к контрактному эндпоинту (через сгенерированный клиент или curl) → корректный ответ.

## Чего НЕ делать

- НЕ возвращай `*sql.DB` из connectDB, если репозитории хотят `*pgxpool.Pool`.
- НЕ создавай второй fiber-app для контрактных эндпоинтов — используй уже
  существующий `app` + `srv.RegisterHandlers(app)`.
- НЕ подключай `config/config.yaml` — он не загружается, живой конфиг = env.
- НЕ редактируй `internal/transport/*` — сгенерировано. Если не хватает опции
  `With*Service` — значит контракта для этого сервиса нет или регенерация нужна.
