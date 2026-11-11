# Outcome: bench-check-suite

## Built

- Suite `bench` in `selfhost/tests/src/bench.nra`, registered in `listSuites()` (`selfhost/tests/src/registry.nra`). It runs `narrac check` on each `bench/<name>/<name>.nra`, one case per program (fib, interp, matmul, particles, sieve, strings), and fails a case on a nonzero exit.
- `NARRAC=<narrac> sh scripts/test.sh --suite bench` runs it.

## Proof

- A scratch edit of `bench/fib/fib.nra` (`return n` to `return "n"`), reverted and not committed, failed the suite: `FAIL bench/fib ... fib.nra:5:12: expected int, found string`, `test.sh: FAIL`.
- On the unedited tree all 6 cases pass. No benchmark program needed a fix.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- A new directory under `bench/` without `<name>/<name>.nra` is skipped silently.

## Proposed learnings

- Add benchmark programs as `bench/<name>/<name>.nra` | the bench suite finds them by folder | bench | yes

## Verification

- `NARRAC=<stage 2> sh scripts/test.sh --suite bench`: 6 passed.
- `--suite sources` and `sh scripts/test.sh`: see log.md.

## Size

- Estimated S, actual S.

Landed on feat/batch-7 as e4dc5a17.
