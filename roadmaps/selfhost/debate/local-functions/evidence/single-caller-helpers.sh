#!/bin/sh
# Counts private top-level functions in selfhost/src and how many have exactly one call site.
cd "$(dirname "$0")/../../../../../selfhost/src" || exit 1
tot=0; one=0
for f in *.nra; do
  for n in $(grep -E "^function\*? +[a-zA-Z_]+" "$f" | sed -E 's/^function\*? +([a-zA-Z_0-9]+).*/\1/'); do
    tot=$((tot+1)); c=$(grep -ohE "\b$n\(" *.nra | wc -l); c=$((c))
    [ $c -le 2 ] && one=$((one+1))
  done
done
echo "private=$tot single_call_site=$one"
