#!/bin/sh
# Throwaway: build Go narra, stage 1 and stage 2 into $1 and keep them.
set -eu
root=$(cd "$(dirname "$0")/../../../.." && pwd)
out=$1
mkdir -p "$out"
cd "$root/compiler"
go build -o "$out/narra" ./cmd/narra
cd "$root"
"$out/narra" build -release -o "$out/narrac1" selfhost
"$out/narra" emit selfhost > "$out/go.c"
"$out/narrac1" emit selfhost > "$out/stage1.c"
cc -O2 -DNARRA_RELEASE -std=c11 -w -o "$out/narrac2" "$out/stage1.c" -lm
"$out/narrac2" emit selfhost > "$out/stage2.c"
cmp "$out/go.c" "$out/stage1.c"
cmp "$out/stage1.c" "$out/stage2.c"
echo "ok: fixed point holds"
