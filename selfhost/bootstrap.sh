#!/bin/sh
# Build the self-hosted compiler from the checked-in seed with `cc` alone, and check the fixed point.
#   stage 0: cc builds seed/narrac.c into narrac0.
#   stage 1: narrac0 emits selfhost/ as C; cc builds it into narrac1.
#   stage 2: narrac1 emits selfhost/ as C again.
#   fixed point: the stage 1 C and the stage 2 C must be identical.
#   stage 3, only when they differ: the seed's runtime text is older than
#   selfhost/runtime, so the stage 1 C carries the old runtime. cc builds the
#   stage 2 C into narrac2, which emits selfhost again; then the stage 2 C and
#   the stage 3 C must be identical, and narrac2 is the result.
# Usage: sh selfhost/bootstrap.sh [outdir]
# With outdir, the resulting compiler (narrac1, or narrac2 after stage 3) is copied there as narrac, and the stage 2 C as stage2.c.
# Exit codes: 0 fixed point holds; 1 it does not, or a step failed; 2 bad usage.
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)
# CC must name a single program. It is not split into words.
cc_cmd=${CC:-cc}
if [ "$#" -gt 1 ]; then
  echo "usage: sh selfhost/bootstrap.sh [outdir]" >&2
  exit 2
fi
outdir=${1:-}
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "stage 0: cc builds the seed"
"$cc_cmd" -O2 -DNARRA_RELEASE -std=c11 -w -o "$work/narrac0" "$root/seed/narrac.c" -lm

echo "stage 1: the seed compiler emits selfhost, cc builds it"
"$work/narrac0" emit "$root/selfhost" > "$work/stage1.c"
"$cc_cmd" -O2 -DNARRA_RELEASE -std=c11 -w -o "$work/narrac1" "$work/stage1.c" -lm

echo "stage 2: the stage 1 compiler emits selfhost"
"$work/narrac1" emit "$root/selfhost" > "$work/stage2.c"

echo "stage 1 binary: $(wc -c < "$work/narrac1" | tr -d ' ') bytes"

result="$work/narrac1"
if cmp -s "$work/stage1.c" "$work/stage2.c"; then
  echo "ok: stage 1 and stage 2 emit identical C ($(wc -l < "$work/stage2.c" | tr -d ' ') lines)"
else
  echo "stage 3: stage 1 C differs from stage 2 (older seed runtime); cc builds the stage 2 C and it emits selfhost"
  "$cc_cmd" -O2 -DNARRA_RELEASE -std=c11 -w -o "$work/narrac2" "$work/stage2.c" -lm
  "$work/narrac2" emit "$root/selfhost" > "$work/stage3.c"
  if ! cmp "$work/stage2.c" "$work/stage3.c"; then
    echo "FAIL: stage 3 C differs from stage 2" >&2
    exit 1
  fi
  echo "ok: stage 2 and stage 3 emit identical C ($(wc -l < "$work/stage3.c" | tr -d ' ') lines)"
  result="$work/narrac2"
fi

# Say when the seed is behind the source. This is not a failure.
if sed '1,/^ \*\/$/d' "$root/seed/narrac.c" | cmp -s - "$work/stage2.c"; then
  echo "seed is current"
else
  echo "note: seed/narrac.c is older than the source; sh scripts/seed-update.sh regenerates it"
fi

if [ -n "$outdir" ]; then
  mkdir -p "$outdir"
  cp "$result" "$outdir/narrac"
  cp "$work/stage2.c" "$outdir/stage2.c"
fi
