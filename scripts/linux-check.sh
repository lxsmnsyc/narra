#!/usr/bin/env bash
# Build the compiler from the seed and run the test suites inside a Linux
# container: sh scripts/test.sh, which builds the compiler from the seed once.
# The repo is mounted read-only; the runner works in /tmp.
# Usage: scripts/linux-check.sh [runner flags]
# PLATFORM picks the architecture, for example PLATFORM=linux/amd64 on an arm64 Mac
# (emulated, slow). Without it the container runs on the native architecture.
# JOBS sets the number of test shards (default: the CPUs of the container).
# Run it from a path under your home folder: colima does not mount /private/tmp.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
image="narra-linux-check"

if ! command -v docker >/dev/null 2>&1 || ! docker info >/dev/null 2>&1; then
  echo "linux-check: Docker is not running." >&2
  echo "  On macOS with colima: colima start --cpu 4 --memory 8" >&2
  exit 2
fi

platform=()
if [ -n "${PLATFORM:-}" ]; then
  platform=(--platform "$PLATFORM")
fi

docker build -q ${platform[@]+"${platform[@]}"} -t "$image" "$root/scripts/linux" >/dev/null

start=$SECONDS
status=0
# SYS_PTRACE and an unconfined seccomp profile let LeakSanitizer run inside the container.
docker run --rm ${platform[@]+"${platform[@]}"} \
  --cap-add SYS_PTRACE \
  -e JOBS="${JOBS:-}" \
  --security-opt seccomp=unconfined \
  -v "$root:/repo:ro" \
  "$image" \
  sh scripts/test.sh "$@" || status=$?

elapsed=$((SECONDS - start))
if [ "$status" -eq 0 ]; then
  echo "linux-check: PASS in ${elapsed}s"
else
  echo "linux-check: FAIL (exit $status) after ${elapsed}s" >&2
fi
exit "$status"
