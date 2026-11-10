#!/bin/sh
# Builds narrac-split: narrac-loop with inlining kept, except that the functions whose C names
# match the pattern get __attribute__((noinline)), so a profile splits their callers' self time.
# Usage: sh split.sh <build dir from profile-compiler/build.sh> <name> <extended regex for C names>
# Example: sh split.sh out effects 'f_dff5bb07_[A-Za-z0-9_]*|M_Map_[0-9a-f]*_lookup'
set -eu
out=$1
sed -E "s/^static (.*[ *])($3)\(/static __attribute__((noinline)) \1\2(/" "$out/loop.c" > "$out/split-$2.c"
grep -q "__attribute__((noinline))" "$out/split-$2.c"
flags="-O2 -g -fno-omit-frame-pointer -DNARRA_RELEASE -std=c11 -w"
cc $flags -Wl,-no_deduplicate -o "$out/narrac-split-$2" "$out/split-$2.c" -lm
