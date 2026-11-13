#!/bin/sh
# Builds each testdata/threads_*.nra sample with ThreadSanitizer and runs it.
# Not part of scripts/test.sh: ThreadSanitizer and AddressSanitizer cannot be
# combined, and a TSan build is slow. A report makes the sample fail.
#   NARRAC=/path/to/narrac sh roadmaps/selfhost/nodes/threads/tsan.sh [samples]
set -u
root=$(cd "$(dirname "$0")/../../../.." && pwd)
narrac=${NARRAC:-narrac}
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cd "$root/testdata"
failed=0
for sample in ${*:-threads_*.nra}; do
  stem=${sample%.nra}
  if ! "$narrac" emit "$sample" > "$work/$stem.c"; then
    echo "FAIL $stem: emit"
    failed=1
    continue
  fi
  # The block heap is kept: its remote frees are what TSan should see.
  if ! ${CC:-cc} -std=c11 -w -g -O1 -fsanitize=thread -pthread -o "$work/$stem" "$work/$stem.c" -lm; then
    echo "FAIL $stem: cc"
    failed=1
    continue
  fi
  if TSAN_OPTIONS=halt_on_error=1 "$work/$stem" > "$work/$stem.out" 2> "$work/$stem.err" && cmp -s "$work/$stem.out" "$stem.out"; then
    echo "ok $stem"
  else
    echo "FAIL $stem"
    head -40 "$work/$stem.err"
    failed=1
  fi
done
exit $failed
