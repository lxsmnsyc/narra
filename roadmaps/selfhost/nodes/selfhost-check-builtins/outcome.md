# Outcome: selfhost-check-builtins

## Built

- `selfhost/src/stdcalls.nra` (about 580 lines): `checkStdCall` dispatches on `StdFunc.check` and panics on an unknown name. It ports `checkArrayLength`, `checkArrayPush`, `checkArrayFilled`, `checkBoxNew`, `checkRangeUpTo/Between`, `checkMapSize/Remove/Iter`, the Option checks (map, andThen, withDefault, orElse, is, okOr), the Result checks (map/mapErr, andThen, withDefault, ok/err), and `checkTestCompare/Panics/Golden` with `testOnly`. Helpers: `checkArgumentCount` (Go's `argCount`), `findPreludeEnum` (`preludeEnumArg`), `checkPreludeArgument` (`preludeArg`), `checkCallback` (uses `checkArrowArgument`), `checkCallbackReturns`.
- Wiring: `Walker.checkStdCall` (check.nra, `BodyWalker` in body.nra); `checkBuiltinCall` in exprs.nra takes `want` and calls it for a named check; `checkFixedCall` exported; calls.nra passes `want`.
- Tests: `compiler/internal/narra/selfhost_stdcalls_test.go` (`TestTypesDumpStdCalls`, `...Coverage`, `TestSelfhostStdCalls`) and goldens `compiler/testdata/typesdump/stdcalls/` (arrays_maps, options, testing, err_arrays_maps, err_options, err_testing).
- `testdata/closures.nra` and `closures_nested.nra` are compared and equal to Go, including `stack` and `captures`.

### Compared and skipped

TestSelfhostBodies: compared 341, skipped 37 (main: 271 and 101). TestSelfhostStdCalls: compared 366, skipped 37 (it also runs the stdcalls cases and the closures inputs; the same skips).

| count | reason | node |
| --- | --- | --- |
| 30 | does not load | (TestSelfhostModules) |
| 6 | a constant declared twice | (bodies stub) |
| 1 | an import of a constructor (E0425) | calls |

No builtin skips remain; the test fails if one appears.

## Deviations

- Beyond the owned and shared paths: `selfhost/src/check.nra` (one Walker function), `selfhost/src/calls.nra` (pass `want`), `compiler/internal/narra/typesdump.go` (doc comment only).
- `typesdump.nra` targs line: now matches Go's quirk of dropping the space after every comma, tuple types included (`#[int,char]`). Found when a tuple-instance input stopped skipping. Go is unchanged.
- No change requests.

## Stubs left

- Unchanged from earlier nodes: a constant declared twice, E0425 (import of a constructor).

## Gotchas

- `f(mutable checker, ... typeString(checker, x))` is rejected; build the message into a local first.
- Prelude enum detection compares the instance's `generic` index with `scopes[0].generics[name]`, so a user enum named `Option` is rejected.
- `checkCallback` counts `checker.errs` before and after the arrow check; errors are deduped, as in Go.

## Proposed learnings

- Reach a checker part that needs closures through a Walker function | modules cannot import in a cycle | selfhost/ | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (internal/narra 113 s).
- The brief's verify command (`TestTypesDump|TestSelfhostBodies|TestSelfhostStdCalls`): pass.
- `narra emit` is unchanged: the only Go change is a doc comment.

## Size

- Estimated M, actual M.
