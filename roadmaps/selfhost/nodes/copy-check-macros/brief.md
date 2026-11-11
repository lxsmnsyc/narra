---
id: copy-check-macros
title: "The debug copy check emits one line per site"
status: done
priority: should
type: refactor
story: batch-7
early: false
outline: false
milestone: m6-performance
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
spent: 190646
owns:
  - roadmaps/selfhost/nodes/copy-check-macros/**
shared:
  - selfhost/src/emitdecls.nra
  - selfhost/src/emitgenerators.nra
  - selfhost/src/stdcopy.nra
  - selfhost/src/runtimecore.nra
reserves: []
provides: []
depends_on: []
---

# The debug copy check emits one line per site

## Goal

The debug copy check (copy-builtin, shared-write-error) writes its code into every build's C: 487 three-line `#ifndef NARRA_RELEASE ... if (narra_copy_on) { ... } ... #endif` blocks, 1036 `NARRA_COPY_GRANT(true/false)` lines around retains, and 76 `credit_`/`uncredit_` functions in the seed. Fold each site into one macro call defined in the runtime (`NARRA_COPY_CREDIT(...)`, a grant macro that wraps the retain), so release builds see one line that expands to nothing. Keep the check's behavior in debug builds exactly. The sites that write this code are few: `emitdecls.nra` (the credit and uncredit functions and their prototypes), `stdcopy.nra` (the grant lines and the `if (narra_copy_on)` wrapper), `emitgenerators.nra` (`takeOwnCopy`), and the macro definitions in `runtimecore.nra`. Say in the outcome whether the grant macro wraps the retain or stays two statements.

## Context

- The seed is one C file for both builds (`-DNARRA_RELEASE` picks), so the code must stay in the output; only its size changes.
- Runtime text lives in `selfhost/src/runtime*.nra`. GOAL.md: editing runtime text may need two seed updates; say in the outcome whether it does.
- Inbox item for this area: `takeOwnCopy` in `emitgenerators.nra` writes `NARRA_COPY_GRANT` without the guard. Fold it in.

## Acceptance

- [ ] Every `copy_*` sample and every program sample under `NARRA_CHECK_COPIES=1` reports the same as before (the test runner already runs them that way).
- [ ] The outcome records lines and bytes of `narrac emit selfhost` before and after, with and without `unifdef -DNARRA_RELEASE`.
- [ ] Fast check passes: `sh scripts/test.sh`
