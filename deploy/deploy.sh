#!/usr/bin/env bash
# deploy.sh — сборка и выкладка релиза на VM.
#
# Что делает:
#   1. Собирает фронтенд (npm ci + build) → web/dist
#   2. Собирает 3 бинарника (go build)
#   3. Кладёт всё в timestamped release: /opt/ducalis/releases/<ts>/
#   4. Обновляет symlink current → release
#   5. Перезапускает сервисы (stop + start)
#
# Использование (из корня репо на VM):
#   bash deploy/deploy.sh
#
# Откат на предыдущий релиз:
#   bash deploy/rollback.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
DEPLOY_DIR="${DEPLOY_DIR:-${SCRIPT_DIR}}"
APP_DIR="${APP_DIR:-/opt/ducalis}"
if [[ ! "${APP_DIR}" =~ ^/[a-zA-Z0-9/_.-]+$ ]]; then
  echo "ERROR: APP_DIR='${APP_DIR}' — подозрительное значение." >&2
  echo "  Проверьте: env | grep APP_DIR" >&2
  exit 1
fi
APP_USER="${APP_USER:-ducalis}"
RELEASES_DIR="${APP_DIR}/releases"
TS=$(date +%Y%m%d-%H%M%S)
RELEASE_DIR="${RELEASES_DIR}/${TS}"

# ── env ────────────────────────────────────────────────────────────
ENV_FILE="${DEPLOY_DIR}/env"
if [[ ! -f "${ENV_FILE}" ]]; then
  echo "ERROR: ${ENV_FILE} не найден." >&2
  echo "  cp deploy/env.example deploy/env — и заполните секреты." >&2
  exit 1
fi
set -a; source "${ENV_FILE}"; set +a

# Проверка секретов
if [[ "${JWT_SECRET}" == "change-me"* ]]; then
  echo "ERROR: JWT_SECRET не сменён. Сгенерируйте: openssl rand -hex 32" >&2
  exit 1
fi

echo "── [1/4] Фронтенд ──"
cd "${REPO_DIR}/web"
npm ci --silent 2>/dev/null || npm install --silent
npm run build
if [[ ! -f dist/index.html ]]; then
  echo "ERROR: web/dist/index.html не создан" >&2
  exit 1
fi

echo "── [2/4] Backend ──"
cd "${REPO_DIR}"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS="-s -w -X main.Version=${VERSION}"
mkdir -p /tmp/ducalis-build
for svc in public admin internal; do
  CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" \
    -o "/tmp/ducalis-build/server-${svc}" "./cmd/server-${svc}"
done
echo "  ✓ v${VERSION}: $(ls /tmp/ducalis-build/ | tr '\n' ' ')"

echo "── [3/4] Release ${TS} ──"
mkdir -p "${RELEASE_DIR}/bin" "${RELEASE_DIR}/web"
mv /tmp/ducalis-build/server-* "${RELEASE_DIR}/bin/"
rm -rf /tmp/ducalis-build
cp -r web/dist/* "${RELEASE_DIR}/web/"
chmod +x "${RELEASE_DIR}/bin/"server-*
chown -R "${APP_USER}:${APP_USER}" "${RELEASE_DIR}" 2>/dev/null || true

# атомарное переключение symlink
ln -sfn "${RELEASE_DIR}" "${APP_DIR}/current"
echo "  ✓ current → ${TS}"

echo "── [4/4] Перезапуск ──"
bash "${SCRIPT_DIR}/stop.sh" 2>/dev/null || true
bash "${SCRIPT_DIR}/start.sh"

echo ""
echo "✓ Deploy ${TS} (v${VERSION}) запущен."
echo "  UI:       http://$(hostname -I 2>/dev/null | awk '{print $1}' || echo localhost):8080"
echo "  Health:   curl localhost:8080/health"
echo "  Логи:     tail -f ${APP_DIR}/logs/*.log"
