---
id: cli-skip-teardown
title: "narrac emit and build skip the final release unless the leak check runs"
status: todo
priority: should
type: feature
story: batch-5
early: false
outline: false
milestone: m6-performance
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
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/cli-skip-teardown/**
shared:
  - selfhost/src/emitter.nra
  - selfhost/src/driver.nra
  - selfhost/runtime/**
reserves: []
provides: []
depends_on:
  - on: profile-compiler-2
    kind: hard
    why: the profile is the evidence for this change
---

# narrac emit and build skip the final release unless the leak check runs

## Goal

After `narrac emit` or `narrac build` has written its output, exit without releasing the `Checker`, the `Emitter` and the AST; the OS frees the memory. Key it on `NARRA_CHECK_LEAKS` (`selfhost/runtime/runtime.h:59`): when it is set, keep the full teardown. Limit it to the emit and build commands; the LSP server and other long-lived paths keep releasing. The maintainer accepted this on 2026-11-10, reversing the first profile's rejection.

## Context

- Evidence: teardown is 6.3% of `narrac emit selfhost` (`release_S_Checker` from `runEmit` and the drop of the AST). Read `git show feat/batch-5:roadmaps/selfhost/nodes/profile-compiler-2/outcome.md`, proposal 2.
- The test runner sets `NARRA_CHECK_LEAKS=1` on every `narrac` run (`selfhost/tests/src/shell.nra:45`, `context.nra:41` and `:53`, `bootstrap.nra:41`), so the leak check and the AddressSanitizer run keep covering the teardown.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, run 3 times each side, reported in the outcome. The gain must exceed the run-to-run spread, or the node closes as not worth it.
- [ ] With `NARRA_CHECK_LEAKS=1`, `narrac emit selfhost` still reports no leaks.
- [ ] The LSP suite passes and the language server path still releases.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Other batch-5 nodes change the compiler at the same time, so measure the base again at your own starting commit.
- Do not regenerate the seed.
