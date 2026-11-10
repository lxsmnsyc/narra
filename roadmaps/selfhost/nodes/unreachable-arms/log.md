# Log: unreachable-arms

- Proof on feat/batch-6 (narrac built from the seed): `narrac check t.nra` with `match (value) { 1 => 1, 1 => 2, _ => 3, 4 => 4 }` exits 0, no diagnostic.
- Added `next: Pos` to `MatchArm` (start of the next arm or the closing brace) so the fix can delete an arm.
- `isUseful` now handles literals of infinite types (same literal or a wildcard covers it); before, a repeated literal was never found.
- checkMatch reports E0983 for every non-or-pattern arm that `isUseful` says is covered, guarded or not. Or-pattern arms keep E0974 per alternative.
- Rejected arms in selfhost/src, std/src, testdata: 0. The only hit was the SPEC.md match sample (`_` after a two-variant enum); the arm was removed.
