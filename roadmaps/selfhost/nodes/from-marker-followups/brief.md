---
id: from-marker-followups
title: "Follow-ups from the batch-5 reviews"
status: active
priority: should
type: fix
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
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/from-marker-followups/**
shared:
  - selfhost/src/sharedresults.nra
  - selfhost/src/stdcopy.nra
  - selfhost/src/parser.nra
  - selfhost/src/diagnostics.nra
  - testdata/**
reserves: [diag E0047]
provides: []
depends_on: []
---

# Follow-ups from the batch-5 reviews

## Goal

Four small fixes from `roadmaps/selfhost/inbox.md` (2026-11-10 entries):

- E0976 on an arrow function with no expected type and no written result type (`const f = (r: Row) => r.cells`) blames the function type. Say instead: write the result type with `from r`, or return a copy with `copy(...)` (`selfhost/src/sharedresults.nra`, near line 826).
- The parse error `` `from` needs the name of a parameter `` uses E0976, but E0001 to E0099 are the lexer and parser range. Give it the parser code E0047.
- `buildLocalName` in `selfhost/src/stdcopy.nra` lowercases every capital of a mixed-case constant (`MaxValue` suggests `let maxvalue`) and can suggest a keyword (`const TEST` suggests `let test`). Keep camel case for mixed-case names (`maxValue`) and add a suffix when the result is a keyword.
- Add an error sample for the invalid recursive struct `struct S { a: S, b: S }` that once hung the type-size walk (E0109).

## Acceptance

- [ ] Each fix has an error or diagnostics sample with the new text.
- [ ] `sh scripts/test.sh` passes.
