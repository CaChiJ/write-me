#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"
RUNTIME_DIR="${ROOT_DIR}/.runtime/bridge"
INSTALL_DIR="${RUNTIME_DIR}/cliproxyapi"
mkdir -p "${INSTALL_DIR}"

if [[ -f "${ENV_FILE}" ]]; then
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
fi

VERSION="${CLIPROXY_VERSION:-v6.8.51}"
API_URL="https://api.github.com/repos/router-for-me/CLIProxyAPI/releases/tags/${VERSION}"

python3 - "${API_URL}" "${INSTALL_DIR}" <<'PY'
import json
import os
import platform
import sys
import urllib.request

api_url, install_dir = sys.argv[1], sys.argv[2]
machine = platform.machine().lower()
system = platform.system().lower()

arch_map = {
    "arm64": ["arm64", "aarch64"],
    "amd64": ["amd64", "x86_64"],
}
os_map = {
    "darwin": "darwin",
    "linux": "linux",
}

if system not in os_map:
    raise SystemExit(f"unsupported os: {system}")

valid_arches = arch_map.get(machine, [machine])
with urllib.request.urlopen(api_url) as resp:
    release = json.load(resp)

assets = release.get("assets", [])
matched = None
for asset in assets:
    name = asset["name"].lower()
    if os_map[system] not in name:
        continue
    if not any(arch in name for arch in valid_arches):
        continue
    matched = asset
    break

if not matched:
    raise SystemExit("could not find a matching release asset for this platform")

url = matched["browser_download_url"]
target = os.path.join(install_dir, matched["name"])
print(target)
urllib.request.urlretrieve(url, target)
PY

ASSET_PATH="$(ls -t "${INSTALL_DIR}" | head -n 1)"
ARCHIVE_PATH="${INSTALL_DIR}/${ASSET_PATH}"
BIN_PATH="${INSTALL_DIR}/cliproxyapi"

case "${ARCHIVE_PATH}" in
  *.tar.gz|*.tgz)
    tar -xzf "${ARCHIVE_PATH}" -C "${INSTALL_DIR}"
    ;;
  *.zip)
    unzip -o "${ARCHIVE_PATH}" -d "${INSTALL_DIR}" >/dev/null
    ;;
  *)
    chmod +x "${ARCHIVE_PATH}"
    cp "${ARCHIVE_PATH}" "${BIN_PATH}"
    ;;
esac

if [[ ! -f "${BIN_PATH}" ]]; then
  CANDIDATE="$(find "${INSTALL_DIR}" -maxdepth 2 -type f | grep -E '/(CLIProxyAPI|cliproxyapi|cli-proxy-api)$' | head -n 1 || true)"
  if [[ -n "${CANDIDATE}" ]]; then
    cp "${CANDIDATE}" "${BIN_PATH}"
  fi
fi

chmod +x "${BIN_PATH}"
echo "installed CLIProxyAPI to ${BIN_PATH}"
