#!/usr/bin/env bash
# SessionStart hook. No-op outside the Claude Code cloud VM.
[ "${CLAUDE_CODE_REMOTE:-}" = true ] || exit 0
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
source .claude/cloud/env.sh

# Make the env visible to every Bash command Claude runs. Source env.sh itself
# rather than copying variables, so new ones are picked up automatically.
if [ -n "${CLAUDE_ENV_FILE:-}" ]; then
  printf 'source %q\n' "$PWD/.claude/cloud/env.sh" >> "$CLAUDE_ENV_FILE"
fi

# Stdout from a SessionStart hook is added to the session context by Claude
# Code, so send start.sh's output to stderr.
.claude/cloud/start.sh >&2
