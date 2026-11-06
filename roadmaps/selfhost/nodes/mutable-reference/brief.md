---
id: mutable-reference
title: "mutable as a second-class mutable reference"
status: todo
priority: should
type: spike
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: S
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-04
spent:
owns:
  - roadmaps/selfhost/nodes/mutable-reference/**
shared:
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [decision D81]
provides: []
depends_on:
  - on: mutable-read-conflict
    kind: hard
    why: the soundness fix ships first; the redesign builds on what it learned (done)
---

# mutable as a second-class mutable reference

## Goal

The maintainer (2026-11-04): "mutable variable shouldn't be a thing anyways, it should be a mutable reference." Write the design below as D81 and SPEC text, measure what it would change in `selfhost/src`, and stop at the gate. Node mutable-bindings builds it.

## Context

- Today: `mutable` marks a parameter and is repeated at the call (`damage(mutable hero, 10)`). The compiler passes a pointer (D50), access is exclusive during the call (E0319), the variable is not used elsewhere in the statement (E0929, `nodes/mutable-read-conflict/outcome.md`), and the parameter cannot escape (E0915, `nodes/mutable-escape/outcome.md`). closure-borrow lets stack closures capture `let` and `mutable` by reference (D77).
- In `selfhost/src`, 126 `mutable` arguments are paths such as `mutable checker.bodies[index]` or `mutable emitter.frames[index]`, repeated at every call. There is no way to name such a place once, as TypeScript code does with `const frame = emitter.frames[index]` and then writes `frame.x = 1`.

## Plan (D81)

`mutable` is a mutable reference: a name for a place owned by someone else. It stays second class, as in Swift `inout` and Hylo, never a type.

1. Words: SPEC section 6 "Mutable parameters" becomes "Mutable references". A `mutable` parameter is a mutable reference; `mutable x` at a call passes one. Behavior of today's programs does not change.
2. New: a local reference binding, `mutable frame = emitter.frames[index]`. The right side is a place: a `let` variable, a `mutable` parameter or binding, or a field or element path of one, the same as a `mutable` argument. Reads and writes of `frame`, its fields and elements, and `mutable frame` arguments go to that place. An element index is evaluated once, at the binding, and the array is made unique there (copied when shared), as for a `mutable` argument.
3. Borrow: from the binding to the end of its block, the place's root cannot be used through any path that overlaps the place: a prefix of it (`emitter`, `emitter.frames`) or an extension (`emitter.frames[j].x`; element indexes always count as overlapping). Disjoint paths stay usable (`emitter.count`). A nested block `{ ... }` ends the borrow sooner.
4. No escape: a reference binding cannot be stored, returned, captured by an escaping closure or passed to an escaping parameter (E0915 rules). A stack closure may capture it by reference (D77). `const copy = frame` copies the value.
5. `x = f(mutable x)` and `x.count = f(mutable x)` become valid: the call returns before the assignment writes, so the emitter evaluates the right side into a temporary first. The rest of E0929 stays.
6. No reference types (`&mut T`), no references in fields or arrays, no functions that return references.

## Acceptance

- [ ] DECISIONS.md D81 with the plan, reasons, prior art (Swift `inout` and exclusivity, SE-0176; Hylo `inout` bindings and projections; C# `ref` locals; Rust `&mut`) and the alternatives below.
- [ ] SPEC section 6 rewritten as "Mutable references", with the binding form and its rules, and section 2 "Bindings" naming `mutable` next to `const` and `let`. `compiler/README.md` "Differences from the spec" lists rules 2, 3 and 5 as not built yet; mutable-bindings removes the line.
- [ ] The outcome counts the `selfhost/src` functions where a binding would remove a repeated path (`mutable checker.x`, `mutable emitter.frames[...]`), with two before and after examples.

## Sizing

S. Documents and a count; the research is in this brief.

## Decision to confirm

Plan: keep `mutable` second class and add local reference bindings (`mutable name = place`, lexical borrow) and `x = f(mutable x)`.

- Reason: second-class references need no lifetimes in signatures, so the checker stays a local scan in both compilers. The rules extend the existing ones (E0319, E0929, E0915) instead of adding a borrow checker.
- Reason: it gives TypeScript's habit of naming a nested object and changing it through the name, without changing value semantics: a `const` copy is still a copy.
- Reason: a lexical borrow is predictable and easy to explain; ending at the last use (Rust's non-lexical lifetimes) needs a liveness pass in both compilers and can come later without breaking programs.
- Rejected: first-class `&mut T` with a borrow checker (Rust). Lifetimes in signatures and in structs, the largest cost in both compilers, and far from TypeScript.
- Rejected: C# `ref` locals and `ref` returns. Returns need escape rules ("safe to escape") in every signature.
- Rejected: Hylo-style projections (functions that yield a reference). Useful for accessors such as "the current frame", but they need a new function kind; possible later on top of this plan.
- Rejected: keeping today's design. Every write to a nested place repeats the full path, and `x = f(mutable x)` stays an error with no reason a user can see.
