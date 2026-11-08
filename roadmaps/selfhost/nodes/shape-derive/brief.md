---
id: shape-derive
title: derives Shape for structs and enums
status: todo
priority: should
type: feature
story: batch-1
early: false
outline: false
milestone: m10-std
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - testdata/stage2/derive_shape*
  - testdata/stage2/errors/derive_shape_*
  - roadmaps/selfhost/nodes/shape-derive/**
shared:
  - selfhost/src/traits.nra
  - selfhost/src/emittraits.nra
  - selfhost/src/declare.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E0531, diag E0532, diag E0533]
provides: []
depends_on:
  - on: shape-trait
    kind: hard
    why: generates implementations of the trait shape-trait adds, reusing its prelude impls; D84 fixes the encoding
---

# derives Shape for structs and enums

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

`struct User derives Shape { ... }` and `enum Event derives Shape { ... }` make the self-hosted compiler write the `implement Shape` block, so `Json.encode(user)` and `Json.decode<User>(text)` work with no hand-written code (SPEC section 4). The Go compiler keeps its "derives Shape is not supported by this compiler yet" error; `selfhost/src` does not derive `Shape`.

## Context

- Read `nodes/shape-trait/outcome.md` first; reuse its mechanism for compiler-provided impls.
- Derived traits today: `derives` handling in `selfhost/src/traits.nra` and `emittraits.nra`.

## Encoding

- Struct: `Record` with one entry per field, in declaration order. A field of type `Option<T>` holding `None` is left out, as `JSON.stringify` leaves out `undefined`.
- Decoding a struct: a missing field takes its default (`age: int = 0`), a missing `Option` field is `None`, any other missing field is an error (`.age: missing field`). Unknown fields are ignored.
- Enum: a `Record` with a `"type"` entry naming the constructor, plus one entry per field: `{"type":"Circle","radius":1.0}`, `{"type":"Empty"}`. This is TypeScript's discriminated union. A constructor field named `type` would clash, so deriving `Shape` on such an enum is an error. An unknown or missing `type` is an error naming the enum.
- Generic types derive `Shape` when their type arguments implement it, as `derives Show` does today.

## Acceptance

- [ ] Stage 2 samples: structs with defaults, `Option` fields, nested structs, arrays and maps of derived types; an enum with unit and fielded constructors; a recursive enum (D55); a generic struct; round trips through `Json.encode` and `Json.decode<T>`, and the exact JSON text for a few values. ASan and leak check clean.
- [ ] Errors: a field whose type does not implement `Shape` (E0531 names the field and type), an enum constructor with a field named `type` (E0532), an opaque struct's derived `Shape` used from another module (E0533, SPEC section 4: only inside its own module).
- [ ] SPEC section 4 "Serialization with `Shape`" describes the encoding above. README "Supported" says `derives Shape` works in the self-hosted compiler only.
- [ ] Fast check passes, including the stage 2 tests and `TestSelfhostBootstrap`.

## Sizing

M. Code generation for two declaration kinds in the self-hosted compiler only; the pattern is set by shape-trait, tests automatic.
