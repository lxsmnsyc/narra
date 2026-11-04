#!/bin/sh
# Counts direct call sites of each private top-level function in selfhost/src,
# excluding the declaration itself and method-style `.name(` matches.
cd "$(dirname "$0")/../../../../../selfhost/src" || exit 1
z=0; one=0; two=0
for f in *.nra; do
  for n in $(grep -E "^function\*? +[a-zA-Z_]+" "$f" | sed -E 's/^function\*? +([a-zA-Z_0-9]+).*/\1/'); do
    c=$(grep -ohE "(^|[^.a-zA-Z_0-9])$n\(" *.nra | wc -l); c=$((c-1))
    [ $c -le 0 ] && z=$((z+1)); [ $c -eq 1 ] && one=$((one+1)); [ $c -eq 2 ] && two=$((two+1))
  done
done
echo "zero_calls=$z one_call=$one two_calls=$two"
