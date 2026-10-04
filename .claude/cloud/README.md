# Claude Code cloud VM setup

The cloud VM has no devcontainer, so the repo provides the equivalent setup here.
The Claude environment's setup script should only run `make cloud_setup`, and
needs no environment variables set in the Claude UI.

- `setup.sh` (`make cloud_setup`): one-time environment setup.
- `start.sh` (`make cloud_start`): per-session startup of the services needed
  for development.
- `session-start.sh`: SessionStart hook (registered in `../settings.json`) that
  exports `env.sh` and runs `start.sh`.

The repo lives at `$ADB_REPO_ROOT` rather than `/workspace`. Go tests need the
Docker daemon (they start MySQL with testcontainers).
