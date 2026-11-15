# Log: narrac-wasm

- The committed seed does not build for wasm: `seed/narrac.c` still has the old `std_process.c` block (`sigaction`, `SIG_IGN`, `pthread_sigmask` undeclared). wasi-runtime fixed the source, not the seed.
- Built a current stage with `sh selfhost/bootstrap.sh <dir>` (stage 3 ran, since the seed is older; stage 2 and 3 C identical, 249,725 lines). `scripts/wasm-build.sh <dir>/stage2.c` builds it for wasm. The script takes the C file as an argument and defaults to the seed.
- `llvm-strip` and `wasm-tools` are not installed. `-Wl,--strip-all` at link time strips the names and gives a 6.3 MB module, the same as the spike's llvm-strip result.
- `emit selfhost` under wasmtime first failed: `could not determine the host: no supported target for this C compiler`. The Wasi row has no host OS or arch, so a wasm narrac cannot tell what host a program that imports `std/build` is for. `--env NARRA_HOST_TARGET=MacOS-Arm64` names it, and the C then matches native.
- `wasm-dis -all` on the module shows no `try`, `try_table`, `throw`, `catch` or `tag`: the compiler never reaches `setjmp`. So it validates and runs with `-W exceptions=n`.
- Load was 13 to 47 during the work; the bootstrap took 8m43s and the wasm build 2m19s.
