---
id: cache-type-names-rest
title: "Use the C type name cache at every call site"
status: todo
priority: should
type: fix
story:
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

cache-type-names moved only the 4 hottest callers of `cnames.typeId` and `cnames.cType` to the cache; 88 call sites, including `functionSignature`, still rebuild names and hashes. Move the rest.

## Context

- Read `git show feat/batch-4:roadmaps/selfhost/nodes/cache-type-names/outcome.md`.
- Source: the batch-4 final report (2026-11-10) and the batch-4 inbox entries.

## Acceptance

- [ ] Emitted C byte-identical; instructions for `narrac emit selfhost` before and after, with the gain above the spread, or the node closes as not worth it.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.
