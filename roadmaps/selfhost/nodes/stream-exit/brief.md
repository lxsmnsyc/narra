---
id: stream-exit
title: "Exit stops a task that ignores its cancel"
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
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent: 215000
owns:
  - roadmaps/selfhost/nodes/stream-exit/**
  - testdata/async_exit_ignored_cancel*
shared:
  - selfhost/runtime/std_async*.c
  - selfhost/src/runtime*.nra
  - DECISIONS.md
  - SPEC.md
reserves: []
provides: []
depends_on: []
---
# Exit stops a task that ignores its cancel

## Goal

A stream body (or any task) that ignores the `Err(Cancelled)` of its `sleep` and loops keeps the process alive after the entry function returns; a manual `Async.next` loop on the base hangs the same way. D109 item 6 only promises that a body which checks its results ends.

## Rule (approved by the maintainer 2026-11-14)

At exit the loop cancels every running task, as now, and gives each one cleanup turns. A cancelled task that suspends again on a new wait after its cancel is dropped at that point: its frame is released without resuming, and exit goes on. A task that loops without suspending is a plain infinite loop, as in synchronous code. Write it as an amendment to D109 item 6 (dated 2026-11-14, maintainer approved) and SPEC section 8, then build it.

## Acceptance

- [ ] `testdata/async_exit_ignored_cancel.nra` hangs on the base (run with a timeout) and exits on the branch; a body that does check its `Err` still runs its cleanup code.
- [ ] Clean under AddressSanitizer and the leak check (dropped frames release their locals). macOS and Linux (`sh scripts/linux-check.sh` at the batch end covers epoll).
- [ ] `sh scripts/test.sh` passes.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
