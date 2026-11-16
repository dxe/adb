# Environment for the Claude Code cloud VM. Sourced by the other scripts here
# and exported into Claude's shell by session-start.sh.
# Keep in sync with .devcontainer/compose.extend.yaml.
ADB_REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
export ADB_REPO_ROOT
export NVM_DIR="${NVM_DIR:-/opt/nvm}"
export DB_HOST=127.0.0.1
export DB_PROTOCOL='tcp(127.0.0.1:3306)'
# DB_USER, DB_PASSWORD and DB_NAME are shared with the devcontainer.
set -a
source "$ADB_REPO_ROOT/server/dev-db.env"
set +a
