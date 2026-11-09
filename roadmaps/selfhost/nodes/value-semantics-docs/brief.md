---
id: value-semantics-docs
title: "Document visible copies and shared writes"
status: todo
priority: should
type: docs
story:
early: false
outline: true
milestone: m12-spec-gaps
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: shared-write-error
    kind: hard
    why: documents the final rule
---

# Document visible copies and shared writes

## Goal

Rewrite SPEC.md section 6 (values, copies, `copy`, `from`, the debug check), amend D1, D27 and D81 rule 5 text, and the guide chapter on values with the TypeScript comparison. Drop "Not built yet" from D94.

## Notes

Outline, from D94 and `roadmaps/selfhost/debate/value-semantics/debate.md` (step 6 of 6). Refine with the debate's claims named here.
