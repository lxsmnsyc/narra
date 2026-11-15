---
id: c-layout
title: "C layouts for structs and enums"
status: todo
priority: could
type: feature
story:
early: false
outline: false
milestone: p5-distribution
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
spent:
owns:
  - roadmaps/selfhost/nodes/c-layout/**
  - selfhost/src/clayout.nra
  - testdata/c_layout*
  - testdata/errors/c_layout*
shared:
  - selfhost/src/parser.nra
  - selfhost/src/externs.nra
  - selfhost/src/emitexterns.nra
  - selfhost/src/emitdecls.nra
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1610-E1619]
provides: []
depends_on:
  - on: c-abi-design
    kind: hard
    why: D122 fixes the attribute spelling and the allowed field types
---
# C layouts for structs and enums

## Goal

The layout attribute D122 picks makes a struct match a C struct (declared field order, C padding, no reference count header) and an enum without fields use fixed integer values. Such types may appear in `extern` member signatures (today E1002 allows only numbers, `bool` and `Ptr<T>`, `externs.nra`) and in exported signatures.

## Acceptance

- [ ] A sample passes a C-layout struct by value and by `Ptr<T>` to a C function in a C file built with it (or libc `div`), and runs clean under AddressSanitizer.
- [ ] `sizeof` and field offsets in the emitted C match a C declaration of the same struct, checked by a C `_Static_assert` the emitter writes.
- [ ] Errors for a field type the layout does not allow, with codes in E1610-E1619.
- [ ] `sh scripts/test.sh` passes.

## Sizing

M: one attribute, one check, one emitter path.
