#!/bin/sh
# Measurements behind D100. Builds narrac from the seed, rewrites its C into
# allocator and counting variants (variant.py), checks every variant emits the
# same C, and prints instructions retired, user seconds and peak memory.
# Then measures a sign check on retain and release for the bench/ programs.
# Usage: sh roadmaps/selfhost/nodes/regions-design/prototype/measure.sh [runs]
# Work files go to $WORK (default: a new temporary directory). macOS only:
# it reads `instructions retired` from /usr/bin/time -l.
set -eu
here=$(cd "$(dirname "$0")" && pwd)
root=$(cd "$here/../../../../.." && pwd)
runs=${1:-7}
work=${WORK:-$(mktemp -d)}
variants="base freelist bump nofree nocount sign"
mkdir -p "$work"
cd "$root"
echo "work: $work"
uptime

# The compiler and its C.
sh selfhost/bootstrap.sh "$work/out" > "$work/bootstrap.log"
cc -O2 -DNARRA_RELEASE -std=c11 -w -o "$work/v-base" "$work/out/stage2.c" -lm
for mode in count freelist bump nofree nocount sign; do
  python3 "$here/variant.py" "$mode" "$work/out/stage2.c" "$work/v-$mode.c"
  cc -O2 -DNARRA_RELEASE -std=c11 -w -o "$work/v-$mode" "$work/v-$mode.c" -lm
done

# Every variant must emit the same C as the base compiler.
"$work/v-base" emit selfhost > "$work/emit-base.c"
for mode in count $variants; do
  "$work/v-$mode" emit selfhost > "$work/emit-$mode.c" 2> "$work/err-$mode.txt"
  if cmp -s "$work/emit-base.c" "$work/emit-$mode.c"; then echo "$mode: same C"; else echo "$mode: DIFFERENT C"; fi
done
echo "counts: $(cat "$work/err-count.txt")"

median() {
  sort -n | awk '{ v[NR] = $1 } END { print v[int((NR + 1) / 2)] }'
}

# Runs the variants alternately on one command line.
time_variants() {
  for v in $variants; do : > "$work/t-$v.txt"; done
  i=0
  while [ "$i" -lt "$runs" ]; do
    for v in $variants; do
      /usr/bin/time -l "$work/v-$v" "$@" 2> "$work/one.txt" > /dev/null || true
      awk '/ user / { u = $3 } /instructions retired/ { n = $1 } /maximum resident/ { m = $1 } END { print u, n, m }' "$work/one.txt" >> "$work/t-$v.txt"
    done
    i=$((i + 1))
  done
  for v in $variants; do
    u=$(awk '{ print $1 }' "$work/t-$v.txt" | median)
    n=$(awk '{ print $2 }' "$work/t-$v.txt" | median)
    m=$(awk '{ print $3 }' "$work/t-$v.txt" | median)
    echo "$v user $u instructions $n peak_bytes $m"
  done
}

echo "== narrac emit selfhost"
time_variants emit selfhost
uptime
echo "== narrac fmt --check selfhost/src"
time_variants fmt --check selfhost/src
uptime

# The sign check on the bench programs.
echo "== bench, sign check on array, map, box and enum retain and release"
mkdir -p "$work/bench"
for p in fib sieve matmul strings interp particles; do
  "$work/out/narrac" emit "bench/$p/$p.nra" > "$work/bench/$p.c"
  python3 "$here/variant.py" sign "$work/bench/$p.c" "$work/bench/$p-sign.c"
  cc -O2 -DNARRA_RELEASE -std=c11 -w -o "$work/bench/$p-base" "$work/bench/$p.c" -lm
  cc -O2 -DNARRA_RELEASE -std=c11 -w -o "$work/bench/$p-sign" "$work/bench/$p-sign.c" -lm
  for v in base sign; do
    n=$(i=0; while [ "$i" -lt 5 ]; do /usr/bin/time -l "$work/bench/$p-$v" 2>&1 > /dev/null | awk '/instructions retired/ { print $1 }'; i=$((i + 1)); done | median)
    echo "$p $v instructions $n"
  done
done
uptime
