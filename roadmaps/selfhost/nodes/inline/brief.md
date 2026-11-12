---
id: inline
title: "@inline for higher-order functions"
status: todo
priority: should
type: feature
story: batch-11
early: false
outline: false
milestone:
target:
ship: auto
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added:
spent:
owns:
  - roadmaps/selfhost/nodes/inline/**
  - selfhost/src/inlining.nra
  - selfhost/src/emitinline.nra
  - testdata/inline*
  - testdata/errors/inline_*
shared:
  - selfhost/src/emit.nra
  - selfhost/src/parser.nra
  - selfhost/src/check.nra
  - selfhost/src/emit.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitclosures.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/diagnostics.nra
  - selfhost/tests/src/emitchecks.nra
  - testdata/astdump/**
  - testdata/fmt/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E1030, diag E1031]
provides: []
depends_on:
  - on: std-function-values
    kind: hard
    why: both change emitIdent (emitexprs.nra:289-302, the Ref.Function line) and readNodeType (emitstate.nra:318); land std-function-values first and reuse one emitter context save and restore helper in emitstate.nra
  - on: closures
    kind: hard
    why: the copied closure argument reads the caller's locals in place of the capture environment that closures built, and must keep its rule that a stack environment never leaves its block
---

# @inline for higher-order functions

## Goal

`@inline` on a function copies its body into each call, and copies the body of an arrow function passed to one of its function-type parameters into the place where the body calls that parameter. The call then has no C call, no closure environment and no indirect call. Today the attribute does not parse:

```
main.nra:1:2: unknown attribute @inline; the known attributes are @target, @trusted, @extern
```

(E0990, from the program below.) The plan: one more row in the attribute table, a checker pass for the rules of D95 (a guarantee, never recursive, never a change in behavior), and an emitter that writes the callee's body at the call site the way block expressions are written.

```ts
@inline
function applyTwice(value: int, transform: (item: int) => int): int {
  return transform(transform(value))
}

export function main(): int {
  const rate = 3
  return applyTwice(2, (item) => item * rate) - 18
}
```

## Context

- Source: SPEC section 2 (`@inline` bullet, around line 126) and section 11 (Closures, around line 1613). D95, "Inline": `@inline` is a guarantee, not a hint; a recursive `@inline` function is an error; it works across packages; inlining never changes behavior.
- SPEC section 11 says `@inline` functions "are emitted as `static inline` C functions". That is not the meaning: C `static inline` is a hint to the C compiler and keeps the closure call. Replace that sentence.
- The attribute list takes `@inline` with no change to its shape. `KNOWN_ATTRIBUTES` in `selfhost/src/parser.nra` (around line 1357) gets `AttributeRow { name: "inline", takesArguments: false }`, like `@trusted`. Code reads it with `findAttribute(funcDecl, "inline")`, as `audit.nra` and `effects.nra` read `@trusted`. The attribute is an `Attribute` node like the others, so there is no new AST kind. `fmt` and `astdump` already print attributes; add one golden each to confirm.
- Two existing rules stay. An implement block accepts only `@trusted` (`parser.nra`, around line 2153), so `@inline` there stays E0990: its calls go through the trait table. Only `@target` is written before an extern block (`parseExternBlock`).
- Checker (`selfhost/src/inlining.nra`, new, called from `check.nra` after the bodies are checked):
  - E1030: an `@inline` function that calls itself, directly or through other `@inline` functions. Walk each `@inline` body for direct calls whose target is a function (`callTarget` in `emittail.nra` shows how to read it), and report the first cycle once, naming the chain (`applyTwice -> helper -> applyTwice`). A call to a function without `@inline` ends the walk, so mutual recursion through an ordinary function is fine.
  - E1031: `@inline` on a function that has no body to copy or cannot be copied: a generator (`function*`), an `async` function, or a member of an `extern { } from` block.
- Emitter (`selfhost/src/emitinline.nra`, new; dispatch from `emitCall` in `emitexprs.nra`):
  - Arguments are evaluated once, left to right, into C temporaries, as the out-of-line call does (node arg-eval-order). A `mutable` parameter binds a pointer to the caller's place, the same C the out-of-line body sees, so the copied body emits unchanged.
  - The copied body runs inside a result temporary and a C label, like a block expression (`BlockInfo` in `emitstate.nra`; `return` there is a `goto` in `emitstmts.nra`, around line 326). A `return` in the copied body stores the result and jumps to the label. The frames and scopes counts in `BlockInfo` release the body's locals on that path.
  - While the body is written, the emitter reads the callee's side tables: save the caller's `body`, `owner`, `scopes`, `frames` and `blocks`, set them as `startFunction` in `emitconsts.nra` does, and restore them after. C names of the copied locals get a suffix per call site, so two sites in one function and an `@inline` call inside an `@inline` body do not clash.
  - An arrow literal argument bound to a function-type parameter is not built. Where the copied body calls that parameter, the emitter switches back to the caller's context and writes the arrow's body with its parameters bound to temporaries. Its captures are the caller's locals themselves: no environment struct, no closure function. A block arrow's own `return` jumps to the arrow's own label, not the callee's.
  - A function-type parameter used other than in a call (passed on to another function) gets a real closure value, built at the call site with `emitArrow` as today. Its stack environment lives in the caller's block at that site, which keeps the closures rule.
  - Any other argument for a function-type parameter (a named function, a local holding a closure) is passed as a value, and the copied body calls it as the out-of-line body would.
  - A generic `@inline` function is written with the call site's type arguments, as `@specialize` does, so no layout table is passed.
  - Reachability (`emitReachableBodies` in `emit.nra`): a call to an `@inline` function does not reach its C function. Only a use as a value does. A program that only calls it has no C function for it.
- Measure against node std-option-result's outcome, which records a capturing callback to `Options.map` as a stack environment and one indirect call per call.

## Acceptance

- [ ] Samples (`testdata/inline*.nra` with `.out`): the program above; a block arrow with its own `return` inside an `@inline` body that also returns early; an `@inline` function with a `mutable` parameter; a generic `@inline` function used with two types; two `@inline` calls in one expression and one nested in another's arrow; an arrow that assigns a captured `let`; a function-type parameter passed on to an ordinary function; an `@inline` function used as a value; `@inline` with `@target` and with `@trusted`; a call across modules.
- [ ] Every sample prints the same output as without `@inline` (check by running each once with the attribute removed while working) and is clean under the leak check and sanitizers.
- [ ] Error tests: direct self-call and a cycle through two `@inline` functions (E1030, chain in the message), a self tail call (E1030), `@inline` on a `function*`, on an `async` function and on an extern block member (E1031), `@inline` in an implement block (E0990).
- [ ] Emit check `inline/no-closure` in `selfhost/tests/src/emitchecks.nra`: the C of the first sample has no C function for `applyTwice`, no closure function and no environment struct for the arrow.
- [ ] A sample measures the per-call cost of a capturing callback through a user `@inline` function (in the shape of `Options.map`) against the same function without `@inline`, recorded in the outcome.
- [ ] astdump and fmt goldens with `@inline`. `narrac fmt` is stable on them.
- [ ] SPEC.md section 2 states the rules (guarantee, E1030, E1031, a value use keeps a C function) and section 11 drops "emitted as `static inline` C functions". D95 gets a "Built" line with the amendments. `compiler/README.md` drops `@inline` from "Not supported yet" and gets a line naming the two new files.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

L. Scores that are not low: emitter state (switching between the callee's and the caller's side tables while one C function is written, nested sites, per-site C names) and the cases that must keep behavior (early returns, `mutable` parameters, generics, a parameter passed on). If it runs long, land non-generic `@inline` with arrow arguments first and split generic sites into a follow-up node.

## Notes for the worker

- Do not emit `static inline`. The C must contain the body at the call site.
- Inlining never changes behavior: evaluation order, copies, releases and effects stay as in the out-of-line call. Copies the out-of-line call makes for a parameter that is written or kept are still made.
- The `copies` and `leaks` suites run every sample; an inlined body that skips a release on its `goto` path shows there.
- Seed rule: do not use `@inline` in `selfhost/src` in this node. The seed does not know it.
- Open question: may an `@inline` function be used as a value? Recommended: yes. The value use gets one ordinary C function, and calls are still copied. D95 asks for a guarantee at calls, and a value has no call site to copy into. The alternative is an error, which makes `@inline` on std helpers break every program that passes them as values.
- Do not mark `std/src` functions `@inline` in this node. The compiler's own source uses std, so std may use `@inline` only after the seed knows it (seed rule). Marking `Options.map`, `andThen` and `orElse` happens after the end-of-batch seed update, as its own step (orchestrator, 2026-11-12).
- Open question: a decision record? Recommended: no new D-number. The rules above are amendments to D95's "Inline" part; write them as its "Built" line, as node contracts did for `requires`.

## Plan review (2026-11-12)

- E1031 covers generators and extern block members only. Top-level `async function` does not parse yet (async is batch-12), so leave the async case and its error test to the async node.
- Reachability: `emitReachableBodies` (emit.nra:177) changes here (calls do not reach an `@inline` body) and in lazy-values (a getter is reached only through `Lazies.force`). Whichever lands second keeps both rules and reruns the `tree-shake` emit checks.
