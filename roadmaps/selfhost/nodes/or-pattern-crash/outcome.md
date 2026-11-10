# Outcome: or-pattern-crash

## Built

- `coversEveryConstructor` and a changed branch in `isUseful` (`selfhost/src/patterns.nra`). The compiler no longer overflows the stack on an or-pattern arm over a recursive enum.
- Test `testdata/or_pattern_nested_bind` (the repro, the same with a guard, two names bound at different depths).

## Proof

- Command: `out/narrac build -o r repro.nra`.
- feat/batch-4 (e64faa2): exit 139. A sanitized build of the compiler shows unbounded recursion in `isUseful`.
- This branch: exit 0; the test prints the expected values.

## Result

Not an experiment.

## Deviations

- None. The cause was the usefulness check, not the emit code, so `emitmatch.nra` is untouched.

## Stubs left

- None.

## Gotchas

- Usefulness with a wildcard vector must not enumerate constructors that no row uses; on a recursive type that never ends.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh` run once before finishing; see the worker report.

## Size

- Estimated S, actual S.
