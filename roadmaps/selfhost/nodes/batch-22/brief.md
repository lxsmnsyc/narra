---
id: batch-22
title: "Batch 22: one form per std API, clean samples, std gaps"
status: active
priority: could
type: story
story: 
early: false
outline: false
milestone: p6-docs
target:
ship: pr
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/batch-22/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-21
    kind: hard
    why: stacked on PR #126
---
# Batch 22: one form per std API, clean samples, std gaps

## Goal

Maintainer requests, 2026-11-20, before the `v0.1.0` tag:

- Remove the `std/async/*` modules. Each std API has one form, sync or async, never both. The `async` language feature stays.
- Doc samples print no warnings that are unrelated to the example (mostly the unused `args` of `run`).
- Find the std modules a first user will miss, and add the ones that matter.
- Add Zig and Odin to the benchmarks (maintainer, 2026-11-20).

The branch `feat/batch-22` starts from `feat/batch-21` (PR #126, not merged yet); it is rebased onto `main` after #126 merges. Worktree `../narra-batch-22`, local only.

## Decisions (orchestrator, 2026-11-20)

The rule: an API is async when the operating system can wait for it in the background (timers, sockets), and sync when it cannot (files, processes, stdin).

- Files: sync only. `std/async/file` goes; async code calls `std/fs` (and may `await Async.idle()` to let other tasks run).
- `Time.sleep`: async only. `std/async/time` goes; `Time.sleep` in `std/time` becomes the async sleep.
- Sockets: async only, as `std/net` (the old `std/async/net`, names unchanged).
- Processes, stdin, `Thread.join` and channels stay sync.
- A thread function may be `async`, so code on a thread can sleep and use `std/net` (found while folding: thread functions could not await, so the async-only sleep left threads with no way to wait).
- The entry function's parameters are never reported as unused: `run`'s signature is fixed, so the warning cannot be acted on.

## Acceptance

- [ ] Every child done or out of the batch with a reason.
- [ ] `sh scripts/test.sh` passes; the seed is current; `pnpm --dir website build` passes; `sh scripts/linux-check.sh` passes.
