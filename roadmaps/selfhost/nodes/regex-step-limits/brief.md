---
id: regex-step-limits
title: "Regex timing cases without wall-clock limits"
status: done
priority: could
type: fix
story: batch-20
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-19
spent:
owns:
  - roadmaps/selfhost/nodes/regex-step-limits/**
  - testdata/regex_linear.*
shared:
  - std/src/regex*
  - testdata/regex*
  - samples/**
  - selfhost/tests/src/**
reserves: []
provides: []
depends_on: []
---
# Regex timing cases without wall-clock limits

## Goal

`testdata/regex_linear` prints "under 1 s" for patterns that backtrack exponentially in other engines, and the regex sample has a 5,000 x case that expects "over 1 s". Both failed four times under load average 20 to 150 during batch-16 and passed when run alone (inbox, 2026-11-15). A wall-clock limit in a shared suite fails whenever other builds run.

Make the cases check the linear-time property without the clock: count engine steps (if `std/regex` can expose a step count without a public API the maintainer has not approved, prefer a test-only hook or a count the case derives itself), or compare the time of input n against 2n so load affects both runs alike. Find the sample with the 5,000 x case (`git grep -n "over 1 s"`). Proving the flake: show that the current case fails under synthetic load (for example several `yes > /dev/null` processes) and the new one passes under the same load.

## Acceptance

- [ ] The proof above, recorded in `log.md`.
- [ ] No regex case compares against a fixed number of seconds.
- [ ] `sh scripts/test.sh` passes.
