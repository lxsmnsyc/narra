---
id: std-iterator
title: std/iterator
status: done
priority: should
type: feature
story: batch-14
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent: 159248
owns:
  - std/src/iterator.nra
  - selfhost/src/stditerator.nra
  - testdata/iterator*.nra
  - testdata/iterator*.out
  - testdata/errors/iterator_*
  - roadmaps/selfhost/nodes/std-iterator/**
shared:
  - selfhost/src/stdlib.nra
  - selfhost/src/stdtext.nra
  - testdata/errors/unknown_module.err
  - testdata/declsdump/imports/err_unknown_module.dump
  - SPEC.md
  - compiler/README.md
  - docs/guide/
reserves: [diag E1300-E1309]
provides: []
depends_on:
  - on: std-design
    kind: hard
    why: D118 item 1 fixes the operator list, the callback rules and the consuming rule
  - on: std-foundation
    kind: hard
    why: source std modules need the generic mismatch errors, std/internal and the stage 2 tests (std-source gate, 2026-11-05)
  - on: generators
    kind: hard
    why: every lazy operator is a generator function over a Generator<T>
  - on: coroutines
    kind: hard
    why: zip steps two generators at once with Co.resume
  - on: std-source
    kind: hard
    why: the operators are Narra source
---

# std/iterator

## Goal

Build `std/iterator` (alias `Iterators`) as D118 item 1 says: lazy operators over `Generator<T>` and eager operators that consume one. Pure Narra source in `std/src/iterator.nra`; the `Arrays.*` functions do not change.

## Context

- Source: D118 item 1 in `DECISIONS.md`; SPEC section 3 "Std modules for tools".
- Model for a source module and its registry row: `std/src/json.nra`, `selfhost/src/stdjson.nra` (one row with `check: "source"`), and the commit that added it (`git show f9d63d62 --stat`).
- Evidence for the signature form: `function* map<T, U, effect E>(source: Generator<T> uses E, escaping transform: (value: T) => U uses E): Generator<U> uses E` builds and runs on the batch-14 seed (`roadmaps/selfhost/nodes/std-design/log.md`).

## Steps

- [ ] Lazy operators, each a `function*`: `map`, `filter`, `flatMap` (the callback returns a `Generator<U>`), `take`, `drop`, `takeWhile`, `dropWhile`, `enumerate` (`#[int, T]`), `zip` (`#[A, B]`, ends with the shorter, uses `Co.resume` on its own copy of the second source), `concat`. Callbacks are `escaping`.
- [ ] `chunk(source, size)` and `window(source, size)`: `Option<Generator<Array<T>>>`, `None` when `size < 1`. `take` and `drop` treat a negative count as 0.
- [ ] Eager operators with non-escaping callbacks and an effect variable: `reduce(source, combine, initial)`, `toArray`, `forEach`, `some`, `every`, `find` (an `Option<T>`), `count`.
- [ ] Registry row in `selfhost/src/stditerator.nra`, registered in `stdlib.nra`; regenerate `stdtext.nra` (`sh scripts/test.sh --suite generated --update`); update the module-list goldens.
- [ ] SPEC: replace "Not built yet" for `std/iterator` in section 3 with the built text; list the module in "Standard modules". Add the module to the guide's standard library chapter (the first batch-14 build to land creates `docs/guide/12-standard-library.md`; ask the orchestrator for the chapter number if another node took it).

## Acceptance

- [ ] `testdata/iterator.nra` prints the result of every operator, including an early `break` out of a `map` over an endless generator, and passes the leak check.
- [ ] `testdata/iterator_consume.nra` shows that `take(lines, 2)` leaves `lines` at its start, and that `Co.resume(mutable lines)` continues where the last resume stopped.
- [ ] `testdata/iterator_effects.nra`: `map` over a source with `uses io` gives a generator with `uses io`; a pure source and callback give a pure one.
- [ ] Error tests: `testdata/errors/iterator_capture_let.nra` (a `let` captured by a closure passed to `Iterators.map`, E0612) and `testdata/errors/iterator_chunk_unwrapped.nra` (a `for ... of` over the `Option` that `chunk` returns).
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: about 300 lines of Narra and four samples; no compiler change beyond one registry row.

## Notes for the worker

- Generators cannot take `mutable` parameters; `zip` keeps its second source in a local and resumes that.
- `@inline` is E1031 on a generator; do not try it on the lazy operators.
- No pipe operator (D12). Samples name each stage with `const`.
- Codes E1300-E1309 are reserved; this node is expected to need none. Add a row only for a check the compiler itself makes.
