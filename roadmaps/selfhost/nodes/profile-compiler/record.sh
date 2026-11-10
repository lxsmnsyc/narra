#!/bin/sh
# Samples narrac-loop while it emits selfhost many times in one process.
# Usage: sh record.sh <narrac-loop> <output.txt> <seconds>
set -eu
root=$(cd "$(dirname "$0")/../../../.." && pwd)
cd "$root"
NARRA_REPEAT=1000 "$1" emit selfhost >/dev/null &
pid=$!
sleep 1
uptime
sample "$pid" "$3" 1 -mayDie -file "$2" >/dev/null 2>&1 || true
kill "$pid" 2>/dev/null || true
uptime
