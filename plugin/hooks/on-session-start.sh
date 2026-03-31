#!/usr/bin/env bash
# on-session-start.sh -- binary bootstrap for claude-code-notification
# Downloads the correct Go binary from GitHub Releases if missing or version-mismatched.
# MUST always exit 0 -- never block a Claude session.

BINARY="${CLAUDE_PLUGIN_DATA}/bin/notification"
BIN_DIR="${CLAUDE_PLUGIN_DATA}/bin"
LOG_DIR="${CLAUDE_PLUGIN_DATA}/logs"
LOG_FILE="${LOG_DIR}/notification.log"
PLUGIN_JSON="${CLAUDE_PLUGIN_ROOT}/.claude-plugin/plugin.json"
REPO="berk-karaal/claude-code-notification"

log_warning() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') [claude-code-notification] $1" >> "${LOG_FILE}"
}

mkdir -p "${BIN_DIR}" "${LOG_DIR}"

# Read version from plugin.json (single source of truth)
PLUGIN_VERSION=$(jq -r '.version' "${PLUGIN_JSON}" 2>/dev/null || python3 -c "import json; print(json.load(open('${PLUGIN_JSON}'))['version'])" 2>/dev/null)
if [[ -z "${PLUGIN_VERSION}" ]]; then
  log_warning "could not read version from ${PLUGIN_JSON}"
  exit 0
fi

# --- Detect OS and architecture ---
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    log_warning "unsupported architecture: ${ARCH}"
    exit 0
    ;;
esac

# --- Check if binary is current ---
if [[ -x "${BINARY}" ]]; then
  CURRENT_VERSION=$("${BINARY}" --version 2>/dev/null)
  if [[ "${CURRENT_VERSION}" == "${PLUGIN_VERSION}" ]]; then
    exit 0
  fi
  log_warning "version mismatch: binary=${CURRENT_VERSION} plugin=${PLUGIN_VERSION}, updating..."
fi

# --- Download and install ---
ARCHIVE_NAME="notification_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${PLUGIN_VERSION}/${ARCHIVE_NAME}"
TMPDIR="$(mktemp -d)"

echo "$(date '+%Y-%m-%d %H:%M:%S') [claude-code-notification] downloading ${DOWNLOAD_URL}" >> "${LOG_FILE}"

if curl -fsSL --max-time 30 "${DOWNLOAD_URL}" -o "${TMPDIR}/${ARCHIVE_NAME}" 2>>"${LOG_FILE}"; then
  tar -xzf "${TMPDIR}/${ARCHIVE_NAME}" -C "${TMPDIR}" 2>>"${LOG_FILE}"
  mv "${TMPDIR}/notification" "${BINARY}"
  chmod +x "${BINARY}"
  echo "$(date '+%Y-%m-%d %H:%M:%S') [claude-code-notification] installed ${PLUGIN_VERSION} (${OS}/${ARCH})" >> "${LOG_FILE}"
else
  log_warning "auto-download failed for ${DOWNLOAD_URL}"
  log_warning "manual download: ${DOWNLOAD_URL}"
fi

rm -rf "${TMPDIR}"

exit 0
