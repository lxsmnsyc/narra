---
id: scope-map
title: "One scope map for locals in the checker"
status: active
priority: should
type: feature
story: batch-5
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
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/scope-map/**
shared:
  - selfhost/src/check.nra
  - selfhost/src/patterns.nra
  - selfhost/src/funcvalues.nra
reserves: []
provides: []
depends_on:
  - on: profile-compiler-2
    kind: hard
    why: the profile is the evidence for this change
---

# One scope map for locals in the checker

## Goal

Shadowing is an error, so a name is normally bound in at most one open scope. `findLocal` (`selfhost/src/check.nra:1774`) looks up one `Map<string, int>` instead of one map per scope level. Keep working: (a) `addLocal` (`check.nra:1905`) still rebinds a name after reporting E0103, so each scope keeps an undo log of (name, previous binding) and restores it at exit; (b) `setNarrowed` and `endNarrowing` (`selfhost/src/patterns.nra:113-127`) write `narrow|N` keys per depth, so they use the same undo log or a separate structure; (c) `isDeclaredBelow` (`selfhost/src/funcvalues.nra:222`) asks at which depth a local was declared, so each local stores its declaration depth.

## Context

- Evidence: `findLocal` 3.1% inclusive; the string-keyed scope map lookup is 4.0% self in the split build. Saves about 2%. Read `git show feat/batch-5:roadmaps/selfhost/nodes/profile-compiler-2/outcome.md`, proposal 5.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, run 3 times each side, reported in the outcome. The gain must exceed the run-to-run spread, or the node closes as not worth it.
- [ ] Every error test passes unchanged, including E0103 follow-on behavior, narrowing samples and closure capture samples.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Other batch-5 nodes change the compiler at the same time, so measure the base again at your own starting commit.
- Do not regenerate the seed.
