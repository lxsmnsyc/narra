# Outcome: cell-capture

## Built

- Decision D90 in `DECISIONS.md`. `std/cell` is a Narra source module (`std/src/cell.nra`): `export opaque struct Cell<T> { slot: Box<T> }` with `create`, `get`, `set` and `update(cell, change)` where `change: (value: T) => T uses E`. Import as `import * as Cells from "std/cell"` and `import { Cell } from "std/cell"` (like `std/json` and `Value`); there is no prelude `Cell.*` spelling.
- One primitive, `CellSlot.store` of `std/internal/cell` (`selfhost/src/stdinternal.nra`: row, `checkCellStore`, `emitCellStore`): writes the box in place whatever its count and releases the old value at the call. Registered in `stdlib.nra`; the module row is `selfhost/src/stdcell.nra`; dispatch lines in `stdcalls.nra` and `emitstd.nra`; `stdtext.nra` regenerated.
- Checker: a bare `None` argument of a generic call takes its type from the earlier arguments (`hintNone` and `learnFromEarlier` in `selfhost/src/closures.nra`), so `Cells.set(cell, None)` works.
- E0612 text (diagnostics table and `funcvalues.nra`) ends with the hint to keep the state in a `Cell` from `std/cell`. Goldens updated.
- Test runner: new `leaks` suite (`selfhost/tests/src/leaks.nra`, registered in `registry.nra`) runs `testdata/leaks/*.nra`, which must print their `.out` and exit 102 with `leak check:` on stderr.
- Tests: `testdata/cell.nra` (counter shared by two escaping closures, `update`, `Cell<string>` in a struct, `Cell<Option<string>>`), `testdata/cell_cycle_broken.nra` (clean after `Cells.set(cell, None)`), `testdata/leaks/cell_cycle.nra` (leaks), `testdata/errors/cell_{let_capture,set_type,field}.nra`.
- Docs: SPEC.md sections 2 (closures, TypeScript table), 3 (module list), 6 and 8 (new "Cell" subsection); `compiler/README.md`.

## Deviations

- E0966 unused: every other misuse is reported by an existing code.
- Edited `selfhost/tests/src/{leaks,registry}.nra` (outside the brief's paths) because the expected-leak form needed runner support, as identity-cycles asked.
- `==` on a Cell is a compile error (E0965) and a Cell is not Hashable, as the coordinator decided after the first report. E0965 is used; E0966 is not.
- The spelling is `Cells.get`, not `Cell.get` (see Built).

## Stubs left

- Identity `==` for `Cell` (needs `implement Equatable` plus `implement Hashable` and pointer primitives).
- Rejecting a `Cell` in thread functions and channels: the threads node must do it (threads are parked).

## Gotchas

- Structs are plain C values, not heap objects, so a struct alone cannot have identity; `Cell` shares through its `Box`.
- `Cells.get` copies, so an array held in a `Cell` is copied by an in-place change made inside `update`.
- The threads node should add its expected-leak samples to `testdata/leaks/`.

## Proposed learnings

- None.

## Verification

- `NARRA_CHECK_LEAKS=1 narrac run -sanitize` on the samples: clean, and exit 102 for the cycle.
- `sh scripts/test.sh`: see the final report.

## Size

- Estimated M, actual M.
