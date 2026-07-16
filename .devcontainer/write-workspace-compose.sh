#!/usr/bin/env bash

set -euo pipefail

# Generates compose.workspace.yaml, which Docker Compose merges with the base
# devcontainer compose file to add workspace-specific volume mounts. This runs
# at devcontainer startup time so the generated file reflects the actual paths
# on the host machine (which vary per developer and per worktree).

# The workspace path is passed in as the first argument. On macOS/Linux hosts
# it is already a POSIX path. On Windows hosts VS Code/Cursor pass a native path
# like c:\Users\foo\repo, which we support only in a reduced, single-checkout
# mode (see the Windows branch below).
raw_workspace_path="${1:?workspace path is required}"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
output_file="${script_dir}/compose.workspace.yaml"

# Detect a Windows-style path (a drive-letter prefix such as C:\ or C:/).
#
# Git worktree support requires bind-mounting host paths so the same absolute
# path resolves both inside and outside the container. That does not work
# across the Windows/WSL/Docker-Desktop path-translation boundary without
# juggling three path formats (native, WSL /mnt/c, and Docker c:/...). Rather
# than carry that complexity, Windows gets a reduced setup: the workspace is
# mounted at /workspace and nothing else. Ordinary (non-worktree) checkouts
# work fully; git worktrees are not supported on Windows.
normalized_path="${raw_workspace_path//\\//}"
if [[ "${normalized_path}" =~ ^[A-Za-z]: ]]; then
  is_windows=1
  # Docker Desktop bind mounts want a lowercase drive letter with forward
  # slashes, e.g. c:/Users/foo/repo.
  drive="$(printf '%s' "${normalized_path:0:1}" | tr '[:upper:]' '[:lower:]')"
  docker_workspace_path="${drive}:${normalized_path:2}"
  workspace_name="$(basename "${normalized_path}")"
else
  is_windows=0
  workspace_path="${raw_workspace_path}"
  docker_workspace_path="${workspace_path}"
  workspace_name="$(basename "${workspace_path}")"
fi

# Derive a DNS-safe project name from the folder name so each worktree gets its
# own isolated Compose project. Without this, VS Code would reattach to whatever
# container happened to share the same default project name.
sanitized_workspace_name="$(printf '%s' "${workspace_name}" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9' '-')"
project_name="dxe-adb-${sanitized_workspace_name}"

# Publish the Go server's port (8080 inside the container) to the host. The
# devcontainer CLI ignores devcontainer.json's "forwardPorts" (that is a VS Code
# editor feature, not part of the devcontainer spec), so for CLI users we publish
# at the Docker layer here instead. Each worktree gets a distinct host port to
# avoid collisions when several run in parallel: honor ADB_HOST_PORT when set,
# otherwise derive a stable port from the workspace name.
if [[ -n "${ADB_HOST_PORT:-}" ]]; then
  host_port="${ADB_HOST_PORT}"
  # When a host port is pinned via ADB_HOST_PORT, the app is expected to run with
  # PORT set to the same value, so publish host->container 1:1. This keeps the
  # port the Go server logs identical to the reachable host port (some tools pick
  # the URL to open by parsing the server's stdout).
  container_port="${ADB_HOST_PORT}"
else
  name_hash="$(printf '%s' "${sanitized_workspace_name}" | cksum | cut -d' ' -f1)"
  host_port=$(( 20000 + (name_hash % 20000) ))
  container_port=8080
fi
echo "write-workspace-compose: publishing container port ${container_port} on host port ${host_port}" >&2

# Escape single quotes so paths with apostrophes don't break the YAML output.
escaped_workspace_path=${docker_workspace_path//\'/\'\'}

# Write the base YAML: name the project and mount the workspace at /workspace.
cat >"${output_file}" <<EOF
# Keep the Compose project name unique per worktree so VS Code does not reattach
# to a container created for a different checkout.
name: ${project_name}

services:
  devcontainer:
    ports:
      - '${host_port}:${container_port}'
    volumes:
      # Long syntax (explicit source/target) so a Windows drive-letter path like
      # c:/Users/foo/repo is not mis-parsed by Compose's colon-delimited short
      # form, which is ambiguous when the source itself contains a colon.
      - type: bind
        source: '${escaped_workspace_path}'
        target: /workspace
        consistency: cached
EOF

# Windows: stop here. The reduced setup above is all we support; git worktrees
# need the extra host-path mounts below, which we deliberately skip.
if [[ "${is_windows}" -eq 1 ]]; then
  echo "write-workspace-compose: Windows host detected; git worktrees are not supported (workspace mounted at /workspace only)" >&2
  exit 0
fi

# Ask git where it stores its data. For a normal repo these two paths are the
# same. For a git worktree they differ: git-dir points to a worktree-specific
# stub, while git-common-dir points to the main repo's .git where objects and
# refs actually live.
abs_git_dir="$(git -C "${workspace_path}" rev-parse --path-format=absolute --git-dir)"
abs_git_common_dir="$(git -C "${workspace_path}" rev-parse --path-format=absolute --git-common-dir)"

# Escape single quotes so paths with apostrophes don't break the YAML output.
escaped_abs_git_common_dir=${abs_git_common_dir//\'/\'\'}

# Extra mounts needed only for git worktrees. A worktree's .git is a pointer
# file, not a full directory, so git commands inside the container must also be
# able to reach the main repo's .git at its original absolute host path. We
# mount both the worktree directory and the common git dir at their real paths
# (in addition to the /workspace alias above) so those absolute paths resolve.
if [[ "${abs_git_dir}" != "${abs_git_common_dir}" ]]; then
  cat >>"${output_file}" <<EOF
      - '${escaped_workspace_path}:${escaped_workspace_path}:cached'
      - '${escaped_abs_git_common_dir}:${escaped_abs_git_common_dir}:cached'
EOF
fi
