---
id: const-functions-std
title: "@const bodies call the rest of the D105 std list"
status: done
priority: must
type: feature
story: batch-15
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent: 245000
owns:
  - roadmaps/selfhost/nodes/const-functions-std/**
  - testdata/const_functions_std*
  - testdata/errors/const_functions_std*
shared:
  - selfhost/src/consteval.nra
  - selfhost/src/constfunctions.nra
  - selfhost/src/*.nra
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on: []
---
# @const bodies call the rest of the D105 std list

## Goal

`@const` bodies may call std string, array, map, int, float and range functions today (const-functions). Add the rest of D105's list: `std/char`, `option`, `result`, `math`, `bytes`, `fixedarray` and the sized number modules (i8 to u64, f32). Functions with `io` stay out.

## Context

- D105 in `DECISIONS.md`; `consteval.nra` and `constfunctions.nra`, how existing std calls are evaluated at compile time. Prefer evaluating the std source bodies over writing a second implementation per function, where the evaluator can; list any function done by hand and why.
- Sized integers must wrap or trap at compile time exactly as at run time: test both sides give the same value.

## Acceptance

- [ ] `testdata/const_functions_std.nra`: one constant per module, each value compared at run time with the same call made at run time.
- [ ] Error tests: a call that would trap at run time (division by zero, overflow that traps) is a compile error at the constant.
- [ ] SPEC lists the modules. `sh scripts/test.sh` passes.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
