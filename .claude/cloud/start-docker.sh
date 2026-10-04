#!/usr/bin/env bash
# Start dockerd (needed by `docker` and by the Go tests' testcontainers).
set -euo pipefail
docker info >/dev/null 2>&1 && exit 0
nohup dockerd >/var/log/dockerd.log 2>&1 &
for _ in $(seq 60); do docker info >/dev/null 2>&1 && exit 0; sleep 1; done
echo "dockerd failed to start; see /var/log/dockerd.log" >&2
exit 1
