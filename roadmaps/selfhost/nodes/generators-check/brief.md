---
id: generators-check
title: "Generators: syntax, types and checking"
status: active
priority: should
type: feature
story:
early: false
outline: false
milestone: m5-concurrency
target:
ship: pr
from:
size: L
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestCompileErrors|TestSelfhost|TestASTDump|TestTypesDump'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent:
owns:
  - compiler/testdata/errors/generator_*
  - compiler/testdata/typesdump/generators/**
  - compiler/testdata/astdump/generators*
  - roadmaps/selfhost/nodes/generators-check/**
shared:
  - compiler/internal/narra/lexer.go
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/astdump.go
  - compiler/internal/narra/check*.go
  - compiler/internal/narra/types.go
  - compiler/internal/narra/typesdump.go
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/**
  - SPEC.md
  - DECISIONS.md
reserves: [diag E0046, diag E0195, diag E0196, diag E0291, diag E0292, diag E0615, decision D78]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes after bootstrap go into both compilers together, and stage 2 must match Go on the new tests
---

# Generators: syntax, types and checking

## Goal

The front half of generators (SPEC.md section 8, D19, D31), in the Go compiler and `selfhost/` together: `function*` declarations, `yield value` and `yield* other` statements, the `Generator<T>` type, `for ... of` over a `Generator<T>` value, and every compile error. The emitter is the next node (`generators`); until it lands, emitting a program that declares a `function*` stops with a clear "generators are not compiled yet" error, so the node can merge alone.

## Context

- `yield` is not a keyword today (`keywords` in `lexer.go`, `selfhost/src/lexer.nra`); no identifier named `yield` exists in `selfhost/src` or `compiler/testdata`.
- for-of sources are dispatched by the checker's `forSource` (`nodes/for-of-sources/outcome.md`); a `Generator<T>` becomes one more source kind.
- `Range` was made a prelude type name by for-of-sources; `Generator` is added the same way (`check_types.go`, `types.go`, `selfhost/src/resolve.nra`, `types.nra`).
- Effects: `check_effects.go`, `selfhost/src/effects.nra` (D58).
- Source: SPEC.md section 8 "Generators", section 7 (the `function*` line), D19, D31.

## Rules

1. `function* name(params): Generator<T> { ... }` declares a generator, top level only (D74). The return type is required and must be `Generator<T>`; `Coroutine<In, Out>` is node `coroutines` and is an error here. Generic generators are allowed. `export function*` works like `export function`. No `function*` arrow form.
2. `yield value` and `yield* other` are statements, valid only directly in a `function*` body or its nested statement blocks (`if`, `while`, `for`, labels, `match` statement arms). Not inside an arrow function, a block expression, a ternary, or a `match` used as a value. `value` must have type `T`; `other` must be a `Generator<T>`.
3. `return` in a generator takes no value. Falling off the end finishes the generator.
4. A generator cannot take a `mutable` parameter, and a function-type parameter of a generator must be `escaping`: the frame outlives the call.
5. Calling a generator function runs none of its body and returns a `Generator<T>`. `Generator<T>` is a value type with no `==`, hashing or `Show`.
6. Effects (D78, maintainer decision 2026-11-05) are carried on the type, as on function types (SPEC section 7, `(item: T) => U uses E`). The type is written `Generator<T> uses io`; `Generator<T>` alone is pure. Creating a generator charges only `alloc`, for its frame. The body's effects are charged where it runs: a `for ... of` over the generator, a `yield*` of it, and `Co.resume` (node coroutines). An exported `function*` writes the effects in its return type (`export function* lines(path: string): Generator<string> uses io, alloc`); a private one gets them inferred into its type. A body that uses an effect its declared type omits is an error. Assigning or passing a `Generator<T> uses io` where `Generator<T>` is expected is an error; the other direction is allowed.
7. An effect variable can be bound by a parameter of generator type as well as one of function type, extending D14: `export function sum<effect E>(numbers: Generator<int> uses E): int uses E`. E0802's rule ("must appear in the function type of a parameter") is widened to generator-typed parameters.
8. Self or mutual calls of generator functions create values and are not recursion for D37 or tail-loops.

## Acceptance

- [ ] AST dump (`compiler/testdata/astdump/generators.nra` and `.ast`): `function*`, `yield`, `yield*`, generic generator. Go and selfhost dumps match.
- [ ] Types dump (`compiler/testdata/typesdump/generators/`): call result types, for-of element type, the inferred `Generator<string> uses io, alloc` type of a private generator that reads a file, the effects of a function that iterates it, and a generic `sum<effect E>` called with a pure and with an io generator. Go and selfhost dumps match.
- [ ] Error tests `compiler/testdata/errors/generator_*.nra`, one per rule: missing or wrong return type; `yield` outside a generator, inside an arrow function, inside a block expression and inside a ternary; wrong `yield` value type; `yield*` of a non-generator; `return value` in a generator; a `mutable` parameter; a non-escaping function-type parameter; `==` on generators; an exported generator whose body does `io` but whose type omits `uses io`; an exported pure function that iterates a `Generator<T> uses io` (for-of and `yield*`); passing a `Generator<T> uses io` where `Generator<T>` is expected; an effect variable in `uses` bound by no function or generator parameter; emitting a generator before node `generators` (removed by that node).
- [ ] SPEC.md section 7 states that generator types carry effects and that generator-typed parameters bind effect variables; section 8 states rules 2 to 8. DECISIONS.md gets D78 (effects on the generator type, charged where the body runs, with `alloc` at creation; the extension of D14; value semantics of `Generator<T>`; `yield` as a statement only; rejected: charging the body's effects at creation, which lets a pure function run io through a generator value it was given) and one line under D14 pointing to D78.
- [ ] Fast check passes, including `TestSelfhostBootstrap`.

## Sizing

L. Uncertainty medium (effects on a value type are new: today only function-type parameters carry effects, and calls through stored function values count as pure), coupling low, blast radius medium (lexer keyword, AST node, prelude type, for-of source and the effect pass in both compilers), verifiability low (dumps and error tests), size medium, context medium.

## Notes for the worker

- Codes reserved: E0046 (parser, for example `function*` in a nested position or `yield*` without an operand), E0195 and E0196 (declarations: return type, `mutable` parameter), E0291 and E0292 (statements: `yield` placement, `return value`), E0615 (non-escaping function parameter of a generator). Reuse existing type-mismatch formats where the text fits.
- New AST nodes (`YieldStmt` with a `Delegate` flag, `FuncDecl.Generator`) need cases in `clone.go`, `astdump.go`, the checker and emitter switches, and the selfhost walkers.
- The emitter stub error must be one place in each compiler, so node `generators` deletes it in one line.
- Record in the outcome the checker facts the emitter will need: which statements contain a `yield` (directly or nested), the generator's element type, and the for-of source kind.

- Effects on types overlap with outline node typed-calls-and-effects, which makes calls through function values charge the effects of their function type (today they count as pure, SPEC lines 702, 791, 805). Store a generator type's effects in the same form the checker uses for `uses` lists on function types, so that node can reuse it for function values. Record the representation in the outcome.

## Open questions

- Should `Arrays.values(xs)` and `Maps.entries(m)` outside a loop head now return `Generator<T>` (D31 says collections return generators), instead of E0920? Proposed: leave that to std-iterator.
