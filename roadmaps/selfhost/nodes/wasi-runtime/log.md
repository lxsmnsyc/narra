# Log: wasi-runtime

- Branch feat/wasi-runtime from feat/batch-16 (rebased onto bf9fa952 for E1510; no new `let`).
- Found: `-O0` fails on wasm (`__c_longjmp`). Coordinator note (D122): defining the tag in `runtime.h` fixes it, so the `-O1` special case is gone.
- E0991 for builtins and whole modules added in buildconditions.nra, since `@target` takes one OS.
- Async: timers-only loop on wasi (nanosleep).
- Sweep over testdata with wasmtime: 295 pass, 53 left out (see outcome).
- Added `-mllvm -wasm-use-legacy-eh=false` and the `__c_longjmp` tag (runtime.h, `__wasi__`); `-O0` links now.
- Found: E0991 on Wasi only ran when the program imported std/build; `chooseNamedTarget` now uses the real target for Wasi.
- Moved `wasi_left_out` out of `testdata/*.nra` (the programs suite built it and wanted a `.out`).
