#!/bin/sh
# Prints the wall time of `narrac emit selfhost` for each binary given, 5 runs each.
# Usage: sh time.sh <narrac> ...
set -eu
root=$(cd "$(dirname "$0")/../../../.." && pwd)
for narrac in "$@"; do
  printf '%s:' "$narrac"
  for i in 1 2 3 4 5; do
    start=$(perl -MTime::HiRes=time -e 'printf "%.3f", time')
    "$narrac" emit "$root/selfhost" >/dev/null
    end=$(perl -MTime::HiRes=time -e 'printf "%.3f", time')
    printf ' %.3f' "$(echo "$end - $start" | bc)"
  done
  echo
done
uptime
