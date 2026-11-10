#!/bin/sh
# Builds the release compiler the way bench/run.sh does, plus two symbolized builds of the same C:
#   narrac-g     -O2 -g -fno-omit-frame-pointer, same code as the release build
#   narrac-loop  the same, but main runs the driver NARRA_REPEAT times (default 1), so `sample`
#                can attach to one long-lived process
#   narrac-noinline  narrac-loop with -fno-inline
# Usage: sh build.sh <outdir>
set -eu
root=$(cd "$(dirname "$0")/../../../.." && pwd)
out=$1
mkdir -p "$out"
sh "$root/selfhost/bootstrap.sh" "$out/rel"
flags="-O2 -g -fno-omit-frame-pointer -DNARRA_RELEASE -std=c11 -w"
cc $flags -o "$out/narrac-g" "$out/rel/stage2.c" -lm
# The driver entry is the only call in main that takes args.
sed 's/^  int64_t code = \(f_[0-9a-f]*_run\)(args);$/  int64_t code = 0; for (int repeat = getenv("NARRA_REPEAT") ? atoi(getenv("NARRA_REPEAT")) : 1; repeat > 0; repeat--) code = \1(args);/' \
  "$out/rel/stage2.c" > "$out/loop.c"
grep -q NARRA_REPEAT "$out/loop.c"
# -no_deduplicate keeps the linker from folding identical functions into <deduplicated_symbol>.
cc $flags -Wl,-no_deduplicate -o "$out/narrac-loop" "$out/loop.c" -lm
# Without inlining, retain_*, release_*, lstr_* and the map helpers keep their own frames,
# so the category shares can be counted. It is slower, so its shares are an upper bound.
cc $flags -fno-inline -Wl,-no_deduplicate -o "$out/narrac-noinline" "$out/loop.c" -lm
