# Outcome: missing-return

## Built

- `selfhost/src/pathends.nra`: `findBlockOpenPath` (the first path that reaches the closing brace, as text), `checkEveryPathReturns` (reports E0969 for a function, E0970 for an arrow function with a block body, at the closing brace) and `isPanicCall` (moved from `stmts.nra`).
- `Block.end` in `ast.nra`: the position of the closing brace, set by `parseBlock`.
- Hooks: `checkFunctionBody` in `body.nra` and `checkArrow` in `closures.nra`.
- The emitted check in `emitconsts.nra` and `emitclosures.nra` now reads "internal compiler error: function ended without returning a value".
- Tests: `testdata/errors/missing_return_{if,match,loop_break,arrow}` and the sample `testdata/missing_return_endings` (if/else chain, match, match with panic arm, `while (true)`, `for (;;)`, inner `break`, labeled `break`, final `panic`, block arrow).
- SPEC.md (Functions) states the rule; D92 says the part is built.

## Proof

- base: dca242d, `narrac check missing.nra` (the sample from the brief) exits 0; `narrac run` panics "function ended without returning a value", exit 101.
- branch: the same `check` exits 1 with E0969 at the closing brace. See `log.md`.

## Deviations

- Compiler sites rejected: 0 in `selfhost/src`, `selfhost/tests/src`, `std/src` and the samples, so nothing needed fixing.
- Only E0969 and E0970 are used; the match case uses E0969 with its own reason text.
- Generators are exempt: `return` in a generator takes no value, so they have no result to miss.
- An exhaustive `match` is accepted when every arm ends; exhaustiveness itself stays with the existing match check.

## Stubs left

- None.

## Gotchas

- The rule is syntactic. A `panic` is recognised by the name `panic` only; a call to a function that always panics does not end a path.
- Statement order is generous: one ending statement ends the block even if code follows it.
- `Block` has a new required field `end`; any new construction site must set it.
- The seed does not have the rule yet. The bootstrap reaches a stage 3 until the seed is regenerated.

## Proposed learnings

- A function with a result must end every path | the compiler rejects the rest (E0969) | selfhost/src | yes

## Verification

- `narrac check selfhost`, `narrac check selfhost/tests`: exit 0.
- `sh selfhost/bootstrap.sh`: fixed point holds at stage 3 (the seed is older than the source).
- `sh scripts/test.sh --filter missing_return`: pass.
- `sh scripts/test.sh`: see the final report.

## Size

- Estimated M, actual S: no existing site needed a fix.
