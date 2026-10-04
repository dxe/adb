# Environment for the Claude Code cloud VM. Sourced by the other scripts here
# and exported into Claude's shell by session-start.sh.
# Keep in sync with .devcontainer/compose.extend.yaml.
ADB_REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
export ADB_REPO_ROOT
export NVM_DIR="${NVM_DIR:-/opt/nvm}"
export DB_HOST=127.0.0.1
export DB_PROTOCOL='tcp(127.0.0.1:3306)'
export DB_USER=adb_user
export DB_PASSWORD=adbpassword
export DB_NAME=adb_db
