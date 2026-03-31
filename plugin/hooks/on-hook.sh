#!/usr/bin/env bash
# on-hook.sh — bash wrapper for Stop, StopFailure, and Notification hook events
# Invoked by hooks.json for all notification-dispatching events.
# The Go binary disambiguates events via hook_event_name in stdin JSON.
# Always exits 0. Never causes Claude Code to hang or show errors.

BINARY="${CLAUDE_PLUGIN_DATA}/bin/notification"
LOG_DIR="${CLAUDE_PLUGIN_DATA}/logs"
LOG_FILE="${LOG_DIR}/notification.log"
PLUGIN_JSON="${CLAUDE_PLUGIN_ROOT}/.claude-plugin/plugin.json"

# Ensure log directory exists
mkdir -p "${LOG_DIR}"

# Read version from plugin.json (single source of truth)
PLUGIN_VERSION=$(jq -r '.version' "${PLUGIN_JSON}" 2>/dev/null || python3 -c "import json; print(json.load(open('${PLUGIN_JSON}'))['version'])" 2>/dev/null)
if [[ -z "${PLUGIN_VERSION}" ]]; then
  echo "$(date '+%Y-%m-%d %H:%M:%S') [claude-code-notification] could not read version from ${PLUGIN_JSON}" >> "${LOG_FILE}"
  echo '{"systemMessage": "[claude-code-notification] could not read plugin version. See logs at '"${LOG_FILE}"'"}'
  exit 0
fi

# Check binary exists and is executable
if [[ ! -x "${BINARY}" ]]; then
  echo "$(date '+%Y-%m-%d %H:%M:%S') [claude-code-notification] binary not found or not executable at ${BINARY}" >> "${LOG_FILE}"
  echo '{"systemMessage": "[claude-code-notification] binary not found. A new session may fix this, or see logs at '"${LOG_FILE}"'"}'
  exit 0
fi

# Verify version matches plugin version
BINARY_VERSION=$("${BINARY}" --version 2>/dev/null)
if [[ "${BINARY_VERSION}" != "${PLUGIN_VERSION}" ]]; then
  echo "$(date '+%Y-%m-%d %H:%M:%S') [claude-code-notification] version mismatch: binary=${BINARY_VERSION} plugin=${PLUGIN_VERSION}" >> "${LOG_FILE}"
  echo '{"systemMessage": "[claude-code-notification] version mismatch: binary='"${BINARY_VERSION}"' plugin='"${PLUGIN_VERSION}"'. A new session may fix this, or see logs at '"${LOG_FILE}"'"}'
  exit 0
fi

# Pipe stdin to binary; redirect stderr to log file; always exit 0
"${BINARY}" 2>>"${LOG_FILE}" || true

exit 0
