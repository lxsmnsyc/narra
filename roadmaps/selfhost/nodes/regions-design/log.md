# Log: regions-design

- Base: feat/batch-4 at `81638c3`. narrac built with `sh selfhost/bootstrap.sh` (fixed point, seed current).
- `narrac emit selfhost` wall time swung 4.8 to 8.7 s at load 8 to 19; user time and instructions retired are the stable metrics.
- Found while reading `emitdecls.nra`: `_unique` copies when `rc != 1`, so a negative count on a region array would copy on every push. This drives the D94 dependency in D100.
- Variants built by `prototype/variant.py` from `stage2.c`; all six emit the same C as base. Counts from the `count` variant.
- First timing (load 19 to 15) and second timing (load 4.7 to 4.1) agree on instructions within 0.2%.
- `fmt --check selfhost/src` (load 4.1 to 3.4): nocount only 5% below freelist, so fmt gains little from a region.
- Sign check on `bench/` (load 2.9 to 2.7): only `interp` moves, +4.9% instructions.
- Per-function emission rejected as a region phase: `Consts.emitFunction` writes `mutable emitter` (literals, closures, counters).
