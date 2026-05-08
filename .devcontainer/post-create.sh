#!/usr/bin/env bash
set -euo pipefail

# Fresh named volumes mount as root:root; chown so vscode can write.
sudo chown -R vscode:vscode /home/vscode/.claude

# Make bash functions available in both interactive and non-interactive shells.
cp /workspace/.devcontainer/.bash_adb_functions ~/.bash_adb_functions
cat /workspace/.devcontainer/.bashrc >> ~/.bashrc
cat /workspace/.devcontainer/.bash_profile >> ~/.bash_profile
