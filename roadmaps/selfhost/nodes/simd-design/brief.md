---
id: simd-design
title: "Design: SIMD vector types"
status: done
priority: must
type: spike
story: batch-13
early: false
outline: false
milestone: p1-syntax
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
added: 2026-11-13
spent: 142467
owns:
  - roadmaps/selfhost/nodes/simd/**
  - roadmaps/selfhost/nodes/simd-design/**
shared:
  - DECISIONS.md
  - SPEC.md
  - selfhost/src/diagnostics.nra
reserves: [decision D112]
provides: []
depends_on: []
---
# Design: SIMD vector types

## Goal

Decide the `std/simd` vector types SPEC lists (around line 418: `F32x4`, `F64x2`, `I32x8`) and how they compile to C vector extensions (around line 1667), and write D112.

## Questions

1. Which types (lane types and widths), and whether widths depend on the target.
2. Operations: arithmetic as built-in operators on vectors (D12 allows built-in operators only), comparisons giving masks, lane access, loads and stores from `Array` and `FixedArray`, shuffles, reductions.
3. Portability: clang and gcc vector extensions on x86-64 and arm64, fallback when a width is wider than the hardware.
4. Effects and panics (D92): no panics for lane access with a constant index; proven indexes as for `FixedArray`.
5. A benchmark program for `bench/` that shows the gain over scalar code.

- Reserved codes: E1250-E1269 (add the range to the `diagnostics.nra` header comment and split it between the build briefs).
- Build brief to refine: `nodes/simd/brief.md`.

## Steps

- [ ] Read the sources above and the code the decision changes.
- [ ] Answer the questions in D112: decision, reason, cost, rejected options. Mark it "Proposed, waiting for the maintainer".
- [ ] Write the SPEC text the decision changes, marked as not built yet where the build node has not landed.
- [ ] Refine the build brief listed below from the template: owns, shared, typed `depends_on` (this spike, hard), codes from the reserved range, acceptance with samples and error tests, README changes, `verify: sh scripts/test.sh`. Split any brief larger than L.
- [ ] Hand D112 to the maintainer. Build nodes start only after approval.

## Acceptance

- [ ] D112 answers every question and is approved by the maintainer.
- [ ] The build brief follow D112, or the outcome says why the feature is dropped and which SPEC lines change.
