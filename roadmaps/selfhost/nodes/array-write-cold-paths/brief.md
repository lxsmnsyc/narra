---
id: array-write-cold-paths
title: "Array writes keep their panic and copy paths out of line"
status: active
priority: should
type: feature
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
exclusive: [benchmarks]
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/array-write-cold-paths/**
shared:
  - selfhost/runtime/**
  - selfhost/src/emitdecls.nra
  - selfhost/src/runtimecore.nra
  - testdata/**
reserves: []
provides: []
depends_on:
  - on: profile-programs
    kind: hard
    why: the spike is the evidence and holds the experiment scripts
---

# Array writes keep their panic and copy paths out of line

## Goal

Mark `narra_panic` noreturn and cold. Split `narra_bounds` into an inline unsigned compare `(uint64_t)i >= (uint64_t)len` and an out-of-line cold `narra_bounds_fail` that formats the message. Split every `_unique` (arrays and boxes in `selfhost/src/emitdecls.nra`, maps in `selfhost/runtime/runtime_map.h`) into an inline `rc == 1` test and a noinline `_copy`. Wrap the GNU attributes in a macro so the C stays valid C11 (`_Noreturn` is C11; `cold` and `noinline` behind `#if defined(__GNUC__)`).

## Context

- Evidence: sieve spends 90% of its user samples in a non-inlined `A_Array_<T>_at`, which carries the copy path and the `snprintf` failure path. Rewriting the emitted C this way took sieve from 0.246 s to 0.131 s (2.26G to 0.68G instructions, level with Rust) and the compiler from 2.65G to 2.53G instructions. Read `git show feat/batch-5:roadmaps/selfhost/nodes/profile-programs/outcome.md`, proposal 1, and reuse `experiments/variant1.py` as the target shape.
- Only the array split was measured; the box and map splits are not. Measure them, and keep any that do not help out of the change.
- A text replacement on the compiler's emitted C also matches the runtime header embedded as a string (`RUNTIME_H` in `runtimecore.nra`).

## Acceptance

- [ ] sieve and particles times and instructions before and after (`perl bench/median.pl 15`, `/usr/bin/time -l`), with the sieve gain above the spread.
- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, run 3 times each side, reported in the outcome. Report the change; no gain is required here.
- [ ] An index-out-of-range sample prints the same panic message and position as on the base.
- [ ] The emitted C compiles with `-std=c11 -pedantic` without new warnings on clang.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- This node edits `selfhost/runtime/`, so the seed must be regenerated twice before the compiler itself runs the new runtime (GOAL constraint). Do it on your branch and record it in the outcome; the orchestrator checks it once.
