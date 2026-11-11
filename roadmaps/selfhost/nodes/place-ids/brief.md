---
id: place-ids
title: "The shared-write analysis keys places by integer ids"
status: todo
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
spent:
owns:
  - roadmaps/selfhost/nodes/place-ids/**
shared:
  - selfhost/src/sharedwrites.nra
reserves: []
provides: []
depends_on: []
---

# The shared-write analysis keys places by integer ids

## Goal

Bring the shared-write analysis (E0978, `selfhost/src/sharedwrites.nra`) back to a small share of every command. Today a place is a string key (`"L3.items[=0]"`), matched by `findRest` with slicing and searches, and the pairs of a flow state are an array that `hasPair` scans with string compares.

## Context

- Profile and C sizes from the batch-7 planning session (2026-11-11, commit 94360a39, Apple M2): `narrac emit selfhost` is 2.49G instructions and 138 MB peak. `reportSharedWrites` is about 23% of it; see the batch-7 brief for the self-time split.
- Batch-6 lessons (wave-19, shared-write-error outcome): a first version that split keys into parts per comparison cost 9.77G. Parse a key once, not per comparison.
- Direction, not a prescription: intern each place per body to an integer id holding its root and an array of part ids (`.field`, `[]`, `[=N]`); keep the text only for messages; keep pairs in a set keyed by the two ids (smaller first) so `hasPair` is one lookup; memoize the prefix relation between ids if `findRest` stays hot. `mergeStates` copies whole pair arrays per branch; check it after the change.

## Acceptance

- [ ] Every `shared_write_*` sample (errors and programs) gives the same diagnostics, fixes and positions as before.
- [ ] `narrac emit selfhost` instructions (`/usr/bin/time -l`, release build, median of 3) drop by at least 0.35G from the batch base, measured on the same source. Record before and after in the outcome, with `check selfhost` too.
- [ ] Peak memory of `emit selfhost` is not higher than at the base.
- [ ] Fast check passes: `sh scripts/test.sh`
