# Claude Code cloud VM setup

The cloud VM has no devcontainer, so the repo provides the equivalent setup here.
`setup.py` runs in one of two ways:

- From the SessionStart hook, on the first session start in an environment that
  didn't already run it (e.g. the default one), for single-repo sessions.
- From the Claude environment's setup script, which should only run
  `.claude/cloud/setup.py`. This is required when sessions have several repos,
  because hooks don't fire then. It also bakes the setup into the environment
  snapshot, so session starts stay fast.

When `setup.py` is the environment's setup script, set
`CLAUDE_ENV_FILE=/root/.dxe-claude-env.sh` in the Claude environment's env vars.
The setup script doesn't get that env var, so `setup.py` uses that path directly
to add mise's shims and `~/go/bin` (for tools installed with `go install`) to
`PATH` (unless another repo's setup script already did).
Sessions do get it, and Claude Code sources the file before every Bash command.
When the SessionStart hook runs `setup.py`, it uses the session's own
`CLAUDE_ENV_FILE` from Claude Code instead.

`setup.py` also adds go and python to `disable_tools` in mise's global config
(see comments in setup.py for details). gopls and delve are skipped in the cloud VM by
the postinstall hook in `mise.toml`.

- `setup.py`: one-time environment setup, including installing project
  dependencies to warm the caches baked into the environment snapshot.
- `start.sh`: per-session startup: installs project dependencies and starts the services needed for development.
- `session-start.sh`: SessionStart hook (registered in `../settings.json`) that
  exports `env.sh`, runs `setup.py` if it hasn't run yet, and runs `start.sh`.
