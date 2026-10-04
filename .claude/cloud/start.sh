#!/usr/bin/env bash
# Per-session startup for the Claude Code cloud VM: dockerd + MySQL 8.4 with a
# migrated, seeded dev database. Idempotent.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
source .claude/cloud/env.sh

.claude/cloud/start-docker.sh

if docker ps -a --format '{{.Names}}' | grep -qx adb-mysql; then
  docker start adb-mysql >/dev/null
else
  # Keep image in sync with server/compose.yaml.
  docker run -d --name adb-mysql -p 3306:3306 \
    -e MYSQL_ROOT_PASSWORD=localdev \
    -v "$PWD/server/scripts/init.sql":/docker-entrypoint-initdb.d/init.sql \
    mysql:8.4 >/dev/null
fi

for _ in $(seq 90); do
  mysqladmin ping -h127.0.0.1 -u"$DB_USER" -p"$DB_PASSWORD" --silent 2>/dev/null && break
  sleep 2
done
mysqladmin ping -h127.0.0.1 -u"$DB_USER" -p"$DB_PASSWORD" --silent

# First start only: migrate and insert fake dev data.
tables=$(mysql -h127.0.0.1 -u"$DB_USER" -p"$DB_PASSWORD" -N -e \
  "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$DB_NAME'" 2>/dev/null)
if [ "$tables" = 0 ]; then
  make dev_db
fi
