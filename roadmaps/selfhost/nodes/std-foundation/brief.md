---
id: std-foundation
title: Groundwork for std modules written in Narra
status: done
priority: should
type: feature
story:
early: false
outline: false
milestone: m10-std
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 -timeout 40m ./internal/narra -run 'TestCompileErrors|TestStage2Programs|TestStdSources|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent: 232000
owns:
  - compiler/testdata/stage2/**
  - compiler/testdata/errors/generic_mismatch_*
  - compiler/testdata/errors/std_internal_*
  - compiler/internal/narra/stage2_test.go
  - roadmaps/selfhost/nodes/std-foundation/**
shared:
  - compiler/internal/narra/generics.go
  - compiler/internal/narra/check_call*.go
  - compiler/internal/narra/modules.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/internal/narra/selfhost_bootstrap_test.go
  - compiler/testdata/errors/std_option_*
  - selfhost/src/generics.nra
  - selfhost/src/calls.nra
  - selfhost/src/modules.nra
  - selfhost/src/scopes.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E0325, diag E0426]
provides: []
depends_on:
  - on: std-source
    kind: hard
    why: builds on the source-module loader and the std/option port that the spike merges
---

# Groundwork for std modules written in Narra

## Goal

The three things every later m10 node needs, decided at the std-source gate on 2026-11-05: generic calls report the real mismatch, `std/internal/...` holds the C primitives behind source modules, and selfhost-only std modules have a test path. After this node, a std module is written as Narra source by default.

## Context

- Answer and measurements: `nodes/std-source/outcome.md` (sections "Error messages that changed", "Primitive form" and "Go options").
- Gate decisions: GOAL.md, Decisions, 2026-11-05.

## Steps

- [ ] **Generic mismatch errors.** When unification fails on a concrete argument (an `int` passed where `Option<T>` is expected), `genericCall` reports `expected Option<T>, found int` at the argument (E0325), instead of "cannot tell type parameter T". Both compilers. The `std_option_*` error tests go back to naming the mismatch. "cannot tell type parameter" stays for calls where nothing pins T.
- [ ] **`std/internal/`.** Primitives that a source module needs are builtin modules under `std/internal/<name>`, reusing `registerModule` (Go) and `STD_FUNCS` (selfhost). Only a module whose path starts with `std/` may import them. Any other import gets E0426, "std/internal/<name> is internal to the standard library". No module moves yet; add one test-only primitive or reuse an existing builtin to prove the import rule.
- [ ] **Stage 2 tests.** `TestStage2Programs` reuses the cached stage 2 binary from `TestSelfhostBootstrap`, emits each sample under `compiler/testdata/stage2/` (and `stage2/errors/`), compiles it with the sanitizer flags, and compares `.out`, the error header and the leak check. The Go tests skip `stage2/`. One sample and one error test prove it.

## Acceptance

- [ ] The `std_option_*` error tests show the mismatch message; a new `generic_mismatch_*` test covers a user generic function.
- [ ] `std_internal_*` error tests: a user module importing `std/internal/...` fails with E0426; a std source module importing it compiles.
- [ ] `TestStage2Programs` runs the stage 2 samples, clean under ASan and the leak check, and fails when an expected output differs.
- [ ] SPEC.md describes `std/internal/` in the std section; `compiler/README.md` says which std modules are selfhost-only and how they are tested.
- [ ] Fast check passes with `-timeout 40m`, including `TestSelfhostBootstrap`.

## Sizing

M. Uncertainty low (the spike mapped all three), coupling low (one dependency), blast radius medium (generic call checking and module loading in both compilers), verifiability low (error and stage 2 tests), size medium, context medium.

## Notes for the worker

- Under option (b), the Go compiler loads a std source module only when `selfhost/src` imports it. Every other new std module is selfhost-only and is tested under `testdata/stage2/`.
- Keep the generic mismatch fix general: no Option-specific cases. It helps every user generic too.
- From the std-source review: remove `std/option` from the `emitOptionCall` condition in `selfhost/src/emitexprs.nra` (about line 443), which can no longer be reached; sort std source modules with `orderModules` in single-file packages (`compiler/internal/narra/modules.go` about line 136, `selfhost/src/modules.nra` about line 88) before a second source module imports another; and `Options.isSome(3)` should report only the mismatch, not also "expected bool, found void".
