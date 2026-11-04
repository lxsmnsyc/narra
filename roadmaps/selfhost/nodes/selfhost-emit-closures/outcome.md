# Outcome: selfhost-emit-closures

## Built

- `selfhost/src/emitclosures.nra` replaces the stubs with the port of `emit_closure.go`:
  - `emitFuncTypeDefinitions`: envRuntime, the typedefs and the structs of every TFunc in `Checker.composites`, sorted by type id.
  - `emitClosureFunctions`: the `// Closures` section (environment structs and drop functions, prototypes, definitions). Written only when a thunk or an arrow made a prototype, as in Go.
  - `emitFuncValue`: thunk `w_...` for a named function used as a value (plain, or generic through `instanceRefs`).
  - `emitArrow`: arrow functions with no captures, stack environments (`.h = {-1, NULL}`) or heap environments with retained captures and a drop function. Counted per `Emitter.owner` from 1. The nested body runs with `out`, `indent`, `frames`, `scopes`, `loopFrames` and `resultType` saved and restored.
  - `emitValueCall`: calls through a function value, holding a reference to the callee until the statement ends.
  - `callFunctionValue(emitter, functionType, callee, arguments): Code` (Go's `callFn`): `callee.fn(callee.env, ...)`, a plain temporary for a non-void result.
- Deleted `compiler/testdata/emit/skips/selfhost-emit-closures.txt` (13 inputs and `packages/closures`). No other skip file listed only this node, so none else changed.

## Deviations

- None. Expression-bodied arrows are wrapped into a block in the emitter because the Narra checker keeps the AST as written.

## Stubs left

- None in this node. `closures.nra` and `closures_nested.nra` still diverge from Go only through the std and match stubs.

## Gotchas

- `callFunctionValue` returns an unowned temporary, like Go's `callFn`; the caller moves it into a payload or result.
- Narra rejects a `mutable emitter` call whose arguments read `emitter` (even in a template): bind the value first.
- Arrow parameters are locals of the enclosing body, found through `BodyInfo.declared`.

## Proposed learnings

- None.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestSelfhostEmit`: pass.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass.
- Whole-output diff against Go: `effects.nra` and `param_after_uses.nra` are byte-identical; the byte-identical count in the harness is 6.

## Size

- Estimated M, actual M (about 330 Narra lines).
