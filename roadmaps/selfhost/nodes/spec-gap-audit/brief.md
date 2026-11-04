---
id: spec-gap-audit
title: Audit SPEC.md against the compiler
status: done
priority: should
type: chore
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: M
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-04
spent: 163504
owns:
  - roadmaps/selfhost/nodes/spec-gap-audit/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Audit SPEC.md against the compiler

## Goal

Find every place where `SPEC.md` promises syntax or behavior that the Go compiler (`narra`) does not accept or gets wrong. One gap is already known: `for (const #[i, item] of Arrays.entries(items))` (SPEC.md line 110) fails to parse with `expected a name, found "#["`. That is node for-destructure; do not fix it here. The maintainer expects more gaps of the same kind.

## Method

- Go through SPEC.md section by section. For each code block and each syntax rule stated in prose, write a minimal program in a scratch directory (not in the repo) and run `go run ./cmd/narra check` or `run` from `compiler/`.
- Mark parts the roadmap already plans or parks (generators, async, threads, regions, SIMD, fixed arrays, c-interop, build conditions, std/json, std/iterator, std/arena, std/unicode, Cell, closure-borrow, compound assignment) as "planned: <node>". Do not test them in depth.
- Also check the "Differences from the spec" list in `compiler/README.md`: a recorded difference is not a gap.
- Note only gaps the Go compiler has. The self-hosted compiler matches Go byte for byte, so its gaps are the same.

## Acceptance

- [ ] `outcome.md` holds a table of findings: spec line, construct, minimal program (inline), actual error or wrong output, and a size guess (S/M/L).
- [ ] Each finding is grouped into a proposed fix node (title, size, files it would touch). Related gaps share one node. Say for each whether it is parser-only (a desugaring) or reaches the checker and emitter.
- [ ] Every SPEC.md section is listed as checked, planned (with node) or recorded difference, so nothing is skipped silently.
- [ ] No code changes outside the node folder. The PR adds only `outcome.md` and `log.md`.

## Notes

The orchestrator writes the fix nodes from this list after the maintainer reviews it. Fixes go after selfhost-bootstrap and must be built in both the Go compiler and `selfhost/`.
