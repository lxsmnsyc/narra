# Outcome: mutable-escape

## Built

- E0915 "mutable parameter %q cannot be returned or stored; copy it first with `const copy = %s`", at the position of `p`. The message names the fix.
- Go: `noMutableEscape` in `compiler/internal/narra/check_expr.go`, called for struct, array, tuple and map literals there, for `return` and field/element assignment in `check_stmt.go`, and for enum constructor arguments in `ctorOf` (`check_call.go`).
- Selfhost: `checkNoMutableEscape` in `selfhost/src/exprs.nra`, called from `exprs.nra`, `stmts.nra` (return, including inferred arrow returns, and assignment) and `ctors.nra` (`checkEnumConstructor`).
- Tests: `compiler/testdata/errors/mutable_escape_{return,struct,array,field,ctor,some,tuple,map}.nra`.
- Still allowed: `const q = p`, `x = p`, `f(p)`, `p.field`. `return n` of a primitive mutable parameter is rejected too (one rule for every type).
- Harness: `TestSelfhostBodies` (`selfhost_bodies_test.go`) takes every `.nra` under `testdata/` (errors included, via `declInputs`) and compares `narrac types` errors and exit code with `TypesDump`. It skips an input only when it does not load.

## Deviations

- `compiler/internal/narra/check_call.go` and `selfhost/src/ctors.nra` are outside the brief's paths: the constructor check lives there, on the maintainer's decision to reject constructor arguments.

## Stubs left

- None.

## Gotchas

- Merge notes: E0915 was taken as the next free code. Another node adding a code at the same time may pick the same number; on conflict renumber in `diagnostic_codes.go` only.
- A constructor call with an unknown variant and a mutable parameter argument reports both errors.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...` (includes TestErrors, TestSelfhost* and the bootstrap).

## Size

- Estimated S, actual S.
