# bench

The same programs in Narra, Go and Rust, and a script that times them.

```sh
bench/run.sh            # all programs, 7 runs each
bench/run.sh -n 3 fib   # 3 runs, one program
```

`run.sh` builds each program in release mode into `bench/build/` (ignored by git), runs it, and prints the median wall-clock seconds per language. It also checks that every language prints the same answer. If `go`, `rustc` or a C compiler is missing, that language is skipped and shown as `-`.

Programs live in `bench/<name>/<name>.nra`, `.go` and `.rs`: `fib` (recursion), `sieve` and `matmul` (arrays), `strings` (string building), `interp` (enum and match), `particles` (structs in an array).

Run it on a quiet machine. Results are in `BASELINE.md`.

The Go and Rust programs are plain idiomatic code that mirror the Narra algorithm. Add a program by adding the three files and its name to the list in `run.sh`.

## Build time (not measured yet)

`run.sh` times the programs, not their builds. Add build times once the compiler is complete. Until then the comparison would mislead, for these reasons:

- `narra build` writes the whole program to one `out.c` (`compiler/internal/narra/driver.go`). There are no per-module translation units, so no parallel C compiles and no incremental rebuilds. Every build is a clean build.
- Narra has no optimization passes of its own. `-release` runs `cc -O2` and debug runs `cc -O0 -g`, so the C compiler does all the work. This affects run time (reference counts clang cannot remove), not build time.
- Debug builds through TCC (`SPEC.md`) are not wired or tested. `CC=tcc` may fail on `runtime.h`.
- The programs here are small, so a build is mostly process start and C compiler overhead.

When the compiler is ready, measure clean and incremental builds:

- Narra debug with TCC and with clang `-O0`, and Narra release with clang `-O2`.
- `go build`.
- `rustc -C opt-level=0` and `rustc -C opt-level=3`, or `cargo build` with and without `--release`.
- One larger generated program as well as the ones here, so the front end and module design show, not only startup.

Public numbers for reference (2026-09, synthetic, about 40,000 small functions, no optimization) from [nordlow/compiler-benchmark](https://github.com/nordlow/compiler-benchmark): tcc 0.12 s, clang 3.2 s, Rust 7.4 s, Go 16 s. The test is one large package, which is Go's worst case and does not show Go's package cache. Check the page before relying on these figures.
