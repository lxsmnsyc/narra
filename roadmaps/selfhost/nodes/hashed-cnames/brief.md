---
id: hashed-cnames
title: Hashed C names
status: todo
priority: must
type: refactor
outline: false
milestone: m3-core
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-28
owns:
  - compiler/internal/narra/cnames*.go
  - compiler/testdata/cnames*
  - compiler/testdata/packages/cnames*/**
shared:
  - compiler/internal/narra/modules.go
  - compiler/internal/narra/modules_check.go
  - compiler/internal/narra/types.go
  - compiler/internal/narra/emit*.go
  - compiler/internal/narra/check_traits.go
  - compiler/internal/narra/check_patterns.go
  - compiler/internal/narra/check_closure.go
reserves: []
provides: [module-graph]
depends_on:
  - on: closures
    kind: hard
    why: renames the closure, environment and thunk C names that closures added
  - on: patterns
    kind: hard
    why: renames the tuple C names that patterns adds, and both edit the same emitter files
---

# Hashed C names

## Goal

C names for user declarations and composite types become short, stable and unique by construction. This replaces the current string mangling, which joins module paths, names and type arguments with `__`.

- **Declarations:** `<kind>_<module hash>_<name>`. The module hash is FNV-1a 32 bit (`hash/fnv`), as 8 lowercase hex digits, over the module path relative to the package. It is computed once per module. Example: `app/text.show` becomes `f_3fa9c1e7_show`.
- **Composite types** (generic instances, tuples, function types, arrays, options and other prelude instances) are interned. Each distinct type is named once. A bigger type refers to its parts by their names, never by their expansion: `#[int, #[int, int]]` is `#[int, T1]` where `T1` names `#[int, int]`.
- The ID of an interned type is `<kind>_<readable head>_<hash>`. The hash is FNV-1a 32 over the folded form: the head, the arity, and the C names of the arguments in order. Examples: `S_Tuple2_4e1a90c3`, or `S_3fa9c1e7_Box_<hash>` for a generic instance of `Box`.
- The generated C carries a comment table that maps each interned name to its Narra type, for example `/* S_Tuple2_b71f0d28 = #[int, #[int, int]] */`, so backtraces and ASan reports can be read.

## Context

The user chose this on 2026-10-28. String mangling caused three collision bugs: `__` inside names, modules named after prelude types, and tuple arity. Each was patched with a new naming rule. Interning plus a folded hash makes names unique without such rules, keeps identifiers short however deep the nesting, and keeps them stable.

- **Why a hash and not sequential IDs (`T1`, `T2`):** sequential IDs depend on the order types are first seen, so a small edit renumbers the whole file and future per-module builds could not share names. A folded hash gives the same type the same name in every build, on every machine, and at both stages of self-hosting.
- **Why FNV-1a 32 and not xxHash32:** both were benchmarked. The hash inputs here are a few short tokens, so speed does not matter. FNV-1a is in the Go standard library and takes a few lines to port to Narra. `hash/maphash` is ruled out because its seed is random per process.
- **No absolute paths** may enter any hash input.
- **Nominal types are never folded into what they wrap.** `opaque` and `distinct` types (node opaque-distinct) keep their own identity, so `distinct UserId = int` must get a different name from `int`. Only true aliases, if the language ever adds them, would be expanded first.
- **Prelude and runtime names** (`lstr_*`, `narra_*`, `E_Ordering`, and others) may stay as they are if no hashed name can equal them. Every hashed name has a hash segment, and prelude names have none.

## Acceptance

- [ ] One helper module (`cnames.go`) owns all C naming: the module hash, declaration names, the type interning table and folded hashes. Every existing name builder uses it. Those are `CName`, `funcCName`, `mangle`, `typeArgsMangle`, `implFuncName`, `traitFuncName`, `tupleCName`, and the closure, environment and thunk names.
- [ ] The compiler keeps a map from C name to type or declaration and stops with an internal error if two different ones get the same name.
- [ ] The rules against `__` and against leading or trailing `_` in declaration names, trait names and module path segments are removed. The tests that checked them now compile, and a sample program uses such names.
- [ ] The tuple-arity, `Option.nra`-module and `b__c`/`a__b` repros from earlier reviews are samples that compile and run.
- [ ] A test emits the same program twice from two different working directories and gets byte-identical C. Another test checks that adding an unrelated function does not change the names of existing types.
- [ ] Output, errors and exit codes for every program in `compiler/testdata/` are unchanged. Generated C is unchanged apart from names and the comment table: compare it with names normalized.
- [ ] Change request `contracts/module-graph.change-request.md` (v3) describes the new naming. If the trait-model C name sections change, add `contracts/trait-model.change-request.md` (v3) as well.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: one naming module, plus the mechanical replacement of the existing name builders.

## Notes for the worker

- Keep the helper small and self-contained, so that the self-hosted compiler can port it line for line. The FNV-1a constants are offset basis `2166136261` and prime `16777619`.
- Intern types by their folded form, so lookup and naming use one key.
- Emit the comment table in a fixed order, sorted by name, so the C stays deterministic.
