#!/bin/sh
# Probe for other match-heavy code: builds the compiler (seed/narrac.c, current on
# feat/batch-5) with and without AArch64 early if-conversion, the pass that turns the
# interp arms into selects, checks both emit the same C, and prints the median cycles
# and instructions of `narrac emit selfhost`, interleaved.
# Usage: compiler.sh REPO WORKDIR [ROUNDS]
set -eu
repo="$1"
work="$2"
rounds="${3:-9}"
mkdir -p "$work"
cd "$work"
cc -O2 -DNARRA_RELEASE -std=c11 -w -o narrac-default "$repo/seed/narrac.c" -lm
cc -O2 -DNARRA_RELEASE -std=c11 -w -mllvm -aarch64-enable-early-ifcvt=false -o narrac-noifcvt "$repo/seed/narrac.c" -lm
./narrac-default emit "$repo/selfhost" >default.c
./narrac-noifcvt emit "$repo/selfhost" >noifcvt.c
cmp default.c noifcvt.c
rm -f times.txt
i=0
while [ "$i" -lt "$rounds" ]; do
  for exe in narrac-default narrac-noifcvt; do
    /usr/bin/time -l "./$exe" emit "$repo/selfhost" >/dev/null 2>"$exe.time"
    cycles="$(awk '/cycles elapsed/ { print $1 }' "$exe.time")"
    instructions="$(awk '/instructions retired/ { print $1 }' "$exe.time")"
    echo "$exe $cycles $instructions" >>times.txt
  done
  i=$((i + 1))
done
for exe in narrac-default narrac-noifcvt; do
  cycles="$(awk -v n="$exe" '$1 == n { print $2 }' times.txt | sort -n | awk '{ v[NR] = $1 } END { print v[int((NR + 1) / 2)] }')"
  instructions="$(awk -v n="$exe" '$1 == n { print $3 }' times.txt | sort -n | awk '{ v[NR] = $1 } END { print v[int((NR + 1) / 2)] }')"
  echo "$exe cycles $cycles instructions $instructions"
done
