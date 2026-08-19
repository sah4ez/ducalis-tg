#!/usr/bin/env bash
# start.sh — запуск всех сервисов из current release.
# Каждый сервис: nohup + PID-файл + лог в ${APP_DIR}/logs/.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="${DEPLOY_DIR:-${SCRIPT_DIR}}"
APP_DIR="${APP_DIR:-/opt/ducalis}"
APP_USER="${APP_USER:-ducalis}"
CURRENT="${APP_DIR}/current"
LOGS="${APP_DIR}/logs"
PIDS="${APP_DIR}/pids"

ENV_FILE="${DEPLOY_DIR}/env"
[[ -f "${ENV_FILE}" ]] || { echo "ERROR: deploy/env не найден" >&2; exit 1; }
set -a; source "${ENV_FILE}"; set +a

if [[ ! -d "${CURRENT}/bin" ]]; then
  echo "ERROR: ${CURRENT}/bin не найден. Сначала bash deploy/deploy.sh" >&2
  exit 1
fi

mkdir -p "${LOGS}" "${PIDS}"
chown -R "${APP_USER}:${APP_USER}" "${LOGS}" "${PIDS}" 2>/dev/null || true

start_service() {
  local name="$1" bind="$2" extra_env="$3"
  local pidfile="${PIDS}/${name}.pid"
  local logfile="${LOGS}/${name}.log"

  if [[ -f "${pidfile}" ]] && kill -0 "$(cat "${pidfile}")" 2>/dev/null; then
    echo "  ${name}: уже запущен (PID $(cat "${pidfile}"))"
    return
  fi

  # rotate лог при >10MB
  if [[ -f "${logfile}" ]] && (( $(stat -c%s "${logfile}" 2>/dev/null || echo 0) > 10*1024*1024 )); then
    mv "${logfile}" "${logfile}.$(date +%Y%m%d)"
  fi

  RUN_AS=""
  if command -v sudo >/dev/null 2>&1 && id -u "${APP_USER}" >/dev/null 2>&1 && [[ "$(id -un)" != "${APP_USER}" ]]; then
    RUN_AS="sudo -u ${APP_USER} -E"
  fi
  ${RUN_AS} env \
    DATABASE_URL="${DATABASE_URL}" \
    BIND="${bind}" \
    JWT_SECRET="${JWT_SECRET}" \
    ADMIN_JWT_SECRET="${ADMIN_JWT_SECRET}" \
    INTERNAL_API_KEY="${INTERNAL_API_KEY}" \
    WEB_DIST="${CURRENT}/web" \
    LOG_LEVEL="${LOG_LEVEL:-info}" \
    ${extra_env} \
    nohup "${CURRENT}/bin/server-${name}" >>"${logfile}" 2>&1 &

  local pid=$!
  echo "${pid}" > "${pidfile}"
  echo "  ${name}: PID ${pid} → ${bind} (лог: ${logfile#${APP_DIR}/})"
}

echo "── Запуск сервисов из release $(basename "$(readlink -f "${CURRENT}")") ──"

start_service public   "${PUBLIC_BIND:-:8080}"   ""
start_service admin    "${ADMIN_BIND:-:8082}"    ""
start_service internal "${INTERNAL_BIND:-:8083}" "KAFKA_BROKERS=${KAFKA_BROKERS:-localhost:9092}"

# ждём health public
sleep 2
for i in $(seq 1 10); do
  if curl -sf "http://localhost${PUBLIC_BIND:-:8080}/health" >/dev/null 2>&1; then
    echo ""
    echo "✓ Все сервисы запущены. UI: http://$(hostname -I 2>/dev/null | awk '{print $1}' || echo 'localhost')${PUBLIC_BIND:-:8080}"
    exit 0
  fi
  sleep 1
done

echo "WARNING: public не отвечает на /health — проверьте логи:" >&2
echo "  tail -20 ${LOGS}/public.log" >&2
exit 1
