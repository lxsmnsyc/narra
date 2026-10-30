# Wave 5

## Merged

- selfhost-ast: [#20](https://github.com/lxsmnsyc/narra/pull/20)
- selfhost-front: [#22](https://github.com/lxsmnsyc/narra/pull/22) (merged by the maintainer, size L)

2 nodes. selfhost-ast: M, 446k tokens, two review rounds. selfhost-front: L, 292k tokens, one review round with nits only. The parser is 1501 Narra lines for about 1180 Go lines (1.27x).

## Side fixes

- [#21](https://github.com/lxsmnsyc/narra/pull/21), merged by the maintainer: the Go parser rejected `g: () => U uses E, mutable out: T` because `paramAfterComma` did not accept the `mutable` keyword. It fails on main and passes on the branch. It is not a roadmap node, so it waits for the maintainer. The selfhost-front brief says to port the fixed version.

## Process lessons

- Fresh review again found real gaps: no golden exercised `\\` or `\u{..}`, and the format comment did not state the position rules that a second implementation must copy. Both were fixed in round 2.
- `gh pr merge 20` ran under an allow rule; the auto mode classifier then blocked the read-only commands after it ("Merge Without Review"), and the orchestrator wrongly reported the merge as blocked. The user moved the `gh pr merge` allow rule from the repo's `settings.local.json` to the global settings.

## Contracts

No change requests.

## Briefs changed

- selfhost-front: points to the merged dump code and its "Positions" paragraph, and tells the worker to port the PR #21 fix.
- selfhost-modules: refined to a full brief (M). A Go `narra pkg` dump, `manifest.nra` and `modules.nra`, `TestSelfhostModules`, one id sequence per package, plus two front review nits (`Fs.writeText` result ignored, stale `narra.toml` comment).
- selfhost-check-decls: notes on the id scheme (gaps, -1 dummies, size tables by `nextId`) and an id-uniqueness test to add.

## Promoted

- `selfhost/CLAUDE.md`: `ok(p)` loop rule, `Tok` prefix for clashing variants, the mutable alias rule; slot lines in dumps, the `astdump.go` doc comment as the format authority, and `narra build selfhost` type-checking unused modules.

## Open follow-ups

- `TestASTDumpDocKinds` checks one direction only (doc kinds are in `astKinds`).
- `quoteGo` in `parser.nra` escapes fewer characters than Go's `strconv.Quote` (private use, U+061C, above U+FFFF). Error text only.
- Nesting about 3000 parentheses crashes `narrac` (C stack); Go handles it. 150 levels work.

## Ready next

selfhost-modules (M).
