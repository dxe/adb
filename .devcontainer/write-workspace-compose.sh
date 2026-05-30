#!/usr/bin/env bash

set -euo pipefail

# Generates compose.workspace.yaml, which Docker Compose merges with the base
# devcontainer compose file to add workspace-specific volume mounts. This runs
# at devcontainer startup time so the generated file reflects the actual paths
# on the host machine (which vary per developer and per worktree).

# The workspace path is passed in as the first argument.
workspace_path="${1:?workspace path is required}"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
output_file="${script_dir}/compose.workspace.yaml"

# Derive a DNS-safe project name from the folder name so each worktree gets its
# own isolated Compose project. Without this, VS Code would reattach to whatever
# container happened to share the same default project name.
workspace_name="$(basename "${workspace_path}")"
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

# Ask git where it stores its data. For a normal repo these two paths are the
# same. For a git worktree they differ: git-dir points to a worktree-specific
# stub, while git-common-dir points to the main repo's .git where objects and
# refs actually live.
abs_git_dir="$(git -C "${workspace_path}" rev-parse --path-format=absolute --git-dir)"
abs_git_common_dir="$(git -C "${workspace_path}" rev-parse --path-format=absolute --git-common-dir)"

# Escape single quotes so paths with apostrophes don't break the YAML output.
escaped_workspace_path=${workspace_path//\'/\'\'}
escaped_abs_git_common_dir=${abs_git_common_dir//\'/\'\'}

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
      - '${escaped_workspace_path}:/workspace:cached'
EOF

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
