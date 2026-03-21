#!/usr/bin/env bash

set -euo pipefail

workspace_path="${1:?workspace path is required}"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
output_file="${script_dir}/compose.workspace.yaml"
workspace_name="$(basename "${workspace_path}")"
sanitized_workspace_name="$(printf '%s' "${workspace_name}" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9' '-')"
project_name="dxe-adb-${sanitized_workspace_name}"
escaped_workspace_path=${workspace_path//\'/\'\'}

cat >"${output_file}" <<EOF
# Keep the Compose project name unique per worktree so VS Code does not reattach
# to a container created for a different checkout.
name: ${project_name}

services:
  devcontainer:
    volumes:
      - '${escaped_workspace_path}:/workspace:cached'
EOF
