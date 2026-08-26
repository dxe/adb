#!/usr/bin/env bash
set -euo pipefail

# Playwright MCP (.mcp.json) needs its browser binary and system libs, neither
# of which npx pulls in automatically.
pnpx @playwright/mcp@latest install-browser chrome-for-testing
sudo env "PATH=$PATH" pnpx playwright install-deps chromium
