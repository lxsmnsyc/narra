---
id: regions
title: "Region runtime: region blocks, the region tag and the region-off points"
status: dropped
priority: must
type: feature
story: batch-13
early: false
outline: false
milestone: p1-syntax
target:
ship: auto
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/regions/**
  - selfhost/runtime/region.c
  - selfhost/src/runtimeregion.nra
  - selfhost/src/stdregion.nra
  - testdata/regions_runtime*.nra
  - testdata/regions_runtime*.out
shared:
  - selfhost/runtime/runtime.h
  - selfhost/runtime/std_async.c
  - selfhost/runtime/std_async_loop.c
  - selfhost/src/runtimetext.nra
  - selfhost/src/runtimecore.nra
  - selfhost/src/runtimestdasync.nra
  - selfhost/src/runtimestdasyncloop.nra
  - selfhost/src/emitconsts.nra
  - selfhost/src/emitgenerators.nra
  - selfhost/src/emitpointers.nra
  - selfhost/src/emitasync.nra
  - std/src/cell.nra
  - selfhost/src/stdinternal.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/stdlib.nra
  - selfhost/src/stdcalls.nra
  - selfhost/src/emitstd.nra
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: regions-plan
    kind: hard
    why: D113 must be approved with regions kept; if the maintainer accepts the drop, this node is dropped
---
# Region runtime: region blocks, the region tag and the region-off points

Dropped 2026-11-13 by D113.

First child of regions (D113 items 3 to 5). The node keeps the id `regions`; its title is region-runtime. Children after it: region-copy-out, region-check.

## Goal

The C runtime can open and close a region on the current thread. While one is open, `narra_alloc` takes blocks from it, `narra_free` and `narra_realloc` tell region blocks apart by the block header, and every place D113 lists runs with the region off. `std/region` exists with `Region.run`, but its copy-out is a stub that returns the value as it is, so the samples of this node return numbers and `bool` only. region-copy-out replaces the stub; region-check adds the errors.

## Context

- Decision: D113 in `DECISIONS.md` (read items 3 to 5). D100 is replaced where they differ: counts are kept, there is no sign check, no store list and no negative count.
- Allocator: `selfhost/runtime/runtime.h`, `narra_block_new`, `narra_block_free`, `narra_realloc` (fast-alloc). The 8-byte header holds the size class; add a region tag value there.
- Region-off points (D113 item 5): the getters of non-static top-level constants and lazy values (`emitconsts.nra`), task start and every resume (`std_async.c`, `std_async_loop.c`, `emitasync.nra`), `Async.then` callbacks, generator and coroutine resume when the frame is not a region block (`emitgenerators.nra`), `Ptr.alloc` (`emitpointers.nra`), `Cell.create`'s slot (`std/src/cell.nra` through a new `std/internal/cell` function in `stdinternal.nra`).
- Generated files: `runtimecore.nra`, `runtimestdasync.nra`, `runtimestdasyncloop.nra` and `stdtext.nra` come from `sh scripts/test.sh --suite generated --update`; a new `region.c` adds one import and one case to `runtimeFor` in `runtimetext.nra` (`selfhost/CLAUDE.md`).
- Runtime text edits need two seed updates; the orchestrator does them at the end of the batch.

## Steps

- [ ] Region state per thread (`NARRA_HEAP_LOCAL`): a stack of open regions, each a list of 1 MiB chunks and a bump pointer. `narra_region_open()`, `narra_region_close()`, `narra_region_active()`, and `narra_region_off()` / `narra_region_restore(saved)` for the region-off points.
- [ ] Header tag: region blocks carry a class value no size class uses, plus the region's depth, so a block of an enclosing region is not taken for the current one. `narra_free` on a region block only lowers `narra_live`; `narra_realloc` on a region block takes a new region block and copies.
- [ ] Plain-malloc builds (`NARRA_PLAIN_MALLOC`, ASan): every block gets the 8-byte header; a region block is its own `malloc`, linked in the region and freed at `narra_region_close`, so ASan reports a use after the region.
- [ ] Copy on write inside a region: the `_unique` paths of arrays, maps, boxes, string builders and generators copy a block that is not of the current region even when its count is 1 (D113 item 4). The debug copy check does not count these copies.
- [ ] `narra_region_close` drains the pending drop list (`narra_pend`) before it frees the chunks.
- [ ] The region-off points above, each restoring the previous state on return.
- [ ] `std/region`, a builtin module (rows, check and emitter in `stdregion.nra`): `Region.run<T, effect E>(body: () => T uses E): T uses E, alloc`. The emitter writes, in this order: open the region, call `body`, copy the result out, release the result, drain pending drops, close the region. It is a builtin and not a source function because the release must come before the close, and a source function releases its locals at its end. The copy-out call is a stub (`narra_region_copy_out` that returns its argument) until region-copy-out. Register the module as `selfhost/CLAUDE.md` says (one `registerAll` line, one `if` in `checkStdCall`, one in `emitNamedCall`).

## Acceptance

- [ ] `testdata/regions_runtime.nra`: `Region.run` builds strings, arrays and maps, grows an array past several chunk sizes and returns a count; output checked; clean under AddressSanitizer and the leak check. On the batch base it fails to compile (no `std/region`).
- [ ] `testdata/regions_runtime_escape.nra`: a constant read for the first time, a task started and a `Cell` created inside an open region still work after it closes (ASan clean).
- [ ] A heap array with count 1 pushed to inside a region is copied, and the original is unchanged after the region (sample output).
- [ ] `bench/` instructions within 2% of the batch base for all six programs (no region is open there; this measures the header and copy-on-write checks).
- [ ] `narrac emit selfhost` instructions within 1% of the batch base.
- [ ] `compiler/README.md` lists the runtime part as built and the language part as not built yet.
- [ ] Fast check passes: `sh scripts/test.sh`.

## Sizing

L. Breadth: one new runtime file, one builtin module, and about ten small hook points across the runtime, emitter and std. Depth is low: no checker rule, no new syntax. Risk: the copy-on-write check sits on hot paths (bench acceptance).

## Notes for the worker

- Do not add a sign check or a negative count. Counts stay as they are (D113 item 3).
- No diagnostic codes in this node. E1270 and E1271 belong to region-check.
- `NARRA_HEAP_LOCAL` is empty today; keep the region state beside `narra_small` so threads make both thread-local together.
- Stub left: the copy-out call returns its argument. A sample that returns a string from `Region.run` would read freed memory until region-copy-out lands; do not add one here.
