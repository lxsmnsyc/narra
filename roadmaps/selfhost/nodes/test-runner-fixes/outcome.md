# Outcome: test-runner-fixes

## Built

- `selfhost/tests/src/main.nra`: `describeEmptyPlan`; a run whose suites and filters match no case exits 1 with `no test case matches: --suite X --filter Y`.
- `scripts/test.sh`: the final line carries the number of cases that ran, for example `test.sh: PASS, 6 cases in 0s` (a `--list` run keeps `PASS in Ns`).
- `selfhost/tests/src/runnertests.nra`: case `runner/no-match-filters` (builds the runner, checks exit 1 and that stderr names the flags).
- `testdata/io_files.nra`: the folder comes from `mktemp -d` under `$TMPDIR`, unique per run.
- `selfhost/tests/src/textops.nra`: dead returns after `panic` removed. `std/src/json.nra`: `cutText` lost its function comment and panics with "cutText: the range is outside the text" on `None`. `selfhost/src/stdtext.nra` regenerated.

## Proof

- base (f62634c1): the 0-case PASS did not reproduce as a total. `--suite bench|errors|packages|emit` ran their cases. Only per-shard `0 passed` lines show (a suite with fewer cases than shards), and the empty-plan message did not name the filters, so the likely cause of the batch-7 and batch-8 sightings is a misreading: shards past the case count print `shard 8/8: 0 passed, 0 failed` right before `test.sh: PASS in 0s`, and agents reading the run through `| tail -2` saw exactly that (see reviews/wave-20.md). The case total on the final line is the fix. An empty plan already failed every shard. io_files: the old sample run 6 at once differed from `io_files.out` in 2 of 6, 1 of 6 and 0 of 6 over 3 rounds.
- branch: `runner/no-match-filters` passes (it fails on base, where the message lacks the flags); `--filter zzz` prints the flags and exits 1; 3 rounds of 6 concurrent io_files runs all match the golden; two concurrent `scripts/test.sh --filter io_files` both pass.

## Deviations

- Touched `selfhost/src/stdtext.nra` (regenerated from `std/src/json.nra`, as the brief said). Item 1 as stated (a total 0-case PASS) did not reproduce; the fix is the message plus the case total on the final line.

## Stubs left

- None.

## Gotchas

- A targeted `--suite` run prints `0 passed` for shards without a case; that is normal when the suite has fewer cases than shards. Read the case total on the final line (`PASS, N cases`), not the last shard line.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh` on the branch: PASS (see the report).

## Size

- Estimated M, actual S.
