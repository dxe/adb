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
# Code, so send the scripts' output to stderr.

# If the session's environment didn't run setup.py as its setup script (e.g. the
# default environment), run it now.
# Keep the marker path in sync with setup.py.
if [ ! -e "$HOME/.local/state/adb/cloud-setup.done" ]; then
  echo "setup.py hasn't run in this environment; running it now." >&2
  .claude/cloud/setup.py >&2
  # setup.py put mise and its shims on PATH via $CLAUDE_ENV_FILE, which only
  # reaches Claude's later Bash commands, not this hook. start.sh needs them now.
  export PATH="$HOME/.local/bin:$HOME/.local/share/mise/shims:$PATH"
fi

.claude/cloud/start.sh >&2
