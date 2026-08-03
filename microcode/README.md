# microcode — ducalis-tg

Манифесты и overlay-навыки для разработки **ducalis-tg** с помощью [microcode](https://github.com/sah4ez/microcode):
loki (cline + GLM) работает внутри отдельной microsandbox VM, мигрирует проект
с tg v2 на **tg v3** (tgp-go плагины) и достраивает сервис до end-to-end
(wiring в `cmd/server-*/main.go`).

> **Важно про название:** `ducalis-tg` — это НЕ Telegram-бот. «tg» = [tg Transport
> Generator](https://github.com/seniorGolang/tg) (`github.com/seniorGolang/tg/v3`),
> контракт-first кодогенератор JSON-RPC 2.0 поверх HTTP (через WASM-плагины tgp-go).
> Telegram-функциональности в проекте нет.

## Структура

```
microcode/
├── platform.yaml          ← APPLY-манифест (boot из snapshot mcd-base)
├── build.yaml             ← BUILD-манифест (один раз построить snapshot)
├── custom-skills/         ← overlay-навыки loki (под специфику tg v3 + ducalis-tg)
│   ├── 00-index.md        ← loki читает ПЕРВЫМ (таблица фаза→модуль)
│   ├── tg-v3-contract.md  ← v3 codegen (tg server, плагины), annotation vocabulary
│   ├── tg-patch/          ← фикс race-condition в tg pkg add (install.go.patched)
│   ├── service-wiring.md  ← миграция v3 + wiring stub'ов в main.go
│   ├── multi-service.md   ← координация public/admin/internal (порты/auth)
│   ├── postgres-repo.md   ← pgxpool, JSONB, NotFoundError, view ranked_tasks
│   ├── infra-in-vm.md     ← docker-compose (pg/redis/kafka) внутри VM
│   ├── tdd-rules.md       ← table-driven, реальная Postgres
│   ├── security-checks.md ← pre-QA гейт
│   └── git-sync.md        ← VM↔remote↔host (git-daemon :9418)
├── src/
│   └── PRD.md             ← спецификация работы для loki
└── README.md              ← этот runbook
```

## Что в VM (snapshot `mcd-base`)

| компонент | версия | зачем |
|---|---|---|
| Debian | bookworm-slim | base image |
| Go toolchain | 1.26.5 | tg v3 требует `go 1.26+` (apt даёт 1.19) |
| tg CLI | v3 (`seniorGolang/tg/v3/cmd/tg@latest`) | кодогенерация (`tg server`) |
| tgp-go плагины | astg + server (v1.0.8) | WASM/wazero генерация Fiber-транспорта |
| tg skills | встроенные agent-skills | входят в состав tg v3 (NB: `tg skills install` — НЕ команда в v3.0.x) |
| golangci-lint | v1.62.0 | `make lint` |
| Docker Engine | 27.3.1 (static, без systemd) | поднять docker-compose проекта |
| docker-compose | v2.30.3 | postgres/redis/kafka/kafka-ui/3 servers |
| Node + bun | для loki-mode/skillkit | оркестратор loki |
| loki-mode + cline (node-shim) | npm global | RARV-цикл на GLM (z.ai) |
| skillkit | npm+bun global | доставка скиллов obra/superpowers |
| postgresql-client | apt | `make migrate-up` + ручные запросы |
| git-daemon | :9418 (0.0.0.0) | read-only отдача результата loki наружу |

> **race-patch fallback:** `tg pkg add` имеет детерминированный download-race
> (`...-skills.tar.gz: EOF`) под microsandbox network layer. Bootstrap
> автоматически патчит и пересобирает tg (clone v3.0.5 → cp install.go.patched →
> cross-compile), затем retry. См. `custom-skills/tg-patch/README.md`.

**Ресурсы VM:** 4 CPU, 8 GB RAM, 12 GB root disk (Kafka + Go module cache + docker images тяжёлые).

## Быстрый старт

### 0. Предварительные требования на хосте

```bash
# microcode CLI + conda env mcd (один раз, из репо microcode):
git clone https://github.com/sah4ez/microcode.git /tmp/microcode
cd /tmp/microcode
conda env create -f environment.yml      # создаёт env "mcd" + ставит microcode
conda activate mcd
microcode doctor                          # проверит msb + skillkit на PATH
# (msb = microsandbox CLI; см. README репо microcode, как поставить)
```

### 1. Секреты на хосте (НИКОГДА не в манифесте)

```bash
# z.ai / GLM (обязательно для cline-shim):
export CLINE_API_KEY=...
export ZAI_BUSINESS_BASE_URL=https://api.z.ai
export ZAI_OAUTH_ORIGIN=https://chat.z.ai
export ZAI_OAUTH_CLIENT_ID=client_...
```

> **git-sync уже настроен в манифесте.** `SYNC_REMOTE_URL` (`https://github.com/sah4ez/ducalis-tg.git`)
> и `SYNC_BRANCH` (`master`) заданы **inline** в `platform.yaml`/`build.yaml` — это не секреты.
> Публичный репо клонируется анонимно по HTTPS, **PAT не нужен** (push из VM всё равно
> запрещён — master read-only, см. `custom-skills/git-sync.md`). Поэтому `SYNC_REMOTE_TOKEN`
> git-sync настраивается **декларативно** в `sandbox.sync` (см. манифест) — host env
> не нужен. `github.com` auto-allowlisted. `SYNC_*` env-переменные убраны.

### 2. PRD живёт в репозитории (microcode/PRD.md в origin/master)

PRD запушен в `origin/master` как `microcode/PRD.md`. `sandbox.sync` git-clone
приносит его в `/workspace/microcode/PRD.md` вместе с кодом — отдельный mount
НЕ нужен (он конфликтовал с clone: примонтированный файл нельзя удалить при
очистке `/workspace`). loki запускается с `--prd microcode/PRD.md`.

```bash
# проверить что PRD в remote:
git show origin/master:microcode/PRD.md | head -3
```

### 3. ПОСТРОИТЬ snapshot (один раз, ~20-30 мин на arm64)

```bash
conda activate mcd
microcode build microcode/build.yaml      # → snapshot 'mcd-base'
```

### 4. APPLY (каждый последующий раз, секунды)

```bash
microcode validate microcode/platform.yaml
microcode plan    microcode/platform.yaml --prd microcode/PRD.md   # dry-run, посмотреть план
microcode apply   microcode/platform.yaml --prd microcode/PRD.md   # boot из snapshot + старт loki
```

### 5. Следить / управлять loki

```bash
microcode status microcode/platform.yaml              # фаза / коммиты / workspace
microcode steer  microcode/platform.yaml "focus on internal service first"
microcode rollback microcode/platform.yaml --to <hash>
microcode destroy microcode/platform.yaml             # снести VM + state
```

### 6. Забрать результат на хост

```bash
# Через git-daemon (read-only, :9418 проброшен):
cd /Users/aleksandrkozlenkov/go/src/github.com/sah4ez/ducalis-tg
git fetch git://localhost:9418/ vm/ducalis-build
git log --oneline FETCH_HEAD -20
git merge FETCH_HEAD            # или cherry-pick
```

## Доступ к сервисам из VM (после wiring)

```bash
curl http://localhost:8080/health   # public API
curl http://localhost:8082/health   # admin API
curl http://localhost:8083/health   # internal API
open http://localhost:8081          # kafka-ui (в браузере; host:8081 → guest:8080)
git fetch git://localhost:9418/ vm/ducalis-build   # git-daemon
```

> **0.0.0.0 bind:** сервисы в VM должны слушать `0.0.0.0` (или `:8080` = то же в Go),
> НЕ `127.0.0.1` — иначе port-forward даёт пустой ответ (curl exit 52). docker-compose
> `BIND=:8080` уже корректен. Подробнее — `custom-skills/infra-in-vm.md`.

## Ключевые design-решения манифеста

- **tg v3 + плагины tgp-go.** Проект мигрирует с v2 на `github.com/seniorGolang/tg/v3`.
  Codegen через WASM-плагины (`astg` + `server`), команда `tg server -o internal/transport`
  (НЕ v2 `tg transport`). Реализовано по образцу [test-todo2 из репо microcode](https://github.com/sah4ez/microcode/tree/master/test-todo2)
  (там v3 проверен end-to-end). Включён race-patch fallback для `tg pkg add`.
- **Go 1.26.5, не 1.24.** tg v3 требует `go 1.26+`; apt даёт 1.19, ставим с go.dev.
- **Старый v2-транспорт удалён.** 42 сгенерированных файла (VersionTg=v2.3.95)
  удалены; hand-written `tracer/`, `context/`, `viewer/` сохранены. Регенерация
  `tg server` — первая задача loki.
- **Контракт плоский.** Файлы в `pkg/contract/*.go`, НЕ в поддиректориях
  `public/admin/internal` (хоть Makefile на них ссылается — это известное
  расхождение, отмечено в overlay).
- **PostgreSQL+Redis+Kafka, не SQLite.** Поэтому memory 8192 + root_disk 12G
  + docker engine в snapshot + registry-1.docker.io в allowlist.
- **start_phase: DEVELOPMENT.** Каркас сервисов, контракты, репозитории УЖЕ есть —
  loki начинает с разработки (миграция v3 + wiring), а не с discovery.
- **effort: high, max_iterations 25, budget 8 USD.** Объёмная задача
  (миграция v3 + wiring 3 сервисов + тесты), нужен полный SDLC.
- **allowlist-сеть.** z.ai (LLM) + npm + go proxy/sum + GitHub + dl.google.com +
  storage.googleapis.com (для circl/go-git транзитивных зависимостей tg) +
  nodejs.org (для `n install 22`) + registry-1.docker.io (образы) +
  download.docker.com (docker static binary). PyPI закомментирован — нужен только
  если включишь `loki.dashboard: true`.

## Чек-лист перед первым apply

- [ ] conda env `mcd` активирован, `microcode doctor` зелёный.
- [ ] `CLINE_API_KEY` + 3 переменные `ZAI_*` заданы на хосте.
- [ ] git-sync уже настроен inline (`SYNC_REMOTE_URL` + `SYNC_BRANCH=master` в манифесте).
      Если хочешь работать без клонирования remote — положи репо в `microcode/src/`.
- [ ] `microcode build microcode/build.yaml` отработал → snapshot `mcd-base` готов.
- [ ] `microcode validate microcode/platform.yaml` → VALID.

## Типичные проблемы → первая проверка

| симптом | причина | фикс |
|---|---|---|
| `apply` висит, loki 0 задач | cline-shim ждёт GLM API | проверь `CLINE_API_KEY` + provider json; используй `--prd` файл |
| host curl → пустой ответ | сервис на 127.0.0.1 | ребинд `0.0.0.0:PORT` |
| `mount: Not a directory` при boot | bind-mount + from_snapshot | ожидаемо; исп. named volume + seed |
| VM disk full (ENOSPC) | Kafka/docker images заполнили overlay | `docker system prune -af`; `go clean -cache` |
| `go install` lookup storage.googleapis.com | GCS-only модуль (circl/go-git) | уже в allowlist |
| go.sum checksum mismatch | stale module hash | `GONOSUMDB=off go mod tidy` через proxy.golang.org |
| steer ignored | нужен PRD-файл вместо inline | `--prd microcode/PRD.md` |

Полный референс — SKILL `microcode` в репо [sah4ez/microcode](https://github.com/sah4ez/microcode):
`references/manifest.md`, `references/msb-operations.md`, `references/git-sync.md`.
