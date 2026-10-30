---
id: selfhost-check-builtins
title: Self-hosted checker: generic builtin check kinds
status: todo
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostStdCalls'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - selfhost/src/stdcalls.nra
  - compiler/internal/narra/selfhost_stdcalls_test.go
  - compiler/testdata/typesdump/stdcalls/**
shared:
  - selfhost/src/main.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/body.nra
  - selfhost/src/exprs.nra
  - compiler/internal/narra/typesdump.go
  - compiler/internal/narra/selfhost_test.go
  - compiler/testdata/typesdump/**
  - compiler/README.md
  - selfhost/src/calls.nra
  - selfhost/src/builtins.nra
  - selfhost/src/stdlib.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-check-bodies
    kind: hard
    why: defines the body context and the dump
  - on: selfhost-check-calls
    kind: hard
    why: owns the builtin call plumbing, `argsLoose`, mutable-argument handling and the `check` dispatch whose stubs this node fills
  - on: selfhost-check-closures
    kind: hard
    why: Option and Result helpers and Test.panics take function arguments, which go through `checkArrowArg`
---

# Self-hosted checker: generic builtin check kinds

## Goal

Port to Narra the type checks of the generic standard library functions, the ones whose `StdFunc.check` names a Go function. Dispatch on that name, as the registry outcome describes. Compare through the `narra types` dump.

## Context

- Read the selfhost-check-imports outcome ("Registry layout", "How bodies resolve names and std calls") and the selfhost-check-calls outcome. Go source (about 500 lines, 680 Narra): `std_array.go` (`checkArrayLength`, `checkArrayPush`, `checkArrayFilled`), `std_map.go` (`mapArg`, `checkMapSize`, `checkMapRemove`, `checkMapIter`), `std_box.go` (`checkBoxNew`), `std_range.go` (`checkRangeUpTo`, `checkRangeBetween`), `std_option.go` (`preludeEnumArg`, `preludeArg`, `checkCallback`, `callbackReturns`, `checkOptionMap`, `AndThen`, `WithDefault`, `OrElse`, `Is`, `OkOr`), `std_result.go` (`checkResultMap`, `AndThen`, `WithDefault`, `ToOption`), `std_testing.go` (`testOnly`, `checkTestCompare`, `checkTestPanics`, `checkTestGolden`), `std_prelude.go` (`argCount`).
- The 22 check names are in the imports outcome. A name without a case must fail loudly (`panic`), so a new std function in Go fails a test until it is ported; `TestSelfhostBuiltins` already fails when the table is missing the entry.
- Dump: you own the `builtin=<Builtin.Name>` attribute and the type of each builtin call. Add them to the Owners list.

## Acceptance

- [ ] `TestSelfhostStdCalls` compares `narrac types` with `narra types` on earlier inputs plus `testdata/typesdump/stdcalls/`: one case per check name with a success and an error, `Array.push` on a mutable local, `Map.remove`, `Map.keys`/`entries` in a `for`, `Box.new`, `Range.upTo`/`between`, every Option and Result helper (capturing and non-capturing callbacks, `Err(...)` with a typed arrow, wrong callback type, void callback, not an Option), `Test.eq`/`ne`/`panics`/`panicsWith`/`golden` inside and outside a `test` block (E0907).
- [ ] Skips: none for std calls; the test lists the remaining `unsupported:` skips by construct.
- [ ] Earlier harnesses pass; `narra emit` output unchanged.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 680 Narra lines, many small independent functions. Coupling medium: three dependencies.

## Notes for the worker

- Carry over from std-option-result (#31): callback arguments go through `checkCallback`, which mirrors `Array.map` (arrow functions get the parameter types as a hint, `Stack = true`, a local function-typed parameter is allowed as non-escaping); the result of `Option.map` and friends comes from the callback; `Result.andThen` takes the result type from the call's expected type when the callback returns `Err(...)`; E0909 to E0913 are the diagnostics. Prelude enum detection is by name and module path of the generic (`preludeEnumArg`); a user enum named `Option` is not accepted.
- From std-test (#24) and test-runner: `Test.eq` takes its type from the first argument and checks the second against it; `T` must implement `Equatable` and `Show` (`typeImplements`); the `panics` function argument is `() => void`; every `Test.*` call must be inside a `test` block (`Checker.inTest`, set by selfhost-check-bodies; arrow functions inside count).
- A builtin with a `mutable` parameter: check the argument expression before `checkMutableArg`.
- From prelude-imports: `Array.*`, `Map.*`, `Range.*`, `Box.*` are ordinary modules; no namespace special case.
- `stdType(mutable c, text)` creates instances; call it only where Go's check would run, in Go's order, or the instance and `order` lines differ.
- Keep `Check` closures out of enum variant fields; the dispatch is a `match` on the `check` string in one function.
- Map key checks (`checkMapKey`) queue; see selfhost-check-traits.
