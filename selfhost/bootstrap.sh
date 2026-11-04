#!/bin/sh
# Bootstrap the self-hosted compiler by hand. Needs only `go` and `cc`.
# Stage 1: the Go compiler builds selfhost/ into narrac1.
# Stage 2: narrac1 emits selfhost/ as C and cc builds it into narrac2.
# Fixed point: narrac2 emits the same C as narrac1 and as Go.
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "stage 0: build the Go compiler"
(cd "$root/compiler" && go build -o "$work/narra" ./cmd/narra)

echo "stage 1: Go compiler builds selfhost"
"$work/narra" build -release -o "$work/narrac1" "$root/selfhost"
"$work/narra" emit "$root/selfhost" > "$work/go.c"

echo "stage 2: narrac1 emits selfhost, cc builds it"
time "$work/narrac1" emit "$root/selfhost" > "$work/stage1.c"
cc -O2 -DNARRA_RELEASE -std=c11 -w -o "$work/narrac2" "$work/stage1.c" -lm

echo "stage 2 binary: $(wc -c < "$work/narrac2" | tr -d ' ') bytes"

echo "fixed point: narrac2 emits selfhost"
time "$work/narrac2" emit "$root/selfhost" > "$work/stage2.c"

status=0
cmp "$work/go.c" "$work/stage1.c" || { echo "FAIL: stage 1 C differs from Go" >&2; status=1; }
cmp "$work/stage1.c" "$work/stage2.c" || { echo "FAIL: stage 2 C differs from stage 1" >&2; status=1; }
if [ "$status" -eq 0 ]; then
  echo "ok: Go, stage 1 and stage 2 emit identical C ($(wc -l < "$work/stage2.c" | tr -d ' ') lines)"
fi
exit "$status"
