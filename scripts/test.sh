#!/bin/sh
# Run the compiler's test suites with the test runner written in Narra
# (selfhost/tests). No Go is involved.
#
#   NARRAC=/path/to/narrac sh scripts/test.sh [runner flags]
#
# NARRAC is the compiler under test. When it is unset, the script builds one
# from seed/narrac.c with `sh selfhost/bootstrap.sh` (a C compiler only, no Go),
# so one command tests a fresh checkout. The compiler builds the runner, and
# the runner runs every sample with it. JOBS is the number of
# shards (default: the number of CPUs); each shard is a runner process that
# takes every JOBS-th case (--shard i/N). In a full run (no flags) the
# sanitized self-build (suite bootstrap), the longest case, is a job of its own
# beside the shards. Flags after the script name go to every shard instead and
# the extra job is not started, for example `--suite errors`,
# `--filter closures`, `--verbose`, `--update`. CC picks the C compiler, as for
# narrac.
set -eu

root=$(cd "$(dirname "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

jobs=${JOBS:-$(getconf _NPROCESSORS_ONLN 2>/dev/null || echo 4)}
case "$jobs" in
  '' | *[!0-9]*)
    echo "test.sh: JOBS must be an integer of at least 1, got: $jobs" >&2
    exit 2
    ;;
esac
if [ "$jobs" -lt 1 ]; then
  echo "test.sh: JOBS must be an integer of at least 1, got: $jobs" >&2
  exit 2
fi

if [ -z "${NARRAC:-}" ]; then
  echo "test.sh: NARRAC is unset; building narrac from the seed"
  sh "$root/selfhost/bootstrap.sh" "$work/compiler" || {
    echo "test.sh: the bootstrap failed" >&2
    exit 2
  }
  narrac=$work/compiler/narrac
else
  narrac=$NARRAC
  case "$narrac" in
    */*)
      if [ ! -f "$narrac" ]; then
        echo "test.sh: NARRAC is set but is not a file: $narrac" >&2
        exit 2
      fi
      narrac=$(cd "$(dirname "$narrac")" && pwd)/$(basename "$narrac")
      ;;
  esac
fi
if ! command -v "$narrac" >/dev/null 2>&1; then
  echo "test.sh: compiler not found: $narrac (set NARRAC)" >&2
  exit 2
fi

echo "test.sh: building the runner with $narrac"
"$narrac" build -release -o "$work/testrunner" "$root/selfhost/tests"

start=$(date +%s)
cd "$root"
pids=""
skip=""
if [ "$#" -eq 0 ]; then
  # A full run: the bootstrap suite gets a job of its own.
  echo "test.sh: $jobs shards and the bootstrap job"
  skip="--skip bootstrap"
  "$work/testrunner" --narrac "$narrac" --repository "$root" --suite bootstrap \
    > "$work/shard-0.log" 2>&1 &
  pids="$pids $!"
else
  echo "test.sh: $jobs shards"
fi
shard=1
while [ "$shard" -le "$jobs" ]; do
  # shellcheck disable=SC2086
  "$work/testrunner" --narrac "$narrac" --repository "$root" $skip \
    --shard "$shard/$jobs" "$@" > "$work/shard-$shard.log" 2>&1 &
  pids="$pids $!"
  shard=$((shard + 1))
done

# Wait for every job; the exit code is 1 when any of them failed.
status=0
for pid in $pids; do
  wait "$pid" || status=1
done

# Print the logs in shard order: shard-0 (bootstrap, if any), then 1, 2, ... N.
shard=0
while [ "$shard" -le "$jobs" ]; do
  if [ -f "$work/shard-$shard.log" ]; then
    cat "$work/shard-$shard.log"
  fi
  shard=$((shard + 1))
done
# The number of cases that ran, summed over the shards, goes on the final line.
ran=$(cat "$work"/shard-*.log | awk '/[0-9]+ passed, [0-9]+ failed/ { for (i = 1; i < NF; i++) if ($(i + 1) == "passed,") total += $i } END { print total + 0 }')
elapsed=$(($(date +%s) - start))
if [ "$status" -eq 0 ]; then
  if [ "$ran" -gt 0 ]; then
    echo "test.sh: PASS, $ran cases in ${elapsed}s"
  else
    echo "test.sh: PASS in ${elapsed}s"
  fi
else
  echo "test.sh: FAIL after ${elapsed}s" >&2
fi
exit "$status"
