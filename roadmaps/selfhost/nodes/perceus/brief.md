---
id: perceus
title: "Borrow loop items and match bindings, move locals at last use"
status: active
priority: should
type: feature
story: batch-4
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: [benchmarks]
fulfilled_by:
gate: false
added:
spent:
owns:
  - testdata/perceus_*
  - roadmaps/selfhost/nodes/perceus/**
shared:
  - selfhost/src/emitstate.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/emitmatch.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitoptions.nra
  - DECISIONS.md
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: bench-compiler
    kind: hard
    why: needs the idle-machine baseline, including the compiler row
  - on: box-recursive
    kind: hard
    why: changes the heap layout this pass works on
  - on: nested-or-guards
    kind: hard
    why: both edit the binding code in emitmatch.nra; the small fix lands first
  - on: closure-arg-temps
    kind: hard
    why: both edit argument emission in emitexprs.nra; the small fix lands first
---

# Borrow loop items and match bindings, move locals at last use

## Goal

The first Perceus step (SPEC section 6, "removes the ones it can prove unnecessary, as in Koka's Perceus"): stop emitting retain and release pairs on values that are already kept alive by something else, and move a local into its last use instead of retaining it. Bring `interp` back near its time before D55 (0.237 s on the old baseline machine), and report the change on the compiler row.

## Context

The ownership rules are at `selfhost/src/emitstate.nra:6-12`: a local owns one reference, parameters are borrowed, call results are owned temps. `consume` (`emitstate.nra:481-511`) is the one place that decides between a move (owned temp) and a retain (borrowed value). Parameters are already borrowed and owned temps already move. What is missing:

1. **for-of items.** `emitArrayFor` (`emitstmts.nra:706-779`) already holds the array for the loop (retained at 727 when borrowed), then retains every `item = arr->items[i]` (757-761) and releases it at the end of each iteration. While the loop holds the array, the item cannot be freed, so the item can be a borrowed local. Copy-on-write on the array inside the body makes a new array and leaves the held one alone, so this stays sound. This is the `interp` hot loop: 150M retain and release pairs on a heap `Op`.
2. **Match bindings.** Pattern bindings each retain (`emitmatch.nra:257-275`), while the subject is held for the whole arm (406-408, and the `ScopeFrame.held` list for `is`). A binding that is only read in the arm can borrow from the held subject.
3. **Last-use moves.** A local whose last use is a consume site (`let`, assignment, `return`, constructor field, array push or literal element, struct literal field, map key or value, tail-call parameter: the call sites the survey lists under `consume`) moves into it: no retain there and no release at scope end. This needs a last-use analysis per function body over the statements, including loops (a use inside a loop body is never the last one unless the loop exits after it) and every unwind path in `unwindAll`, `unwindBlock` and `unwindToLoop`.

Keep the rules visible: add them to the ownership comment at `emitstate.nra:6-12` and to D55's cost line (drop "until node perceus" when it holds).

- Source: SPEC.md line 864, D55 (DECISIONS.md line 465).
- Evidence: `bench/BASELINE.md` interp 0.393 s against 0.237 s before D55 (busy machine, 2026-10-31). bench-compiler records the fresh numbers; measure again at its commit before changing code.

## Steps

- [ ] Borrowed for-of items, then measure `interp` and the compiler.
- [ ] Borrowed match and `is` bindings, then measure.
- [ ] Last-use moves of locals, then measure.
- [ ] Samples `testdata/perceus_*` for each rule, with values that would be freed early if the rule were wrong: a loop body that reassigns the array it iterates, an arm that reassigns the matched local, a moved local inside a loop, an early `return` and a `break` after a move.
- [ ] Regenerate the seed (`sh scripts/seed-update.sh`); the bootstrap fixed point must hold.

## Acceptance

- [ ] `bench/run.sh interp` is within 10% of 0.237 s scaled to the new baseline machine, or the outcome explains what remains and where the time goes.
- [ ] `bench/run.sh compiler` is reported before and after; no program in `bench/` gets slower beyond noise.
- [ ] `sh scripts/test.sh` passes: every sample clean under AddressSanitizer and the leak check, and the sanitized `bootstrap` suite emits identical C.
- [ ] `sh scripts/linux-check.sh` passes.

## Sizing

L: uncertainty medium (known technique, new to this emitter), blast radius high (the ownership core of every emitted function), size medium. The three steps share `consume` and the scope stack, so they stay one node with a measurement after each step.

## Notes for the worker

- Do not change the enum layout or the RC functions in `emitdecls.nra`; option-inline owns that next.
- `--report-copies`, `@unique` and `NARRA_TRACE_COPIES` are out of scope: D94 replaces them with `copy(x)` and the debug copy check (node copy-builtin).
- Large constant parameters by pointer (D50) are out of scope.
