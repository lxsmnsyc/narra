---
id: selfhost-check-decls
title: Self-hosted checker: type model and declarations
status: done
spent: 401000
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: L
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestDeclDump|TestSelfhostDecls|TestSelfhostNodeIds|TestSelfhostParser|TestSelfhostModules'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - selfhost/src/types.nra
  - selfhost/src/check.nra
  - selfhost/src/declare.nra
  - selfhost/src/resolve.nra
  - selfhost/src/generics.nra
  - selfhost/src/declsdump.nra
  - selfhost/src/idcheck.nra
  - compiler/internal/narra/decldump.go
  - compiler/internal/narra/decldump_test.go
  - compiler/internal/narra/selfhost_decls_test.go
  - compiler/testdata/declsdump/**
shared:
  - compiler/cmd/narra/main.go
  - compiler/internal/narra/selfhost_test.go
  - compiler/internal/narra/check.go
  - compiler/internal/narra/modules_check.go
  - selfhost/src/main.nra
  - selfhost/src/modules.nra
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: selfhost-front
    kind: hard
    why: checks the AST the parser produces and uses its node id scheme
  - on: selfhost-modules
    kind: hard
    why: takes a loaded `Package` (modules in load order, `nextId`) as input
---

# Self-hosted checker: type model and declarations

## Goal

Port the core of the declaration half of the checker to Narra: the `Type` model, the checker context with its side tables, declaration of structs, enums, distinct types, functions and constants, type resolution, and generic templates with instantiation, substitution, unification and the polymorphic-recursion check. Compare it with Go through `narra decls <path>` and `narrac decls <path>`, as the lexer, parser and loader are compared.

## Context

- Split from selfhost-check-decls (outline) on 2026-10-30. The declaration half is about 2,900 Go lines, 3,800-4,000 Narra lines, so it is three nodes: this one, `selfhost-check-traits` and `selfhost-check-imports`. Read `selfhost/CLAUDE.md`, `compiler/internal/narra/CLAUDE.md`, DECISIONS.md D63, and the outcomes of selfhost-spike (gaps 3, 12, 13), selfhost-front (node id scheme, error flow) and selfhost-modules (`Package`, id sequence).
- Go source for this node (about 1,050 lines): `types.go` (type structs, `sameType`, `containsTypeParam`, not the `CName` methods, which belong to selfhost-emit), `check_types.go` except `isConstant` and `expectType` (bodies), `check_opaque.go` declaration half (`DistinctType`, `underlying`, `declareDistinct`, `resolveDistinct`, `hiddenStruct`), `check.go` (`Program`, `genericType`, `declareModule` minus its import loop and trait calls, `finish`, `errorf`), `diagnostic.go` (`report` dedup; no fixes or codes), and the prelude (`preludeSrc`, parsed by the Narra parser).
- Left to the other two nodes: traits, `derives` and bounds -> `selfhost-check-traits`; module scopes across a package, imports, exports, qualified types and the builtin registry -> `selfhost-check-imports`. This node handles the modules of a package one by one with a simple scope (own declarations plus the prelude) and treats a `std/...` import as "record it, add `StringBuilder` for `std/string`". Package-module imports and import errors are `selfhost-check-imports`.
- Left to `selfhost-check-bodies`: everything that visits a statement or expression (`funcBody`, `block`, `expr`, `expectType`, constant initializer types, pending generic function instances, effects).
- Type model (D63): `enum Type` is recursive. Prims are variants. Structs and enums are `TStruct(index)` and `TEnum(index)` into declaration tables held in the checker context (`StructInfo { name, module, opaque, args: Array<Type>, fields }`, `EnumInfo { ..., variants }`), so a type compares with `==` by index and never by name alone. Generic instances get their own table entry, found through a `Map<string, int>` keyed by the same text Go uses for `instances`. `Array`, `Box`, `Map`, `Func` are variants holding child types. `Map` has no `Equatable`/`Hashable` (gap 13), so declaration tables are `Array`s and lookups by name are `Map<string, int>` fields kept in the context struct.
- Side tables are sized by `Package.nextId`. Ids have gaps and dummy nodes have id -1, so never size by node count and never index with -1. This node fills the parameter types (`paramTypes: Array<Type>` by `Param.id`) and keeps `localN` for later.
- Dump (Go side): `DeclDump(path) (string, bool)` in `decldump.go` and hidden command `narra decls <path>`. Define the format in a doc comment at the top, as `astdump.go` and `pkgdump.go` do. It runs only the declaration passes of the existing checker, so refactor `CheckPackage` in `modules_check.go` into a declare phase and the rest with no change in output (every `testdata/` output stays byte-identical). Line kinds, each owned by one node (`traits` and `imports` append theirs; keep the kind name as the first word so the harness can filter):
  - `type` (this node): one line per struct, enum or distinct type in declaration order, then its `field` / `variant` lines, then generic instances created while declaring, sorted by name. Print types with `Type.String()` text, so structs are by name and module, not by index.
  - `generic` (this node): the template name, parameters, fields or variants.
  - `func` (this node): module, name, type parameters, parameter types, result type. Bounds come from `selfhost-check-traits`.
  - `const` (this node): name and declared type (an unannotated constant prints `?`; its type comes from the body checker).
  - `order` (this node): the type list after `finish`/`orderTypes`, so the C definition order is covered.
  - `trait`, `impl`, `bound` (selfhost-check-traits); `module`, `import`, `builtin` (selfhost-check-imports).
  - Errors print as `error <file>:<line>:<col>: <message>`, in report order, then exit 1.
- Harness (`selfhost_decls_test.go`, reuse `buildSelfhost` and `runNarrac`): run both dumps and compare only this node's line kinds. Skip an input only when Go reports an error from a stage this node does not port (trait or import errors), and list the skips in the outcome. Inputs: every single-file sample and every `compiler/testdata/errors/` file, `selfhost/src/*.nra` (each as a single file is impossible, so use `selfhost/` as one package), the single-module packages under `testdata/packages/`, and edge cases in `testdata/declsdump/` (declaration errors: duplicate type, duplicate function, unknown type, bad defaults, polymorphic recursion, `Box<T>` of a growing self-reference, opaque struct field access, distinct types, nested generic instances, forward references, shadowed prelude names).
- Id walker test (deferred from the selfhost-modules review): a Narra walker in `idcheck.nra` visits every `Expr`, `Pattern`, `Param`, `VarStmt` and `ForStmt` of a parsed file and checks that ids are unique and lie in the file's `[firstId, endId)` range. Extend `narrac pkg --ids` with it (it prints `ids ok <nextId>`) and add `TestSelfhostNodeIds` over the same inputs as `TestSelfhostModuleIds`. A new AST node type must be added to the walker, so say so in its doc comment.

## Acceptance

- [x] `narra decls <path>` prints the dump, its format is defined in the doc comment of `decldump.go`, and `TestDeclDump` compares goldens in `compiler/testdata/declsdump/`. The `CheckPackage` split changes no output.
- [x] `TestSelfhostDecls`: `narrac decls` equals `narra decls` on this node's line kinds for every input listed above, with the skipped inputs named in the outcome.
- [x] `TestSelfhostNodeIds` passes and fails if two nodes get the same id (shown by a unit check that corrupts an id on purpose).
- [x] `TestSelfhostLexer`, `TestSelfhostParser`, `TestSelfhostModules` still pass.
- [x] The outcome describes the `Type` enum, the declaration tables, the checker context and which fields later nodes extend, and the side tables with their sizes.
- [x] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L: about 1,400-1,600 Narra lines (1,050 Go lines at 1.3-1.45x), plus a Go dump of about 200 lines and a `CheckPackage` split. Uncertainty medium: the recursive `Type` enum and the index-based tables are new in this repo, and Go's pointer identity (`*StructType`, `map[*TypeParamType]Type`) has to become indexes. Blast radius medium: `check.go` and `modules_check.go` are edited for the split.

## Notes for the worker

- Types are a recursive `enum Type`; structs and enums are an index into a declaration table. Do not add a pointer-like handle. Type parameters need identity (`map[*TypeParamType]Type` in Go): give each a unique index or a (owner, position) pair and use it as the `Map` key.
- Keep `Check` closures out of enum variants (selfhost/CLAUDE.md). This node has none; the registry is `selfhost-check-imports`.
- Thread one `mutable` context struct and keep each array in it unique (a second owner copies on every push). Declaration tables are written in place through `mutable`.
- The prelude is Narra source parsed by the self-hosted parser; give it module path `""` as Go does, with file name `prelude`.
- Errors go through one `errorf(mutable c, pos, message)` that records `{file, pos, message}` and drops exact duplicates (Go's `report`). Do not port error codes, fixes or JSON; the codes table (`diagnostic_codes.go`, 280 lines) and fixes go with the driver in selfhost-bootstrap. Say so in the outcome.
- Do not port `CName`. The xxHash32 module collision check (`LoadPackage`) is deferred to selfhost-emit.
- Go behavior to keep: `orderTypes` ordering, instance cache keys, `checkPolyRecursion` message text, default-value checks run after all field types are known.
- Edit Go files other than the dump only for the `CheckPackage` split or a bug the comparison finds; record it in the outcome.
