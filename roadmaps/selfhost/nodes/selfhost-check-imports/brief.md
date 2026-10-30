---
id: selfhost-check-imports
title: Self-hosted checker: module scopes and builtin registry
status: active
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: L
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestDeclDump|TestSelfhostDecls|TestSelfhostImports'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - selfhost/src/scopes.nra
  - selfhost/src/builtins.nra
  - selfhost/src/stdlib.nra
  - compiler/internal/narra/selfhost_imports_test.go
  - compiler/testdata/declsdump/imports/**
shared:
  - compiler/internal/narra/decldump_test.go
  - compiler/internal/narra/selfhost_decls_test.go
  - compiler/internal/narra/selfhost_traits_test.go
  - selfhost/src/resolve.nra
  - compiler/testdata/declsdump/**
  - compiler/internal/narra/decldump.go
  - compiler/internal/narra/selfhost_test.go
  - compiler/internal/narra/check.go
  - compiler/internal/narra/modules_check.go
  - selfhost/src/declare.nra
  - selfhost/src/check.nra
  - selfhost/src/types.nra
  - selfhost/src/declsdump.nra
  - selfhost/src/main.nra
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: selfhost-check-decls
    kind: hard
    why: extends its Type model, checker context and declaration passes; no frozen interface exists yet
  - on: selfhost-check-traits
    kind: hard
    why: shares declare.nra, check.nra, types.nra, declsdump.nra and decldump.go, and cross-module traits (importTrait) need both; the orchestrator serialized the two on 2026-10-30 to avoid conflicts in those files
---

# Self-hosted checker: module scopes and builtin registry

## Goal

Port the cross-module half of the declaration checker to Narra: a scope per module, package and std imports, exports and visibility, qualified types, entry function lookup, and the builtin registry (the signature data of the standard library). Compare it with Go through the `module`, `import` and `builtin` lines of the `narra decls` dump, on every package under `testdata/packages/`.

## Context

- Split from selfhost-check-decls on 2026-10-30. Read the selfhost-check-decls outcome, `contracts/builtins.md` (v1, frozen) and `contracts/module-graph.md` (v3, frozen).
- Go source (about 750 lines, 1,000 Narra lines at 1.35x): `modules_check.go` (`modScope`, `enter`, `newScope`, the module loop of `CheckPackage`, `entryFunc`, `checkEntry`, `importPackageModule`, `exported`, `qualifiedType`, `genericOf`, `namesEnum`, `typeName`, `typeErrorf`, `declPos`); the import loop of `declareModule`; `useModule`; `builtins.go` (`Builtin`, `Module`, the registries, `moduleList`); the data half of `std_*.go` (module paths, function names, parameter types, `Mutable` flags, result types, effects, the `StringBuilder` type); and the declaration parts of `check_const.go` (`declareConsts`, `importConst`, `moduleHasConst`, `typeNamed`).
- The registry is data. Do not port `Emit` closures (selfhost-emit) or the `Check` closures that type-check call arguments (`checkArrayPush`, `checkMapRemove`, ...; selfhost-check-bodies). Record for each generic builtin the check kind it needs (a name or enum variant), so bodies can dispatch on it. Keep it in structs, not enum variant fields, so it can be called after narrowing (selfhost/CLAUDE.md).
- Decide how the registry reaches Narra code. Go fills it in `init()` from many files. Options: a hand-written `stdlib.nra` of table functions, or a Go-side generator that writes `stdlib.nra` from the Go registry (commit the generated file and add a test that it is current, so a new std function in Go fails until regenerated). Prefer the generator if the table is over about 300 lines of Narra. Record the choice in the outcome.
- Dump: add `module`, `import` and `builtin` line kinds in `decldump.go` and document them in the top doc comment: per module its scope names (types, generics, functions, constants, imported names, aliases), per import the resolved target or the error, and per registered std module its functions with signatures and effects, in `moduleList` order. Errors print as in the existing format.
- Harness (`selfhost_imports_test.go`): compare the whole dump (all line kinds) for every package under `testdata/packages/` and `errors/`, std import samples, and edge cases in `testdata/declsdump/imports/` (unknown module, unknown export, private name, same name in two modules, alias clash, `std` as a package name, entry function missing or wrong signature, constant and type name clash). selfhost-check-traits merges before this node: remove its skips, compare multi-module trait packages too, and port `importTrait`.
- Not here: the xxHash32 module collision check (selfhost-emit).

## Acceptance

- [ ] `TestDeclDump` goldens cover the new line kinds, and `TestSelfhostImports` passes: `narrac decls` equals `narra decls` on every input above (all line kinds, or all but the trait kinds when traits has not merged).
- [ ] `TestSelfhostDecls` still passes.
- [ ] A new std function added only in Go fails a test until the Narra registry matches (generated file or a registry-equality check).
- [ ] The outcome describes the scope layout, the registry shape and how bodies dispatch generic builtins.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L: about 1,000 Narra lines, plus a registry that reads nine std files' signatures and possibly a generator. Uncertainty medium: the registry in Narra has no Go precedent. Coupling medium: it extends the context and dump owned by selfhost-check-decls and runs beside selfhost-check-traits.

## Notes for the worker

- Types are a recursive `enum Type`; structs and enums are an index into a declaration table. A qualified type `mod.Name` resolves through the module scope to an index; never compare by name alone, since two modules can both declare `Token`.
- Before resolving another module's types, enter that module's scope (`enter`), as `funcInstance` and `buildGeneric` do in Go. In Narra, keep the current scope as an index into `Array<ModScope>` in the context.
- Register a std type name such as `StringBuilder` only in modules that import the std module that defines it (compiler/internal/narra/CLAUDE.md).
- Every `Builtin.Name` is unique; keep the duplicate check in the Narra registry.
- Ids have gaps and dummy nodes have id -1; size side tables by `Package.nextId`.
- Edit the Go checker only to expose what the dump needs or to fix a bug the comparison finds; record it in the outcome.
- From the selfhost-check-decls review (#26): `decldump.go` prints field, parameter and result types without their module, so two modules that both declare `Token` dump alike (`testdata/declsdump/two_modules.dump`). Print `module.Name` for types from a non-prelude module, in Go and Narra, before comparing cross-module packages.
- Register `std/test` (from std-test, #24) in the Narra registry, and the namespaces prelude-imports moves to `std/array`, `std/map`, `std/range`, `std/box`.
- From the selfhost-check-traits review (#32): `dumpImpl` in `declsdump.nra` returns silently on a position mismatch and drops every later impl block. It cannot happen while the dump runs only without errors, but make it fail loudly when you extend impls across modules.
- If std-option-result (#31) has merged, register `std/option` (7 functions) and `std/result` (6 functions) in the Narra registry too.
