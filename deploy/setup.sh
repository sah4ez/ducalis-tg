#!/usr/bin/env bash
# setup.sh — однократная подготовка выделенной VM (Ubuntu/Debian).
#
# Устанавливает: Go 1.26, Node 20, PostgreSQL 16, psql-клиент.
# Создаёт БД ducalis + пользователя, применяет миграции.
#
# Использование (от root или с sudo):
#   sudo bash deploy/setup.sh
#
# После setup.sh — deploy/deploy.sh для сборки и запуска.

set -euo pipefail

# ── Конфигурация (переопределяется env-переменными) ─────────────────
DB_NAME="${DB_NAME:-ducalis}"
DB_USER="${DB_USER:-ducalis}"
DB_PASS="${DB_PASS:-ducalis123}"
GO_VERSION="${GO_VERSION:-1.26.5}"
NODE_MAJOR="${NODE_MAJOR:-20}"
APP_DIR="${APP_DIR:-/opt/ducalis}"
APP_USER="${APP_USER:-ducalis}"

echo "═════════════════════════════════════════════════════════════"
echo "  Ducalis VM Setup"
echo "  DB: ${DB_USER}@localhost/${DB_NAME}"
echo "  App: ${APP_DIR} (user: ${APP_USER})"
echo "═════════════════════════════════════════════════════════════"

if [[ $EUID -ne 0 ]]; then
  echo "ERROR: запустите от root: sudo bash deploy/setup.sh" >&2
  exit 1
fi

# ── 1. Системные пакеты ────────────────────────────────────────────
echo "── apt-пакеты ──"
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq curl git ca-certificates postgresql postgresql-contrib \
  build-essential jq unzip >/dev/null

# ── 2. Go (из go.dev — apt даёт старую версию) ─────────────────────
if ! command -v go >/dev/null || ! go version | grep -qE "go1\.(2[4-9]|[3-9])"; then
  echo "── Go ${GO_VERSION} ──"
  ARCH=$(dpkg --print-architecture)
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" -o /tmp/go.tgz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tgz
  rm -f /tmp/go.tgz
  ln -sf /usr/local/go/bin/go /usr/local/bin/go
  ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
fi
echo "  Go: $(go version)"

# ── 3. Node.js 20 (для сборки фронтенда) ───────────────────────────
if ! command -v node >/dev/null || [[ "$(node -v | cut -d. -f1 | tr -d v)" -lt ${NODE_MAJOR} ]]; then
  echo "── Node.js ${NODE_MAJOR} ──"
  curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR}.x" | bash - >/dev/null
  apt-get install -y -qq nodejs >/dev/null
fi
echo "  Node: $(node -v), npm: $(npm -v)"

# ── 4. PostgreSQL ──────────────────────────────────────────────────
echo "── PostgreSQL ──"
systemctl enable --now postgresql 2>/dev/null || true

# ждём готовности
for i in $(seq 1 15); do
  if sudo -u postgres pg_isready -q 2>/dev/null; then break; fi
  sleep 1
done

# пользователь + БД (idempotent)
sudo -u postgres psql -tc "SELECT 1 FROM pg_roles WHERE rolname='${DB_USER}'" | grep -q 1 || \
  sudo -u postgres psql -c "CREATE ROLE ${DB_USER} LOGIN PASSWORD '${DB_PASS}' SUPERUSER"
sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" | grep -q 1 || \
  sudo -u postgres psql -c "CREATE DATABASE ${DB_NAME} OWNER ${DB_USER}"
echo "  PostgreSQL: $(sudo -u postgres psql -tAc 'SELECT version()' | head -c 40)..."

# ── 5. Пользователь и директории приложения ────────────────────────
echo "── Пользователь ${APP_USER} ──"
id -u "${APP_USER}" >/dev/null 2>&1 || useradd --system --create-home --shell /bin/bash "${APP_USER}"

mkdir -p "${APP_DIR}"/{bin,web,logs, releases}
chown -R "${APP_USER}:${APP_USER}" "${APP_DIR}"

# ── 6. Миграции (idempotent — CREATE IF NOT EXISTS) ────────────────
echo "── Миграции ──"
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PGPASSWORD="${DB_PASS}" psql -h localhost -U "${DB_USER}" -d "${DB_NAME}" \
  -f "${REPO_DIR}/migrations/init.sql" 2>&1 | grep -cE "ERROR" | {
  read count
  if [[ "$count" -gt 0 ]]; then
    echo "  WARNING: ${count} ошибок миграции (чаще всего 'already exists' — это нормально при повторном запуске)"
  else
    echo "  ✓ Миграции применены"
  fi
}

TABLES=$(PGPASSWORD="${DB_PASS}" psql -h localhost -U "${DB_USER}" -d "${DB_NAME}" \
  -tAc "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'")
echo "  Таблиц в БД: ${TABLES}"

echo ""
echo "═════════════════════════════════════════════════════════════"
echo "  ✓ Setup завершён"
echo ""
echo "  Далее:"
echo "    1. cp deploy/env.example deploy/env  # и заполнить секреты"
echo "    2. bash deploy/deploy.sh             # сборка + запуск"
echo "═════════════════════════════════════════════════════════════"
