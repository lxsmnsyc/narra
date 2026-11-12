---
id: array-patterns
title: "Array patterns in match, is and const"
status: todo
priority: should
type: feature
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/array-patterns/**
  - testdata/array_patterns*
  - testdata/errors/array_pattern*
shared:
  - selfhost/src/ast.nra
  - selfhost/src/parser.nra
  - selfhost/src/astdump.nra
  - selfhost/src/idcheck.nra
  - selfhost/src/trivia.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/patterns.nra
  - selfhost/src/match.nra
  - selfhost/src/emitmatch.nra
  - selfhost/src/contracts.nra
  - selfhost/src/querysites.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/fixedarrays.nra
  - selfhost/src/diagnostics.nra
  - testdata/astdump/**
  - testdata/fmt/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E0722, diag E0723, diag E0724, diag E0726]
provides: []
depends_on:
  - on: destructuring-design
    kind: hard
    why: builds D103 part 1, which the maintainer approves at the gate
---

# Array patterns in match, is and const

## Goal

D103 part 1. `[]`, `[only]`, `[a, b]`, `[first, ..._]` and `[first, ...rest]` as patterns in `match` arms and `is` tests over `Array<T>` and `FixedArray<T, N>`, with exhaustiveness over lengths, and irrefutable patterns in `const`, `let` and `for ... of`.

- A pattern without `...` matches exactly that length. `...` after `k` elements matches `k` elements or more. `...` at most once and last (E0722). Elements nest and take any pattern, literals included: `[_, "hash", text, output]`.
- On an `Array<T>`, `...rest` binds a new `Array<T>` with the remaining elements, retained: linear time in the tail and one allocation per match that binds it, like a slice (D60). `..._` copies nothing. On a `FixedArray` a named rest is E0724 (its type would need arithmetic on `N`, D95).
- Bound elements are retained, read-only copies (D81). The emitter tests the length once, then reads elements without a bounds check.
- Exhaustiveness: lengths join the usefulness matrix in `patterns.nra` (`specialize`, `isUseful`, `findMissingCases`). `L` is the largest element count in any arm, with or without `...`. The cases are lengths 0 to `L` and "more than `L`". `...` after `k` elements covers every length from `k`, so `[]` with `[_, _, _, ..._]` leaves lengths 1 and 2 missing. Missing cases print as patterns (`[_]`, `[_, _, _, _, ..._]`). Guarded arms do not count. Unreachable arms stay E0983. Or-patterns (D98) work inside and around array patterns.
- Irrefutable: `[..._]` and `[...rest]` alone on an `Array` (any length), so `const [...rest] = items` compiles and copies. Every other pattern on an `Array` is refutable: `const [a, b] = items` and `for (const [a, b] of rows)` are E0715 with the existing message. No new panic (D92).
- On a `FixedArray<T, N>` with a known `N`, a pattern of exactly `N` elements, or `..._` after at most `N`, is irrefutable when its elements are. It works in `const`, `let`, `for ... of`, and nested inside tuple and struct patterns there. Another length is E0723, also in `match` and `is`. `FixedArray<T, 0>` matches `[]`.
- In a generic body where `N` is a type parameter, `const [a, b] = values` is E0723: no single pattern length holds for every `N`. Only `[..._]` is irrefutable there. In `match` and `is` the length test is made per instance, and exhaustiveness needs an arm that covers every length (`[..._]` or `_`).
- An array pattern on any other type (`string`, `bytes`, `Map`, `Range`, a generator, a struct) is E0726.
- `narrac fmt` prints `[a, b]` and `[first, ...rest]` like array literals, breaking one element per line when the pattern does not fit.

## Context

- D103 in DECISIONS.md. SPEC section 4 "Pattern matching" and section 2 "Bindings" get the new rows; section 6 states the rest copy's cost next to the slice rule.
- `Pattern` is the enum in `selfhost/src/ast.nra`. A new variant needs a case in the checker, emitter, `fmt`, `astdump`, `idcheck` and `trivia` switches (roadmap CLAUDE.md).
- E0715 lives in `checkDestructure` in `match.nra`. `isIrrefutable(checker, pattern)` in `patterns.nra:165` takes no type, so it cannot tell an `Array` pattern from a `FixedArray` one. Pass the type in, or read the type `resolvePattern` records for each pattern id, so patterns nested in tuples and structs are judged by their own type.
- `contracts.nra` reads patterns for `ensures`. An array pattern in `ensures` is out of scope here and is reported as E0987 (the `ensures` pattern error).
- `FixedArray` knows `N` per instance (`fixedarrays.nra`, constant type parameters).
- Size L: parser and AST, matrix, emitter (length test, element reads, tail copy), FixedArray checks, fmt and dumps. If it runs long, split FixedArray patterns into a follow-up node and land `Array` patterns in `match` and `is` first.
- Later sweeps, not this node: the argv dispatch in `selfhost/src/main.nra` (15 sites) and the 15 `Ranges.between(1, Arrays.length(x))` tail loops, after the seed has the feature.

## Acceptance

- [ ] Samples: `[]`, `[x]`, `[first, ..._]`, `[first, ...rest]` with the rest used, `const [...rest] = items`, nested and literal elements, a guard, an or-pattern of two array patterns, `is` with array patterns, string arrays.
- [ ] Samples of nesting: `Some([a, ...rest])` in `match`, `#[[x], y]` in `match`, FixedArray patterns in `const`, in `for ... of` and inside a tuple pattern in `const`, `FixedArray<int, 0>` with `[]`.
- [ ] Every sample is clean under the leak check and sanitizers, the rest copies included.
- [ ] Error tests: a non-exhaustive match with the missing pattern printed, a match whose only length cover is a guarded arm, an unreachable array arm (E0983), `...` twice or not last (E0722), a FixedArray length mismatch and `const [a, b]` over a generic `N` (E0723), a named rest on a FixedArray (E0724), an array pattern on a `string` and on a `Map` (E0726), `const [a, b] = items` on an Array (E0715).
- [ ] astdump and fmt goldens for the new pattern. `narrac fmt` is stable on them.
- [ ] SPEC.md sections 2, 4 and 6 describe the forms and the rest copy's cost. D103 gets a "Built" line.
- [ ] Fast check passes: `sh scripts/test.sh`
