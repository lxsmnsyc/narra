---
id: selfhost-check-calls
title: Self-hosted checker: constructors, generic calls and trait calls
status: active
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: L
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostCalls'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - selfhost/src/calls.nra
  - selfhost/src/ctors.nra
  - selfhost/src/instances.nra
  - compiler/internal/narra/selfhost_calls_test.go
  - compiler/testdata/typesdump/calls/**
shared:
  - selfhost/src/main.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/body.nra
  - selfhost/src/exprs.nra
  - compiler/internal/narra/typesdump.go
  - compiler/internal/narra/selfhost_test.go
  - compiler/testdata/typesdump/**
  - compiler/README.md
  - selfhost/src/stmts.nra
  - selfhost/src/check.nra
  - selfhost/src/types.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-check-bodies
    kind: hard
    why: defines the body context, per-body side tables, the dispatch arms to replace and the types dump; no frozen interface exists yet
  - on: selfhost-names-frontend
    kind: hard
    why: renames the files this node edits; building first would conflict on every line it touches
  - on: selfhost-names-checker
    kind: hard
    why: renames the files this node edits; building first would conflict on every line it touches
---

# Self-hosted checker: constructors, generic calls and trait calls

## Goal

Port the call half of the body checker to Narra: enum constructors, calls of generic functions with type argument inference and per-instance body checking (the pending queue), `Trait.function(args)` calls, impl function bodies, distinct type calls, package-module calls, argument exclusivity, and the E0908 missing-import path. Compare it with Go through the `narra types` dump.

## Context

- Read the selfhost-check-bodies outcome first (side-table layout, `unsupported` convention, dispatch arms), then the selfhost-check-traits outcome ("What selfhost-check-bodies calls") and selfhost-check-imports ("How bodies resolve names and std calls").
- Go source (about 800 lines, 1,080 Narra): `check_call.go` (`ctor`, `ctorOwners`, `visibleEnums`, `isEnumName`, `containsOpen`, `ctorArgs`, `genericCall`, `funcInstance`, `argsLoose`, `builtinCall` (plumbing only), `checkExclusive`, `usesLocal`, `missingImport`), `check_traits.go` (`traitCall`, `implBodies`, the call-site `checkBounds`), `check_opaque.go` (`distinctCall`, `distinctOf`, `hiddenStruct`, `unwrapName`), `check.go` (`checkPending`), `moduleCall` and the module half of `check_const.go` (`moduleConstValue`), and the call-site use of `unify` and `substitute`.
- `builtinCall` here dispatches on `StdFunc.check`: "" and "fixed" check arguments against `stdType` of each `params` entry (Go's `checkFixed` and the plain `Params` path, with `mutables`). Any other `check` value reports `unsupported: builtin <check>`; selfhost-check-builtins ports those.
- Hint arguments: Go's `hintArrow` for arrow arguments of generic calls is selfhost-check-closures. Leave an `unsupported: arrow argument` arm in the argument loop.
- Dump: you own the line attributes `func=`, `ctor=`, `module=` on calls and identifiers, and the `body ... impl`, `targs=` headers. Add them to the doc comment's Owners list and to both printers.

## Acceptance

- [ ] `TestSelfhostCalls` (in `selfhost_calls_test.go`) compares `narrac types` with `narra types` on the inputs of `TestSelfhostBodies` plus the cases in `testdata/typesdump/calls/`: generic functions with one and several instances, nested generic calls, a generic call inside a generic body, constructors with and without expected types, a constructor name in two enums, trait calls on structs, enums, generics and type parameters with bounds, impl bodies, distinct and opaque calls, module calls across packages, and the error cases (bound not met, cannot tell type parameter, ambiguous constructor, private call, missing import with its E0908).
- [ ] A skip is only for `unsupported:` (arrows, patterns, named builtin kinds); the test prints the count and reasons.
- [ ] The order of instance bodies in the dump equals Go's (`checkPending`, discovery order), proven on a case where two generic functions call each other's instances.
- [ ] `TestSelfhostBodies` and the earlier harnesses still pass; `narra emit` output is unchanged.
- [ ] The outcome states where the pending queue lives, how an instance body is keyed, and the stubs left.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L: about 1,080 Narra lines. Uncertainty medium: instance bodies against per-body tables are new to Narra. Coupling medium: it shares the dispatch files with three parallel nodes. Verifiability low: a dump.

## Notes for the worker

- Instances: Go clones the function declaration per instance (`funcInstance`, `cloneArrow`). In Narra the AST is immutable and shared, so an instance is a pending entry `{ func index, module scope, env: Map<string, Type> }` and its results go into their own body table (selfhost-check-bodies). Key the instance cache by `typeKey` text of the type arguments, as `instances` already does for types. The Go cache key is a C name; the Narra key must still dedupe the same instances, with the same count.
- Instances of the same generic function in different modules: enter the declaring module's scope (`enter`) before checking the instance, as `checkPending` does.
- From selfhost-check-traits: `implFunc` returns None when the compiler writes the function; `c.impls` holds only blocks that passed their checks; `typeImplements` takes instances too and never reads the dump. Implementation function names are not mangled yet (`ImplFunc.name`).
- From literal-type-args: a constructor can name type arguments only through the expected type; `Pair<A, B> { ... }` is a struct literal (selfhost-check-bodies), not a call.
- From std-option-result: `Result.andThen` with `Err(...)` needs the callback's return type written; a constructor cannot infer the other type argument. The behavior is in `ctor` through the expected type; keep it.
- Namespaces for E0908: the namespace is the part of `StdFunc.name` before the dot, and its module is `StdFunc.module`. Go's `namespaceModules` is gone.
- From prelude-imports: a user type named `Box` or `Map` hides the prelude one without the import; with `import * as Box` the alias wins for `Box.new`. `Array<int>` goes through type resolution, never `modules`.
- Keep Go's order of argument checks: Go checks the argument expression before `checkMutableArg` for a `mutable` parameter.
- The dump node ids: every line you add must be printed on both sides in the same walk order.
