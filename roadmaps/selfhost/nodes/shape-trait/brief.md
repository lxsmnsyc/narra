---
id: shape-trait
title: "The Shape trait, Json.encode and Json.decode"
status: active
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
gate: true
added: 2026-11-07
spent:
owns:
  - testdata/stage2/shape*
  - testdata/stage2/errors/shape_*
  - roadmaps/selfhost/nodes/shape-trait/**
shared:
  - std/src/shape.nra
  - std/src/json.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/traits.nra
  - selfhost/src/emittraits.nra
  - selfhost/src/generics.nra
  - selfhost/src/stdtext.nra
  - SPEC.md
  - compiler/README.md
  - DECISIONS.md
reserves: [diag E0530, decision D84]
provides: []
depends_on:
  - on: std-json
    kind: hard
    why: adds the trait to std/shape and encode and decode to std/json, which std-json creates
  - on: call-type-args
    kind: hard
    why: Json.decode<User>(text) and Shape.fromValue<T>(value) need explicit type arguments in calls
---

# The Shape trait, Json.encode and Json.decode

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

Typed JSON for any type that implements `Shape`. `std/shape` gets the trait, the compiler provides it for the prelude types, and `std/json` gets `encode` and `decode`. A user type implements `Shape` by hand here; `derives Shape` is node shape-derive.

```ts
// std/shape
export trait Shape {
  function toValue(value: Self): Value uses alloc
  function fromValue(value: Value): Result<Self, string> uses alloc
}

// std/json
export function encode<T: Shape>(value: T): string uses alloc
export function decode<T: Shape>(text: string): Result<T, string> uses alloc

const user = Json.decode<User>(text)
```

## Context

- Read `nodes/std-json/outcome.md` and `nodes/call-type-args/outcome.md` first.
- The prelude impls exist in the self-hosted compiler only. The Go compiler loads `std/shape` and `std/json` (`selfhost/src` imports `std/json`), so both modules must still check under Go: Go needs only the trait declaration and the generic bodies, never an instance. Samples that call `encode` or `decode` go in `testdata/stage2/` (README "Tests").
- Go's `check_traits.go` has a special case for `derives Shape` ("not supported by this compiler yet"). A source trait named `Shape` must not clash with it.

## Acceptance

- [ ] The compiler provides `Shape` for `bool`, `int`, `f64`, `char` (one-character `Text`), `string`, `Value` itself, `Array<T>`, `Map<string, V>`, `Option<T>` and tuples (`List`) when their contents implement it. `Map` with a non-string key and function types do not implement it (E0530 names the type that does not implement `Shape`).
- [ ] Decoding is strict on types and lenient on extra data: `int` accepts only `Value.Int`; `f64` accepts `Int` and `Float`; `Option<T>` decodes `Null` as `None`. Errors carry a path from the root: `.users[2].age: expected int, found Text`.
- [ ] `Json.encode(value)` is `stringify(Shape.toValue(value))`; `Json.decode<T>(text)` is `parse`, then `Shape.fromValue<T>`, with parse errors passed through.
- [ ] Stage 2 samples: a struct with a hand-written `implement Shape` round-trips through `Json.encode` and `Json.decode<T>`; nested arrays, maps and options; every decode error form. ASan and leak check clean. Error tests for E0530.
- [ ] SPEC section 4 "Serialization with `Shape`" describes the trait and uses `Json.decode<User>(text)` instead of `Json.decode(text, User)`. DECISIONS.md D84 records the decision below.
- [ ] Fast check passes, including `TestSelfhostBootstrap` and the stage 2 tests.

## Sizing

M. Uncertainty medium (prelude impls for generic containers are new; today only `Show` and friends are compiler-provided), size medium, tests automatic.

## Notes for the worker

- Compiler-provided impls: follow how `Show` is provided for `Array<T>` and tuples in `selfhost/src/traits.nra` and `emittraits.nra`. If generating the bodies as Narra source text and checking them like a std module is simpler, do that, and say so in the outcome: shape-derive can reuse the same mechanism.

## Decision to confirm

Plan (D84): `Shape` converts to and from `Value` (from `std/shape`, D83) with two trait functions, and the caller names the target type with an explicit type argument, `Json.decode<User>(text)`.

- Reason: an explicit type argument is TypeScript's form for a generic call and is already in SPEC section 4 (`id<int>(1)`), so no new kind of value is needed.
- Reason: going through `Value` keeps `Shape` independent of JSON, and every impl is ordinary Narra code with no runtime type layout.
- Rejected: SPEC's `Json.decode(text, User)`, a type name passed as a value. It needs a new type-witness type, and the natural name `Type` is taken by the self-hosted compiler's own `enum Type`, while prelude names cannot be redeclared.
- Not the only form: inference from the expected type (`genericCall` unifies the result with it) also works, `const user: Result<User, string> = Json.decode(text)`. It stays allowed, but it cannot be the documented form: a call in a `match` subject or an argument has no annotation to read.
- Rejected: SPEC's "constant description" read by a generic runtime walker. It needs layout tables (parked node layout-tables) and C code that builds reference-counted values from raw memory.
- Cost: one intermediate `Value` tree per encode or decode.
- D84 also records the derived encoding from `nodes/shape-derive/brief.md`: a `None` field is left out, and an enum value is `{"type": "Circle", ...fields}`, TypeScript's discriminated union. Rejected: Rust serde's default `{"Circle": {...}}`, which TypeScript code cannot narrow with a `switch`.
