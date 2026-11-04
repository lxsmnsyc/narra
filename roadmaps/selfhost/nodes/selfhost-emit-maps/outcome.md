# Outcome: selfhost-emit-maps

## Built

- `selfhost/src/emitmaps.nra`: map typedefs, structs, prototypes and functions (template `RUNTIME_MAP_H` with the `@...@` names replaced, sorted by "<key id>|<value id>", type-parameter maps skipped); `emitMapLiteral`, `emitMapGet` (Option via lookup), `emitMapBuiltin` (Map.size, remove, keys, entries), `emitMapSet`, `emitMapFor`. The module no longer calls the `hasWork` probe; the `selfhost-emit-maps` arm in `emitstate.nra` `typeNeeds` is now dead and left alone.
- `compiler/testdata/emit/skips/selfhost-emit-maps.txt` deleted.

## Deviations

- `emitMapFunctions` calls `useEq` before `useHash`, as Go's replacer arguments are evaluated in that order.
- After rebasing on main with selfhost-emit-traits merged, the traits skip file is gone, so no skip entries were added. `compiler/testdata/emit/skips/` no longer exists.

## Stubs left

- None.

## Gotchas

- `emitMapSet` and `emitMapFor` do not call `beginStatement`/`endStatement`; emitstmts.nra does.
- None is written as `(Option){.tag = N}` directly, not via `enumAllocation` (which wraps in extra parentheses).
- `Map.entries` tuple fields are `f_0` and `f_1`.
- With traits landed, `maps.nra` and `literal_type_args.nra` are byte-identical to Go over the whole output.

## Proposed learnings

- None.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestSelfhostEmit`: pass.
- After the rebase on main (traits merged), `narrac emit` equals Go's output byte for byte on maps.nra and literal_type_args.nra.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass.

## Size

- Estimated S, actual S (about 330 Narra lines).
