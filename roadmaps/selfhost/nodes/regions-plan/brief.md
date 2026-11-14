---
id: regions-plan
title: "Plan: regions from D100"
status: todo
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
spent:
owns:
  - roadmaps/selfhost/nodes/regions/**
  - roadmaps/selfhost/nodes/regions-plan/**
shared:
  - DECISIONS.md
  - SPEC.md
  - selfhost/src/diagnostics.nra
reserves: [decision D113, diag E1270-E1279]
provides: []
depends_on: []
---
# Plan: regions from D100

## Goal

D100 parked regions and wrote the design they need; the maintainer moved regions into p1 on 2026-11-13. Check D100 against the code as it is now (size-class allocator, D104, async tasks, `Cell`, `Dynamic`, `lazy`), record what changed as D113, and split node regions into the three children `nodes/regions-design/outcome.md` proposes: region-runtime, region-copy-out and region-check.

## Questions

1. Does each D100 measurement and rule still hold after the allocator and batches 5 to 12? Re-measure the one realistic use (`formatSource` per file) if the numbers moved.
2. New value kinds since D100 that copy-out must handle: `Async<T>` handles, `Dynamic<T>`, lazy values, `FixedArray`, `Ptr`.
3. The two escape errors (captured `let` written inside the region, an identity type in the result) and their codes.
4. SPEC section 6 "Regions" (around line 1177) and D48: the text that changes (`Region.run(() => ...)`).
5. Child briefs: rename `nodes/regions/brief.md` as region-runtime or set it dropped with the children listed.

- Reserved codes: E1270-E1279 (add the range to the `diagnostics.nra` header comment and split it between the build briefs).
- Build brief to refine: `nodes/regions/brief.md`.

## Steps

- [ ] Read the sources above and the code the decision changes.
- [ ] Answer the questions in D113: decision, reason, cost, rejected options. Mark it "Proposed, waiting for the maintainer".
- [ ] Write the SPEC text the decision changes, marked as not built yet where the build node has not landed.
- [ ] Refine the build brief listed below from the template: owns, shared, typed `depends_on` (this spike, hard), codes from the reserved range, acceptance with samples and error tests, README changes, `verify: sh scripts/test.sh`. Split any brief larger than L.
- [ ] Hand D113 to the maintainer. Build nodes start only after approval.

## Acceptance

- [ ] D113 answers every question and is approved by the maintainer.
- [ ] The build brief follow D113, or the outcome says why the feature is dropped and which SPEC lines change.
