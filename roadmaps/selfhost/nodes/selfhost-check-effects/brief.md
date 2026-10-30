---
id: selfhost-check-effects
title: Self-hosted checker: effect pass
status: todo
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostEffects'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - selfhost/src/effects.nra
  - compiler/internal/narra/selfhost_effects_test.go
  - compiler/testdata/typesdump/effects/**
shared:
  - selfhost/src/main.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/body.nra
  - selfhost/src/exprs.nra
  - compiler/internal/narra/typesdump.go
  - compiler/internal/narra/selfhost_test.go
  - compiler/testdata/typesdump/**
  - compiler/README.md
  - selfhost/src/check.nra
  - selfhost/src/declare.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-check-bodies
    kind: hard
    why: defines the body tables and the dump
  - on: selfhost-check-calls
    kind: hard
    why: effects follow the resolved callee of each call, including generic instances and trait calls
  - on: selfhost-check-patterns
    kind: hard
    why: the walker visits match arms, narrowing and aggregate literals
  - on: selfhost-check-closures
    kind: hard
    why: effects of arrow bodies and function values flow to the caller
  - on: selfhost-check-builtins
    kind: hard
    why: builtin effects and `callbacks` indexes feed the pass
---

# Self-hosted checker: effect pass

## Goal

Port the effect pass (`uses alloc`, `uses io`, inference for private functions, effect variables for function parameters, signature checks) to Narra, add it to the dump, and make the whole body checker complete: after this node no `unsupported:` skip remains in `TestSelfhostBodies`.

## Context

- Go source (about 512 lines, 700 Narra): `check_effects.go` (`effectState`, `checkEffects`, `eachBody`, `walkAll`, `analyse`, `checkSignature`, `funcEffects`, `valueEffects`, `concrete`, `effectWalker`: `block`, `stmt`, `expr`, `call`, `userCall`).
- The pass runs after all bodies and reads the resolution recorded by the earlier nodes (callee of each call, builtin, instance, trait function, function value). Read the outcomes of selfhost-check-bodies, -calls, -closures and -builtins for where those tables live.
- Dump: add `effects` lines (the final effect set of each function, sorted, and for each inferred private function) and print the effect errors with the other errors. The Go dump (`typesdump.go`) starts running `checkEffects` here; before this node it did not. Update the doc comment.
- Builtins: `Builtin.Effects` is set on every std builtin (`StdFunc.effects`); `callbacks` lists argument indexes whose effects are added to the call (std-option-result); `io` in a callback reaches the caller; `test` bodies discard their effects (test-runner).

## Acceptance

- [ ] `TestSelfhostEffects` compares `narrac types` with `narra types` (now with effects) on every input of `TestSelfhostBodies` plus `testdata/typesdump/effects/`: inferred and declared effects, higher-order functions with effect variables, `io` callbacks to `Option.map`, recursion and mutual recursion, generic instances, trait and impl functions, exported functions that must declare, test blocks, and the errors (missing `uses`, effect not declared, strict private higher-order function).
- [ ] `TestSelfhostBodies` skips only inputs that do not load. If any `unsupported:` is left, the test fails; remove the last stubs or list them in the outcome and keep the failure.
- [ ] `narrac check`-style run over `selfhost/` itself (the selfhost package) gives no error and matches Go.
- [ ] Earlier harnesses pass; `narra emit` output unchanged.
- [ ] The outcome states the Narra line total for the body half, the timing of `narrac types selfhost/` against Go, and what selfhost-emit may rely on.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 700 Narra lines, one self-contained pass. Context load medium: it reads four earlier outcomes. It is last in the chain, so it also closes the skip ratchet.

## Notes for the worker

- Fixpoint over a call graph: loop with a changed flag; do not recurse on call depth.
- Instance bodies: the effects of a generic function are those of its instances; see `concrete` in Go.
- `uses` on exported functions is mandatory (selfhost/CLAUDE.md); the pass must accept the selfhost sources.
- Private higher-order functions: the effect variable rule from the effects node (private functions with function parameters are inferred, not strict).
- Time `narrac types selfhost/`. Bodies of the whole selfhost package are the first large input; a quadratic cost shows here. The `sortStrings` insertion sort in `modules.nra` is a known one (inbox); fix it if it shows in the timing.
