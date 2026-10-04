#!/usr/bin/env bash
# SessionStart hook. No-op outside the Claude Code cloud VM so the devcontainer
# (which has its own DB and env) is unaffected.
[ "${CLAUDE_CODE_REMOTE:-}" = true ] || exit 0
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
source .claude/cloud/env.sh

# Make the env visible to every Bash command Claude runs.
if [ -n "${CLAUDE_ENV_FILE:-}" ]; then
  for v in ADB_REPO_ROOT NVM_DIR DB_HOST DB_PROTOCOL DB_USER DB_PASSWORD DB_NAME; do
    printf 'export %s=%q\n' "$v" "${!v}" >> "$CLAUDE_ENV_FILE"
  done
fi

make cloud_start >&2
