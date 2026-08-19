#!/usr/bin/env bash
# rollback.sh — откат на предыдущий релиз (или указанный).

set -euo pipefail

APP_DIR="${APP_DIR:-/opt/ducalis}"
CURRENT="${APP_DIR}/current"
RELEASES_DIR="${APP_DIR}/releases"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="${DEPLOY_DIR:-${SCRIPT_DIR}}"

# целевой релиз: аргумент или предпоследний
if [[ -n "${1:-}" ]]; then
  TARGET="${1}"
else
  # сортируем по имени (timestamp), берём предпоследний
  TARGET=$(ls -1 "${RELEASES_DIR}" 2>/dev/null | sort -r | sed -n 2p)
fi

if [[ -z "${TARGET}" ]] || [[ ! -d "${RELEASES_DIR}/${TARGET}" ]]; then
  echo "ERROR: релиз '${TARGET}' не найден." >&2
  echo "Доступные:" >&2
  ls -1 "${RELEASES_DIR}" 2>/dev/null | sort -r | head -10 >&2
  exit 1
fi

echo "── Откат: $(basename "$(readlink "${CURRENT}" 2>/dev/null || echo '?')") → ${TARGET} ──"

bash "${SCRIPT_DIR}/stop.sh"
ln -sfn "${RELEASES_DIR}/${TARGET}" "${CURRENT}"
bash "${SCRIPT_DIR}/start.sh"

echo "✓ Откат на ${TARGET} завершён."
