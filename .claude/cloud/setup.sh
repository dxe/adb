#!/usr/bin/env bash
# One-time setup for the Claude Code cloud VM. Run via `make cloud_setup` from
# the environment's setup script. Safe to re-run.
set -euxo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
source .claude/cloud/env.sh

# `mysql` client for the `db` shell function (the server runs in Docker).
apt-get update -qq
apt-get install -y -qq mysql-client

make deps

# Same shell functions the devcontainer's post-create.sh installs.
cp .devcontainer/.bash_adb_functions ~/.bash_adb_functions
grep -qF '.bash_adb_functions' ~/.bash_profile 2>/dev/null ||
  cat .devcontainer/.bash_profile >> ~/.bash_profile

# Best effort: pre-pull so session start is fast.
.claude/cloud/start-docker.sh && docker pull mysql:8.4 || true
