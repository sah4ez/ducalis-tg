---
name: tdd-rules
description: >-
  Test-driven development for ducalis-tg (Go 1.24, pgx, gofiber). RED-GREEN-
  REFACTOR on `go test`. Table-driven tests, real Postgres via docker-compose
  (no DB mocks), persistence-reopen checks. Use before writing any feature/fix.
---

# TDD for ducalis-tg

RED-GREEN-REFACTOR. Тест пишется ПЕРВЫМ, падает (RED), реализация делает его
зелёным (GREEN), потом рефакторинг. `make test` = `go test -v -race ./...`.

## Текущее состояние тестов

В репо **НЕТ** `*_test.go` файлов — `make test`/`make test-cover` это stub'ы.
Первая задача при wiring (service-wiring.md) — покрыть smoke-тестами хотя бы
запущенный сервис + ключевые репозитории.

## Паттерны

### Table-driven
```go
func TestTaskRepository_Create(t *testing.T) {
    cases := []struct{
        name    string
        task    types.Task
        wantErr error
    }{ ... }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) { ... })
    }
}
```

### Реальная Postgres (НЕ мок БД)
Тестируй репозитории против живой Postgres. Два варианта:
1. **docker-compose** — подними `docker-compose up -d postgres`, коннектись к
   `localhost:5432` (см. infra-in-vm.md). Подходит для локальных/VM-тестов.
2. **testcontainers-go** (`github.com/testcontainers/testcontainers-go`) — если
   хочешь изолированную БД на каждый тест-ран. Добавь зависимость в go.mod.

НЕ мокай `*pgxpool.Pool` — это скрывает SQL-баги. Лучше медленный тест с реальной
БД, чем быстрый мок, пропускающий регрессии.

### Persistence-reopen check
После Create → Close pool → New pool → Get: данные должны сохраниться.
Подтверждает что commit реально записан, а не висит в транзакции.

### Сервисы
`AuthService`, `TaskService`, `WorkspaceService` — тестируй с реальными или
in-memory фейк-репозиториями (реализующими service-интерфейс). Фейк репозитория
ОК для бизнес-логики; для SQL-слоя — реальная БД.

## Чего НЕ делать

- НЕ коммить код без зелёного `make test` (или явного обоснования, почему тест
  откладывается — например, infra-тест требует VM).
- НЕ мокай `pgxpool` — используй реальную Postgres или testcontainers.
- НЕ пиши один гигантский тест — table-driven с именованными кейсами.
- НЕ забывай `-race` (он в `make test` по умолчанию) — pgx/concurrency-sensitive.
