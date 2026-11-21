#!/usr/bin/env bash
# One-time setup for the Claude Code cloud VM. Run from the environment's
# setup script. Safe to re-run.
set -euxo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
source .claude/cloud/env.sh

# mise installs pinned tool versions (see mise.toml), including the `mysql`
# client, and runs project tasks.
command -v mise >/dev/null 2>&1 || curl -fsSL https://mise.run | sh

# Warm the caches baked into the environment snapshot. This command should be
# re-run in each session so deps still match the checked-out branch; with warm
# caches that only fetches what changed since the snapshot.
mise run deps

# Best effort: pre-pull so session start is fast.
.claude/cloud/start-docker.sh && docker pull mysql:8.4 || true
