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
local_bin = Path.home() / ".local" / "bin"  # already on PATH in the cloud VM
os.chdir(repo_root)

# --- Install mise, unless it's already installed ---------------------------------------------
if not shutil.which("mise"):
    env = {**os.environ, **({"MISE_VERSION": args.mise_version} if args.mise_version else {})}
    run("bash", "-c", "curl -fsSL https://mise.run | sh", env=env)

# --- Install the tools in mise.toml ----------------------------------------------------------
# Don't let mise install/shim these:
# - python: the VM already ships one, and linking mise's python into ~/.local/bin below would
#   interfere with python calls made outside mise projects.
# - go: the VM already ships one, and go manages its own toolchain version (it downloads the
#   version from the `go`/`toolchain` lines in go.mod as needed).
os.environ["MISE_DISABLE_TOOLS"] = "go,python"

run("mise", "trust", "--yes", "mise.toml")
# Installs the tools and warms the dependency caches baked into the environment snapshot.
# start.sh re-runs this each session so deps still match the checked-out branch; with warm
# caches that only fetches what changed since the snapshot.
run("mise", "run", "deps")

# --- Make every mise-managed tool available to Claude's tool calls ---------------------------
# Their shell is non-interactive and non-login, so it never reads ~/.bashrc or ~/.profile
# (PATH edits there don't work) and can't see the mise shims. Instead, for each executable in
# the installed tools' bin dirs, create a symlink to mise itself under that name in
# ~/.local/bin, which is already on its PATH. Invoked under a tool's name, mise acts as a
# shim: it checks the current directory and loads the environment accordingly.
# `mise bin-paths` lists those dirs for every tool in mise.toml (skipping anything in
# MISE_DISABLE_TOOLS), so tools added to mise.toml later are linked automatically.
# These links are identical across repos, so in multi-repo environments another repo's setup
# script overwriting them is harmless.
mise_bin = shutil.which("mise")
bin_dirs = run("mise", "bin-paths", capture_output=True).stdout.split("\n")

local_bin.mkdir(parents=True, exist_ok=True)
for bin_dir in filter(None, bin_dirs):
    for exe in sorted(Path(bin_dir).iterdir()):
        if exe.is_file() and os.access(exe, os.X_OK):
            link = local_bin / exe.name
            link.unlink(missing_ok=True)
            link.symlink_to(mise_bin)

# --- Pre-pull the MySQL image so session start is fast (best effort) -------------------------
try:
    run(".claude/cloud/start-docker.sh")
    # Keep image in sync with start.sh and server/compose.yaml.
    run("docker", "pull", "mysql:8.4")
except subprocess.CalledProcessError as e:
    print(f"Skipping MySQL image pre-pull: {e}", file=sys.stderr)
