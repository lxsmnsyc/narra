---
id: std-option-result
title: std/option and std/result helpers
status: done
spent: 375000
priority: should
type: feature
outline: false
milestone: m3-core
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestRun|TestErrors'
exclusive: []
fulfilled_by:
gate: true
added: 2026-10-30
owns:
  - compiler/internal/narra/std_option.go
  - compiler/internal/narra/std_result.go
  - compiler/testdata/std_option*.nra
  - compiler/testdata/std_option*.out
  - compiler/testdata/std_result*.nra
  - compiler/testdata/std_result*.out
  - compiler/testdata/errors/std_option_*.nra
  - compiler/testdata/errors/std_result_*.nra
shared:
  - compiler/internal/narra/builtins.go
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/emit_call.go
  - compiler/internal/narra/check_effects.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/README.md
  - SPEC.md
  - DECISIONS.md
reserves: [decision D69]
provides: []
depends_on:
  - on: prelude-imports
    kind: hard
    why: follows the std-module pattern for function namespaces that prelude-imports sets, and both edit SPEC.md's prelude and std sections
---

# std/option and std/result helpers

## Goal

Two std modules with combinators for the prelude types:

- `std/option` (`import * as Option from "std/option"`): `map`, `andThen`, `withDefault`, `orElse`, `isSome`, `isNone`, `okOr` (to `Result`).
- `std/result` (`import * as Result from "std/result"`): `map`, `mapErr`, `andThen`, `withDefault`, `ok` and `err` (to `Option`).

## Context

- User decision 2026-10-30: add these; no `unwrap` or `expect`. `SPEC.md` already documents `Option.withDefault(x, fallback)` but the compiler does not implement it (no `Option` namespace in `std_*.go`); this node closes that gap.
- Callbacks take effect parameters (`f: (value: T) => U uses E`), like `Array.map`, so `io` and `alloc` closures work. Set `Builtin.Effects` on every builtin (`compiler/internal/narra/CLAUDE.md`).
- `Option`/`Result` the import alias equals the global type name; prelude-imports makes that work for `Array` and `Map`; reuse it.
- Spec: keep `is` and `match` as the main style for error paths (section 5, "every error path is a visible `if` or `match`"); present these as helpers that never return early from the caller. Record D69. Gate: the user approves the spec text.
- Closure cost: each call with a capturing callback allocates a closure environment. The user noted "we'll have to do better on closure management"; record what a call costs in the outcome (allocations per call in debug builds) so the `inline` node can measure against it.

## Acceptance

- [x] Every function works on `Some`/`None` and `Ok`/`Err`, with capturing and non-capturing callbacks and with an `io` callback; samples with `.out` pass in debug (ASan, UBSan, leak check) and release.
- [x] Error tests: a callback of the wrong type, a missing import, `Option.andThen` with a callback that does not return an `Option`.
- [x] `SPEC.md` (the Optional values section and section 5), D69 and `compiler/README.md` describe the modules.
- [x] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: thirteen generic builtins in two modules, following `Array.map`'s check and emit.

## Notes for the worker

- If selfhost-check-imports has already ported the builtin registry when this starts, list the new modules in the outcome so the Narra registry picks them up.
