# Claude Code cloud VM setup

The cloud VM has no devcontainer, so the repo provides the equivalent setup here.
`setup.py` runs in one of two ways:

- From the SessionStart hook, on the first session start in an environment that
  didn't already run it (e.g. the default one). That first start is slower.
- From the Claude environment's setup script, which should only run
  `.claude/cloud/setup.py`. This is required when sessions have several repos,
  because hooks don't fire then. It also bakes the setup into the environment
  snapshot, so session starts stay fast.

When `setup.py` is the environment's setup script, set `CLAUDE_ENV_FILE` in the
Claude environment's env vars to an absolute path, e.g. `/root/.claude-env.sh`
(`setup.py` fails if it's missing). `setup.py` adds mise's shims to `PATH` in
that file (unless another repo's setup script already did), and Claude Code
sources it before every Bash command. The SessionStart hook gets its own
`CLAUDE_ENV_FILE` from Claude Code.

`setup.py` also sets `disable_tools = ["go", "python"]` in mise's global config:
the VM ships its own go and python, and this keeps mise from installing them or
putting a python shim on `PATH`, which would shadow the system python everywhere,
not just in this repo.

- `setup.py`: one-time environment setup, including installing project
  dependencies to warm the caches baked into the environment snapshot.
- `start.sh`: per-session startup: installs project dependencies and starts the services needed for development.
- `session-start.sh`: SessionStart hook (registered in `../settings.json`) that
  exports `env.sh`, runs `setup.py` if it hasn't run yet, and runs `start.sh`.
