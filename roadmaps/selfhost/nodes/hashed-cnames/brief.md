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

Every C name the compiler makes for a user declaration becomes `<kind>_<readable name>_<hash>`, where the hash is FNV-1a 32 bit (`hash/fnv`) in 8 lowercase hex digits. The hash input is the fully qualified name: the module path relative to the package, the declaration name, the type arguments, and the kind (function, struct, enum, closure, environment, trait implementation, tuple, function type). Example: `app/text.show` becomes `f_show_3fa9c1e7`.

## Context

The user chose this on 2026-10-28. Today names are made injective by rules: no `__` in names or module path segments, no leading or trailing `_`, and a planned ban on modules named after prelude types (node `prelude-cnames`, now dropped). The review of PR #6 found two more collisions of the same kind. A hash makes names unique by construction, and the readable part keeps ASan reports and backtraces readable.

- Decided: FNV-1a 32, not xxhash or `hash/maphash`. It is in the Go standard library, and the self-hosted compiler can reproduce it in a few lines. `maphash` is seeded at random, so its names would differ between builds.
- The hash input must not contain absolute paths, so that two machines and the stage 1 and stage 2 compilers emit the same C.
- Prelude and runtime names (`lstr_*`, `narra_*`, prelude types) may keep their current names if they cannot collide with hashed names.

## Acceptance

- [ ] Every user-declared C name has the form above. A test checks that the names are deterministic across two runs and two working directories.
- [ ] The compiler keeps a map from C name to qualified name and stops with an internal error on a collision.
- [ ] The rules against `__` in declaration names, trait names and module path segments are removed or relaxed where the hash makes them unnecessary. `testdata/` cases that tested those rules are updated, and a sample program uses such names.
- [ ] Output, errors and exit codes for every program in `compiler/testdata/` are unchanged. Generated C is unchanged except for names: compare it with names normalized.
- [ ] Change request `contracts/module-graph.change-request.md` for v3 describes the new naming. Also update the C name sections of `contracts/trait-model.md` if needed, through a change request.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: one naming helper used everywhere, plus mechanical replacement of the existing name builders (`CName`, `funcCName`, `mangle`, `implFuncName`, `traitFuncName`, closure names, tuple names).

## Notes for the worker

- Keep one helper, for example `cname(kind, qualified string) string`, so the self-hosted compiler can port it.
- The readable part may be shortened (for example to 32 bytes) to keep identifiers manageable; the hash carries uniqueness.
