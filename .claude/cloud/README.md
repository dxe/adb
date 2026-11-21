# Claude Code cloud VM setup

The cloud VM has no devcontainer, so the repo provides the equivalent setup here.
The Claude environment's setup script should only run `.claude/cloud/setup.py`, and
needs no environment variables set in the Claude UI.

- `setup.py`: one-time environment setup, including installing project
  dependencies to warm the caches baked into the environment snapshot.
- `start.sh`: per-session startup: installs project dependencies and starts the services needed for development.
- `session-start.sh`: SessionStart hook (registered in `../settings.json`) that
  exports `env.sh` and runs `start.sh`.
