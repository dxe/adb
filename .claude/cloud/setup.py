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
os.chdir(repo_root)

# --- Install mise, unless it's already installed ---------------------------------------------
if not shutil.which("mise"):
    env = {**os.environ, **({"MISE_VERSION": args.mise_version} if args.mise_version else {})}
    run("bash", "-c", "curl -fsSL https://mise.run | sh", env=env)

# --- Install the tools in mise.toml ----------------------------------------------------------
# MISE_DISABLE_TOOLS=go,python must be set in the Claude environment's env vars (see
# README.md), so mise doesn't install/shim these, here or in any later session:
# - python: the VM already ships one, and mise's python shim (on PATH, see below) would
#   interfere with python calls made outside mise projects.
# - go: the VM already ships one, and go manages its own toolchain version (it downloads the
#   version from the `go`/`toolchain` lines in go.mod as needed).
disabled_tools = set(filter(None, os.environ.get("MISE_DISABLE_TOOLS", "").split(",")))

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
# automatically. CLAUDE_ENV_FILE must be set in the Claude environment's env vars (see
# README.md). In multi-repo environments, another repo's setup script may already have added
# the shims, so only add them if the file doesn't mention them yet.
def add_mise_shims_to_env_file():
    """Return False if CLAUDE_ENV_FILE isn't set (reported as an error at the end)."""
    env_file = os.environ.get("CLAUDE_ENV_FILE")
    if not env_file:
        return False
    env_path = Path(env_file)
    existing = env_path.read_text() if env_path.exists() else ""
    if "mise/shims" in existing:
        return True
    env_path.parent.mkdir(parents=True, exist_ok=True)
    with env_path.open("a") as f:
        if existing and not existing.endswith("\n"):
            f.write("\n")
        f.write('export PATH="$HOME/.local/share/mise/shims:$PATH"\n')
    return True


env_file_ok = add_mise_shims_to_env_file()

# --- Pre-pull the MySQL image so session start is fast (best effort) -------------------------
try:
    run(".claude/cloud/start-docker.sh")
    # Keep image in sync with start.sh and server/compose.yaml.
    run("docker", "pull", "mysql:8.4")
except subprocess.CalledProcessError as e:
    print(f"Skipping MySQL image pre-pull: {e}", file=sys.stderr)

# --- Report problems that shouldn't stop the rest of setup -----------------------------------
# Fail last, so everything above still runs and the errors are the final thing shown.
errors = []
if not env_file_ok:
    errors.append("CLAUDE_ENV_FILE is not set, so mise tools won't be on PATH.")
if not {"go", "python"} <= disabled_tools:
    errors.append("MISE_DISABLE_TOOLS must include go and python.")
if errors:
    for error in errors:
        print(f"ERROR: {error}", file=sys.stderr)
    sys.exit("Fix the Claude environment's env vars; see .claude/cloud/README.md.")
