#!/usr/bin/env bash
# Per-session startup for the Claude Code cloud VM: project dependencies, shell
# helper functions, and dockerd + MySQL with a migrated, seeded dev
# database. Idempotent.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
source .claude/cloud/env.sh

# Project dependencies (node, go, `migrate`), kept in sync with the checked-out
# branch. Must precede `mise run db:init` below.
mise run deps
mise run db:up
mise run db:init
