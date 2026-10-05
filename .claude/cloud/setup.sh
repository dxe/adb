#!/usr/bin/env bash
# One-time setup for the Claude Code cloud VM. Run from the environment's
# setup script. Safe to re-run.
set -euxo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
source .claude/cloud/env.sh

# `mysql` client for the `db` shell function (the server runs in Docker).
apt-get update -qq
apt-get install -y -qq mysql-client

# Node + pnpm. Unlike project deps (installed per session by start.sh so they
# match the checked-out branch), these change rarely, so bake them in here.
make node_toolchain

# Best effort: pre-pull so session start is fast.
.claude/cloud/start-docker.sh && docker pull mysql:8.4 || true
