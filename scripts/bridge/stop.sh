#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PID_PATH="${ROOT_DIR}/.runtime/bridge/bridge.pid"

if [[ ! -f "${PID_PATH}" ]]; then
  pkill -f 'cliproxyapi --config' >/dev/null 2>&1 || true
  echo "bridge pid file was missing; attempted process cleanup"
  exit 0
fi

PID="$(cat "${PID_PATH}")"
if kill -0 "${PID}" 2>/dev/null; then
  kill "${PID}"
  echo "stopped bridge pid ${PID}"
else
  echo "bridge pid ${PID} was not running"
fi

rm -f "${PID_PATH}"
pkill -f 'cliproxyapi --config' >/dev/null 2>&1 || true
