#!/usr/bin/env bash

# Runs from conductor.json's scripts.archive, immediately before Conductor
# archives this workspace. Tears down the per-workspace devcontainer resources
# that Conductor itself does not clean up: the Docker Compose project (its
# containers, network, and named volumes) plus the devcontainer image the CLI
# built for this worktree (~2.9GB each).
#
# Best effort: every step is guarded so a missing resource never blocks the
# archive.

set -uo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
compose_file="${script_dir}/compose.workspace.yaml"

# Prefer the project name compose actually used (written by
# write-workspace-compose.sh at setup time). Fall back to deriving it from the
# workspace name with the same sanitization as that script.
project=""
if [[ -f "${compose_file}" ]]; then
  project="$(awk '/^name:[[:space:]]*/{print $2; exit}' "${compose_file}")"
fi
if [[ -z "${project}" ]]; then
  ws="$(printf '%s' "${CONDUCTOR_WORKSPACE_NAME:-}" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9' '-')"
  [[ -n "${ws}" ]] && project="dxe-adb-${ws}"
fi

# Safety: only ever act on our own projects, never an empty name.
if [[ -z "${project}" || "${project}" != dxe-adb-* ]]; then
  echo "workspace-archive: could not determine a dxe-adb compose project; skipping cleanup" >&2
  exit 0
fi

echo "workspace-archive: tearing down compose project '${project}'" >&2
docker compose -p "${project}" down --volumes --remove-orphans 2>/dev/null || true

# Remove the devcontainer image the CLI built for this worktree. Its name is
# vsc-<workspace>-<hash>; the workspace segment is the project name minus the
# dxe-adb- prefix.
image_prefix="vsc-${project#dxe-adb-}-"
images="$(docker image ls --format '{{.Repository}}' 2>/dev/null | grep "^${image_prefix}" || true)"
if [[ -n "${images}" ]]; then
  echo "workspace-archive: removing image(s): ${images}" >&2
  printf '%s\n' "${images}" | xargs -r docker image rm 2>/dev/null || true
fi

exit 0
