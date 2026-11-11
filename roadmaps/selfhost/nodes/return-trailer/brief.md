---
id: return-trailer
title: "No unreachable trailer after a function's final return"
status: active
priority: should
type: refactor
story: batch-7
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
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/return-trailer/**
shared:
  - selfhost/src/pathends.nra
  - selfhost/src/check.nra
  - selfhost/src/emitclosures.nra
  - selfhost/src/emitconsts.nra
reserves: []
provides: []
depends_on: []
---

# No unreachable trailer after a function's final return

## Goal

The emitter ends every function that returns a value with `narra_panic(<pos>, "internal compiler error: function ended without returning a value"); return (T){0};` (`emitclosures.nra`, `emitconsts.nra`). The seed has 1878 of them, about 3.7k lines, most after a final `return`. Skip the trailer when the body cannot reach its end: its last statement is a `return`, a `panic`, or an `if`/`match` whose every branch ends that way. The checker's missing-return rule (E0969/E0970) is in `pathends.nra`: have it record the bodies whose end is unreachable (for example a set of body ids on the Checker), and let both trailer sites (`emitclosures.nra`, `emitconsts.nra`) call one helper that reads it. The check sees every branch of a build-condition `if`, so a fold never makes the record wrong: if every branch returns, the kept one does too.

## Acceptance

- [ ] The trailer stays where the body's end is reachable as far as the emitter can tell; a C compiler warning or error for a missing return must not appear (`cc -Wall` on the emitted selfhost C shows no new `-Wreturn-type`).
- [ ] The seed-size drop (lines of `narrac emit selfhost`) is in the outcome.
- [ ] Fast check passes: `sh scripts/test.sh`
