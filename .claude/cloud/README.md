# Claude Code cloud VM setup

The cloud VM has no devcontainer, so the repo provides the equivalent setup here.
The Claude environment's setup script should only run `make cloud_setup`, and
needs no environment variables set in the Claude UI.

- `make cloud_setup` (`setup.sh`): one-time. Installs the `mysql` client, runs
  `make deps`, and installs the `db` and `adb` shell functions.
- `make cloud_start` (`start.sh`): per session. Starts `dockerd` and a
  `mysql:8.4` container on `127.0.0.1:3306`, and runs `make dev_db` on first
  start. Idempotent.
- `session-start.sh`: SessionStart hook (registered in `../settings.json`). Does
  nothing unless `CLAUDE_CODE_REMOTE=true`. Exports the variables in `env.sh` to
  Claude's shell, then runs `make cloud_start`.

The repo lives at `$ADB_REPO_ROOT` rather than `/workspace`. Go tests need the
Docker daemon (they start MySQL with testcontainers).
