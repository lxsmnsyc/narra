# Outcome: selfhost-emit-decls

## Built

- `selfhost/src/emitdecls.nra`: port of `emit_types.go` and `emit_box.go`. `emitHeapTypedefs`, `emitTypeDefinitions` (structs and enums; array structs stay in `emit.nra`), `emitHeapDefinitions` (enum payloads, boxes), `emitBoxPrototypes`, `emitRcFunctions` (retain/release prototypes of types and arrays, `narra_drop` worklist drop functions, retain/release, array functions, box functions), and `enumAllocation(emitter, enumType, variant, fields): Code`.
- `enumAllocation` returns `owned: false`; it writes the payload allocation lines with `writeLine` and bumps `tempCount` (`q<n>`) as Go does.
- Arrays and boxes are listed from `Checker.composites`, sorted by element type id with a local merge sort (a copy of the one in `emit.nra`, which cannot be imported without a cycle).
- Each section returns the `program facts (selfhost-emit-program)` marker while `Checker.programRecorded` is false, because the names come from the placeholder `cnames.nra`.

## Deviations

- `isBoxedEnum` is reused from `types.nra` (already exported); `emitstate.nra` is untouched. 
- The skip file `compiler/testdata/emit/skips/selfhost-emit-decls.txt` is not deleted: `TestSelfhostEmit` opens a file for every node and fails if it is missing. It keeps only its comment line (empty list). The orchestrator may delete it if the harness is changed to skip missing files.
- Acceptance 1 ("match Go on every input") cannot be run until selfhost-emit-program sets `programRecorded`, fills `composites` and ships the real `typeId`, `cType`, `payloadName` and `opFuncName`. Today the code is checked by type-checking and by a smoke run with the guard disabled and stub names (output shape matches Go). After emit-program merges, run `TestSelfhostEmit` and fix any diff in the Types and Reference counting parts.

## Stubs left

None in this file. It depends on the real `cnames.nra` and `programRecorded` from selfhost-emit-program.

## Gotchas

- `is` with bindings works only as the whole `if` condition: use nested ifs.
- `for ... of` needs `Arrays.values([...])` even for a literal array.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostEmit'`: pass.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass (go vet clean; narra/internal/narra ok).

## Size

Estimated M, actual M (about 400 Narra lines).
