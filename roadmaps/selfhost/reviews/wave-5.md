# Wave 5

## Merged

- selfhost-ast: [#20](https://github.com/lxsmnsyc/narra/pull/20)

1 node, estimated M and finished M. 446k tokens (worker 281k, reviewer 165k), two review rounds.

## Side fixes

- [#21](https://github.com/lxsmnsyc/narra/pull/21), open: the Go parser rejected `g: () => U uses E, mutable out: T` because `paramAfterComma` did not accept the `mutable` keyword. It fails on main and passes on the branch. It is not a roadmap node, so it waits for the maintainer. The selfhost-front brief says to port the fixed version.

## Process lessons

- Fresh review again found real gaps: no golden exercised `\\` or `\u{..}`, and the format comment did not state the position rules that a second implementation must copy. Both were fixed in round 2.
- The auto mode classifier blocked the orchestrator's `gh pr merge` and the read-only commands after it ("Merge Without Review"), although `can-merge` returned `auto`. The user confirmed the merge. Until a permission rule allows it, the maintainer merges.

## Contracts

No change requests.

## Briefs changed

- selfhost-front: points to the merged dump code and its "Positions" paragraph, and tells the worker to port the PR #21 fix.
- selfhost-modules stays an outline. `plan.py refine` lists it; refine it after selfhost-front fixes the node id scheme, as wave 4 planned.

## Promoted

- `selfhost/CLAUDE.md`: slot lines in dumps, the `astdump.go` doc comment as the format authority, and `narra build selfhost` type-checking unused modules.

## Open follow-ups

- `TestASTDumpDocKinds` checks one direction only (doc kinds are in `astKinds`).
- The Narra printer is untested until selfhost-front adds the parser.

## Ready next

selfhost-front (L, opus).
