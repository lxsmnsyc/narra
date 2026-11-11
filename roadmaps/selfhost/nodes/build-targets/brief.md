---
id: build-targets
title: "@target declarations and check --all-targets"
status: active
priority: should
type: feature
story: batch-7
early: false
outline: false
milestone:
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
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/build-targets/**
shared:
  - selfhost/src/lexer.nra
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/check.nra
  - selfhost/src/fmt.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/diagnostics.nra
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
  - docs/guide/**
  - selfhost/src/declare.nra
  - selfhost/src/resolve.nra
  - selfhost/src/driver.nra
reserves: [diag E0990, diag E0991]
provides: []
depends_on:
  - on: build-conditions
    kind: hard
    why: needs the target table and the build-constant folding
---

# @target declarations and check --all-targets

## Goal

SPEC section 12: `@target(os = Os.Windows)` on a declaration makes it exist only on that target. Calling it outside a branch that guarantees that target is a compile error. `narrac check --all-targets` type-checks the package for every target in the table. The `extern function` form in the SPEC example waits for c-interop; this node covers Narra declarations only and says so in SPEC.

## Acceptance

- [ ] Samples: a call inside a guaranteeing branch (ok), outside (error), and `check --all-targets` finding an error only on a non-host target.
- [ ] Fast check passes: `sh scripts/test.sh`
