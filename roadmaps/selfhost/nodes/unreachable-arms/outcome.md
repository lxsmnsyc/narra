# Outcome: unreachable-arms

## Built

- E0983 "this arm is unreachable", reported in `checkMatch` (`selfhost/src/match.nra`) for every plain arm the earlier unguarded arms already cover: after `_`, a repeated constructor, a repeated literal, a `_` after arms that cover every constructor. A guarded arm before the same pattern stays allowed; a guarded arm that is already covered is rejected.
- The fix "remove the unreachable arm" deletes the arm up to the next arm or the closing brace. `MatchArm` gained `next: Pos` (parser).
- `isUseful` (`selfhost/src/patterns.nra`) compares literals of infinite types, so repeated integer, string, char and float literals are found, in or-pattern alternatives too (E0974).
- Tests: `testdata/errors/unreachable_arm_{wildcard,constructor,literal}`, `testdata/diagnostics/unreachable_arm_fix` (habits fix suite), `testdata/unreachable_arm_guard` (allowed shapes, with `.out`). SPEC.md updated.

## Proof

- base: feat/batch-6, `narrac check` on `match (value) { 1 => 1, 1 => 2, _ => 3, 4 => 4 }` exits 0.
- branch: the same file exits 1 with E0983 at the second `1` and at `4`.

## Result

Not an experiment.

## Deviations

- SPEC.md edited (shared path), including removing a `_` arm from a match sample that the rule rejects.
- Arms rejected in the repo: 1 (that SPEC.md sample; `selfhost/src`, `std/src` and `testdata` had none).

## Stubs left

- None.

## Gotchas

- Literals are compared by source text, so `16` and `0x10` count as different; that only misses a report.
- Or-patterns nested in a constructor are checked as part of the whole arm, not per alternative.
- The seed is not regenerated.

## Proposed learnings

- None.

## Verification

- `NARRAC=<built> sh scripts/test.sh`: see the worker report.

## Size

- Estimated S, actual S.
