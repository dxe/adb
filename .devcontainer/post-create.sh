#!/usr/bin/env bash
set -euo pipefail

# Fresh named volumes mount as root:root; chown so vscode can write.
sudo chown -R vscode:vscode /home/vscode/.claude

# Install mise (tool versions and tasks; see /workspace/mise.toml) and the
# tools it pins.
command -v mise >/dev/null 2>&1 || [ -x ~/.local/bin/mise ] || curl -fsSL https://mise.run | sh
~/.local/bin/mise -C /workspace install

cat /workspace/.devcontainer/.bashrc >> ~/.bashrc
