---
id: selfhost-check-traits
title: Self-hosted checker: traits and implementations
status: active
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestDeclDump|TestSelfhostDecls|TestSelfhostTraits'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - selfhost/src/traits.nra
  - compiler/internal/narra/selfhost_traits_test.go
  - compiler/testdata/declsdump/traits/**
shared:
  - compiler/testdata/declsdump/**
  - compiler/internal/narra/decldump.go
  - compiler/internal/narra/selfhost_test.go
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
  - on: prelude-imports
    kind: hard
    why: the user moved prelude function namespaces behind std imports on 2026-10-30; port the checker against the new prelude
---

# Self-hosted checker: traits and implementations

## Goal

Port the trait half of the declaration checker to Narra: trait declarations and the prelude traits, `derives`, `implement` blocks and their signature checks, function bounds, `implements(type, trait)` and the Map key rules for signatures. Compare it with Go through the `trait`, `impl` and `bound` lines of the `narra decls` dump.

## Context

- Split from selfhost-check-decls on 2026-10-30. Read the selfhost-check-decls outcome (Type model, context, dump line kinds) and `contracts/trait-model.md` (v3, frozen).
- Go source (about 620 lines of `check_traits.go` plus about 60 of `std_map.go`, 840 Narra lines at 1.35x): `Trait`, `TraitFunc`, `typeTraits`, `traitState`, `declareTraits`, `resolveTraits`, `resolveTraitDecl`, `resolveImpl`, `sameSignature`, `signatureText`, `typeExprString`, `declaredName`, `fieldsOf`, `implements`, `implFunc`, `funcBounds`, `checkBounds`, `checkEquatable`, `containsFunc`, and `mapKeyProblem`/`checkMapKey` (queued; flush after all modules are declared, as `flushMapKeys` does).
- Left to `selfhost-check-bodies`: `traitCall` (calls on trait functions), `implBodies` (impl function bodies), and the call-site use of `checkBounds`. Expose them as functions that bodies can call.
- Dump: the Go side already prints `type`, `generic`, `func`. This node adds `trait`, `impl` and `bound` line kinds in `decldump.go` and adds the `bounds` to `func` lines; document them in the doc comment at the top of the file. The harness compares those kinds plus the whole dump for inputs that need no import resolution.
- Inputs: every single-file sample that declares or implements a trait or uses `derives`, `testdata/errors/` trait and bound errors (missing function, wrong signature, duplicate implementation, orphan implementation, unknown trait, derive on a type that cannot derive, `Equatable` on a function type, Map key errors in signatures), single-module packages, and edge cases in `testdata/declsdump/traits/`.
- The prelude declares `Equatable`, `Hashable`, `Ordered`, `Show` and derives them for `Ordering`, `Option` and `Result`. Those lines must match too.
- Cross-module traits (`importTrait`, `ownsType` across modules) need module scopes from `selfhost-check-imports`. Port the single-module logic here, keep the `module` field on every trait and implementation key, and leave `importTrait` to selfhost-check-imports. selfhost-check-imports runs after this node and adds the combined comparison on multi-module packages; skip those packages here and list the skips in the outcome.

## Acceptance

- [ ] `TestDeclDump` goldens cover the new line kinds, and `TestSelfhostTraits` passes: `narrac decls` equals `narra decls` on the trait lines and errors for every input above, skips named in the outcome.
- [ ] `TestSelfhostDecls` still passes with the extra `bound` data on `func` lines.
- [ ] The outcome lists the `Trait`/`Impl` table layout and the functions selfhost-check-bodies calls.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 850 Narra lines, one Go dump extension, one harness. Coupling medium: it extends the context, dump and declaration driver owned by selfhost-check-decls and runs beside selfhost-check-imports.

## Notes for the worker

- Types are a recursive `enum Type`; structs and enums are an index into a declaration table. Trait lookup is keyed by (module, type name) as in Go (`typeKey`); keep it a `Map<string, int>` with a joined key, since `Map` is not hashable (gap 13).
- Compare struct and enum types by module and name through the table index, never by name alone.
- Output order must match Go: `Map` iteration is insertion order in Narra but Go map order is random, so find where Go sorts and sort the same way; where Go relies on declaration order, use it.
- Edit `check_traits.go` only to expose what the dump needs or to fix a bug the comparison finds; record it in the outcome.
- From the selfhost-check-decls review (#26): `errorIn` in `selfhost/src/check.nra` rescans every earlier error on each report (quadratic with many errors); Go uses a map. Switch to a `Map<string, bool>` keyed by file, line, column and message.
