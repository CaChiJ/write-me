#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"
BIN_PATH="${ROOT_DIR}/.runtime/bridge/cliproxyapi/cliproxyapi"
CONFIG_PATH="${ROOT_DIR}/.runtime/bridge/config.yaml"

if [[ -f "${ENV_FILE}" ]]; then
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
fi

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <codex|gemini|claude>" >&2
  exit 1
fi

if [[ ! -x "${BIN_PATH}" ]]; then
  echo "bridge binary not found: ${BIN_PATH}" >&2
  echo "run ./scripts/bridge/install.sh first" >&2
  exit 1
fi

if [[ ! -f "${CONFIG_PATH}" ]]; then
  echo "bridge config not found: ${CONFIG_PATH}" >&2
  echo "run ./scripts/bridge/start.sh once to generate the config" >&2
  exit 1
fi

case "$1" in
  codex)
    exec "${BIN_PATH}" --config "${CONFIG_PATH}" --codex-login
    ;;
  gemini)
    exec "${BIN_PATH}" --config "${CONFIG_PATH}" --login
    ;;
  claude)
    exec "${BIN_PATH}" --config "${CONFIG_PATH}" --claude-login
    ;;
  *)
    echo "unsupported provider: $1" >&2
    exit 1
    ;;
esac
