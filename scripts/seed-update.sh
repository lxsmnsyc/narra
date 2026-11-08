#!/bin/sh
# Regenerate seed/narrac.c from the current source of selfhost/.
# It runs the bootstrap from the current seed. When the fixed point holds, the stage 2 C
# becomes the new seed, under a header that names the commit of the source.
# On an unchanged tree the seed stays byte-identical.
# Usage: sh scripts/seed-update.sh
# Exit codes: 0 seed written (or already current); 1 the bootstrap failed, the seed is untouched; 2 bad usage.
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)
if [ "$#" -ne 0 ]; then
  echo "usage: sh scripts/seed-update.sh" >&2
  exit 2
fi
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

sh "$root/selfhost/bootstrap.sh" "$work/out"

# The commit that last changed selfhost/, so the header does not change when only the seed changes.
commit=$(cd "$root" && git log -1 --format=%h -- selfhost 2>/dev/null || true)
[ -n "$commit" ] || commit=unknown

{
  cat <<HEADER
/*
 * seed/narrac.c: generated, never edit by hand.
 * It is the C that the self-hosted compiler (selfhost/) emits for itself.
 * Source: selfhost/ at commit $commit.
 * Made by: sh scripts/seed-update.sh
 * Build it with: cc -O2 -DNARRA_RELEASE -std=c11 -w -o narrac seed/narrac.c -lm
 * Read seed/README.md before changing how it is made.
 */
HEADER
  cat "$work/out/stage2.c"
} > "$work/narrac.c"

if cmp -s "$work/narrac.c" "$root/seed/narrac.c"; then
  echo "seed unchanged"
else
  cp "$work/narrac.c" "$root/seed/narrac.c"
  echo "seed updated: seed/narrac.c"
fi
