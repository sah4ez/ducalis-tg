#!/usr/bin/env bash
# stop.sh — остановка всех сервисов (graceful SIGTERM, затем SIGKILL).

set -euo pipefail

APP_DIR="${APP_DIR:-/opt/ducalis}"
PIDS="${APP_DIR}/pids"

stop_service() {
  local name="$1"
  local pidfile="${PIDS}/${name}.pid"

  if [[ ! -f "${pidfile}" ]]; then
    echo "  ${name}: не запущен"
    return
  fi

  local pid
  pid=$(cat "${pidfile}")

  if ! kill -0 "${pid}" 2>/dev/null; then
    echo "  ${name}: процесс ${pid} уже мёртв"
    rm -f "${pidfile}"
    return
  fi

  # graceful: SIGTERM, ждём до 5 сек
  kill -TERM "${pid}" 2>/dev/null || true
  for i in $(seq 1 5); do
    kill -0 "${pid}" 2>/dev/null || break
    sleep 1
  done

  # если жив — SIGKILL
  if kill -0 "${pid}" 2>/dev/null; then
    kill -KILL "${pid}" 2>/dev/null || true
    echo "  ${name}: PID ${pid} убит (KILL)"
  else
    echo "  ${name}: PID ${pid} остановлен"
  fi
  rm -f "${pidfile}"
}

echo "── Остановка сервисов ──"
stop_service public
stop_service admin
stop_service internal
echo "✓ Все остановлены."
