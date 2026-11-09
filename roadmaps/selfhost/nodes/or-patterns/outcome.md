# Outcome: or-patterns

## Built

- `Pattern.PatOr` in `selfhost/src/ast.nra`, with cases in the parser (`parsePattern`), checker (`match.nra`, `patterns.nra`), emitter (`emitmatch.nra`), `fmtexprs.nra`, `astdump.nra`, `typesdump.nra`, `idcheck.nra`, `querysites.nra` and `trivia.nra`.
- Codes: E0967 (alternatives bind different names), E0968 (different types), E0974 (alternatives in `is`, and an unreachable alternative).
- Tests: `testdata/or_pattern_basic`, `testdata/or_pattern_strings` (with `.out`), `testdata/errors/or_pattern_{names,types,is,unreachable}`, `testdata/fmt/or_pattern_layout`, `testdata/astdump/or_pattern`.
- Docs: SPEC.md, DECISIONS.md D98 (now "Built"), compiler/README.md, docs/guide/05-pattern-matching.md.

## Proof

Not a fix node.

## Result

Not an experiment.

## Deviations

- The repo had no unreachable-arm check. `isUseful` is new and runs only on or-pattern arms, so an unreachable plain arm is still accepted.
- The unreachable message shares E0974 with the `is` error because only three codes were reserved.
- fmt long layout keeps the first alternative on the arm line; later ones start with `|`.
- Contract patterns (D95) do not exist yet, so nothing was done for them.

## Stubs left

- None.

## Gotchas

- Later alternatives create locals that nothing reads (marked used). Tools that list every local of a body see them.
- The seed is not regenerated; `adopt-new-syntax` needs the orchestrator to run `sh scripts/seed-update.sh` first.

## Proposed learnings

- A new pattern kind needs cases in `collectPatternLocals`, `bindRest`, `collectConditions`, `bindPattern`, `isIrrefutable` and the exhaustiveness expansion, besides the dump/fmt switches | the `_ => {}` arms hide a missing case | selfhost/src | no

## Verification

- `sh scripts/test.sh` (full run): see the report of the worker.

## Size

- Estimated M, actual M.
