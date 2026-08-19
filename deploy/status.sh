#!/usr/bin/env bash
# status.sh — состояние сервисов, версий и health-чеков.

set -euo pipefail

APP_DIR="${APP_DIR:-/opt/ducalis}"
PIDS="${APP_DIR}/pids"
CURRENT="${APP_DIR}/current"
LOGS="${APP_DIR}/logs"

DEPLOY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${DEPLOY_DIR}/env"
if [[ -f "${ENV_FILE}" ]]; then
  set -a; source "${ENV_FILE}"; set +a
fi

BOLD=$(tput bold 2>/dev/null || echo "")
RESET=$(tput sgr0 2>/dev/null || echo "")
GREEN=$(tput setaf 2 2>/dev/null || echo "")
RED=$(tput setaf 1 2>/dev/null || echo "")
YELLOW=$(tput setaf 3 2>/dev/null || echo "")

echo "═════════════════════════════════════════════════════════════"
echo "  Ducalis — Status"
echo "═════════════════════════════════════════════════════════════"

# Release
if [[ -L "${CURRENT}" ]]; then
  REL=$(readlink "${CURRENT}")
  echo "  Release: $(basename "${REL}")"
  if [[ -x "${CURRENT}/bin/server-public" ]]; then
    echo "  Version: (в /health ответе)"
  fi
else
  echo "  Release: ${RED}нет (deploy.sh не запускался)${RESET}"
fi
echo ""

# Сервисы
for svc in public admin internal; do
  pidfile="${PIDS}/${svc}.pid"
  port=":808$([[ "${svc}" == "admin" ]] && echo 2 || ([[ "${svc}" == "internal" ]] && echo 3 || echo 0))"

  if [[ -f "${pidfile}" ]] && kill -0 "$(cat "${pidfile}")" 2>/dev/null; then
    pid=$(cat "${pidfile}")
    mem=$(ps -o rss= -p "${pid}" 2>/dev/null | awk '{printf "%.0fMB", $1/1024}')
    cpu=$(ps -o %cpu= -p "${pid}" 2>/dev/null | tr -d ' ')

    # health check
    health="—"
    if curl -sf -m 2 "http://localhost${port}/health" 2>/dev/null | grep -q '"ok"'; then
      health="${GREEN}healthy${RESET}"
    else
      health="${RED}no response${RESET}"
    fi

    # последние ошибки
    recent_errors=$(tail -50 "${LOGS}/${svc}.log" 2>/dev/null | grep -c '"level":"error"' 2>/dev/null) || recent_errors=0

    echo "  ${BOLD}${svc}${RESET}  PID ${pid}  ${port}  ${health}  mem=${mem} cpu=${cpu}%  err(50)=${recent_errors}"
  else
    echo "  ${BOLD}${svc}${RESET}  ${RED}остановлен${RESET}"
  fi
done

# БД
echo ""
if command -v pg_isready >/dev/null && pg_isready -q 2>/dev/null; then
  echo "  PostgreSQL: ${GREEN}доступен${RESET}"
else
  echo "  PostgreSQL: ${RED}недоступен${RESET}"
fi

# Место на диске
echo "  Диск: $(df -h "${APP_DIR}" 2>/dev/null | tail -1 | awk '{print $4}' || echo "?") свободно"
echo "  Релизов: $(ls "${APP_DIR}/releases" 2>/dev/null | wc -l)"
echo "═════════════════════════════════════════════════════════════"
