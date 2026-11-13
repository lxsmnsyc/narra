---
id: parameter-defaults-design
title: "Design: parameter defaults"
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
  - roadmaps/selfhost/nodes/parameter-defaults/**
  - roadmaps/selfhost/nodes/parameter-defaults-design/**
shared:
  - DECISIONS.md
  - SPEC.md
  - selfhost/src/diagnostics.nra
reserves: [decision D115, diag E1210-E1219]
provides: []
depends_on: []
---
# Design: parameter defaults

## Goal

Decide whether functions get default parameter values, and write D115. D12 rejected optional parameters ("a struct with defaults covers optional settings with one rule"), and SPEC says "no varargs and no optional parameters" (around line 89); D87 added named arguments with every argument passed. The maintainer put defaults in p1 on 2026-11-13, so the spike weighs them against D12 and D87 with evidence and recommends one answer. Run it as a short goal debate (agent-skills:goal-debate) if the answer is not clear.

## Questions

1. Allowed or not. If allowed: which parameters (trailing only, or any with named arguments), what a default may be (compile-time constants as struct field defaults, or any expression evaluated at the call).
2. Function types and function values: does `(a: int, b?: int) => int` exist, or does a value of such a function take every argument.
3. Interaction with named arguments (D87), generics, `mutable` parameters, traits and `extern`.
4. Effects of a default expression and where they are charged.
5. What changes in SPEC, the guide and the std APIs that use option structs today.

- Reserved codes: E1210-E1219 (add the range to the `diagnostics.nra` header comment and split it between the build briefs).
- Build brief to refine: `nodes/parameter-defaults/brief.md`.

## Steps

- [ ] Read the sources above and the code the decision changes.
- [ ] Answer the questions in D115: decision, reason, cost, rejected options. Mark it "Proposed, waiting for the maintainer".
- [ ] Write the SPEC text the decision changes, marked as not built yet where the build node has not landed.
- [ ] Refine the build brief listed below from the template: owns, shared, typed `depends_on` (this spike, hard), codes from the reserved range, acceptance with samples and error tests, README changes, `verify: sh scripts/test.sh`. Split any brief larger than L.
- [ ] Hand D115 to the maintainer. Build nodes start only after approval.

## Acceptance

- [ ] D115 answers every question and is approved by the maintainer.
- [ ] The build brief follow D115, or the outcome says why the feature is dropped and which SPEC lines change.
