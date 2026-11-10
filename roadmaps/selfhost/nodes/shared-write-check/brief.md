---
id: shared-write-check
title: "Shared-write analysis, as a warning"
status: done
priority: should
type: feature
story: batch-6
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns:
  - selfhost/src/sharedwrites.nra
  - testdata/errors/shared_write_*
  - testdata/diagnostics/shared_write_*
  - roadmaps/selfhost/nodes/shared-write-check/**
shared:
  - selfhost/src/**
  - selfhost/tests/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E0978]
provides: []
depends_on:
  - on: copy-builtin
    kind: hard
    why: the fix the warning offers is copy(x), and copy results are unshared
  - on: from-marker
    kind: hard
    why: call results share their from arguments (findSharedArgument)
---

# Shared-write analysis, as a warning

## Goal

Build the shared-write rule of D94 (debate `roadmaps/selfhost/debate/value-semantics/debate.md`, settled structure item 1, claims P1, P17, O17) as a warning, so the real number of sites is known before code must change (open risk O1).

- A write to a place warns while another live name shares that place's root or a containing path. The check runs before the write clears the facts on the place.
- Tracked types: the ones `isCopyTracked` accepts (`selfhost/src/stdcopy.nra`). Strings, scalars and enums are not tracked.
- Sharing goes both ways and through chains: after `const items = a.items`, a write through `items` conflicts with a later use of `a.items`; after `const b = a; const c = b`, all three share.
- Passing `mutable R` writes `R` and everything below it; `mutable R.f` writes only `R.f` and below.
- A non-escaping closure that writes a captured `const` counts as a write where the closure is passed.
- Always shared with their source: plain parameters, pattern and `for (const ...)` bindings, captured values, values from identity types (`Cells.get`).
- Assigning a whole place ends its sharing (`const saved = checker.labels; checker.labels = []` is fine).
- A call result shares the argument `findSharedArgument` returns (`selfhost/src/sharedresults.nra`); `copy(...)` is unshared (`isCopyCall`, `stdcopy.nra`).
- Liveness inside one function body only. The message names the write, the shared place, the sharing name and its later use, as in the debate's sample. The fix wraps the sharing name's source in `copy(...)` and is applied by `narrac apply-fix`.

Decide and record in the outcome:

- Whether std builtin rows (`Arrays.get`, `Maps.get`) need a `shared` column. from-marker treats them as fresh (its outcome, Gotchas).
- How the walk handles reassigning a local whole. from-marker's walk is flow-insensitive; this rule needs liveness, so it cannot be.
- Whether the analysis reuses the walk in `sharedresults.nra` or replaces it.

The warning is on for every package, including `selfhost` (it does not fail the build). `narrac check --json` lists it with its fix.

## Acceptance

- [ ] One warning sample per bullet of the rule above, in `testdata/errors/` or a warnings suite, each with the expected message.
- [ ] Samples that must stay clean: whole-place assignment, a `copy(...)` source, a scalar or string, a write after the sharing name's last use.
- [ ] The fix is applied by `narrac apply-fix` and the result checks clean (habits suite).
- [ ] The outcome lists the warning count for `selfhost/src`, `std/src`, `selfhost/tests/src` and `testdata`, and how many are save/restore sites with a `mutable` call between save and reset (O9). The orchestrator uses it to size shared-write-sweep.
- [ ] `sh scripts/test.sh` passes.
