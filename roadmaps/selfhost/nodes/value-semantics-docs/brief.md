---
id: value-semantics-docs
title: "Document visible copies and shared writes"
status: active
priority: should
type: docs
story: batch-6
early: false
outline: false
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
owns:
  - roadmaps/selfhost/nodes/value-semantics-docs/**
shared:
  - SPEC.md
  - DECISIONS.md
  - docs/guide/**
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: shared-write-error
    kind: hard
    why: documents the final rule
---

# Document visible copies and shared writes

## Goal

D94 step 6. Describe the final rule where readers look for it:

- SPEC.md section 6: the shared-write rule, with checked samples for an error and its `copy(...)` fix.
- DECISIONS.md: D81 rule 5 and D27 updated to point to D94; D94's status line says everything is built.
- The guide (`docs/guide/`): the memory chapter explains why a write to a shared value is an error and how `copy(...)` fixes it, with a sample the guide suite checks. Follow `docs/guide` rules: no D-numbers and no assumed TypeScript knowledge outside the TypeScript chapter, which gets a short comparison with JavaScript's shared references.
- shared-write-error edited five guide samples (chapters 2, 4, 8, 9, 11) to compile under the error, but left the prose that says the compiler cannot warn about shared copies. Rewrite that prose.
- State which index places count as the same (different literal indexes are separate; identifier indexes and map keys match every element), as the D94 amendment does.

## Acceptance

- [ ] The samples added to SPEC.md and the guide pass the spec and guide suites.
- [ ] `sh scripts/test.sh` passes.
