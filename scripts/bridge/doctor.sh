#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"

if [[ -f "${ENV_FILE}" ]]; then
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
fi

PORT="${BRIDGE_PORT:-43110}"
TOKEN="${LLM_BRIDGE_TOKEN:-change-me-bridge-token}"

run_with_timeout() {
  python3 - "$@" <<'PY'
import subprocess
import sys

timeout_seconds = int(sys.argv[1])
command = sys.argv[2:]
try:
    completed = subprocess.run(command, capture_output=True, text=True, timeout=timeout_seconds)
    sys.stdout.write(completed.stdout)
    sys.stderr.write(completed.stderr)
    sys.exit(completed.returncode)
except subprocess.TimeoutExpired:
    sys.exit(124)
PY
}

echo "== binaries =="
for bin in codex gemini claude; do
  if command -v "${bin}" >/dev/null 2>&1; then
    echo "${bin}: $(command -v "${bin}")"
  else
    echo "${bin}: missing"
  fi
done

echo
echo "== bridge =="
if curl -sf -H "Authorization: Bearer ${TOKEN}" "http://127.0.0.1:${PORT}/v1/models" >/dev/null; then
  echo "models endpoint: ok"
else
  echo "models endpoint: unavailable"
fi

echo
echo "== models =="
curl -s -H "Authorization: Bearer ${TOKEN}" "http://127.0.0.1:${PORT}/v1/models" || true

echo
echo "== docker -> bridge =="
API_CONTAINER_ID="$(run_with_timeout 5 docker compose ps -q api 2>/dev/null || true)"
if [[ -z "${API_CONTAINER_ID}" ]]; then
  echo "api container not running yet"
elif run_with_timeout 8 docker compose exec -T api wget --header="Authorization: Bearer ${TOKEN}" -qO- "http://host.docker.internal:${PORT}/v1/models" >/dev/null 2>&1; then
  echo "api container can reach bridge models endpoint"
else
  echo "api container cannot reach bridge yet"
fi
