---
id: call-type-args
title: "Explicit type arguments in calls: f<T>(x) and Trait.f<T>(x)"
status: active
priority: should
type: feature
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - compiler/testdata/call_type_args*
  - compiler/testdata/errors/call_type_args_*
  - roadmaps/selfhost/nodes/call-type-args/**
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/check_traits.go
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/calls.nra
  - selfhost/src/generics.nra
  - selfhost/src/traits.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/querysites.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E0326, diag E0327, diag E0529]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Explicit type arguments in calls: f<T>(x) and Trait.f<T>(x)

## Goal

Spec-gap F21 (SPEC section 4 "Generics", and `Channel.create<int>(16)` in section 8): `id<int>(1)` today parses as two comparisons. Add explicit type arguments to calls in both compilers. Also let a trait call name `Self`: `Shape.fromValue<User>(value)`. A trait function whose `Self` appears only in its result needs that form. Node shape-trait needs it for `Json.decode<User>(text)`.

## Context

- Split from typed-calls-and-effects, which keeps F20 (effects of function values).
- Parser: the balanced `<...>` lookahead of literal type arguments (D68, `Map<K, V> {}`) decides between type arguments and `<`. Here the closing `>` must be followed directly by `(`, as in TypeScript.
- Generic calls: `genericCall` in `compiler/internal/narra/check_call.go` and `checkGenericCall` in `selfhost/src/calls.nra`. Given arguments fix the type parameters before unification; the rest is unchanged.
- Trait calls dispatch on the argument in the `Self` position (`check_traits.go`, `selfhost/src/traits.nra`).
- Spec-gap audit row F21: `nodes/spec-gap-audit/outcome.md`.

## Acceptance

- [ ] `f<int>(1)`, `Mod.f<int, string>(x)` and `f<Array<int>>(xs)` call with the given type arguments; `a < b > (c)` with values on both sides still parses as comparisons. Type arguments may be fewer than the parameters only when none are given (all or nothing, as in TypeScript without defaults).
- [ ] `Trait.f<T>(args)` sets `Self = T`. A trait may declare a function whose `Self` appears in no parameter (for example `function fromValue(value: Value): Result<Self, string>`); calling it without `<T>` is an error that says to name the type. Inside a generic body `<T: Shape>`, `Shape.fromValue<T>(value)` resolves at instantiation.
- [ ] Errors: type arguments on a non-generic function (E0326), the wrong number of type arguments (E0327), a trait call that cannot find `Self` (E0529). An argument that does not fit the given type keeps E0119.
- [ ] Samples `call_type_args.nra`: explicit arguments where inference cannot tell (`Arrays.filled` style, a generic `none<T>(): Option<T>`), nested generic arguments, a trait function with `Self` only in the result, called directly and from a generic body. Error tests, one per code, plus `a < b > (c)` staying a comparison.
- [ ] SPEC section 4 "Generics" shows the call form and the trait form; README drops F21 from "Not supported yet". Go and `selfhost/` emit the same C; fast check passes.

## Sizing

M. Parser lookahead and generic-call changes in both compilers; known pattern (D68), tests automatic.

## Notes for the worker

- Every new syntax form reaches the formatter (`selfhost/src/fmt*.nra`), the AST dumps (`astdump.go`, `astdump.nra`) and the query site index (`querysites.nra`).
