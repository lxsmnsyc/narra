# Outcome: mutable-escape

## Built

- E0915 "mutable parameter %q cannot be returned or stored; copy it first with `const copy = %s`", at the position of `p`.
- Go: `noMutableEscape` in `compiler/internal/narra/check_expr.go`, called for struct, array, tuple and map literals there and for `return` and field/element assignment in `check_stmt.go`.
- Selfhost: `checkNoMutableEscape` in `selfhost/src/exprs.nra`, called from `exprs.nra` and `stmts.nra` (return, including inferred arrow returns, and assignment).
- Tests: `compiler/testdata/errors/mutable_escape_{return,struct,array,field}.nra`.

## Deviations

- None. Spec ambiguity: only a bare identifier is checked. Copies stay allowed per the spec text, so constructor calls such as `Some(p)` and passing `p` to a function are not rejected, although a constructor call stores its argument. Decision for the maintainer.

## Stubs left

- None.

## Gotchas

- Merge notes: E0915 was taken as the next free code. Another node adding a code at the same time may pick the same number; on conflict renumber in `diagnostic_codes.go` only.
- `return n` of a primitive mutable parameter (`mutable n: int`) is also rejected.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...` (includes TestErrors, TestSelfhost* and the bootstrap).

## Size

- Estimated S, actual S.
