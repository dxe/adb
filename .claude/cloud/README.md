# Claude Code cloud VM setup

The cloud VM has no devcontainer, so the repo provides the equivalent setup here.
The Claude environment's setup script should only run `.claude/cloud/setup.py`.

Set these in the Claude environment's env vars (`setup.py` fails if they're missing):

- `CLAUDE_ENV_FILE`: an absolute path, e.g. `/root/.claude-env.sh`. `setup.py`
  adds mise's shims to `PATH` in that file (unless another repo's setup script
  already did), and Claude Code sources it before every Bash command.
- `MISE_DISABLE_TOOLS=go,python`: the VM ships its own go and python. This keeps
  mise from installing them or putting a python shim on `PATH`, which would shadow
  the system python everywhere, not just in this repo.

- `setup.py`: one-time environment setup, including installing project
  dependencies to warm the caches baked into the environment snapshot.
- `start.sh`: per-session startup: installs project dependencies and starts the services needed for development.
- `session-start.sh`: SessionStart hook (registered in `../settings.json`) that
  exports `env.sh` and runs `start.sh`.
