# Outcome: bench-zig-odin

## Built

- `bench/<program>/<program>.zig` and `.odin` for `fib`, `sieve`, `matmul`, `strings`, `interp`, `particles` and `dot`.
- `bench/run.sh` has `zig` and `odin` columns. A missing toolchain is skipped and shown as `-`. Zig builds with `zig build-exe -O ReleaseFast`, Odin with `odin build <file> -file -o:speed`.
- `bench/README.md` lists five languages and a "Build flags" section with reasons.
- `bench/BASELINE.md` has a 2026-11-20 section with Zig numbers (load 1.97 at the start, 2.65 at the end).

## Proof

Not a fix node.

## Result

Not an experiment node.

## Deviations

- Odin is not installed. The seven Odin programs are written against the documented core library and have not been compiled or run; the `odin` column in `BASELINE.md` is `-`.
- Zig `-O ReleaseFast` drops bounds checks, which Go and Rust keep. Odin keeps them by default. Documented in the README.
- The load was 1.97 at the start of the run and 2.65 at the end, as in earlier sections.

## Stubs left

- Odin numbers: run `bench/run.sh` where Odin is installed and add the column to `BASELINE.md`. `odin check` of the seven files is also pending.

## Gotchas

- Zig 0.17 has no `std.io.getStdOut`; the programs take `std.process.Init` in `main` and write through `std.Io.File.stdout().writer(init.io, &buffer)`. Older Zig will not compile them.
- Zig rejects `[_]f32{0.0} ** 8` here; `dot.zig` uses `@splat(0.0)`.
- `interp.zig` and `interp.odin` take `acc % n` without a zero check, as Go and Rust do; Narra checks in `remainder`.
- Rust's `black_box` becomes `std.mem.doNotOptimizeAway` in Zig and `intrinsics.volatile_load` in Odin.

## Proposed learnings

- None.

## Verification

- `zig build-exe -O ReleaseFast` of all seven programs; each prints the Narra answer (39088169, 1270607, 3547800000, 100000, 972838, 6996184, 800).
- `bash -n bench/run.sh`.
- `bench/run.sh -n 7`: every row `ok`, Odin skipped.

## Size

- Estimated M, actual M.
