---
id: const-eval-design
title: "Design: how @const functions run at compile time"
status: todo
priority: should
type: spike
story: batch-11
early: false
outline: false
milestone:
target:
ship: auto
land: auto
from:
size: S
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-12
spent:
owns:
  - roadmaps/selfhost/nodes/const-eval-design/**
  - roadmaps/selfhost/nodes/const-functions/**
shared:
  - DECISIONS.md
reserves: [decision D105]
provides: []
depends_on: []
---

# Design: how @const functions run at compile time

## Goal

Decide how a top-level `const` can call an `@const` function whose result the compiler computes, and write it as D105 for the maintainer to approve. Today the attribute does not parse:

```
main.nra:1:2: expected a name, found "const"
```

(E0012, from `@const function square(value: int): int { return value * value }` and `const area = square(4)`.) Without the attribute, the call is E0192 "constant cannot call function". The spike ends with D105 approved and a refined brief for a `const-functions` node, or a reason to drop the feature.

## Context

- Source: SPEC section 10 (around lines 1590-1596: "a call to an `@const` function" is a valid top-level value) and section 12 (around line 1633: compile-time evaluation is limited to top-level constants, `@const` functions and build conditions). Split from const-lazy on 2026-11-12; `lazy` values go to node lazy-values.
- Inbox, 2026-11-07, user approved for planning: compile-time evaluation of pure functions. A function whose inferred or declared effects are empty (D58) can run at compile time. Open there: an interpreter in the compiler or compiling and running C, limits on time and memory, which builtins count as pure, `alloc`, and how results are written into the generated C.
- What exists: D61 builds non-scalar top-level constants on first use behind a getter (`selfhost/src/emitconsts.nra`), with no static C data for counted values. Constant expressions are checked by `checkConstValue` in `selfhost/src/exprs.nra` (E0192, E0193). Effects are inferred by a fixed point (`selfhost/src/effects.nra`, D58). Build conditions already fold `Build.os` and friends (`selfhost/src/buildconditions.nra`). Node lazy-values covers the case where computing at first use is enough.
- D92: a program panics only through `panic`, `assert`, an index out of range, out of memory and stack overflow. At compile time each of these must become a compile error at the constant.

## Questions D105 must answer

Each has a recommended answer. The spike checks it, changes it if the evidence says so, and writes the result into D105.

1. **Interpreter or compiled C?** Recommended: a tree-walking interpreter over the checked AST in the compiler (`selfhost/src/consteval.nra`). Compiling and running C needs `cc` during `narrac check` and in the LSP, runs host code when building for another target, and makes the build depend on a program's runtime. The cost is a second semantics to keep equal to the emitted C. Answer that with a test that runs each `@const` sample both at compile time and at runtime and compares.
2. **Marked or inferred?** Recommended: marked. Only an `@const` function may be called from a top-level constant, and its body is checked for the rules below where it is declared. An exported function's body can change without changing its signature; inference would let such an edit break a constant in another module. SPEC already writes `@const`.
3. **Which effects and builtins count?** Recommended: an `@const` function's effects are at most `alloc` (E-code in the node). Allowed: arithmetic, comparisons, `std/string`, `std/array`, `std/map`, `std/option`, `std/char`, `std/int` and the other number modules, and calls to other `@const` functions. Not allowed: `io`, `unsafe`, `std/random`, `std/time`, `std/env`, `std/fs`, `std/process`, `Ptr`, extern functions, closures that escape, generators and async.
4. **`alloc`?** Recommended: allowed in `@const`. It is an effect in D58 but has no outside observer at compile time, and tables (the main use) need arrays.
5. **Limits.** Recommended: a step budget (for example 10 million evaluation steps per top-level constant), a memory budget (for example 64 MiB of interpreter values) and a call depth limit. Each is a compile error at the constant with the call chain. The spike measures one realistic sample (a 256-entry CRC table, a keyword map) against the budgets in a capped prototype or by counting steps by hand, and states the numbers.
6. **Panics.** Recommended: `panic`, a failed `assert`, an index out of range and a budget overrun are compile errors at the constant, showing the message and the call chain.
7. **Writing results into C.** Recommended: the evaluator turns the result back into a constant expression (literals, constructors, struct and array literals, map literals), and the existing D61 path emits it: scalars at each use, others behind the getter. Static C data for counted values stays out, as D61 decided. State what this costs at startup for a large table, and when static data would be worth a node.
8. **Where calls are allowed.** Recommended: top-level `const` values only, in the first node. Not in types (D95: no arithmetic on `N` in types), not in build conditions, not in `lazy` values (they run at first force anyway). An `@const` function is also an ordinary function at runtime.
9. **Determinism and the build model.** Recommended: results depend only on the source. Map iteration order and float formatting must match the runtime exactly. Section 12 asks for byte-identical output; say how the interpreter keeps that.
10. **Tooling.** Recommended: `narrac check` and the LSP evaluate constants, since an evaluation error is a compile error. Cache results per constant within one run; incremental caching waits for interface files.

## Steps

- [ ] Read the sources above and the parts of `exprs.nra`, `emitconsts.nra` and `effects.nra` that D105 changes.
- [ ] Answer the ten questions. For question 5, measure or count; do not guess the budgets.
- [ ] Write D105 in DECISIONS.md: decision, reason, cost, and what is rejected (compiled C, inferred purity) with why.
- [ ] Write `roadmaps/selfhost/nodes/const-functions/brief.md` as a full brief from the template: `outline: false`, owns, shared, typed `depends_on` (`const-eval-design`, hard), codes left for the orchestrator to reserve, acceptance with samples and error tests, SPEC and README changes, `verify: sh scripts/test.sh`. Split it if it is larger than L. If the answer is to drop the feature, write the reason in the outcome instead and leave no brief.
- [ ] Hand D105 to the maintainer. Dependents start only after approval.

## Acceptance

- [ ] D105 in DECISIONS.md answers the ten questions and is approved by the maintainer.
- [ ] `roadmaps/selfhost/nodes/const-functions/brief.md` is a full brief that follows D105, or the outcome states why the feature is dropped and which SPEC lines change (sections 10 and 12).
- [ ] The outcome records the measured or counted cost of the sample against the chosen budgets.

## Sizing

S. A design record and one brief; the only code is an optional capped prototype for the budgets, thrown away after.

## Notes for the worker

- Do not build the feature. Do not edit `selfhost/src`, SPEC.md or compiler/README.md in this node.
- D61 says "a constant may not call a function". D105 amends it for `@const` calls; say so in D105.
- Keep D12 in view: an `@const` call at a top-level constant is a visible call, so no call is hidden.
