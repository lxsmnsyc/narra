#!/bin/sh
# Regenerates the two emitted C files the forms start from:
#   OUT/inline.c: bench/interp/interp.nra emitted by the compiler at 9f8ee5f1 (D99, inline payloads).
#   OUT/heap.c:   the same file emitted by the compiler at 9501a3cc^ (before option-inline).
# Each compiler is bootstrapped from its own commit's seed. Both emit 9f8ee5f1's interp.nra.
# Usage: gen.sh REPO OUT
set -eu
repo="$1"
out="$2"
mkdir -p "$out"
out="$(cd "$out" && pwd)"
work="$out/build"
for commit in 9f8ee5f1 9501a3cc^; do
  name="$(echo "$commit" | tr -d '^')"
  rm -rf "$work/$name"
  mkdir -p "$work/$name/src"
  git -C "$repo" archive "$commit" | tar -x -C "$work/$name/src"
  (cd "$work/$name/src" && sh selfhost/bootstrap.sh "$work/$name/narrac" >"$work/$name/bootstrap.log" 2>&1)
done
program="$work/9f8ee5f1/src/bench/interp/interp.nra"
"$work/9f8ee5f1/narrac/narrac" emit "$program" >"$out/inline.c"
"$work/9501a3cc/narrac/narrac" emit "$program" >"$out/heap.c"
echo "wrote $out/inline.c and $out/heap.c"
