# Outcome: nested-or-guards

## Built

- `expandOrPatterns` and `containsOrPattern` in `selfhost/src/emitmatch.nra`. A guarded arm with bindings and an or-pattern at any depth tries every alternative combination until the guard holds.
- Test `testdata/or_patterns_nested_guard` (depth 1 and 2, two lists in one pattern, with and without bindings, a string guard).

## Proof

- Command: `NARRAC=<narrac> sh scripts/test.sh --filter or_patterns_nested_guard`.
- feat/batch-4 (81638c3): fails, line 1 prints `small 1` instead of `big 10`.
- This branch: passes.

## Result

Not an experiment.

## Deviations

- The inbox entry is not removed (batch rule: the orchestrator does it).

## Stubs left

- None.

## Gotchas

- The expansion is a product: an arm with k nested or-lists of n alternatives emits n^k tests.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh` run once before finishing; see the worker report.

## Size

- Estimated S, actual S.
