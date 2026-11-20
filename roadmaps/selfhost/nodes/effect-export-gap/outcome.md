# Outcome: effect-export-gap

## Built

- Not reproduced. No code changed.

## Proof

- base: `feat/batch-20` (0fd1e813), seed and stage 1 compilers. An exported function with `uses alloc` that prints reports "uses `io` (through `print`)" in every shape tried: direct print, through a local arrow, through a private helper, in a nested arrow, in a package library file under `check` and `build`, and passing a printing callback to an exported function. Shapes and results are in `log.md`.
- branch: not built, the problem did not reproduce.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- The entry function (`run` of a single file, or the `[[binary]]` entry) needs no `uses`, even when it prints. A report of an exported `uses alloc` function that prints and compiles most likely met the entry.

## Proposed learnings

- None.

## Verification

- Shapes run with the seed and stage 1 compilers, see `log.md`. No code changed, so `sh scripts/test.sh` was not run.

## Size

- Estimated S, actual S.
