#!/usr/bin/env bash
# Run the fast check (go vet and go test) inside a Linux container.
# The repo is mounted read-only; only the Go caches (a named volume) and /tmp are writable.
# Usage: scripts/linux-check.sh [extra go test args]
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
image="narra-linux-check"
cache="narra-linux-check-cache"

if ! command -v docker >/dev/null 2>&1 || ! docker info >/dev/null 2>&1; then
  echo "linux-check: Docker is not running." >&2
  echo "  On macOS with colima: colima start --cpu 4 --memory 8" >&2
  exit 2
fi

docker build -q -t "$image" "$root/scripts/linux" >/dev/null

start=$SECONDS
status=0
# SYS_PTRACE and an unconfined seccomp profile let LeakSanitizer run inside the container.
docker run --rm \
  --cap-add SYS_PTRACE \
  --security-opt seccomp=unconfined \
  -v "$root:/repo:ro" \
  -v "$cache:/cache" \
  "$image" \
  bash -c 'go vet ./... && go test -timeout 45m ./... "$@"' bash "$@" || status=$?

elapsed=$((SECONDS - start))
if [ "$status" -eq 0 ]; then
  echo "linux-check: PASS in ${elapsed}s"
else
  echo "linux-check: FAIL (exit $status) after ${elapsed}s" >&2
fi
exit "$status"
