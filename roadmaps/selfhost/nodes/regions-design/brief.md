---
id: regions-design
title: "Spike: how regions skip reference counts and stay safe"
status: todo
priority: should
type: spike
story: batch-4
early: false
outline: false
milestone: m4-systems
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
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/regions-design/**
shared:
  - DECISIONS.md
reserves: [decision D100]
provides: []
depends_on:
  - on: closures
    kind: hard
    why: Region.run takes a closure
---

# Spike: how regions skip reference counts and stay safe

## Goal

Answer the open questions in D48 and SPEC section 6 "Regions" (SPEC.md lines 902-920) before anyone builds `std/region`, and write the answer as decision D100 plus refined child briefs for node regions. No compiler code ships from this node; a small throwaway C prototype under the node folder is fine.

## Questions

1. **Skipping counts.** Region allocations could carry a negative count, like string literals (`lstr`, `selfhost/runtime/runtime.h:180-204`) and stack closure envs (rc -1). Today enum retain is `if (v.u) v.u->rc++` with no sign check, and array and box retains are similar. What does a sign check cost on the `bench/` programs outside regions, and is a per-thread "in region" flag at `narra_alloc` (`runtime.h:65`) enough to route allocations?
2. **Escape.** SPEC says a region cannot escape, "with the same rules as a `Scope`", but nothing implements `Scope`. Closures capture by reference (D77), so the body of `Region.run` can assign a region value to an outer `let`, push it into an outer array, or store it in a `Cell`. Which writes does the checker forbid, which ones copy, and what error codes are needed?
3. **Copying out.** The result of `f` is copied to the normal heap, but the emitter has no deep copy today (enum retain only bumps `u->rc`). Which types need generated deep-copy functions, and what does a copy cost for the compiler's use (a parse tree summarised per file)?
4. **Growth and runtime buffers.** `narra_realloc` (array growth) and raw `malloc` in `selfhost/runtime/std_*.c` bypass `narra_alloc`. How do they behave inside a region?
5. **Value to the compiler.** The compiler keeps its AST and types for the whole compile, so name a real phase that has a clear end (for example parsing or checking one module and keeping only a summary, or one `narrac fmt` file). Estimate the gain from the bench-compiler numbers. If no phase fits, say so; the node then stays a user feature with priority could.

## Acceptance

- [ ] D100 written with the answers, the costs and the rejected options.
- [ ] Node regions refined into child briefs (or dropped) from D100, with sizes.
- [ ] The maintainer has read D100 before any child starts (gate).

## Sizing

M: research and a prototype, no shipped code; uncertainty high, which is why it is a spike.
