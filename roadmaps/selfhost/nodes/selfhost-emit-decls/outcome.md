# Outcome: selfhost-emit-decls

## Built

- `selfhost/src/emitdecls.nra`: port of `emit_types.go` and `emit_box.go`. `emitHeapTypedefs`, `emitTypeDefinitions` (structs and enums; array structs stay in `emit.nra`), `emitHeapDefinitions` (enum payloads, boxes), `emitBoxPrototypes`, `emitRcFunctions` (retain/release prototypes of types and arrays, `narra_drop` worklist drop functions, retain/release, array functions, box functions), and `enumAllocation(emitter, enumType, variant, fields): Code`.
- `enumAllocation` returns `owned: false`; it writes the payload allocation lines with `writeLine` and bumps `tempCount` (`q<n>`) as Go does.
- Arrays and boxes are listed from `Checker.composites`, sorted by element type id with a local merge sort (a copy of the one in `emit.nra`, which cannot be imported without a cycle).
- Rebased on main after selfhost-emit-program (PR #49): `programRecorded` is true, so the sections run for real. `TestSelfhostEmit` compares them on 63 inputs with no diff from Go; the Types and Reference counting parts agree up to the first marker of other nodes.

## Deviations

- `isBoxedEnum` is reused from `types.nra`; `emitstate.nra` is untouched.
- The skip file `compiler/testdata/emit/skips/selfhost-emit-decls.txt` is deleted (the harness now tolerates a missing file).
- The temporary program-facts guards are removed.

## Stubs left

None in this file. It depends on the real `cnames.nra` and `programRecorded` from selfhost-emit-program.

## Gotchas

- `is` with bindings works only as the whole `if` condition: use nested ifs.
- `for ... of` needs `Arrays.values([...])` even for a literal array.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostEmit'`: pass.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass (go vet clean; narra/internal/narra ok, after the rebase).

## Size

Estimated M, actual M (about 400 Narra lines).
