---
id: prelude-imports
title: Move prelude function namespaces behind std imports
status: todo
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestRun|TestErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: true
added: 2026-10-30
owns:
  - compiler/internal/narra/std_array.go
  - compiler/internal/narra/std_range.go
  - compiler/internal/narra/std_box.go
  - compiler/testdata/errors/prelude_*.nra
shared:
  - compiler/internal/narra/std_prelude.go
  - compiler/internal/narra/std_map.go
  - compiler/internal/narra/builtins.go
  - compiler/internal/narra/check.go
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/check_types.go
  - compiler/internal/narra/modules_check.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/internal/narra/diagnostic_fixes.go
  - compiler/internal/narra/decldump.go
  - compiler/testdata/**
  - selfhost/src/**
  - bench/**
  - compiler/README.md
  - SPEC.md
  - DECISIONS.md
reserves: [decision D67]
provides: []
depends_on:
  - on: selfhost-check-decls
    kind: hard
    why: edits the same prelude handling (`Checker.prelude`, check.go) and the selfhost/src files that node adds; running after it avoids a rebase of a 1800-line port
---

# Move prelude function namespaces behind std imports

## Goal

No global functions except `print`, `assert` and `panic`. The function namespaces `Array.*`, `Map.*`, `Range.*` and `Box.*` move to `std/array`, `std/map`, `std/range` and `std/box` and need an import (`import * as Array from "std/array"`). The types and constructors stay global: `Array`, `Map`, `Range`, `Box`, `Option`, `Some`, `None`, `Result`, `Ok`, `Err` and the primitive types.

## Context

- User decision 2026-10-30: "I don't really want to expose global functions". Types stay global because the syntax produces them (`[]`, map literals, `is`, `match`, `derives`); `print`, `assert` and `panic` stay global (language level: contracts, test blocks, runtime panics). Run right after selfhost-check-decls and before selfhost-check-traits, so selfhost-check-imports ports the builtin registry in its new shape once.
- Today: `SPEC.md` section on imports says the prelude holds `Option`, `Some`, `None`, `Result`, `Ok`, `Err`, `Array`, `Map`, `Range`, the primitive types, `print`, `assert` and `panic`. The compiler also exposes the `Array`, `Map`, `Range` and `Box` namespaces through `registerNamespace` (`std_prelude.go`, `std_map.go`); `Box` is missing from the spec list. `check_call.go` and `check_types.go` special-case `Box` and `Map` when the user declares their own.
- Register each namespace as a std module (`registerModule`), as `std/string` and `std/int` are. A namespace name may equal a global type name (`Array` the type, `Array` the import alias); keep that working, and keep user types named `Box` or `Map` shadowing the prelude.
- A call such as `Array.push(...)` without the import is a compile error with a fix that adds the import line (diagnostic code from the format table; a fix like the existing ones in `diagnostic_fixes.go`).
- Migrate every user: `compiler/testdata/**` (programs, errors, packages, goldens), `selfhost/src/**`, `bench/**` and the examples in `SPEC.md`. Do it with a script, record the script in the log, and keep diffs to added import lines. Regenerate goldens (`NARRA_UPDATE_GOLDEN=1`) only where the import lines shift positions; check every regenerated diff is position-only.
- If selfhost-check-decls's Narra port or its `decldump.go` handle prelude namespaces, update them so `TestSelfhostDecls` and the other selfhost dumps still match.
- Spec: update the prelude bullet and the std module list; record the rule (a type in the prelude does not bring its functions; modules hold functions) as D67. Gate: the user approves the spec text.

## Acceptance

- [ ] `Array.*`, `Map.*`, `Range.*` and `Box.*` resolve only through their `std/*` import; without it, a compile error with an add-import fix (`testdata/errors/prelude_*.nra`). Types and constructors remain global; `print`, `assert`, `panic` remain global.
- [ ] Every sample, package, bench program and `selfhost/src` file is migrated; all run tests pass in debug (ASan, UBSan, leak check) and release; selfhost dump tests pass.
- [ ] `SPEC.md`, D67 and `compiler/README.md` describe the new prelude.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: four namespace registrations moved, one diagnostic with a fix, and a scripted migration across many files.

## Notes for the worker

- The migration touches many files but each change is an import line; keep it mechanical so review can skim it.
- Effects: moving a builtin must keep its `Builtin.Effects`.
