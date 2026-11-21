# Environment for the Claude Code cloud VM. Sourced by the other scripts here.
# Also exported into Claude's shell by session-start.sh, but note that hook only
# fires when this is the only repo in the session, so with several repos
# Claude's shell doesn't get these.
# Keep in sync with .devcontainer/compose.extend.yaml.
ADB_REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
export ADB_REPO_ROOT
export MISE_YES=1
export DB_HOST=127.0.0.1
export DB_PROTOCOL='tcp(127.0.0.1:3306)'
# DB_USER, DB_PASSWORD and DB_NAME are shared with the devcontainer.
set -a
source "$ADB_REPO_ROOT/server/dev-db.env"
set +a
