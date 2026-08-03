---
name: infra-in-vm
description: >-
  Run the ducalis-tg docker-compose stack (postgres, redis, kafka, zookeeper,
  kafka-ui, 3 servers) INSIDE the microsandbox VM. Docker Engine + compose
  plugin are in the snapshot. Use when starting the app with real dependencies
  for testing, or when the app needs DATABASE_URL/REDIS_URL/KAFKA_BROKERS.
---

# Infrastructure inside the VM

Docker Engine + compose-plugin установлены в snapshot `mcd-base` (static binary
`dockerd` без systemd + compose v2.30.3). `loki` пользователь — в группе `docker`,
может запускать `docker`/`docker-compose` напрямую.

## Запуск стека

```bash
cd /workspace                  # репо ducalis-tg (примонтирован/склонирован)
docker-compose up -d            # postgres, redis, zookeeper, kafka, kafka-ui
docker-compose ps               # проверить что все healthy
docker-compose logs -f postgres # если что-то не поднялось
```

Образы тянутся из `registry-1.docker.io` (в allowlist: registry-1, auth.docker.io,
production.cloudflare.docker.com). **Первый pull медленный** (postgres:16,
redis:7, cp-kafka:7.5.0, kafka-ui) — Kafka образ ~1.5GB. Поэтому root_disk: 12G
и memory: 8192 в манифесте.

## Переменные окружения для сервисов

docker-compose.yaml уже задаёт их для server-public/admin/internal:

```
DATABASE_URL=postgres://ducalis:ducalis123@postgres:5432/ducalis?sslmode=disable
REDIS_URL=redis://redis:6379/0
KAFKA_BROKERS=kafka:9092
BIND=:8080   (или :8082/:8083)
JWT_SECRET=change-me-in-production
```

При запуске сервисов **вручную** (`make run-public` вне docker) — `postgres`/`redis`
в URL заменяй на `localhost` (compose пробрасывает порты на host VM):

```bash
export DATABASE_URL=postgres://ducalis:ducalis123@localhost:5432/ducalis?sslmode=disable
export REDIS_URL=redis://localhost:6379/0
export KAFKA_BROKERS=localhost:9093   # PLAINTEXT_HOST listener в compose = 9093
make run-public
```

## Port-forward (host → VM)

Манифест форвардит: 8080/8082/8083 (сервисы), 9090 (metrics), 8081 (kafka-ui),
9418 (git-daemon). **ВАЖНО:** сервис внутри VM должен биндить `0.0.0.0`
(или `:8080` = `0.0.0.0:8080` в Go fiber). Биндинг на `127.0.0.1` даст
«порт открыт, но пустой ответ» (curl exit 52) — msb port-forward идёт на eth0 VM.

docker-compose `BIND=:8080` уже корректен (= 0.0.0.0). Не меняй на 127.0.0.1.

## Доступ с хоста

```bash
curl http://localhost:8080/health          # public API
curl http://localhost:8082/health          # admin API
curl http://localhost:8083/health          # internal API
open http://localhost:8081                 # kafka-ui (в браузере)
git fetch git://localhost:9418/ vm/ducalis-build   # git-daemon (см. git-sync.md)
```

## Ресурсы и типичные проблемы

- **disk full (ENOSPC):** Go-сборка + docker images заполняют overlay. Чисти:
  `docker system prune -af`, `go clean -cache` (`$(go env GOCACHE)` в tmpfs).
- **Kafka не поднимается:** memory. cp-kafka хочет ≥1G. Если OOMKill —
  подними `memory` в манифесте (сейчас 8192) и пересобери snapshot.
- **postgres не готов к коннекту:** healthcheck в compose есть, но при ручном
  запуске делай `sleep 3` или poll `pg_isready`.
- **`pgrep`/`ps` недоступны в VM:** скань `/proc/[0-9]*/cmdline`.

## Чего НЕ делать

- НЕ бинди сервисы на 127.0.0.1 — только 0.0.0.0 / `:port`.
- НЕ поднимай отдельные postgres/redis/kafka — используй docker-compose проекта.
- НЕ хорди registry-1.docker.io убирай из allowlist — без него pull образов упадёт.
