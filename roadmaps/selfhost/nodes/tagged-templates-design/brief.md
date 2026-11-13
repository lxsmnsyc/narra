---
id: tagged-templates-design
title: "Design: tagged template literals"
status: active
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
  - roadmaps/selfhost/nodes/tagged-templates/**
  - roadmaps/selfhost/nodes/tagged-templates-design/**
shared:
  - DECISIONS.md
  - SPEC.md
  - selfhost/src/diagnostics.nra
reserves: [decision D114, diag E1200-E1209]
provides: []
depends_on: []
---
# Design: tagged template literals

## Goal

Decide `tag\`text ${value}\`` with TypeScript's meaning (approved for planning 2026-11-07), and write D114. The tag is a function named at the use site, so no call is hidden (D12).

## Questions

1. The tag's signature: an `Array<string>` of parts plus the values, and how values of different types pass without varargs (D12). Candidates: one value type per tag (`Array<T>`), a `Shape`-based `Value` per hole, or the compiler calling one tag function per value type.
2. Raw strings (`String.raw`-like) and escapes in parts.
3. Effects and `@const` tags (a compile-time checked tag, for example for regex or SQL).
4. Whether std ships tags (JSON, shell quoting, `std/terminal` styles) and where.
5. fmt and the language service for tagged templates.

- Reserved codes: E1200-E1209 (add the range to the `diagnostics.nra` header comment and split it between the build briefs).
- Build brief to refine: `nodes/tagged-templates/brief.md`.

## Steps

- [ ] Read the sources above and the code the decision changes.
- [ ] Answer the questions in D114: decision, reason, cost, rejected options. Mark it "Proposed, waiting for the maintainer".
- [ ] Write the SPEC text the decision changes, marked as not built yet where the build node has not landed.
- [ ] Refine the build brief listed below from the template: owns, shared, typed `depends_on` (this spike, hard), codes from the reserved range, acceptance with samples and error tests, README changes, `verify: sh scripts/test.sh`. Split any brief larger than L.
- [ ] Hand D114 to the maintainer. Build nodes start only after approval.

## Acceptance

- [ ] D114 answers every question and is approved by the maintainer.
- [ ] The build brief follow D114, or the outcome says why the feature is dropped and which SPEC lines change.
