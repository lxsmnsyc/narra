# Outcome: nested-function-error

## Built

- Go: `parser.go` `stmt()` reports E0930 at a `function` keyword that starts a statement, and at the `function` of `async function`. `diagnostic_codes.go` has the E0930 entry and its `explain` text.
- Selfhost: `selfhost/src/parser.nra` `parseStmt` does the same with the same message and position.
- Tests: `testdata/errors/nested_function_{body,block,arrow,impl,test,generator,async}.nra` (also run by the selfhost harness), two `TestHabitCases` entries (no fix), E0930 in `TestHabitExplain`.
- SPEC.md: line 62 rewritten, new row in "Differences from TypeScript".

## Deviations

- None. `async function` is reported at the `function` keyword, not at `async`.

## Stubs left

- None.

## Gotchas

- `function` at module level and in implement blocks goes through `decl`/`implDecl`, not `stmt`, so it is unaffected.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: see PR.

## Size

- Estimated S, actual S.
