#!/bin/sh
# The build on Vercel. The code samples get their hovers from narrac, which the site compiles
# once from the C seed of the repository (the seed is current, so no bootstrap stages are needed). The Vercel build image has no C compiler, so this installs gcc
# first. Without a compiler the site still builds, with plain highlighted samples.
set -eu

out=${TMPDIR:-/tmp}/narra-site-narrac
if ! command -v cc >/dev/null 2>&1; then
  if command -v dnf >/dev/null 2>&1; then
    dnf install -y gcc >/dev/null 2>&1 || echo "vercel-build: could not install gcc"
  fi
fi
if command -v cc >/dev/null 2>&1 && cc -O2 -DNARRA_RELEASE -std=c11 -w -o "$out" ../seed/narrac.c -lm; then
  export NARRAC="$out"
else
  echo "vercel-build: no narrac, the samples build without hovers"
fi
pnpm build
