#!/usr/bin/env bash
# SessionStart hook. No-op outside the Claude Code cloud VM.
[ "${CLAUDE_CODE_REMOTE:-}" = true ] || exit 0
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

# Make the env visible to every Bash command Claude runs. Source env.sh itself
# rather than copying variables, so new ones are picked up automatically. This
# hook also fires on resume/clear/compact, and CLAUDE_ENV_FILE may be a
# persistent file set in the Claude environment (see README.md), so don't add
# the line twice.
if [ -n "${CLAUDE_ENV_FILE:-}" ]; then
  line=$(printf 'source %q' "$PWD/.claude/cloud/env.sh")
  grep -qxF "$line" "$CLAUDE_ENV_FILE" 2>/dev/null || echo "$line" >> "$CLAUDE_ENV_FILE"
fi

# Stdout from a SessionStart hook is added to the session context by Claude
# Code, so send start.sh's output to stderr.
.claude/cloud/start.sh >&2
