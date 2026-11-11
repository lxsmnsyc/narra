---
id: c-audit-target
title: "narra audit and @target on extern functions"
status: done
priority: should
type: feature
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
spent: 141625
owns:
  - roadmaps/selfhost/nodes/c-audit-target/**
  - selfhost/src/audit.nra
shared:
  - selfhost/src/lexer.nra
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/check.nra
  - selfhost/src/fmt.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/idcheck.nra
  - selfhost/src/trivia.nra
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
  - selfhost/tests/spec-skip.txt
  - selfhost/src/driver.nra
  - selfhost/src/buildconditions.nra
  - selfhost/src/main.nra
reserves: [diag E1003]
provides: []
depends_on:
  - on: c-interop
    kind: hard
    why: audits extern blocks and puts @target on them; unsafe-ptr comes in through c-interop
---

# narra audit and @target on extern functions

## Goal

Finish C interop step 1: `narrac audit` lists every `unsafe` block and every `@trusted` function in a package with its position. `@target` on an extern function (or block) links and declares it only on that target, completing the form build-targets left for C interop (SPEC section 12 sample `GetTickCount64`, the `extern` line in `selfhost/tests/spec-skip.txt`).

## Context

- The command: add `audit` to the dispatch in `main.nra` (next to `hash`, `fmt`, `query`), implemented in a new `audit.nra` that walks each module's AST in position order. Output, one line each, sorted by file, line and column: `src/x.nra:12:3: unsafe block in function name` and `src/x.nra:4:1: @trusted function name`. Exit 0. SPEC writes `narra audit`; the binary is `narrac`, as for every other command.
- `@target` on a block: the block-level attribute from c-interop. The per-target check is `checkTargetUse` in `buildconditions.nra`, and the per-target grouping of E0991 is in `driver.nra` (`--all-targets`). An extern member of a `@target` block follows the same rules as a `@target` function, and its library is linked only on that target.

## Acceptance

- [ ] Samples: `narrac audit` output for a package with both kinds (golden), a `@target(os = Os.Windows)` extern function used in a guaranteeing branch (no link error on the host), and E0991 outside one; the rewritten SPEC section 12 sample (`@target(os = Os.Windows) extern { function GetTickCount64(): u64 } from "kernel32"`) passes the spec suite on macOS and Linux and its skip line is removed.
- [ ] Fast check passes: `sh scripts/test.sh`
