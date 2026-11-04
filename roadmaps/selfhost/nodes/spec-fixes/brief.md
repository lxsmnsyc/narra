---
id: spec-fixes
title: "Fix spec text that is wrong, not the compiler"
status: active
priority: should
type: docs
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: S
spike: false
hypothesis:
verify: 
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - roadmaps/selfhost/nodes/spec-fixes/**
shared:
  - SPEC.md
reserves: []
provides: []
depends_on: []
---

# Fix spec text that is wrong, not the compiler

## Goal

The audit found spec text that names things that do not exist. Line 329 uses `F64.toString` and line 534 uses `I32.from`, `F64.fromInt` and `Int.truncate`, but the std list (line 193) has no `std/f64` or `std/i32`; the compiler has `Float.toString`, `Float.fromInt` and `Float.truncate`. Line 578 writes the bare constructor `IoFailed(e)`, which line 222 (D71) makes an error. Line 552: the exported `parsePair` needs `uses alloc`. Fix the text to match the language as decided. Where a sample uses a planned feature (`u8` at line 625, sized-number conversions at line 534), keep the feature but name the real module, and leave a note in the outcome for sized-numbers.

## Acceptance

- [ ] Each item above is fixed in SPEC.md; each fixed sample checks clean with `narra check` (except parts that need a planned feature, listed in the outcome).
- [ ] SPEC.md line 57 (unused locals and parameters are errors) stays as is: the maintainer chose the spec over lint (2026-11-04).
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap` (the fixed point still holds).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
