# Claude Code cloud VM setup

The cloud VM has no devcontainer, so the repo provides the equivalent setup here.
The Claude environment's setup script should only run `.claude/cloud/setup.sh`, and
needs no environment variables set in the Claude UI.

- `setup.sh`: one-time environment setup.
- `start.sh`: per-session startup of the services needed
  for development.
- `session-start.sh`: SessionStart hook (registered in `../settings.json`) that
  exports `env.sh` and runs `start.sh`.
