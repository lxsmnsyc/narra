#!/bin/sh
# Times the variants built by build.sh, alternating, median of $1 runs.
# Usage: WORK=<dir> sh measure.sh runs repo-root   (after build.sh)
# nocount is left out: it reports false errors on emit selfhost. Run it on
# `fmt --check selfhost/src` by hand.
set -eu
here=$WORK
runs=$1
root=$2
variants="base plain bump nofree sign signfl"
cd "$root"
uptime
"$here/v-base" emit selfhost > "$here/emit-base.c"
for v in count $variants; do
  "$here/v-$v" emit selfhost > "$here/emit-$v.c" 2> "$here/err-$v.txt"
  if cmp -s "$here/emit-base.c" "$here/emit-$v.c"; then echo "$v: same C"; else echo "$v: DIFFERENT C"; fi
done
echo "counts: $(cat "$here/err-count.txt")"

median() {
  sort -n | awk '{ v[NR] = $1 } END { print v[int((NR + 1) / 2)] }'
}

time_variants() {
  for v in $variants; do : > "$here/t-$v.txt"; done
  i=0
  while [ "$i" -lt "$runs" ]; do
    for v in $variants; do
      /usr/bin/time -l "$here/v-$v" "$@" 2> "$here/one.txt" > /dev/null || true
      awk '/ user / { u = $3 } /instructions retired/ { n = $1 } /maximum resident/ { m = $1 } END { print u, n, m }' "$here/one.txt" >> "$here/t-$v.txt"
    done
    i=$((i + 1))
  done
  for v in $variants; do
    u=$(awk '{ print $1 }' "$here/t-$v.txt" | median)
    n=$(awk '{ print $2 }' "$here/t-$v.txt" | median)
    m=$(awk '{ print $3 }' "$here/t-$v.txt" | median)
    echo "$v user $u instructions $n peak_bytes $m"
  done
}

echo "== narrac emit selfhost"
time_variants emit selfhost
uptime
echo "== narrac fmt --check selfhost/src"
time_variants fmt --check selfhost/src
uptime
