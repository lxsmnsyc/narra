# Outcome: effect-error-impl-name

## Built

- `compiler/internal/narra/check_effects.go`: `userCall` names an implement block function `<Type>.<Trait>.<function>` in the "through" part of an effect error (for example ``(through `Cat.Named.name`)``), not its C name. `narra check`, `narra check --json` and `narra types` share this path.
- `selfhost/src/effects.nra`: `Callee.label`, set by `findCallee` (`describeImplementation` for implement functions, the source name otherwise) and used by `walkUserCall`. The "Differences from Go" header note is removed.
- Tests: `testdata/typesdump/effects/err_impl_through.{nra,dump}` (compared by `TestTypesDumpEffects` and `TestSelfhostEffects`) and `testdata/errors/effects_impl_through.nra`.

## Deviations

- The text has no module prefix (`Cat.Named.name`), as the maintainer confirmed. The types dump `func=` form keeps its prefix (`basics.Cat.Named.name`), so the two differ in the module part.

## Stubs left

- None.

## Gotchas

- Type names come from `Type.String()` (Go) and `typeString` (Narra); a generic implementing type prints with its arguments.

## Proposed learnings

- None.

## Verification

- Fails on main (sources reverted, new goldens kept): `TestTypesDumpEffects/err_impl_through.nra` got ``(through `Cat_Named_name_71b62159`)``; `TestSelfhostEffects/.../err_impl_through.nra` got ``(through `name`)``.
- Passes on the branch: `go test -count=1 ./internal/narra -run 'TestTypesDump|TestSelfhostEffects|TestCheck|TestErrors|TestSelfhostBodies'` ok.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass (go test about 185 s).
- Only error text changes; no emit code touched.

## Size

- Estimated S, actual S.
