---
id: dead-code-check
title: "Unreachable statements are a compile error"
status: done
priority: could
type: feature
story: batch-16
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent: 279923
owns:
  - roadmaps/selfhost/nodes/dead-code-check/**
  - selfhost/src/deadcode.nra
  - testdata/dead_code*
  - testdata/errors/dead_code*
shared:
  - selfhost/src/*.nra
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - compiler/README.md
  - docs/guide/*.md
reserves: [diag E1500-E1509]
provides: []
depends_on: []
---
# Unreachable statements are a compile error

## Goal

A statement that can never run is a compile error, the same way an unused name is. Today `narrac check` accepts code after `return`, after `panic(...)`, after `break` or `continue`, and after an `if`/`else` or exhaustive `match` whose every branch leaves the block. This node replaces the "unreachable code" rule of the parked lint node (see lint, dropped).

## Context

- Severity: every diagnostic is `"severity":"error"` (`selfhost/src/diagnostics.nra`, the JSON writer near line 2074). There is no warning severity, and unused names are errors by design. Follow that rule unless the maintainer picks warnings (open question in batch-16).
- Unreachable `match` arms are already E-coded errors (diagnostics.nra, "this arm is unreachable"). Reuse their wording style.
- Put the analysis in its own module, `selfhost/src/deadcode.nra`, with one dispatch row in the checker.

## Acceptance

- [ ] Statements after `return`, `panic`, `break`, `continue`, and after a branch statement whose every path leaves, are E1500 with the range of the first dead statement only.
- [ ] `narrac apply-fix` offers deleting the dead statements.
- [ ] A loop with no exit (`while (true)` with no `break`) counts as leaving.
- [ ] The selfhost compiler, std and every sample compile clean (fix what the new rule finds in the same node).
- [ ] SPEC.md lists the rule next to "Unused names are errors"; the guide mentions it where it teaches `return`.
- [ ] Error tests in `testdata/errors/dead_code*`.

## Sizing

M: one checker pass over statement lists, one code, and a sweep of existing code.
