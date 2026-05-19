#!/bin/sh
set -eu

APP_NAME="nexperf"
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

INSTALL_DIR="${NEXPERF_INSTALL_DIR:-${HOME}/.local/bin}"
CONFIG_DIR="${XDG_CONFIG_HOME:-${HOME}/.config}/nexperf"
CONFIG_FILE="${NEXPERF_CONFIG_FILE:-${CONFIG_DIR}/config.toml}"
BASHRC="${NEXPERF_BASHRC:-${HOME}/.bashrc}"
RUNTIME_DIR="${HOME}/.nexperf"
BIN_PATH="${INSTALL_DIR}/${APP_NAME}"
PATH_BLOCK_START="# >>> nexperf path >>>"
PATH_BLOCK_END="# <<< nexperf path <<<"

log() {
  printf '%s\n' "$*"
}

need_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'Missing required command: %s\n' "$1" >&2
    exit 1
  fi
}

write_config() {
  mkdir -p "${CONFIG_DIR}"

  if [ -f "${CONFIG_FILE}" ] && ! grep -q "Managed by NexPerf installer" "${CONFIG_FILE}"; then
    backup="${CONFIG_FILE}.bak.$(date +%Y%m%d%H%M%S)"
    cp "${CONFIG_FILE}" "${backup}"
    log "Existing config backed up to ${backup}"
  fi

  cat >"${CONFIG_FILE}" <<'CONFIG'
# Managed by NexPerf installer.
# Edit through the NexPerf Settings screen or update this file directly.

polling_interval_seconds = 5
retention_hours = 24
dashboard_refresh_ms = 3000
enabled_collectors = ["system", "processes", "storage"]
observation_windows = ["5m", "15m", "1h"]
usage_mode = "balanced"

[storage_limits]
max_scan_depth = 3
max_directory_rows = 80

[insight_thresholds]
cpu_warning = 80.0
memory_warning = 80.0
disk_warning = 80.0
process_memory_mb = 500.0
CONFIG
}

update_bashrc() {
  touch "${BASHRC}"

  tmp="$(mktemp)"
  awk -v start="${PATH_BLOCK_START}" -v end="${PATH_BLOCK_END}" '
    $0 == start { skip = 1; next }
    $0 == end { skip = 0; next }
    skip != 1 { print }
  ' "${BASHRC}" >"${tmp}"

  cat >>"${tmp}" <<EOF

${PATH_BLOCK_START}
export PATH="${INSTALL_DIR}:\${PATH}"
${PATH_BLOCK_END}
EOF

  mv "${tmp}" "${BASHRC}"
}

main() {
  need_command go
  need_command npm

  log "Building NexPerf web dashboard..."
  npm --prefix "${REPO_ROOT}/web" ci
  npm --prefix "${REPO_ROOT}/web" run build

  log "Building NexPerf CLI..."
  mkdir -p "${INSTALL_DIR}" "${RUNTIME_DIR}"
  go build -o "${BIN_PATH}" "${REPO_ROOT}/cmd/nexperf"

  log "Writing NexPerf config..."
  write_config

  log "Adding NexPerf PATH entry to ${BASHRC}..."
  update_bashrc

  cat <<EOF

NexPerf installed successfully.

Binary:
  ${BIN_PATH}

Config:
  ${CONFIG_FILE}

Runtime data:
  ${RUNTIME_DIR}

Start a new shell or run:
  source "${BASHRC}"

Start NexPerf:
  nexperf start

Open the dashboard:
  nexperf open

Stop NexPerf:
  nexperf stop

Dashboard URL:
  http://127.0.0.1:8756/nexperf
EOF
}

main "$@"
