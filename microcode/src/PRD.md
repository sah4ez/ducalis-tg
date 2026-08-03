# PRD — ducalis-tg: подключение сгенерированного транспорта и запуск сервисов

> Этот PRD — вводная для loki (microcode). Полная методология — в overlay-модулях
> `/workspace/skills/*.md` (читай `00-index.md` ПЕРВЫМ). Стек и текущее состояние
> описаны ниже. Цель: довести скаффолд до работающего end-to-end.

## Контекст проекта

**ducalis-tg** — open-source self-hostable платформа приоритизации задач
(альтернатива hi.ducalis.io). Команды скорят задачи фреймворками (RICE, ICE,
WSJF, кастомные взвешенные критерии), голосуют с настраиваемыми весами, добавляют
оценки (story points / часы), трекают зависимости/блокеры, синхронизируют задачи
из GitHub/Jira/Linear/Внутренние трекеры. Multi-tenant через изоляцию workspace. Event-driven через Kafka.

**Название НЕ значит Telegram:** «tg» = [tg Transport Generator](https://github.com/seniorGolang/tg)
(`github.com/seniorGolang/tg/v3`) — контракт-first кодогенератор JSON-RPC 2.0
поверх HTTP. Telegram-бота в проекте НЕТ.

## Стек

| слой | технология |
|---|---|
| язык | Go 1.26 (`go.mod`, миграция с 1.24 — tg v3 требует 1.26+) |
| web | gofiber/fiber/v2 (v2.52.8) |
| кодогенерация | tg v3 (`github.com/seniorGolang/tg/v3`) + плагины tgp-go (astg+server, WASM/wazero). Команда `tg server -o internal/transport` (НЕ v2 `tg transport`). |
| transport | JSON-RPC 2.0 over HTTP (батчинг) + REST routes + OpenAPI |
| DB | PostgreSQL 16 (`jackc/pgx/v5` через `pgxpool`) |
| кэш/сессии | Redis 7 |
| события | Apache Kafka (confluent cp-kafka 7.5.0 + Zookeeper) |
| auth | golang-jwt/jwt/v5 (HS256), bcrypt (golang.org/x/crypto) |
| ID | google/uuid |
| логи | rs/zerolog (JSON в stdout) |
| метрики | prometheus/client_golang (:9090) |
| трейсинг | OpenTelemetry (OTLP/gRPC) |

## Архитектура (3 сервиса по портам)

| сервис | порт | контейнер | auth | зависимости |
|---|---|---|---|---|
| public | 8080 | ducalis-public | JWT (`JWT_SECRET`) | postgres, redis, kafka |
| admin | 8082 | ducalis-admin | admin JWT (`ADMIN_JWT_SECRET`) | postgres, redis |
| internal | 8083 | ducalis-internal | API-key (`INTERNAL_API_KEY`) | postgres, redis, kafka |
| kafka-ui | 8081 (host) | ducalis-kafka-ui | — | kafka |

Подробности — `multi-service.md`.

## ТЕКУЩЕЕ СОСТОЯНИЕ: МИГРАЦИЯ v2→v3 + WIRING

Проект был на tg v2 (`VersionTg=v2.3.95`, `go 1.24`, `tg/v2/cmd/tg`,
`tg transport --services`). **Старый v2-сгенерированный транспорт УДАЛЁН** —
его нужно регенерировать через v3 `tg server`. Затем — wireить в main.go.

| компонент | файлы | статус |
|---|---|---|
| контракты | `pkg/contract/{task,workspace,integration}.go` | v2-формат `// @tg` (проверь совместимость с v3 через `tg plugin doc astg`) |
| доменные типы | `pkg/types/{task,workspace,integration}.go` | ✅ готовы |
| **сгенерированный транспорт** | `internal/transport/` | ❌ **УДАЛЁН (v2)** — регенерировать `tg server -o internal/transport`. Hand-written `tracer/`, `context/`, `viewer/` сохранены. |
| клиенты | `pkg/client/{go,ts}` | hand-written, ✅ (НЕ трогать; опц. регенерировать `tg client-go`) |
| бизнес-логика | `internal/service/` (AuthService, TaskService, WorkspaceService, AdminService + репо-интерфейсы) | ✅ готов |
| postgres-репозитории | `internal/storage/postgres/` (db.go: `New(url)` pgxpool, repositories.go: User/Workspace/Member/Task/Vote/Estimation/Integration) | ✅ готовы |
| DDL | `migrations/init.sql` | ✅ готов |
| docker-compose | postgres, redis, zookeeper, kafka, kafka-ui, 3 servers | ✅ готов |
| **go.mod** | `go 1.24`, tg не подключён | ❌ нужно `go 1.26` + `go get github.com/seniorGolang/tg/v3@latest` |
| **Makefile** | v2-команды (`tg transport`, `install-tg → /v2/`), `*_CONTRACT` указывают на несуществующие поддиректории | ❌ обновить под v3 |
| **entry points** | `cmd/server-{public,admin,internal}/main.go` | ❌ **STUB** (connectDB=nil, transport не подключён) |
| тесты | `*_test.go` | ❌ отсутствуют |

### Что не так в main.go (одинаково во всех трёх)

1. `connectDB(url string) (*sql.DB, error)` возвращает `nil, nil` — stub.
2. Использует `database/sql`, но репозитории хотят `*pgxpool.Pool`.
3. Создаёт голый fiber-app только с `/health` + `/` (info).
4. Создание stores/services и `transport.New(...)` — **закомментированы** (TODO).
5. `srv.RegisterHandlers(app)` — закомментирован.

## ЦЕЛЬ (Definition of Done)

Две фазы: **(1) миграция на tg v3**, **(2) wiring + end-to-end запуск**. Три
сервиса ducalis-tg запускаются end-to-end: коннектятся к postgres/redis/kafka
(docker-compose внутри VM), обслуживают контрактные эндпоинты через
**v3-сгенерированный** (`tg server`) транспорт, отвечают на health/ready,
покрыты тестами.

### Критерии приёмки — Фаза 1: миграция v2→v3

- [ ] `go.mod`: `go 1.26` (с 1.24); `github.com/seniorGolang/tg/v3@latest` подключён;
      `go mod tidy` чистый (`GONOSUMDB=off GONOSUMCHECK=* GOPROXY=https://proxy.golang.org,direct`).
- [ ] `tg pkg list` показывает `astg` + `server`.
- [ ] `tg server -o internal/transport` выполнено — top-level файлы сгенерированы
      (server.go, options.go, *service-*.go). `tracer/`, `context/`, `viewer/` целы.
- [ ] `Makefile`: `install-tg` → `/v3/cmd/tg`; `transport` → `tg server -o`;
      `*_CONTRACT` указывают на реальный плоский `pkg/contract/`.
- [ ] `go build ./...` чистый (после миграции контрактов/транспорта).

### Критерии приёмки — Фаза 2: wiring + запуск

- [ ] `cmd/server-internal/main.go`: `connectDB()` → `postgres.New(url)` (возвращает
      `*pgxpool.Pool`); сервисы смонтированы через `transport.New(logger, ...)`
      (точные имена опций v3 — СВЕРЬ в regenerated options.go); auth через
      `INTERNAL_API_KEY`; `/ready` check; `make build-internal` ✓;
      `curl :8083/health` → 200.
- [ ] `cmd/server-public/main.go`: AuthService + WorkspaceService + TaskService
      смонтированы; `make build-public` ✓; `curl :8080/health` → 200;
      `POST /api/v1/auth/register` → создаёт юзера, возвращает JWT.
- [ ] `cmd/server-admin/main.go`: AdminService на `ADMIN_JWT_SECRET`; `make build-admin` ✓;
      `curl :8082/health` → 200.
- [ ] `make build` (все 3 бинарника) ✓.
- [ ] `make test` — smoke-тесты на каждый сервис + репозитории (table-driven,
      реальная Postgres через docker-compose). См. `tdd-rules.md`.
- [ ] Сервисы биндят `0.0.0.0:8080/8082/8083`. См. `infra-in-vm.md`.
- [ ] `security-checks.md` пройден.

## План работ (рекомендованный порядок)

### Фаза 1: миграция v2→v3 (ПЕРВИЧНО)
1. **Изучить overlay-модули** (`00-index.md` → `tg-v3-contract.md` → `service-wiring.md`).
2. **go.mod**: `go 1.24` → `go 1.26`; `go get github.com/seniorGolang/tg/v3@latest`;
   `GONOSUMDB=off GONOSUMCHECK=* GOPROXY=https://proxy.golang.org,direct go mod tidy`.
3. **Makefile**: `install-tg` → `/v3/cmd/tg`; `transport` → `tg server -o internal/transport`;
   `*_CONTRACT` → реальный плоский `pkg/contract/`.
4. **Регенерировать транспорт**: `tg server -o internal/transport` (после правки
   контрактов под v3-синтаксис, если `tg plugin doc astg` покажет расхождения).
5. **Проверить**: `tg pkg list` (astg+server), `go build ./...` чистый.

### Фаза 2: wiring + запуск
6. **Поднять инфру в VM:** `cd /workspace && docker-compose up -d`, дождаться healthy.
7. **Сверить сигнатуры:** postgres-репозитории vs service-интерфейсы vs опции
   `transport.New` в regenerated `internal/transport/options.go` (v3 API!).
8. **Wiring internal (:8083)** — простейший. TDD: smoke-тест.
9. **Wiring public (:8080)** — самый объёмный (auth + workspace + task + vote). TDD.
10. **Wiring admin (:8082)** — admin-операции. TDD.
11. **Полный `make build` + `make test` + `curl`-проверки всех эндпоинтов.**
12. **`security-checks.md`** перед переходом в QA.
13. **Коммиты на `vm/ducalis-build`**, git-daemon :9418 отдаёт результат хосту (`git-sync.md`).

## Не-цели (out of scope для этой итерации)

- Реализация новых фич сверх того, что уже в контрактах (`pkg/contract/`).
- CI/CD pipeline (отдельная задача; сейчас нет `.github/`).
- Telegram-бот (его нет и не планируется — «tg» = Transport Generator).
- Деплой в реальное облако (работаем внутри VM через docker-compose).

## Источники правды (читай в этом порядке)

1. `README.md` в корне репо — обзор проекта.
2. `/workspace/skills/00-index.md` — overlay-модули loki (ЧИТАТЬ ПЕРВЫМ).
3. `/workspace/skills/tg-v3-contract.md` — v3 codegen, annotation vocabulary, never-rules.
4. `/workspace/skills/service-wiring.md` — миграция v3 + пошаговый wiring main.go.
5. `docker-compose.yaml` — env-переменные, порты, зависимости.
6. `migrations/init.sql` — схема БД, view `ranked_tasks`.
7. `internal/transport/options.go`, `server.go` — API сгенерированного транспорта (ПОСЛЕ `tg server`).
8. `internal/service/*.go` — сигнатуры `New*` + репо-интерфейсы.
9. `internal/storage/postgres/db.go`, `repositories.go` — API репозиториев.
