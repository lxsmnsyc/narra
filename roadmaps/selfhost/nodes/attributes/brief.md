---
id: attributes
title: "General declaration attributes"
status: todo
priority: should
type: refactor
story: batch-8
early: false
outline: false
milestone:
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
  - roadmaps/selfhost/nodes/attributes/**
shared:
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/fmt.nra
  - selfhost/src/trivia.nra
  - selfhost/src/buildconditions.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/check.nra
  - testdata/**
  - SPEC.md
reserves: []
provides: []
depends_on: []
---

# General declaration attributes

## Goal

Batch-7 added `@target` as the only attribute: `parseTargetAttribute` (`parser.nra`) rejects any other name with E0990 "the only attribute is @target", the AST holds one `FuncDecl.target: Option<TargetAttribute>` (`ast.nra`), `fmt` prints it with `printTargetAttribute`, and `buildconditions.nra` (`resolveFunctionTarget`) checks it. Make attributes a general list on declarations, so the next nodes add `@trusted` (unsafe-ptr) and `@extern("name")` (c-interop) as a name and a check, not another parser path. Behavior is unchanged: `@target` works as before, and an unknown attribute stays E0990 with a message that lists the known ones.

Also add a range row for E1000 to E1099 to the diagnostics table header (it lists ranges up to E0999); c-interop and c-audit-target use codes from it.

## Acceptance

- [ ] Every `@target` sample, fmt sample (comments around the attribute included) and astdump golden gives the same result; the E0990 goldens change only in their message text.
- [ ] Fast check passes: `sh scripts/test.sh`
