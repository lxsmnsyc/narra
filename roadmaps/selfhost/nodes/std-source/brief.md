---
id: std-source
title: Std modules as Narra source
status: todo
priority: should
type: spike
outline: true
milestone: m10-std
target:
from:
size: M
spike: true
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-01
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: std grows on the self-hosted compiler, so nothing is built in Go and ported
---

# Std modules as Narra source

## Goal

Answer one question: can std modules ship as Narra source files bundled with the compiler, with only primitives (memory, syscalls, string bytes) as builtins? Find the smallest primitive set, how a source std module reaches C functions without full c-interop, and what it costs in build time and binary size. Port one small existing module (`std/option`) as the proof. Every compiler builtin is currently emitted C, so each new std function means editing the emitter; this spike decides the cost of everything else in m10.

## Notes

Outline, gated: the user reviews the answer before the other m10 nodes are refined. If the answer is "needs c-interop first", raise c-interop out of `later`.
