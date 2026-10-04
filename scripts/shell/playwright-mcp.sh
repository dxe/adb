#!/usr/bin/env bash
# Launches the Playwright MCP. In the Claude Code cloud VM, use the image's
# preinstalled Chromium (/opt/pw-browsers/chromium is a version-independent
# symlink) instead of the build @latest wants, which the VM can't download.
set -euo pipefail
if [ "${CLAUDE_CODE_REMOTE:-}" = true ] && [ -x /opt/pw-browsers/chromium ]; then
  set -- --executable-path /opt/pw-browsers/chromium "$@"
fi
exec pnpx @playwright/mcp@latest "$@"
