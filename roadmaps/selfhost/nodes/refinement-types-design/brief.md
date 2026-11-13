---
id: refinement-types-design
title: "Design: refinement types on distinct types"
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
  - roadmaps/selfhost/nodes/refinement-types/**
  - roadmaps/selfhost/nodes/refinement-types-design/**
shared:
  - DECISIONS.md
  - SPEC.md
  - selfhost/src/diagnostics.nra
reserves: [decision D116, diag E1220-E1239]
provides: []
depends_on: []
---
# Design: refinement types on distinct types

## Goal

Decide `distinct type Port = int requires >= 1 && <= 65535` and `distinct type Method = string requires "GET" | "POST"` (inbox, 2026-11-09, from the D95 discussion), and write D116.

## Context

- D95 contracts, D102 (contract logic: `||`, relations), the prover in `selfhost/src/contracts.nra`, distinct types (`opaque-distinct`, D108 traits on distinct types), `std/shape` and `Json.decode`.

## Questions

1. The predicate language (the D95 and D102 patterns).
2. `Port.from(raw): Option<Port>` checks at runtime; where else a value is built (literals proven at compile time, arithmetic on a refined type gives the base type).
3. What the prover knows inside: a `Port` parameter is a fact.
4. `Json.decode` and `Shape`: a failing check is `Err` (standard-schema style).
5. Generics and refined types; `derives`; diagnostics naming the failed predicate.

- Reserved codes: E1220-E1239 (add the range to the `diagnostics.nra` header comment and split it between the build briefs).
- Build brief to refine: `nodes/refinement-types/brief.md`.

## Steps

- [ ] Read the sources above and the code the decision changes.
- [ ] Answer the questions in D116: decision, reason, cost, rejected options. Mark it "Proposed, waiting for the maintainer".
- [ ] Write the SPEC text the decision changes, marked as not built yet where the build node has not landed.
- [ ] Refine the build brief listed below from the template: owns, shared, typed `depends_on` (this spike, hard), codes from the reserved range, acceptance with samples and error tests, README changes, `verify: sh scripts/test.sh`. Split any brief larger than L.
- [ ] Hand D116 to the maintainer. Build nodes start only after approval.

## Acceptance

- [ ] D116 answers every question and is approved by the maintainer.
- [ ] The build brief follow D116, or the outcome says why the feature is dropped and which SPEC lines change.
