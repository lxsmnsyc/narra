---
id: slot-shared-writes
title: "Cells.set and Channel.send escape the shared-write check"
status: done
priority: must
type: fix
story: batch-15
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent: 190000
owns:
  - roadmaps/selfhost/nodes/slot-shared-writes/**
  - testdata/errors/slot_shared*
  - testdata/slot_shared*
shared:
  - selfhost/src/*.nra
  - std/src/cell.nra
  - std/src/channel.nra
  - std/src/shared.nra
reserves: []
provides: []
depends_on: []
---
# Cells.set and Channel.send escape the shared-write check

## Goal

`Cells.set(cell, items)` then `Arrays.push(mutable items, 1)` compiles, and the debug copy check exits 103; `Channel.send(ch, items)` then a write to `items` does the same. Values stored through internal slots are not tracked by the D94 shared-write check. Make the check see them: a store into a slot shares the argument with the container, like a field store.

## Context

- The shared-write check and `from` markers: `grep -rn 'E0601\|sharedWrite\|fromMarker' selfhost/src | head`. Probably the right fix is a `from` style marker on the std parameters that store (the value parameter of `Cells.set`, `Channel.send`, `Shared.new` and similar), so the rule is in std signatures, not in a special case per function. Check every std function that stores its argument (arena insert, map set, array push already handled?) and list them in `log.md`.
- The fix must not add copies to programs that do not write after storing.

## Acceptance

- [ ] Error tests fail on the base (compile) and pass (the shared-write error) for Cells.set, Channel.send, and any other store found.
- [ ] A run test where writing after `copy(...)` works.
- [ ] `sh scripts/test.sh` passes; the debug copy check on `emit selfhost` exits 0.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
