---
id: guide-multifile-samples
title: "The guide test compiles multi-file package samples"
status: done
priority: should
type: feature
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
spent: 243000
owns:
  - roadmaps/selfhost/nodes/guide-multifile-samples/**
  - selfhost/tests/src/guide.nra
shared:
  - docs/guide/**
  - roadmaps/selfhost/contracts/guide-samples.md
reserves: []
provides: []
depends_on: []
---

# The guide test compiles multi-file package samples

## Goal

Inbox, from the review of PR #87: the guide suite (`selfhost/tests/src/guide.nra`) skips `narra fragment` blocks, so multi-file package samples are never compiled (a chapter 4 fragment that did not compile slipped through). Add a multi-file sample form, for example a named package with several fragments that the test assembles into a temporary folder and builds, then convert the package samples in chapters 1, 4 and 8.

The form is a change to `contracts/guide-samples.md`. Write it as `contracts/guide-samples.change-request.md`; the orchestrator applies it before the node lands.

## Acceptance

- A broken fragment in a converted sample makes the guide suite fail (show it once, then revert).
- Every package sample in chapters 1, 4 and 8 is compiled by the suite.
- `sh scripts/test.sh` passes.
