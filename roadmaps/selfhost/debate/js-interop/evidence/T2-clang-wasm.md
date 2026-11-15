# Proof T2: a C compiler in wasm on emitted Narra C

Scratch folder: /private/tmp/claude-501/js-interop-design/t2 (thrown away). Installed there, and nowhere else: `npm install @yowasp/clang@22.0.0-git20542-10` (Clang and LLD 22 built for WebAssembly by the YoWASP project, run through its JS API). tcc was not tried: it is not on the machine and upstream tcc has no wasm backend.

## Size

| file | bytes |
| --- | --- |
| node_modules total | 109 MB |
| gen/llvm.core.wasm | 72.0 MB (22,637,160 gzip -9; 16,917,886 brotli -q 9) |
| gen/llvm-resources.tar (wasi sysroot, headers, libc) | 28.3 MB (2,849,518 brotli -q 9) |

A page would download about 19.8 MB brotli for the C compiler, against 0.99 MB for narrac.wasm.

## Compile

`compile.mjs` calls `runClang(["clang", "--target=wasm32-wasip1", "-O0" or "-O2", "-std=c11", "-w", "-mexception-handling", "-mllvm", "-wasm-enable-sjlj", "program.c", "-o", "program.wasm", "-lsetjmp", "-lm"])` on the C that native narrac emits. Each run is a fresh Node process, so it includes loading and compiling the 72 MB module. Load average 15 to 48 (other workers building), so treat the times as upper bounds.

| input | flags | time | output | RSS |
| --- | --- | --- | --- | --- |
| hello.c (44,700 bytes) | -O0 | 6.2 s | 35,640 bytes | 236 MB |
| regex.c (236,827 bytes) | -O0 | 17.2 s | 134,251 bytes | 194 MB |
| regex.c | -O2 | 62.3 s | 139,104 bytes | 277 MB |

`wasmtime run -W exceptions=y` on both outputs prints `Hello, world` and output identical to `testdata/regex.out`.

Most of hello.c is the runtime header every emitted file starts with, so a fixed few seconds per run in the page are the runtime, not the user's code.
