# Log: inline-benchmark

- Branch renamed to `spike/inline-benchmark`, reset to `feat/batch-15` (aa6720f2).
- Counted: `grep -rn '@inline' selfhost/src std/src` gives 3 real sites, all in `std/src/option.nra` (the rest are messages and comments). Nothing in `selfhost/src` or `bench/` calls `Options.map`, `andThen` or `orElse`.
- Measurement patch: `isInlineFunction` in `selfhost/src/inlining.nra` returns `false` (parameters renamed with `_`, the checker rejects unused ones). Built with `sh selfhost/bootstrap.sh`, as `with/` and `without/` in the scratch folder. Patch reverted with `git checkout` before the commit.
- Added `bench/inline/inline.nra` because no bench program uses an `@inline` function.
- `measure.sh`: 7 interleaved runs per program and variant, `/usr/bin/time -l` for instructions retired and cycles elapsed; 5 runs for `emit selfhost`. It waits while the 1-minute load is above 4 (it never was; 2.3 to 2.5).
- Emitted C identical for the seven older programs; differs only for `inline`.
- Result: no gain anywhere, all inside the spread.
