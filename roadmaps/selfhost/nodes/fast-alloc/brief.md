---
id: fast-alloc
title: "Size-class free lists in narra_alloc"
status: done
priority: should
type: feature
story: batch-4
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
spent: 123080
owns:
  - roadmaps/selfhost/nodes/fast-alloc/**
shared:
  - selfhost/runtime/runtime.h
  - selfhost/src/runtimecore.nra
  - DECISIONS.md
reserves: []
provides: []
depends_on:
  - on: regions-design
    kind: hard
    why: D100 measured the free-list prototype this node builds properly
---

# Size-class free lists in narra_alloc

## Goal

Replace `malloc` and `free` for small blocks in `narra_alloc` and `narra_free` (`selfhost/runtime/runtime.h:65-72`) with per-thread size-class free lists carved from large chunks. regions-design measured a prototype at 5.45G instructions for `narrac emit selfhost` against 9.27G today (0.57 s to 0.42 s user time): `malloc` and `free` are about 41% of the compiler's instructions, reference counts at most 9%.

## Context

- Read `git show feat/batch-4:roadmaps/selfhost/nodes/regions-design/outcome.md` and D100 in `DECISIONS.md` on feat/batch-4. The prototype is `roadmaps/selfhost/nodes/regions-design/prototype/variant.py freelist` on that branch. It keeps an 8-byte size header and never returns chunks to the system; it is a measurement, not the design.
- Every generated heap object goes through `narra_alloc`. `narra_realloc` (array growth, `runtime.h:112`) must move small blocks between classes or out of the free lists.
- The leak check (`NARRA_CHECK_LEAKS`, `narra_live`, `narra_report_leaks` at `runtime.h:58`) must still count live objects exactly.
- AddressSanitizer builds (`-sanitize`) keep plain `malloc` and `free` so ASan still sees every block. Choose by a macro the sanitize build already defines, or add one.
- `runtime.h` is embedded in the compiler as `RUNTIME_H` in `selfhost/src/runtimecore.nra`. After editing it, run `NARRAC=<built narrac> sh scripts/test.sh --suite generated --update`. GOAL.md: editing runtime text needs two seed updates; the orchestrator does them.
- Threads are not built yet, so per-thread means one set of lists today; write it so a `_Thread_local` later is a one-word change.

## Acceptance

- [ ] `narrac emit selfhost` at most 6.0G instructions retired (macOS: `/usr/bin/time -l` prints "instructions retired"), measured against the base at the same commit; peak memory within 20% of the base.
- [ ] `bench/run.sh -n 9` before and after (time is noisy while other workers run; instructions are the main number), reported in the outcome with the load.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.
- [ ] A short note in D55's cost line or a new paragraph under D100 saying the allocator now uses size classes, with the measured gain.

## Sizing

S: one runtime file and its embedded copy; the design is measured already.

## Notes for the worker

- Do not change the emitter; node perceus edits it at the same time.
- Do not regenerate the seed.
