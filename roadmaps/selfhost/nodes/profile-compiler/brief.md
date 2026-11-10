---
id: profile-compiler
title: "Spike: profile the compiler compiling itself"
status: active
priority: should
type: spike
story: batch-4
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: M
spike: true
hypothesis:
verify: 
exclusive: [benchmarks]
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/profile-compiler/**
shared:
  - roadmaps/selfhost/inbox.md
reserves: []
provides: []
depends_on:
  - on: bench-compiler
    kind: hard
    why: profiles the same release build and command the compiler row times
---

# Spike: profile the compiler compiling itself

## Goal

Find where `narrac emit selfhost` spends its time and memory, so compiler speed work targets measured costs instead of guesses. Write the findings to `outcome.md` and turn each hotspot worth fixing into a node brief proposal.

## Context

- Use the release build from `bench/run.sh compiler`. On macOS, `sample` or Instruments (`xctrace record --template 'Time Profiler'`); on Linux, `perf record -g` in `sh scripts/linux-check.sh`'s container. Build with `-g` and frame pointers (`-fno-omit-frame-pointer`) so C frames map back to Narra functions through the hashed names (`selfhost/src/cnames.nra`).
- Split the cost into: reference counting (`retain_*`, `release_*`, `narra_drop`), allocation (`narra_alloc`, `narra_free`, `narra_realloc`), string building (`lstr_*`, `narra_sb_*`), map lookups (`runtime_map.h`), and the compiler's own phases (lexer, parser, checker, emitter in `driver.nra`). Time each phase with a clock around the driver calls if the profile does not show phases.
- Earlier quadratic finding: node decls-quadratic (done). Look for others of the same kind: linear scans in lookups, repeated string concatenation in loops.

## Acceptance

- [ ] `outcome.md` has a table of the top 10 functions by self time and by inclusive time, the share of time per category above, and the share per phase, at a named commit.
- [ ] Each finding worth fixing becomes a proposed brief (title, goal, evidence, size) in `outcome.md`. The orchestrator decides which join batch-4 and which go to `inbox.md`.
- [ ] It states whether const-params-pointer is worth doing: how much time goes to copying large structs passed by value.

## Sizing

M: measurement only, but reading profiles of 180k lines of generated C through hashed names takes care.
