---
id: uses-list-design
title: "Design: effect lists inside type lists"
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
  - roadmaps/selfhost/nodes/uses-list-syntax/**
  - roadmaps/selfhost/nodes/uses-list-design/**
shared:
  - DECISIONS.md
  - SPEC.md
  - selfhost/src/diagnostics.nra
reserves: [decision D117, diag E1240-E1249]
provides: []
depends_on: []
---
# Design: effect lists inside type lists

## Goal

Fix the `uses` ambiguity from the PR #82 review, and write D117. Outside a parameter list, a `uses` list on a type takes every following `, name` as another effect (`parseUsesList` in `selfhost/src/parser.nra`), so `Map<Generator<int> uses io, int>` parses as one generator with effects `io, int`, and `#[Generator<int> uses io, Array<int>]` is a parse error. A trailing comma after `... uses E` in a parameter list gives `expected a name, found ")"`. Function types have the same problem.

## Questions

1. The syntax: parentheses (`uses (io, alloc)`), `+` between effects, effects only before `=>`, or another form. Keep it TypeScript-like and readable.
2. Where the old form stays valid (a parameter list, a function's own `uses`) and where the new one is required. No deprecation path (the language is unreleased): the old form is removed where it changes.
3. fmt output, the guide, SPEC, std and compiler sources that change.

- Reserved codes: E1240-E1249 (add the range to the `diagnostics.nra` header comment and split it between the build briefs).
- Build brief to refine: `nodes/uses-list-syntax/brief.md`.

## Steps

- [ ] Read the sources above and the code the decision changes.
- [ ] Answer the questions in D117: decision, reason, cost, rejected options. Mark it "Proposed, waiting for the maintainer".
- [ ] Write the SPEC text the decision changes, marked as not built yet where the build node has not landed.
- [ ] Refine the build brief listed below from the template: owns, shared, typed `depends_on` (this spike, hard), codes from the reserved range, acceptance with samples and error tests, README changes, `verify: sh scripts/test.sh`. Split any brief larger than L.
- [ ] Hand D117 to the maintainer. Build nodes start only after approval.

## Acceptance

- [ ] D117 answers every question and is approved by the maintainer.
- [ ] The build brief follow D117, or the outcome says why the feature is dropped and which SPEC lines change.
