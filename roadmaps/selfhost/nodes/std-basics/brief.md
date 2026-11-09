---
id: std-basics
title: "Math, time, random and hashing"
status: todo
priority: should
type: feature
story: batch-2
early: false
outline: false
milestone: m10-std
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
added: 2026-11-01
spent:
owns:
  - std/src/math*
  - std/src/time*
  - std/src/random*
  - std/src/hash*
  - testdata/std_math*
  - testdata/std_time*
  - testdata/std_random*
  - testdata/std_hash*
  - testdata/ranges_stepped*
  - roadmaps/selfhost/nodes/std-basics/**
shared:
  - selfhost/src/**
  - std/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: std-foundation
    kind: hard
    why: source std modules need the generic mismatch errors, std/internal and the stage 2 tests
  - on: std-source
    kind: hard
    why: std modules are Narra source over std/internal primitives (std-source gate, 2026-11-05)
---

# Math, time, random and hashing

## Goal

Four std modules, each Narra source in `std/src/` over primitives under `std/internal/` (std-source gate, 2026-11-05). Read `nodes/std-foundation/outcome.md` and look at `std/src/option.nra` and `json.nra` for the pattern.

- `std/math`: `abs`, `min`, `max`, `clamp` for int and float, `pow`, `sqrt`, and the float functions of C's math.h that programs use (`floor`, `ceil`, `round`, `trunc`, `sin`, `cos`, `tan`, `atan2`, `exp`, `log`), constants `PI` and `E`. Primitives call math.h; link `-lm` where Linux needs it.
- `std/time`: a monotonic clock (`Time.now()` as an `Instant`), elapsed time as a `Duration` (milliseconds and nanoseconds), the wall clock as Unix milliseconds, and `Time.sleep(duration)`. Reading a clock or sleeping is an effect; use the existing effect names (check SPEC.md section 7 for which).
- `std/random`: a seeded generator (`Random.seeded(seed)`, a value type, so results repeat) with `nextInt(range)`, `nextFloat`, `shuffle`, and a system-seeded one. Pick one small generator (for example xorshift or PCG) and document the choice in SPEC.md. Not for cryptography; say so.
- `std/hash`: re-export the existing `Hash.of`, plus a stable hash (FNV-1a 64) whose output does not change between runs or builds.
- `Ranges.stepped(start, end, step)`, which SPEC.md section 2 documents but no module registers. A zero step panics; a negative step counts down.

Deferred from std-foundation: when a std source module first imports another one, sort std source modules with `orderModules` in single-file packages (`selfhost/src/modules.nra`); today they skip it.

## Acceptance

- [ ] Samples for each module with `.out`, clean under AddressSanitizer and the leak check. Time and system-random samples print only values that do not change between runs (for example "elapsed >= 0").
- [ ] Stage 2 std tests cover each module.
- [ ] SPEC.md lists the four modules and `Ranges.stepped` as built; compiler/README.md matches.
- [ ] `sh scripts/test.sh` passes, and `sh scripts/linux-check.sh` passes (math.h linking).
