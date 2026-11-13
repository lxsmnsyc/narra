# Outcome: regions-plan

## Built

- D113 in `DECISIONS.md`: "Approved by the maintainer, 2026-11-13: regions are dropped". It keeps the measurements and the reason, and records the kept design only as a rejected option. D48 now says it is superseded by D113.
- SPEC section 6 "Regions" removed. The section 11 "Output is C" line no longer mentions regions.
- `selfhost/tests/spec-skip.txt`: the Regions line removed.
- `selfhost/src/diagnostics.nra` header: `E1270-E1279  retired (regions were dropped); never reuse`.
- `nodes/regions/brief.md`: "Dropped 2026-11-13 by D113" (status unchanged; the orchestrator sets it). The region-copy-out and region-check nodes proposed in the first version were deleted.
- `nodes/regions-plan/prototype/`: D100's `variant.py` patched for the fast-alloc runtime, plus `build.sh` and `measure.sh`.

## Result

- Question 1: D100's numbers no longer hold. The compiler went from 9.27 G to 3.03 G instructions since D100. A region that keeps counts is 1.7% slower on `emit selfhost` and 4% slower on `fmt --check`, at 3 to 15 times the peak memory. Skipping counts now gives wrong output on the compiler. `formatSource` per file gains at most 1.6% (with no counts at all).
- Questions 2 to 4: answered in the first version of D113 (the kept design). With the drop, only the summary under "Rejected: building regions with counts kept" remains.
- Question 5: node regions is dropped; no children.
- reproduce: `feat/batch-13` b23bbd4f. Run `sh selfhost/bootstrap.sh $WORK/out`, then `WORK=$WORK sh roadmaps/selfhost/nodes/regions-plan/prototype/build.sh` and `WORK=$WORK sh roadmaps/selfhost/nodes/regions-plan/prototype/measure.sh 5 <repo>`. Apple M2, 8 GB, macOS 27.

## Deviations

- The brief expected a split into three build children. The maintainer approved the drop instead.

## Stubs left

- None.

## Gotchas

- The orchestrator sets `nodes/regions` to dropped and removes regions from the batch-13 child list and the GOAL.md coverage line.
- The `count` variant prints its counters only with `NARRA_CHECK_LEAKS=1`, because the compiler exits without teardown otherwise.

## Proposed learnings

- Re-measure an allocator idea against the current allocator | free lists made bump allocation no faster | selfhost/runtime | no

## Verification

- `NARRAC=<bootstrap narrac> sh scripts/test.sh --suite spec` and `--suite sources`: PASS, 61 and 1 cases. `--suite codes`: PASS.

## Size

- Estimated S, actual M. The re-measurement needed a patched prototype.
