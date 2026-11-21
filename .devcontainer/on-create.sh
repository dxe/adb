#!/usr/bin/env bash
set -euo pipefail

# Fresh named volumes mount as root:root; chown so vscode can write.
sudo chown -R vscode:vscode /home/vscode/.claude /mnt/mise-data

# Install Claude Code with Anthropic's native installer, into ~/.local/bin
# (on PATH via remoteEnv in devcontainer.json), where it can update itself.
# Before `mise install`, whose postinstall hook installs the gopls-lsp Claude
# Code plugin if `claude` is on PATH. Not the claude-code devcontainer feature,
# which installs a system Node 18 and a root-owned npm install that can't
# self-update.
export PATH="$HOME/.local/bin:$PATH"
command -v claude >/dev/null \
  || curl -fsSL https://claude.ai/install.sh | bash \
  || echo "warning: couldn't install Claude Code" >&2

# Install the tools pinned in /workspace/mise.toml. mise itself comes from the
# devcontainer feature; installs land in the mise-data volume (MISE_DATA_DIR).
mise -C /workspace install

cat /workspace/.devcontainer/.bashrc >> ~/.bashrc
