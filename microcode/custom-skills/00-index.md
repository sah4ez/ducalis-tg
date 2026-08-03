# Skill Modules Index (microcode overlay) — ducalis-tg

Overlay поверх встроенных модулей loki (`skills/*.md`) + переведённых скиллов из
obra/superpowers. Loki грузит модули по фазе задачи — 1-3 за раз, по этой
таблице. Файлы с тем же именем здесь переопределяют встроенные.

**Проектный стек: Go 1.26 + `github.com/seniorGolang/tg/v3` (миграция с v2) +
gofiber/fiber/v2 + PostgreSQL 16 (pgx/v5) + Redis 7 + Apache Kafka.** НЕ Telegram
(несмотря на имя репо `ducalis-tg`; «tg» = Transport Generator). JSON-RPC 2.0
поверх HTTP.

**Текущее состояние репо: МИГРАЦИЯ v2→v3 + WIRING.** Контракты в `pkg/contract/`,
postgres-репозитории в `internal/storage/postgres/`, бизнес-логика в
`internal/service/` — есть. НО: старый **v2-сгенерированный транспорт УДАЛЁН**
(нужно регенерировать через `tg server`), `go.mod` ещё на `go 1.24` (нужен 1.26),
Makefile ссылается на v2-команды. `cmd/server-*/main.go` НЕ подключают транспорт
(`connectDB()` возвращает `nil, nil`). Две задачи: (1) МИГРАЦИЯ на v3, (2) WIRING.

## ⚠️ tg v3 codegen — ОБЯЗАТЕЛЬНО (не опционально)

Проект **мигрирует на tg v3** (`github.com/seniorGolang/tg/v3`, суффикс `/v3`
обязателен). v3 — codegen через WASM-плагины `tgp-go` (`astg` + `server`),
команда `tg server -o internal/transport` (НЕ v2 `tg transport --services`).
Плагины tgp-go (`astg` + `server`) уже стоят в VM (проверка: `tg pkg list`).
Контракты лежат в `pkg/contract/*.go` (ПЛОСКО — task.go, workspace.go,
integration.go, **НЕ** в public/admin/internal поддиректориях, хоть Makefile на
них и ссылается). Старый v2-транспорт удалён — регенерация `tg server` ПЕРВЫМ шагом.

Модуль `tg-v3-contract.md` грузится ВСЕГДА перед любым кодом, трогающим транспорт.

Load 1-3 modules based on your current task. Do not load all modules.

## Module Selection Rules

| If your task involves...                                    | Load these modules          | origin     |
|------------------------------------------------------------|-----------------------------|------------|
| **ЛЮБОЙ код, трогающий transport/контракты/main.go/go.mod**| **tg-v3-contract.md**       | overlay    |
| Миграция v2→v3 (go.mod, Makefile, `tg server`) — ПЕРВАЯ задача| **tg-v3-contract.md** + **service-wiring.md** | overlay |
| Подключение сервисов в cmd/server-*/main.go (вторая задача)| **service-wiring.md**       | overlay    |
| Работа с postgres-репозиториями, SQL, миграции             | postgres-repo.md            | overlay    |
| Координация 3 сервисов (public/admin/internal), порты, auth| multi-service.md            | overlay    |
| Запуск инфры (docker-compose: pg/redis/kafka) внутри VM    | infra-in-vm.md              | overlay    |
| Writing tests (table-driven, pgx, testcontainers)          | tdd-rules.md                | overlay    |
| Security checks, pre-QA transition                         | security-checks.md          | overlay    |
| Sync VM↔host, pulling remote, exposing loki result         | git-sync.md                 | overlay    |
| Model / tool selection                                     | model-selection.md          | loki built-in |
| Code review, quality checks                                | quality-gates.md            | loki built-in |
| Debugging, errors, failures                                | troubleshooting.md          | loki built-in |
| Architecture decisions                                     | patterns-advanced.md        | loki built-in |

## Overlay modules

### tg-v3-contract.md (ОБЯЗАТЕЛЬНЫЙ при работе с транспортом)
Когда: **ВСЕГДА** перед кодом, трогающим `internal/transport/`, `pkg/contract/`,
`cmd/server-*/main.go`, или `go.mod`/Makefile. Объясняет v3 codegen (`tg server`,
плагины tgp-go), annotation vocabulary, миграцию v2→v3 (go.mod → go 1.26,
`go get tg/v3@latest`), `go mod tidy` под прокси, never-правила (не импортировать
`tg/v3/skills`, не использовать v2 `tg transport`). **Старый v2-транспорт удалён —
регенерация `tg server` ПЕРВЫМ шагом.** tg v3 CLI + плагины tgp-go (astg+server)
уже стоят в VM.

### service-wiring.md
Когда: **главная задача проекта** — подключить сгенерированный transport +
services + postgres-репозитории в `cmd/server-public/main.go`,
`cmd/server-admin/main.go`, `cmd/server-internal/main.go`. Заменить stub
`connectDB()` на реальный `postgres.New()`, зарегистрировать `transport.New(...)`
с нужными `With*Service(...)` опциями. Пошаговый чек-лист на каждый main.

### postgres-repo.md
Когда: работа с `internal/storage/postgres/repositories.go`, SQL-запросами,
`migrations/init.sql`. Паттерны pgxpool, JSONB-колонки (`scoring_config`,
`scores`, `metadata`), `NotFoundError`, view `ranked_tasks`. Связь
сервис-интерфейсов (`TaskRepository`, `WorkspaceRepository`) и postgres-реализаций.

### multi-service.md
Когда: координация трёх сервисов. public (:8080, JWT), admin (:8082,
ADMIN_JWT), internal (:8083, INTERNAL_API_KEY + Kafka consumer). Разделение
контрактов, какие сервисы куда монтируются, env-переменные на каждый.
Конфиг через env (НЕ config.yaml — он вестигиальный).

### infra-in-vm.md
Когда: запуск `docker-compose up` внутри VM. Docker Engine + compose-plugin уже
в snapshot. Нюансы: образы тянутся из registry-1.docker.io (в allowlist), Kafka
тяжёлая — memory: 8192 подобран под это, переменная `BIND=:8080` в compose
означает 0.0.0.0 (доступно через port-forward).

### tdd-rules.md
Когда: реализация любой фичи/багфикса, ДО написания кода.
RED-GREEN-REFACTOR на `go test`. Table-driven, реальная Postgres через
testcontainers-go или docker-compose, без моков БД.

### security-checks.md
Когда: фаза DEVELOPMENT, перед переходом в QA.
Security-скан изменившихся файлов, проверка JWT-secrets/api-keys в коде.

### git-sync.md
Когда: **перед первым git-действием в VM** и при любой синхронизации с host.
**microcode уже склонировал** remote в `/workspace` (через `sandbox.sync`) до
старта loki — НЕ клонируй/не init'и сам. Работай на ветке `vm/<sandbox-name>`
(она уже активна), коммить туда, git-daemon :9418 отдаёт результат наружу для
host (fetch+merge). Никогда не пушить в remote из VM (master — read-only).

## How to Load
1. Read this index.
2. **Always read `tg-v3-contract.md` first when touching transport/contracts/main.go/go.mod.**
3. Pick 1-2 more modules matching the current phase/task.
4. Read those files.
5. Execute with loaded context.
