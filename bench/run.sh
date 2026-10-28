#!/usr/bin/env bash
# Builds every benchmark program in Narra, Go and Rust in release mode, runs each
# several times and prints a table of median wall-clock seconds.
#
# Usage: bench/run.sh [-n RUNS] [program ...]
#
# Needs a quiet machine. A language whose toolchain is missing is skipped with a note.
set -u

here="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "$here/.." && pwd)"
build="$here/build"
runs=7

if [ "${1:-}" = "-n" ]; then
  runs="$2"
  shift 2
fi

if [ "$#" -gt 0 ]; then
  programs=("$@")
else
  programs=(fib sieve matmul strings interp particles)
fi

mkdir -p "$build"

have_narra=1
have_go=1
have_rust=1

command -v cc >/dev/null 2>&1 || { echo "note: no C compiler found, skipping Narra" >&2; have_narra=0; }
command -v go >/dev/null 2>&1 || { echo "note: go not found, skipping Go (and the Narra compiler build)" >&2; have_go=0; have_narra=0; }
command -v rustc >/dev/null 2>&1 || { echo "note: rustc not found, skipping Rust" >&2; have_rust=0; }

if [ "$have_narra" = 1 ]; then
  if ! (cd "$root/compiler" && go build -o "$build/narra" ./cmd/narra); then
    echo "note: the Narra compiler failed to build, skipping Narra" >&2
    have_narra=0
  fi
fi

printf '%-12s %10s %10s %10s   %s\n' program narra go rust check
failed=0

for p in "${programs[@]}"; do
  dir="$here/$p"
  cols=()
  outputs=()

  for lang in narra go rust; do
    exe="$build/$p-$lang"
    rm -f "$exe"
    ok=1
    case "$lang" in
      narra)
        [ "$have_narra" = 1 ] || ok=0
        [ "$ok" = 1 ] && { "$build/narra" build -release -o "$exe" "$dir/$p.nra" >"$build/$p-$lang.log" 2>&1 || ok=0; }
        ;;
      go)
        [ "$have_go" = 1 ] || ok=0
        [ "$ok" = 1 ] && { go build -o "$exe" "$dir/$p.go" >"$build/$p-$lang.log" 2>&1 || ok=0; }
        ;;
      rust)
        [ "$have_rust" = 1 ] || ok=0
        [ "$ok" = 1 ] && { rustc -C opt-level=3 -o "$exe" "$dir/$p.rs" >"$build/$p-$lang.log" 2>&1 || ok=0; }
        ;;
    esac
    if [ "$ok" = 0 ]; then
      cols+=("-")
      outputs+=("")
      continue
    fi
    result="$(perl "$here/median.pl" "$runs" "$exe")" || { cols+=("FAIL"); outputs+=(""); failed=1; continue; }
    cols+=("${result%% *}")
    outputs+=("${result#* }")
  done

  # All languages that ran must print the same answer.
  check=ok
  ref=""
  for o in "${outputs[@]}"; do
    [ -z "$o" ] && continue
    if [ -z "$ref" ]; then ref="$o"; elif [ "$o" != "$ref" ]; then check=MISMATCH; failed=1; fi
  done
  printf '%-12s %10s %10s %10s   %s\n' "$p" "${cols[0]}" "${cols[1]}" "${cols[2]}" "$check ($ref)"
done

echo
echo "median of $runs runs, seconds, wall clock. '-' means the language was skipped."
exit "$failed"
