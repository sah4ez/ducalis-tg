---
name: security-checks
description: >-
  Pre-QA security gate for ducalis-tg. Scan changed files for hardcoded
  JWT/api-key secrets, missing auth on endpoints, SQL injection via string
  concat, and weak crypto. Blocks DEVELOPMENT→QA transition on HIGH findings.
---

# Security checks — ducalis-tg

Запускается перед переходом DEVELOPMENT → QA. HIGH-находки блокируют переход.

## Что проверять (по изменившимся файлам)

1. **Hardcoded secrets.** В коде/коммитах НЕ должно быть:
   - `JWT_SECRET=change-me-in-production` (это пример из docker-compose — в коде
     никогда). Реальные секреты — только env (`${JWT_SECRET}`).
   - `ghp_...` (GitHub PAT), пароли, API-keys. Все через env/secrets.
   `grep -rnE '(jwt_secret|api_key|password|token)\s*[:=]\s*["'\'']\w' --include=*.go .`

2. **SQL injection.** Репозитории (`internal/storage/postgres/`) ДОЛЖНЫ
   использовать параметризованные запросы (`$1`, `$2` в pgx), НЕ строковую
   конкатенацию (`fmt.Sprintf("... WHERE id = " + id)`). Запрет:
   `grep -rn 'fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT\|fmt.Sprintf.*WHERE' --include=*_repo.go .`

3. **Auth на эндпоинтах.** Контрактные эндпоинты public (`/api/v1/...`) —
   требуют JWT (кроме register/login). Admin (`/admin/v1/...`) — admin JWT.
   Internal (`/internal/v1/...`) — `INTERNAL_API_KEY` header. Новый эндпоинт
   без auth = HIGH. Сверь middleware в `internal/transport/` + `services.go`.

4. **Weak crypto.** Пароли — `bcrypt.GenerateFromPassword` (DefaultCost),
   НЕ md5/sha1/plain. JWT — HS256 с достаточным секретом (≥32 байт).
   `grep -rn 'md5\|sha1\|math/rand' --include=*.go .` (crypto/rand для токенов).

5. **CORS.** `AllowOrigins: "*"` в main'ах — для прод с кредами небезопасно.
   Для dev OK; для deployment-фазы — вынести в env, restrictive origins.

## Чего НЕ делать

- НЕ коммить `*.env` / файлы с реальными секретами (.gitignore должен покрывать).
- НЕ оставлять `JWT_SECRET=change-me-in-production` как финальное значение.
- НЕ использовать `math/rand` для security-sensitive токенов (только `crypto/rand`).
- НЕ пропускать auth-middleware на новых эндпоинтах без явного обоснования.
