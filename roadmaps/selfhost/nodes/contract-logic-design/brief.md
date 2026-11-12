---
id: contract-logic-design
title: "Spike: logical operators in requires and ensures"
status: todo
priority: should
type: spike
story: batch-11
early: false
outline: false
milestone: 
target:
ship: auto
from:
size: M
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/contract-logic-design/**
shared:
  - DECISIONS.md
reserves: [decision D102]
provides: []
depends_on:
  - on: contracts-ensures
    kind: hard
    why: the design extends the requires and ensures rules and the prover as built in batch-7
---

# Spike: logical operators in requires and ensures

## Goal

Maintainer request, 2026-11-11. Contracts (D95, SPEC section 13) have only part of the logical operators:

- `requires` joins integer comparisons with `&&` only. `||`, two parameters compared (`low < high`), a field path, a call and a float parameter are E0984. `!` is not named in SPEC; the condition parser (`selfhost/src/contracts.nra`) accepts no `!`.
- `ensures` is a pattern. `|` gives OR inside a pattern; AND exists only across lines (`ensures Some(>= 0) | None` plus `ensures Some(< 64) | None`), and there is no `&&` inside a pattern (`Some(>= 0 && < 64)`).

The prover keeps one interval of facts per integer name (the D92 division proofs, D95 contracts), so `x < 0 || x > 10` has no place to live today. Decide, as D102:

1. `||` in `requires`: what a caller must prove (one side, from its own facts), what the body knows (nothing per name, or a set of intervals), and the cost of storing more than one interval per name in the prover.
2. Comparisons between two parameters (`requires low <= high`): a relation fact between names, how the body uses it (for example `high - low` cannot be negative), and how a caller proves it.
3. `!` in `requires`: allowed as negation of a comparison only, or not at all (it is always rewritable).
4. `&&` inside an `ensures` pattern (`Some(>= 0 && < 64)`), or a range pattern form, against keeping several lines.
Out of scope (maintainer, 2026-11-11): field paths and floats in conditions and facts. Bind a field to a local first (`const value = foo.value`), which proves today.

For each: the syntax, the facts kept, the errors (codes from the diagnostics table at build time), the effect on proof cost (`narrac emit selfhost` instructions), and samples. Prefer the smallest set that keeps facts as intervals unless a real program in `selfhost/src` or `std/src` needs more; count how many existing checks would use each form.

## Acceptance

- D102 in `DECISIONS.md`, approved by the maintainer (gate).
- The child briefs for node contract-logic, or the reason it should wait.
