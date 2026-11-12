---
id: hidden-copy-fix
title: "The hidden copy in loadFile under the debug copy check"
status: done
priority: should
type: fix
story: batch-9
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
spent: 159000
owns:
  - roadmaps/selfhost/nodes/hidden-copy-fix/**
shared:
  - selfhost/src/modules.nra
  - selfhost/src/emitgenerators.nra
  - selfhost/src/sharedwrites.nra
  - testdata/**
  - SPEC.md
reserves: []
provides: []
depends_on: []
---

# The hidden copy in loadFile under the debug copy check

## Goal

Inbox, from build-targets: a compiler built without `-DNARRA_RELEASE` trips the debug copy check (`NARRA_CHECK_COPIES`) in `loadFile` (`selfhost/src/modules.nra`) when the test runner runs it. That is a hidden copy in the compiler's own source that the static shared-write rule (E0978) did not catch. Find it. If E0978 should have caught it, fix the rule and the source; if it is a case the rule cannot see, reduce it to a small sample and add it to the known gaps in SPEC section 6. Then a debug-built compiler must run the test suites without exit 103.

Also from the shared-write-error inbox item: `takeOwnCopy` in `selfhost/src/emitgenerators.nra` writes `NARRA_COPY_GRANT` without the `#ifndef NARRA_RELEASE` guard the other sites use; add it.

## Acceptance

- Fix proof: the exit 103 on main with the command that shows it (in `log.md`), gone on the branch.
- A sample for the reduced case (an error case if the rule now catches it, a known-gap note if not).
- The compiler source must not need syntax the seed lacks.
- `sh scripts/test.sh` passes.
