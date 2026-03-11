#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"
RUNTIME_DIR="${ROOT_DIR}/.runtime/bridge"
INSTALL_DIR="${RUNTIME_DIR}/cliproxyapi"
CONFIG_PATH="${RUNTIME_DIR}/config.yaml"
LOG_PATH="${RUNTIME_DIR}/bridge.log"
PID_PATH="${RUNTIME_DIR}/bridge.pid"
AUTH_DIR="${RUNTIME_DIR}/auth"
BIN_PATH="${INSTALL_DIR}/cliproxyapi"

mkdir -p "${RUNTIME_DIR}" "${AUTH_DIR}"

if [[ -f "${ENV_FILE}" ]]; then
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
fi

if [[ ! -x "${BIN_PATH}" ]]; then
  echo "bridge binary not found: ${BIN_PATH}" >&2
  echo "run ./scripts/bridge/install.sh first" >&2
  exit 1
fi

if [[ -f "${PID_PATH}" ]] && kill -0 "$(cat "${PID_PATH}")" 2>/dev/null; then
  echo "bridge already running (pid $(cat "${PID_PATH}"))"
  exit 0
fi

PORT="${BRIDGE_PORT:-43110}"
TOKEN="${LLM_BRIDGE_TOKEN:-change-me-bridge-token}"

cat > "${CONFIG_PATH}" <<EOF
host: 127.0.0.1
port: ${PORT}
api-keys:
  - ${TOKEN}
auth-dir: ${AUTH_DIR}
oauth-model-alias:
  gemini-cli:
    - name: gemini-2.5-pro
      alias: write-me-gemini
  codex:
    - name: gpt-5
      alias: write-me-codex
  claude:
    - name: claude-sonnet-4-5-20250929
      alias: write-me-claude
EOF

nohup "${BIN_PATH}" --config "${CONFIG_PATH}" < /dev/null > "${LOG_PATH}" 2>&1 &
echo $! > "${PID_PATH}"

for _ in $(seq 1 20); do
  if ! kill -0 "$(cat "${PID_PATH}")" 2>/dev/null; then
    echo "bridge exited unexpectedly; recent log:" >&2
    tail -n 40 "${LOG_PATH}" >&2 || true
    exit 1
  fi
  if curl -fsS -H "Authorization: Bearer ${TOKEN}" "http://127.0.0.1:${PORT}/v1/models" >/dev/null 2>&1; then
    echo "bridge started on 127.0.0.1:${PORT} (pid $(cat "${PID_PATH}"))"
    exit 0
  fi
  sleep 0.5
done

echo "bridge process is running but models endpoint is not ready yet; recent log:" >&2
tail -n 40 "${LOG_PATH}" >&2 || true
exit 1
