#!/bin/sh
# Runs `narrac emit selfhost` N times in a row, for a profiler to attach to.
# Usage: sh loop.sh <narrac> <count>
set -eu
root=$(cd "$(dirname "$0")/../../../.." && pwd)
i=0
while [ "$i" -lt "$2" ]; do
  "$1" emit "$root/selfhost" >/dev/null
  i=$((i + 1))
done
