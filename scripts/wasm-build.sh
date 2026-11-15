#!/bin/sh
# Build the compiler for Wasi-Wasm32 with zig cc and write dist/narrac.wasm.
#
#   sh scripts/wasm-build.sh [compiler.c]
#
# compiler.c is the compiler as C. It defaults to seed/narrac.c. While the seed
# is older than the runtime, pass the C that `selfhost/bootstrap.sh <dir>` writes
# as <dir>/stage2.c. The module is stripped at link time (wasm-ld --strip-all),
# which keeps the name section out. wasm-opt --strip-debug --all-features is not
# used: wasmtime 50 refuses its output.
#
# Run it with `wasmtime run --dir . dist/narrac.wasm <command>`. The compiler
# never calls setjmp, so the module holds no exception instruction and runs with
# exceptions off too. Programs it emits do need `-W exceptions=y`. A program
# that imports std/build also needs `--env NARRA_HOST_TARGET=<row>`, because a
# wasm compiler cannot detect its host. Not part of scripts/test.sh: it needs zig.
#
# Environment:
#   OUT  output path (default dist/narrac.wasm)
set -eu

root=$(cd "$(dirname "$0")/.." && pwd)
source_file=${1:-$root/seed/narrac.c}
output=${OUT:-$root/dist/narrac.wasm}

if [ "$#" -gt 1 ]; then
  echo "usage: sh scripts/wasm-build.sh [compiler.c]" >&2
  exit 2
fi
if ! command -v zig >/dev/null 2>&1; then
  echo "wasm-build: zig is not installed" >&2
  exit 2
fi
if [ ! -f "$source_file" ]; then
  echo "wasm-build: $source_file does not exist" >&2
  exit 2
fi

mkdir -p "$(dirname "$output")"

# The exception flags are those of targetCompilerArguments for the Wasi row
# (buildtarget.nra); they are copied here, so change both together. It is a
# command, so no -mexec-model=reactor.
zig cc -target wasm32-wasi -O2 -DNARRA_RELEASE -std=c11 -w \
  -mexception-handling -mllvm -wasm-enable-sjlj -mllvm -wasm-use-legacy-eh=false \
  -Wl,--strip-all \
  -o "$output" "$source_file" -lsetjmp -lm

raw_size=$(wc -c <"$output" | tr -d ' ')
if command -v brotli >/dev/null 2>&1; then
  brotli_size=$(brotli -q 11 -c "$output" | wc -c | tr -d ' ')
  echo "wasm-build: $output $raw_size bytes, $brotli_size bytes brotli"
else
  echo "wasm-build: $output $raw_size bytes (brotli is not installed)"
fi
