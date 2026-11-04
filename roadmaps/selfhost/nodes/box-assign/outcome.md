# Outcome: box-assign

## Built

- `b.value = v` on a `Box` works. `lvaluePtr` (`compiler/internal/narra/emit_expr.go`) calls `<box>_at(&box)` for a field of a Box; `emit_box.go` defines `<box>_unique` (copies the cell and retains the payload when the count is above 1) and `<box>_at`. Same in `selfhost/src/emitexprs.nra` (`fieldPointer`) and `selfhost/src/emitdecls.nra` (`boxFunctions`).
- `compiler/testdata/box_assign.nra` (+ `.out`): int, struct and string payloads, and shared boxes for int and string (the other holder keeps its value).
- `compiler/testdata/errors/box_assign_const.nra`: assigning to `b.value` on a `const` box. The checker already rejected it ("is constant"), so no checker change.

## Deviations

- None. Files outside the brief: none.

## Stubs left

- None.

## Gotchas

- In `selfhost/`, `containerStruct(emitter, ...)` cannot sit in the same call as `mutable emitter`; hoist it into a const.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (includes `TestSelfhostBootstrap`).

## Size

- Estimated S, actual S.
