# Proof T1: seed/narrac.c on wasm32-wasi

Scratch folder: /private/tmp/claude-501/js-interop-design/t1 (thrown away). Base: feat/batch-16 at 9f4017d7. Apple M2, 8 GB, macOS 27, zig 0.17.1, wasmtime 50.0.1, Node 26.12.1. Started and finished in about 9 minutes.

## Build

```sh
zig cc -target wasm32-wasi -O2 -DNARRA_RELEASE -std=c11 -w \
  -mexception-handling -mllvm -wasm-enable-sjlj \
  -o narrac.wasm narrac_wasi.c -lsetjmp -lm
```

The unpatched seed fails only in the `std_process.c` block (`signal.h`, `sys/wait.h`, `pipe`, `sigaction`). Three patches to a copy of the seed made it build and run:

1. `#ifndef __wasi__` around the process block, with a `narra_proc_run` that returns an error.
2. Arguments: the runtime reads `/proc/self/cmdline` on every OS but macOS. WASI has no `/proc`, so `Env.args` was empty and every command but `check <file>` exited 1 with no output. Patched to `__wasi_args_sizes_get` and `__wasi_args_get`.
3. Output: the compiler writes through `Fs.writeText("/dev/stdout")` and `"/dev/stderr"`. WASI has no `/dev`. Patched `narra_fs_write` to send those two paths to `stdout` and `stderr`.

Build time: 68 s at load 4 to 10, 190 s at load 10 to 15.

## Size

| file | bytes |
| --- | --- |
| narrac native (macOS arm64, -O2) | 5,080,736 |
| narrac.wasm (with names and debug sections) | 27,455,976 |
| narrac.wasm after llvm-strip | 6,227,176 |
| stripped, gzip -9 | 1,580,267 |
| stripped, brotli -q 11 | 989,803 |

(`wasm-opt --strip-debug --all-features` wrote a module that wasmtime 50 refused: "compact imports proposal disabled". llvm-strip works.)

## Runs

| command | result |
| --- | --- |
| `wasmtime run -W exceptions=y --dir . narrac.wasm check hello.nra` | exit 0 |
| same, `check bad.nra` (`let x: int = "no"`) | `bad.nra:2:16: expected int, found string`, exit 1 |
| `emit hello.nra`, `emit regex.nra` under wasmtime | byte-identical to the native seed's C |
| `check regex.nra` (211 lines) under wasmtime, first run | 1.38 s real, 239 MB peak (Cranelift compiles the module, then caches it) |
| same, cached | 0.07 s real, 45 MB peak (native: 0.01 s, 8.5 MB) |
| `node run-node.mjs check regex.nra` (node:wasi, no flags) | 219 ms total, 29 ms of it `WebAssembly.compile`; exit 0 |
| `node run-node.mjs check bad.nra` | the same diagnostic, 53 ms, exit 1 |
| `node run-node.mjs emit regex.nra` | byte-identical to native |

Node prints "ExperimentalWarning: WASI is an experimental feature".

## Running the emitted C (server path)

```sh
zig cc -target wasm32-wasi -O2 -std=c11 -w -mexception-handling -mllvm -wasm-enable-sjlj -o regex.wasm regex.c -lsetjmp -lm
```

2.6 s at load 10, 1,425,533 bytes unstripped; `wasmtime run -W exceptions=y regex.wasm` matches `testdata/regex.out`. The zig install is 250 MB (`du -sh /opt/homebrew/Cellar/zig/0.17.1`).

## WASI imports

`WebAssembly.Module.imports` on the stripped module lists 18 preview1 calls: args_get args_sizes_get environ_get environ_sizes_get fd_fdstat_get fd_fdstat_set_flags fd_prestat_get fd_prestat_dir_name fd_read fd_readdir fd_seek fd_write path_filestat_get path_open path_remove_directory path_unlink_file proc_exit fd_close. No clock, random, poll_oneoff or socket call.
