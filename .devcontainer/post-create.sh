#!/usr/bin/env bash
set -euo pipefail

cp /workspace/.devcontainer/.bash_adb_functions ~/.bash_adb_functions
cat /workspace/.devcontainer/.bashrc >> ~/.bashrc
cat /workspace/.devcontainer/.bash_profile >> ~/.bash_profile
