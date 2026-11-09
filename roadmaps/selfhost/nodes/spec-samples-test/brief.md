---
id: spec-samples-test
title: "Every SPEC.md sample is checked by a test"
status: active
priority: should
type: feature
story: batch-2
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - selfhost/tests/src/spec*
  - selfhost/tests/spec-skip*
  - roadmaps/selfhost/nodes/spec-samples-test/**
shared:
  - selfhost/src/**
  - std/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: spec-fixes
    kind: hard
    why: the spec text must be right before samples become tests
  - on: c-style-for
    kind: hard
    why: its SPEC samples then pass instead of joining the skip list
  - on: compound-assign
    kind: hard
    why: same
  - on: cell-capture
    kind: hard
    why: same
  - on: std-basics
    kind: hard
    why: same
---

# Every SPEC.md sample is checked by a test

## Goal

Audit learning: F4 and F8 were bugs in features marked done. A test suite extracts the fenced `ts` blocks from SPEC.md and runs `narrac check` on each, with a skip list of samples that need planned features. Removing a skip is part of each later fix node, like the emit skip files were.

- Wrap a fragment that is not a whole module (statements, expressions) in a function before checking. Decide the rule by the first line, and keep it simple; samples that cannot be wrapped go on the skip list with a reason.
- Samples that show an error on purpose need a marker. Use an HTML comment before the block (for example `<!-- error: E0123 -->`), so the rendered spec does not change, and check that the error is the one named.
- The skip list is a file in `selfhost/tests/`, one line per sample: section, first line, and the node or reason.
- Look at how the guide test (`TestGuide` and its port in `selfhost/tests/src`) extracts samples and reuse it.
- `roadmaps/selfhost/nodes/spec-gap-audit/outcome.md` has the findings table with minimal programs.

Where a sample is wrong rather than unsupported, fix the sample in SPEC.md and list the fix in the outcome.

## Acceptance

- [ ] A suite run by `sh scripts/test.sh` checks every fenced `ts` block in SPEC.md, wrapped or skipped.
- [ ] Every skip has a reason; the outcome gives the count of checked and skipped samples.
- [ ] Error samples are checked for their code.
- [ ] `sh scripts/test.sh` passes.
