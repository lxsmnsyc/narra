---
id: selfhost-check-bodies
title: Self-hosted checker: body core and the types dump
status: active
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: L
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostDecls|TestSelfhostTraits|TestSelfhostImports'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - selfhost/src/body.nra
  - selfhost/src/exprs.nra
  - selfhost/src/stmts.nra
  - selfhost/src/typesdump.nra
  - compiler/internal/narra/typesdump.go
  - compiler/internal/narra/typesdump_test.go
  - compiler/internal/narra/selfhost_bodies_test.go
  - compiler/testdata/typesdump/bodies/**
shared:
  - selfhost/src/check.nra
  - selfhost/src/declare.nra
  - selfhost/src/main.nra
  - selfhost/src/types.nra
  - compiler/internal/narra/check.go
  - compiler/internal/narra/modules_check.go
  - compiler/cmd/narra/main.go
  - compiler/internal/narra/selfhost_test.go
  - compiler/README.md
  - selfhost/src/scopes.nra
  - compiler/internal/narra/selfhost_imports_test.go
reserves: [decision D70]
provides: []
depends_on:
  - on: selfhost-check-decls
    kind: hard
    why: extends its Type model, Checker context, side tables and declaration driver; no frozen interface exists
  - on: selfhost-check-traits
    kind: hard
    why: bodies use typeImplements, checkEquatable, checkBounds, trait tables and impls
  - on: selfhost-check-imports
    kind: hard
    why: names resolve through module scopes, and std calls use the registry
  - on: qualified-constructors
    kind: hard
    why: the migration to qualified constructors and plural aliases (D71, D72) rewrites selfhost/src; the body checker ports the new constructor lookup instead of the old one
---

# Self-hosted checker: body core and the types dump

## Goal

Port the core of the body checker to Narra: locals and scopes, the expression and statement switches, operators, literals, field and index access, struct literals, plain calls of non-generic functions, top-level constants and the `test` and function body drivers. Define the `narra types` dump once (Go side `typesdump.go`, Narra side `typesdump.nra`) so that every later body node compares itself with Go through it.

## Context

- Split from the outline on 2026-10-30 into six nodes: this one (core and dump), selfhost-check-calls, selfhost-check-patterns, selfhost-check-closures, selfhost-check-builtins and selfhost-check-effects. After this node the other four except effects can run in parallel; effects needs all of them. Read the outcomes of selfhost-check-decls (Checker, side tables, HOOK comments, driver), selfhost-check-traits ("What selfhost-check-bodies calls") and selfhost-check-imports ("How bodies resolve names and std calls").
- Go source (about 920 lines, 1,250 Narra at 1.35x, plus about 100 for the dump printer): `check.go` (`checkBodies`, `lookup`, `declare`, `push`, `pop`, `Local`, `scope`), `check_expr.go` (`expr`, `exprInner`, `binary`, `structLit`), `check_stmt.go` (`funcBody`, `block`, `diverges`, `stmt`, `lvalueRoot`, `isResult`, `inGenericTemplate`), the plain path of `check_call.go` (`call` for a non-generic user function, `args`, `checkMutableArg`, `ctorArgs` is calls'), `check_const.go` (`constType`, `checkConstValue`, `constWrite`, `moduleConstValue`, `calleeName`), `checkDefaults` in `check_types.go`, `isConstant`, `expectType`, and the `checkBodies` loop of `modules_check.go`.
- Not here, with the node that ports it: enum constructors, generic calls and instances (`pending`), `funcInstance`, trait calls, impl bodies, distinct calls, exclusivity rules (`checkExclusive`, `usesLocal`), package-module calls and the E0908 import fix: selfhost-check-calls. `is`, `match`, destructuring, narrowing, map, tuple and spread literals: selfhost-check-patterns. Arrow functions, function values, `valueCall`: selfhost-check-closures. The named builtin check kinds (`StdFunc.check`): selfhost-check-builtins. The effect pass: selfhost-check-effects.
- Driver: replace `declarePackage` with Go's interleaving: for each module, `declareSourceModule(m)`, then m's bodies; then the pending queue, `flushMapKeys` and `finish`. `narrac decls` keeps working (declarations only), `narrac types` runs the whole thing.

## The types dump (frozen by this node)

`narra types <path>` and `narrac types <path>`. The format lives in the doc comment of `typesdump.go`, in the style of `decldump.go`. Plan:

- The dump runs what `narra check` runs except the effect pass (selfhost-check-effects adds it): the declaration passes, every body, the pending generic instances, `flushMapKeys`, `finish`. If there is any error, print only the errors (`file:line:col: message`, in report order after dedup) on stderr and exit 1, as `decldump.go` does. Otherwise print on stdout.
- One block per checked body, in checking order: module order, then declaration order, with the generic instances queued by `funcInstance` after the module's bodies in discovery order (Go's `checkPending`). Header: `body module="<m>" func <name>`, `body ... test "<name>"`, `body ... impl <Trait> <Type>.<fn>`, `body ... const <name>`, `body ... default <Type>.<field>`, and `targs=<type>,<type>` for an instance.
- Inside a block, one line per expression, pattern and local, in pre-order with children in source order (the order `astdump.go` and `idcheck.nra` walk): `<line>:<col> <Kind> <type> [attr...]`. `Kind` is the AST node name that `astdump.go` prints. Position conventions are the AST dump's (a `CallExpr` is at its `(`). A local is `<line>:<col> Local <name> [mutable] <type>` at its declaring position (variable, parameter, loop variable, later pattern binding). Resolution attributes: `local=<name>`, `const=<module>.<name>`, `func=<module>.<name>`, `ctor=<Enum>.<Variant>`, `builtin=<Builtin.Name>`, `module=<path>`. Each node owns the line kinds and attributes of its own constructs and says so in the doc comment's Owners list.
- Go has no ids, so the Go side keys by walk order and position, not by id. The Narra side stores results in per-body side tables (below) and prints by walking the AST.
- Positions and types are enough to find the first divergence with a text diff. Do not print the C name or hash.

## Acceptance

- [ ] `typesdump.go` defines the format in its doc comment; hidden commands `narra types` and `narrac types` exist; goldens in `testdata/typesdump/bodies/` cover each line kind of this node and the error output.
- [ ] `TestSelfhostBodies` compares `narrac types` with `narra types` (stdout, or stderr and exit 1) on every input of `declInputs` plus the new cases. It skips an input only when the load fails or the Narra output holds `unsupported:`; it logs every skip with the construct, and prints the compared and skipped counts.
- [ ] The ported error formats are tested: unknown name, type mismatch, wrong argument count, assignment to a constant or an immutable local, missing return, `break` outside a loop, constant rules, integer literal overflow, template part not a string, opaque field access.
- [ ] `TestSelfhostDecls`, `TestSelfhostTraits` and `TestSelfhostImports` still pass; `narra emit`, `narra check` and `narra check --json` are byte-identical to main on all `compiler/testdata` inputs.
- [ ] The outcome lists: the per-body side-table layout, the `unsupported` convention and every stub with the node that replaces it, the compared and skipped counts per skip reason, and the Go changes.
- [ ] The Narra checker reports E0421 (alias equals a visible type name), E0423 (renamed import clashes with a declaration, a type name or an alias) and E0424 (a name renamed and imported twice), with the same messages as the Go checker. `TestSelfhostImports` no longer skips inputs with those errors, and the sentence about them in `compiler/README.md` (line 54, "The self-hosted compiler reports the duplicate alias but not ...") is removed.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L: about 1,350 Narra lines plus a 300-line Go dump. Uncertainty medium: the per-body side tables and the dump format have no precedent. Coupling medium: five later nodes build on the layout. Blast radius medium: it restructures the package driver and defines the shared dispatch files.

## Notes for the worker

- Side tables are per body, not per package. A generic body is checked once per instance (Go clones the AST per instance), and the same ids get different types in each, so key results by body: `BodyInfo { header, first: int, types: Array<Type>, ... }` indexed by `id - first`. Take the id range of a body from the nodes it contains (`collectIds` in `idcheck.nra` shows how to walk them; verify the range is contiguous enough, since ids have gaps and dummy nodes have id -1). Do not allocate `Array.filled(_, nextId)` per body: with many bodies that is quadratic. A package-wide table sized by `nextId` is fine for data that does not depend on the instance (`paramTypes`, already there).
- Local variables: an id per `Local`; scopes as an `Array<Map<string, int>>` stack in `Checker`, `push`/`pop` as in Go. Narrowing (`setNarrow`) is selfhost-check-patterns; leave the `Local` field for it.
- Keep Go's order of checks and of error reports, and Go's dedup (`errorf` already ports it). A different order shows up as a dump mismatch.
- Dispatch: `exprs.nra` has `exprInner` as one `match` on `Expr`; each later node replaces its `unsupported` arms with a call into its own module. Keep each arm to one line so later branches merge cleanly. A Narra module cannot import in a cycle, so put each cycle (expr -> call -> args -> expr) in one module or pass the entry point through a struct field closure; record the choice in the outcome. Mutual recursion between `expr` and the later nodes' modules is the main layout risk: decide it here.
- Integer literal check: Go uses `strconv.ParseInt` with base 0; use `Int.parse` and handle `0x`, `0b`, `0o`, `_`.
- From literal-type-args: port `structLit` with `typeArgs`: resolve them with `resolveType`, compare with the expected type (E0237 text `the type arguments make %s, but the expected type is %s`), use the result as the literal's type. The `MapLit` half goes to selfhost-check-patterns.
- From selfhost-check-decls: run `checkDefaults` (field default expressions) at the declaration point where Go does (HOOK in `declare.nra`), since a default that names a generic instance changes the `order` lines. `Checker.localN` is yours. Parameter types are in `paramTypes`; arrow parameters are selfhost-check-closures.
- From selfhost-check-imports: resolve a plain name as local, constant, function, imported std function, prelude builtin. Never compare struct types by name; `==` on `Type` is identity.
- From selfhost-check-traits: `==` and `!=` call `checkEquatable`; `typeImplements` is `mutable c`. `flushMapKeys` runs after all bodies.
- From std-test: the test-only rule (`Checker.inTest`, E0907) is set while a `test` body is checked; the check itself is in selfhost-check-builtins, but set the flag here.
- Narra rules: `f(mutable c, g(c))` is rejected, and an `errorf(mutable c, ...)` message must not read `c` in the same call; build it into a local first. Do not use `implements`, `derives`, `trait`, `implement` as names. Loop over statements; recurse only on nesting.
- Edit the Go checker only to expose what the dump needs or to fix a bug the comparison finds; list each in the outcome. Add a CLAUDE.md-style comment for new Go files only as the repo does.
- Inbox, `narrac decls` quadratic cost: located during this refinement. It is not in the declaration passes: `sortStrings` in `selfhost/src/modules.nra` is an insertion sort over a module's exported names, quadratic on descending names (`narrac pkg` on 16000 exports: 3 s; single file, no exports: linear). Bodies do not make it worse, because they add no sort over module-wide lists. You may fix it (a merge sort, as traits did for impl blocks) as a small extra commit if it touches no other file; otherwise leave it in the inbox.
- From import-alias-rules (PR #35): the Narra checker ports only E0422. Port E0421, E0423 and E0424 from `compiler/internal/narra/import_names.go` (`checkImportNames`, `aliasTaken`) to `selfhost/src/scopes.nra`, where E0422 already lives, and add their formats to `portedImportErrors` in `selfhost_imports_test.go`, which is how the harness stops skipping those inputs. The Go tests in `compiler/testdata/errors/import_*.nra` and `compiler/testdata/packages/errors/import_rename_type_clash/` are the inputs. Fixes (the rename suggestion) are not part of the Narra port. Added 2026-10-31.
- Record the per-body side-table design (tables indexed by `id - first` per checked body, because a generic function is checked once per instance) as D70 in `DECISIONS.md`.
