---
id: checker-emitter-copies
title: "A debug-built compiler compiles itself with no hidden copies"
status: active
priority: should
type: fix
story: batch-10
early: false
outline: false
milestone: 
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
added: 2026-11-12
spent:
owns:
  - roadmaps/selfhost/nodes/checker-emitter-copies/**
  - selfhost/src/sharedwrites.nra
  - selfhost/tests/src/copies.nra
shared:
  - selfhost/src/**
  - selfhost/CLAUDE.md
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [decision D104]
provides: []
depends_on: []
---

# A debug-built compiler compiles itself with no hidden copies

## Goal

A compiler built from the seed without `-DNARRA_RELEASE` and run with `NARRA_CHECK_COPIES=1` reports copies on write that no `copy(...)` explains, and exits 103. Measured during batch-9 (hidden-copy-fix, 2026-11-11): `check testdata/arrays.nra` gives `array 3, map 1`, and `emit selfhost` about 950 array and 964 map copies. They come from the checker and the emitter. Until they are gone, the debug copy check cannot run in CI.

Two E0978 gaps let such copies compile, and are likely behind most of them:

1. **A write after another name's last use.** `const team = teams[0]; resetTeam(mutable team)` with `teams` unused afterwards passes E0978, but `teams` keeps the value until its scope ends, so the write copies. `selfhost/CLAUDE.md` already warns about `let x = holder.field` followed by a write to `x`. Choose one fix: treat a name as live until its scope ends (more errors, simple rule), or release a value at its last use in the emitter (no new errors, emitter work). Record the choice as D104.
2. **Array elements that share one value.** `if (Arrays.filled([0, 0], 3) is Some(rows)) { let grid = rows; grid[0][0] = 5 }` compiles and copies at run time; `[row, row]` is the same. SPEC section 6 lists it as a known gap. Treat heap elements of such arrays as shared, or keep it documented in SPEC and say why in D104.

Then fix the patterns in `selfhost/src` that the closed rule now rejects, or that the copy check still reports, until the debug-built `emit selfhost` exits 0. Add a `copies` suite case (`selfhost/tests/src/copies.nra`) that runs the debug-built compiler on `selfhost` and fails on exit 103, if it fits in the suite's time; otherwise one on a smaller program and a note in the outcome.

## Acceptance

- Proof on `feat/batch-10` before any change: the debug-built `emit selfhost` exit code and copy counts, in `log.md`.
- D104 in `DECISIONS.md`, with error tests in `testdata/errors/` for every case the new rule rejects.
- The debug-built compiler runs `check testdata/arrays.nra` and `emit selfhost` with `NARRA_CHECK_COPIES=1` and exits 0.
- `narrac emit selfhost` instructions (release build) before and after, in `outcome.md`.
- The `selfhost/CLAUDE.md` line about `let x = holder.field` is updated or removed to match the new rule.
- `sh scripts/test.sh` passes.
