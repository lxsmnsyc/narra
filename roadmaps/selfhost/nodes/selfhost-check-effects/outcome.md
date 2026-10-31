# Outcome: selfhost-check-effects

## Built

- `selfhost/src/effects.nra` (743 lines): the effect pass, a port of `check_effects.go`. `checkEffects(mutable checker): EffectState` (signature checks, then the fixpoint over the call graph, then the reporting pass) and `dump(checker, state)`. `narrac types` runs it when the bodies have no errors (`main.nra`).
- The Go dump (`typesdump.go`) now runs `checkEffects` and ends with one `effects module="<m>" <kind> <name> <set>` line per func or impl body that is not an instance. The set is what callers see: declared for exported functions, inferred for private ones, all three for entry functions. The doc comment is updated. All `testdata/typesdump/*/*.dump` goldens gained these lines.
- `check_effects.go`: `checkEffects` returns the `*effectState`.
- Tests: `TestSelfhostEffects` (every input of `TestSelfhostBodies`, the cases in `testdata/typesdump/effects/`, and the `selfhost/` package; logs the timing), `TestTypesDumpEffects`, `TestTypesDumpEffectsCoverage`, `TestSelfhostPackageIsClean`. New cases: `basics.nra` (inferred and declared sets, recursion and mutual recursion, generic instances, effect variables, `io` callback to `Option.map`, impl function, function values in locals, Map write, test block), `err_undeclared.nra`, `err_signature.nra`.
- The last `unsupported:` stubs are gone. `TestSelfhostBodies` now fails on an `unsupported:` line: 352 compared, 30 skipped (all "does not load"). Main had 341 compared, 37 skipped.
  - Duplicate constants (`const_twice` and others, and `Alias.NAME` of a duplicate): `findConstInfo` in `exprs.nra` gives a rejected duplicate its own `Checker.consts` entry, as Go checks every `ConstDecl`.
  - E0425 constructor import: `ImportName` has a `pos` now (`ast.nra`, `parser.nra`); `scopes.nra` reports at the name (`findConstructorOwner`).

## Deviations

- Touched `ast.nra`, `parser.nra`, `scopes.nra` (not in the brief's paths), minimally, for the E0425 position. `exprs.nra` and `body.nra` are shared paths.
- The dump format gained the `effects` section; the brief left the format open.
- An implement block function appears by its source name in the "through" part of an effect error; Go shows the C name `resolveImpl` gave it. Only an exported function calling an impl function with an effect that it does not declare shows this.
- "string +" in the Go pass is ported but cannot fire (Narra rejects `+` on strings before the pass).

## Stubs left

- None in the checker. `TestSelfhostBodies` has no `unsupported:` skip.

## Gotchas

- The pass walks only open bodies (kind func or impl, no type arguments); instances share the syntax and are not walked, as in Go. Keys: a function by module and position, a local by body and local index, an arrow by node id.
- A Map write is found from the assignment target (an index whose target has a Map type); Narra records no `MapSet` flag.
- Narra string templates cannot hold a string with a backtick inside `${}`; compute it first.
- The Go `go test` cache ignores the selfhost sources; use `-count=1` after editing them.

## Proposed learnings

- Run `go test -count=1` after editing `selfhost/` | the test cache does not see those files | compiler/internal/narra | no

## Verification

- `cd compiler && go vet ./... && go test -count=1 ./...`: pass (about 130 s).
- `go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostEffects'`: pass. TestSelfhostBodies: 352 compared, 30 skipped.
- `narrac types selfhost/` (AddressSanitizer debug build, in parallel with other tests) takes about 4.0 s against 0.31 s for Go; no quadratic cost showed, so `sortStrings` in `modules.nra` is unchanged. The output equals Go's and has no error.

## For selfhost-emit

- The body checker is complete for what `narra check` runs: declarations, bodies, instances, effects. `checker.bodies` holds per-body `types`, `refs`, `locals` and `closures` by node id, and `EffectState` has the final sets (`functions` by function key).
- Body half size: about 6,400 Narra lines (exprs, stmts, body, typesdump, calls, ctors, closures, funcvalues, match, patterns, literals, stdcalls, generics, instances, effects).
- Not ported, because only the emitter or the driver needs it: C names (`CName`), diagnostic codes and fixes (`CheckError` has only file, position and message), the `Program` that Go hands to the emitter (Files, Entries, Modules map), and the link from a call to its generic instance. The Narra checker keeps instances in `Checker.instanceList` with their own bodies.

## Size

- Estimated M, actual M. The stubs and the golden churn added a little to the pass itself.
