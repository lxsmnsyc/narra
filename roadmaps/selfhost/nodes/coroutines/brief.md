---
id: coroutines
title: "Coroutine<In, Out> and Co.resume"
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
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent:
owns:
  - compiler/testdata/coroutine*
  - compiler/testdata/errors/coroutine_*
  - compiler/internal/narra/std_coroutine.go
  - roadmaps/selfhost/nodes/coroutines/**
shared:
  - compiler/internal/narra/emit_generator.go
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/check*.go
  - compiler/internal/narra/types.go
  - compiler/internal/narra/builtins.go
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/**
  - SPEC.md
  - compiler/README.md
  - DECISIONS.md
reserves: [diag E0197, diag E0293, decision D79]
provides: []
depends_on:
  - on: generators
    kind: hard
    why: extends the state-machine lowering and the Generator value of node generators
---

# Coroutine<In, Out> and Co.resume

## Goal

The rest of SPEC.md section 8 "Generators", in both compilers: `Coroutine<In, Out>` declared with `function*`, `const input = yield output` inside it, and `std/coroutine` with `Co.resume(mutable g, input)` for generators and coroutines, returning `Yielded(value)`, `Done` or `Panicked(message)`.

## Context

- Read `nodes/generators/outcome.md` and `nodes/generators-check/outcome.md` first.
- std modules today are compiler builtins (`std_*.go`, `selfhost/src/stdcalls.nra`, `emitstd.nra`). If std-source has landed and decided otherwise, follow its outcome.
- `Panicked` is only reported in test builds; `std/test` already catches panics in a child process (D65). Outside test builds a panic aborts.

## Acceptance

- [ ] `function* f(): Coroutine<In, Out>` declares a coroutine. `const x = yield v` is valid only in a coroutine and only as a `const` or `let` initializer statement; the first resume's input is dropped, as in TypeScript.
- [ ] `Co.resume(mutable g, input)` for a `Coroutine<In, Out>` and `Co.resume(mutable g)` for a `Generator<T>` change `g` in place (count above 1 clones first, as in node generators). It returns `Step<Out>` from `std/coroutine` (maintainer decision 2026-11-05), with constructors `Step.Yielded(value)`, `Step.Done` and `Step.Panicked(message)`, always qualified (D71). SPEC section 8 keeps the names Yielded, Done and Panicked and names the enum.
- [ ] `Co.resume` charges the effects carried on the type of `g` (D78): `Coroutine<In, Out> uses io` is written like `Generator<T> uses io`, and resuming one from a function without `io` is an error.
- [ ] `for ... of` over a coroutine is a compile error (it needs inputs).
- [ ] Samples `compiler/testdata/coroutine*.nra` and `.out`: a running-total coroutine; manual resume of a generator until `Done`; zipping two generators by hand; heap inputs and outputs released correctly; a coroutine dropped unfinished. Clean under ASan and the leak check. A test in a `test` block shows `Panicked`.
- [ ] Error tests: `const x = yield v` in a generator, `Co.resume` without `mutable`, a wrong input type, for-of over a coroutine, a pure exported function resuming an io coroutine.
- [ ] `DECISIONS.md` gets D79, "Coroutines are a separate type from generators" (maintainer request 2026-11-07), with the reasons for keeping `Coroutine<In, Out>` apart from `Generator<T>` where TypeScript has one `Generator<T, TReturn, TNext>`: `for ... of` has no input to send, since Narra has no `undefined` (D21); only coroutines need `yield` in a `const` initializer, so plain generators keep the statement-only `yield` of D78; the common lazy-sequence case keeps one type parameter. Name the rejected alternative (one generator type with an input parameter, as in TypeScript) and the cost (two types, and `Co.resume` takes both).
- [ ] SPEC.md section 8 matches; Go and `selfhost/` emit the same C; fast check passes including `TestSelfhostBootstrap`.

## Sizing

M. Uncertainty medium (panic capture in test builds), coupling low, blast radius low (extends node generators' file plus one std module), verifiability low (samples and error tests), size medium, context medium.

## Notes for the worker

- Codes reserved: E0197 (declarations), E0293 (statements, `yield` value used outside a coroutine). Decision number D79. Std errors reuse existing call formats where they fit; ask the orchestrator for an E09xx code otherwise.
- The input is passed to `resume` as an extra parameter; the generator's `resume` signature from node generators may need an input slot. Keep the C of plain generators unchanged if possible, so their samples' C does not churn.
