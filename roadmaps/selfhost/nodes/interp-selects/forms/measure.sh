#!/bin/sh
# Builds every form from make.py with the flags of `narrac build -release`, then runs
# them interleaved and prints the median user seconds, cycles and instructions retired.
# SRC holds inline.c and heap.c from gen.sh.
# Usage: measure.sh SRC WORKDIR [ROUNDS]
set -eu
here="$(cd "$(dirname "$0")" && pwd)"
source_dir="$1"
work="$2"
rounds="${3:-15}"
python3 "$here/make.py" "$source_dir" "$work"
cd "$work"
for c in *.c; do
  cc -O2 -DNARRA_RELEASE -std=c11 -w -o "${c%.c}" "$c" -lm
done
rm -f times.txt
i=0
while [ "$i" -lt "$rounds" ]; do
  for c in *.c; do
    exe="${c%.c}"
    /usr/bin/time -l "./$exe" >"$exe.out" 2>"$exe.time"
    user="$(awk '/ user / { print $3 }' "$exe.time")"
    instructions="$(awk '/instructions retired/ { print $1 }' "$exe.time")"
    cycles="$(awk '/cycles elapsed/ { print $1 }' "$exe.time")"
    echo "$exe $user $instructions $cycles $(cat "$exe.out")" >>times.txt
  done
  i=$((i + 1))
done
# Median per form.
sort times.txt | awk '
  { name[$1] = 1; users[$1, ++count[$1]] = $2; instr[$1, count[$1]] = $3; cyc[$1, count[$1]] = $4; out[$1] = $5 }
  END {
    for (n in name) {
      k = count[n]
      for (a = 1; a <= k; a++) { u[a] = users[n, a]; s[a] = instr[n, a]; y[a] = cyc[n, a] }
      for (a = 1; a <= k; a++) for (b = a + 1; b <= k; b++) {
        if (u[b] < u[a]) { t = u[a]; u[a] = u[b]; u[b] = t }
        if (s[b] < s[a]) { t = s[a]; s[a] = s[b]; s[b] = t }
        if (y[b] < y[a]) { t = y[a]; y[a] = y[b]; y[b] = t }
      }
      m = int((k + 1) / 2)
      printf "%-20s user %.3f s  cycles %.3f G  instructions %.3f G  output %s\n", n, u[m], y[m] / 1e9, s[m] / 1e9, out[n]
    }
  }' | sort
