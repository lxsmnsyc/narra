# Outcome: decls-quadratic

## Built

- Not reproduced. `narrac decls` and `narrac check` grow linearly with the declarations of one module on the self-hosted compiler. No code changed.

## Proof

- base: dca242d (origin/feat/batch-3), `narrac decls <pkg>` and `narrac check <pkg>` on generated single-module packages, 5 runs each, two declaration mixes (see log.md). Shape B (all exported: generic structs with derives, enums, functions, consts) `decls`: 2000 declarations 0.06-0.07 s, 4000 0.11-0.14 s, 8000 0.24-0.26 s. `check`: 0.05, 0.10-0.11, 0.23-0.24 s. Shape A is faster and also linear (8000: 0.10-0.13 s). Each doubling costs 2.0x to 2.3x; a quadratic step would cost 4x. The spread between runs is under 0.02 s.
- branch: not built, the problem did not reproduce.

## Deviations

- None. No test added, because there is no quadratic step to guard.

## Stubs left

- None.

## Gotchas

- The old figures (0.3 s, 1.15 s, 4.9 s) came from the Go-era compiler, which no longer exists. The quadratic step probably died with it.
- Shapes not covered: many functions that call each other, many `implement` blocks, deep generic nesting inside one declaration. A new report with such a module would need its own generator.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh <out>`: fixed point holds, seed current.
- Timings as above. `sh scripts/test.sh` not run, no code changed.

## Size

- Estimated M, actual S. The first measurement showed linear growth.
