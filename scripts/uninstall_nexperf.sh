#!/bin/sh
set -eu

APP_NAME="nexperf"
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

remove_bashrc_block() {
  [ -f "${BASHRC}" ] || return 0

  tmp="$(mktemp)"
  awk -v start="${PATH_BLOCK_START}" -v end="${PATH_BLOCK_END}" '
    $0 == start { skip = 1; next }
    $0 == end { skip = 0; next }
    skip != 1 { print }
  ' "${BASHRC}" >"${tmp}"
  mv "${tmp}" "${BASHRC}"
}

main() {
  if command -v "${BIN_PATH}" >/dev/null 2>&1; then
    "${BIN_PATH}" stop >/dev/null 2>&1 || true
  elif command -v "${APP_NAME}" >/dev/null 2>&1; then
    "${APP_NAME}" stop >/dev/null 2>&1 || true
  fi

  log "Removing NexPerf binary..."
  rm -f "${BIN_PATH}"

  log "Removing NexPerf PATH entry from ${BASHRC}..."
  remove_bashrc_block

  if [ -f "${CONFIG_FILE}" ]; then
    if grep -q "Managed by NexPerf installer" "${CONFIG_FILE}"; then
      log "Removing NexPerf config..."
      rm -f "${CONFIG_FILE}"
    else
      log "Keeping existing config because it is not marked as installer-managed: ${CONFIG_FILE}"
    fi
  fi

  if [ -d "${CONFIG_DIR}" ] && [ -z "$(find "${CONFIG_DIR}" -mindepth 1 -maxdepth 1 -print -quit)" ]; then
    rmdir "${CONFIG_DIR}"
  fi

  log "Removing NexPerf runtime data..."
  rm -rf "${RUNTIME_DIR}"

  cat <<EOF

NexPerf cleanup complete.

Start a new shell or run:
  source "${BASHRC}"

Removed:
  ${BIN_PATH}
  ${RUNTIME_DIR}

Installer-managed shell entry removed from:
  ${BASHRC}
EOF
}

main "$@"
