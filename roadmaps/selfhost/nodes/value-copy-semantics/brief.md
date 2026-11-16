---
id: value-copy-semantics
title: "Assigning a shared array: alias, silent copy, or error"
status: done
priority: should
type: spike
story: 
early: false
outline: false
milestone: p3-review
target:
ship: auto
land: auto
from:
size: M
spike: true
hypothesis:
verify: 
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/value-copy-semantics/**
  - roadmaps/selfhost/debate/value-copy/**
shared:
reserves: []
provides: []
depends_on: []
---
# Assigning a shared array: alias, silent copy, or error

## Goal

Maintainer, 2026-11-16, on

    const prices = [100, 200];
    const synonym = prices;
    Arrays.push(mutable prices, 300);
    print(Int.toString(synonym[0]));

"this code snippet is still counter-intuitive. Why can't `synonym` refer to `prices`' value like every other language does? why does it imply copying?" and "why is it a bug when that's what the user knows?"

Today arrays are values. The push must copy on write so `synonym` keeps `[100, 200]`, and the rule that every such copy is written as `copy(...)` makes it error E0978. Decide, as a goal debate with the maintainer (goal-debate skill), between:

- (a) Swift model: copy on write happens without an error. Keep `NARRA_CHECK_COPIES`, maybe warn on copies inside loops.
- (b) Keep the error, with a message that says why.
- (c) Reference semantics for arrays (and maps), as TypeScript and Go. The user's mental model, but it touches the memory model: `mutable` parameters, in-place reuse (part of the benchmark results), sending values between threads, and the "values never change behind your back" claim on the home page.

Weigh each against the TypeScript learner, the costs (compiler work in E0976 and E0978 and the shared-write rules, guide chapters 2 and 9, the home page section 02, benchmarks), and safety. Measure, do not assume: count E0978 sites in `selfhost/src` and `testdata`, and run the benchmarks for any option that changes copies.

## Acceptance

- [x] debate.md with the options, evidence and the settled choice; the maintainer approves.
- [x] A decision record and follow-up nodes for the chosen option.

Settled 2026-11-16: D124, debate `roadmaps/selfhost/debate/value-copy/debate.md` (approved). Follow-up node: restrict-renames.
