#!/usr/bin/env bash
set -euo pipefail

# Fresh named volumes mount as root:root; chown so vscode can write.
sudo chown -R vscode:vscode /home/vscode/.claude /mnt/mise-data

# Install the tools pinned in /workspace/mise.toml. mise itself comes from the
# devcontainer feature; installs land in the mise-data volume (MISE_DATA_DIR).
mise -C /workspace install

cat /workspace/.devcontainer/.bashrc >> ~/.bashrc
