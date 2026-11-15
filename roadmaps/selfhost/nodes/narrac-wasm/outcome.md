# Outcome: narrac-wasm

## Built

- `scripts/wasm-build.sh [compiler.c]`: `zig cc -target wasm32-wasi -O2` with `-mexception-handling -mllvm -wasm-enable-sjlj -mllvm -wasm-use-legacy-eh=false`, `-lsetjmp -lm`, stripped at link time (`-Wl,--strip-all`). Writes `dist/narrac.wasm` (gitignored; `OUT` overrides) and prints the raw and brotli size. POSIX sh, `shellcheck` clean. Not part of `scripts/test.sh`.
- The input defaults to `seed/narrac.c`. The committed seed predates wasi-runtime and fails to build for wasm (`sigaction`, `SIG_IGN`, `pthread_sigmask` undeclared in the old process block). Until the batch seed update, build from a current stage: `sh selfhost/bootstrap.sh <dir>`, then `sh scripts/wasm-build.sh <dir>/stage2.c`. After the seed update the default should work; the default path was not run, and every number below comes from a bootstrap stage2.c and a native stage compiler, not from the seed.
- `llvm-strip` is not installed here, so the strip happens in `wasm-ld`. Brief text changed to say so.

## Acceptance

- Byte-identical `emit`, `wasmtime run -W exceptions=y --dir .` against the native stage compiler: `testdata/hello.nra` 45,585 bytes, `testdata/regex.nra` 237,712 bytes, `selfhost` 11,942,282 bytes. `selfhost` needs `--env NARRA_HOST_TARGET=MacOS-Arm64`: the Wasi row has no host OS, so a wasm compiler cannot detect its host, and without the variable `emit selfhost` stops with E0989 (`could not determine the host`). `emit selfhost` under wasm: 7.7 s, 154 MB peak, cached module.
- Size: 6,332,654 bytes raw, 1,001,930 bytes brotli -q 11 (limit 1.2 MB). Native stage compiler 5,153,080 bytes.
- Exceptions off: `wasmtime run -W exceptions=n --dir . dist/narrac.wasm check testdata/hello.nra` exits 0. `wasm-dis -all` finds no `try`, `try_table`, `throw`, `catch` or tag, because the compiler never calls `setjmp`. This does not contradict D122: programs the compiler emits use the flags and need `-W exceptions=y`; the compiler itself does not. The flags stay in the script so a later compiler that uses `setjmp` keeps working.
- The script is not in `scripts/test.sh`.

## Not done

- cross-targets running `check` and `emit` with the module: the goal line mentions it, but no acceptance item does and `scripts/cross-check.sh` belongs to cross-targets. Not touched. The seed row of cross-check for Wasi waits for the seed update.

## For the orchestrator

- A wasm narrac cannot detect its host, so any program that imports `std/build` needs `NARRA_HOST_TARGET` (D123 item 8 ships the wasm compiler as a fallback package, so npm-narrac must set it or the Wasi row needs a host answer).
- Brief text for the strip tool, the env variable and the exceptions item was changed to match.

## Tests

- `sh scripts/test.sh`: `test.sh: PASS, 9915 cases in 503s` (first run; no timing case failed, so no rerun).
- `emit` byte-compare of hello, regex and selfhost, and the `-W exceptions=n` check: see Acceptance.
- `shellcheck scripts/wasm-build.sh`: clean.

## Size

- Estimated S, actual S.
