---
id: cache-type-names-rest
title: "Use the C type name cache at every call site"
status: done
priority: should
type: fix
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
spent: 129597
owns:
  - roadmaps/selfhost/nodes/cache-type-names-rest/**
shared:
  - selfhost/src/cnames.nra
  - selfhost/src/emit*.nra
reserves: []
provides: []
depends_on:
  - on: batch-4
    kind: hard
    why: builds on the batch-4 compiler and its measurements
---

# Use the C type name cache at every call site

## Goal

cache-type-names moved only the 4 hottest callers of `cnames.typeId` and `cnames.cType` to the cache; 88 call sites, including `functionSignature`, still rebuild names and hashes. Move the rest, and also cache `declName` per declaration and the module hash per module (widened by the maintainer on 2026-11-10 after profile-compiler-2: about 1.2% for the brief as first written, inside the 1.3% spread; about 2.5% widened).

## Context

- Read `git show feat/batch-4:roadmaps/selfhost/nodes/cache-type-names/outcome.md`.
- Evidence: read `git show feat/batch-5:roadmaps/selfhost/nodes/profile-compiler-2/outcome.md`, proposal 7 and the verdict on this node.
- Source: the batch-4 final report (2026-11-10) and the batch-4 inbox entries.

## Acceptance

- [x] Emitted C byte-identical; instructions for `narrac emit selfhost` before and after, with the gain above the spread, or the node closes as not worth it.
- [x] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.
