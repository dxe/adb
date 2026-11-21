#!/usr/bin/env python3
"""One-time setup for the Claude Code cloud VM. Safe to re-run.

Tool versions are defined in mise.toml at the repo root.
Optionally pin mise with MISE_VERSION=<version> or --mise-version <version>.
"""
import argparse
import os
import shutil
import subprocess
import sys
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument(
    "--mise-version",
    default=os.environ.get("MISE_VERSION"),
    help="mise version to install, e.g. v2025.1.0 (default: latest). Ignored if mise is installed.",
)
args = parser.parse_args()


def run(*cmd, **kwargs):
    """Echo the command (like `set -x`), run it, and stop the script if it fails."""
    print("+", *cmd, file=sys.stderr, flush=True)
    return subprocess.run(cmd, check=True, text=True, **kwargs)


repo_root = Path(__file__).resolve().parents[2]
# Keep in sync with session-start.sh.
SETUP_DONE_MARKER = Path.home() / ".local/state/adb/cloud-setup.done"
os.chdir(repo_root)

# --- Install mise, unless it's already installed ---------------------------------------------
if not shutil.which("mise"):
    env = {**os.environ, **({"MISE_VERSION": args.mise_version} if args.mise_version else {})}
    run("bash", "-c", "curl -fsSL https://mise.run | sh", env=env)

# --- Install the tools in mise.toml ----------------------------------------------------------
# Disable these in mise's global config (~/.config/mise/config.toml), so mise doesn't install
# or shim them, here or in any later session, in this repo or any other:
# - python: the VM already ships one, and mise's python shim (on PATH, see below) would
#   interfere with python calls made outside mise projects.
# - go: the VM already ships one, and go manages its own toolchain version (it downloads the
#   version from the `go`/`toolchain` lines in go.mod as needed).
# - gopls: it's slow to install (built from source), which isn't worth it for ephemeral cloud
#   sessions; it's only in mise.toml for VS Code's Go extension.
# - delve (dlv): Claude usually wouldn't use an interactive debugger (it debugs via logs and
#   tests), and it's built from source, so it's not worth the install time.
for tool in ("go", "python", "go:golang.org/x/tools/gopls", "go:github.com/go-delve/delve/cmd/dlv"):
    run("mise", "settings", "add", "disable_tools", tool)

run("mise", "trust", "--yes", "mise.toml")
# Installs the tools and warms the dependency caches baked into the environment snapshot.
# start.sh re-runs this each session so deps still match the checked-out branch; with warm
# caches that only fetches what changed since the snapshot.
run("mise", "run", "deps")

# --- Make every mise-managed tool available to Claude's tool calls ---------------------------
# Their shell is non-interactive and non-login, so it never reads ~/.bashrc or ~/.profile
# (PATH edits there don't work). It does source $CLAUDE_ENV_FILE before each command, so put
# mise's shims dir on PATH there. Invoked under a tool's name, a shim checks the current
# directory and loads the environment accordingly; tools added to mise.toml later get shims
# automatically.
#
# The environment's setup script doesn't get CLAUDE_ENV_FILE, so use the path the Claude
# environment's env vars set it to for sessions (see README.md). When the SessionStart hook runs
# this instead, Claude Code sets CLAUDE_ENV_FILE to the session's own env file, so use that.
# In multi-repo environments, another repo's setup script may already have added the shims, so
# only add them if the file doesn't mention them yet.
env_path = Path(os.environ.get("CLAUDE_ENV_FILE") or "/root/.dxe-claude-env.sh")
existing = env_path.read_text() if env_path.exists() else ""
if "mise/shims" not in existing:
    env_path.parent.mkdir(parents=True, exist_ok=True)
    with env_path.open("a") as f:
        if existing and not existing.endswith("\n"):
            f.write("\n")
        f.write('export PATH="$HOME/.local/share/mise/shims:$PATH"\n')

# --- Pre-pull the MySQL image so session start is fast (best effort) -------------------------
try:
    run(".claude/cloud/start-docker.sh")
    # Keep image in sync with start.sh and server/compose.yaml.
    run("docker", "pull", "mysql:8.4")
except subprocess.CalledProcessError as e:
    print(f"Skipping MySQL image pre-pull: {e}", file=sys.stderr)

# Tells session-start.sh that setup already ran (e.g. as the environment's setup script).
SETUP_DONE_MARKER.parent.mkdir(parents=True, exist_ok=True)
SETUP_DONE_MARKER.touch()
