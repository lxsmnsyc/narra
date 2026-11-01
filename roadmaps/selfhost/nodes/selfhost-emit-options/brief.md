---
id: selfhost-emit-options
title: Self-hosted emitter: Option, Result and std/test builtins and the test main
status: todo
priority: must
type: feature
story:
early: false
outline: false
milestone: m8-selfhost
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostEmit'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns:
  - selfhost/src/emitoptions.nra
  - compiler/testdata/emit/skips/selfhost-emit-options.txt
  - roadmaps/selfhost/nodes/selfhost-emit-options/**
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-emit-core
    kind: contract
    via: emitter
  - on: selfhost-emit-bodies
    kind: hard
    why: its samples compare only once plain bodies, decls and program have landed
  - on: selfhost-emit-closures
    kind: hard
    why: the test main and Option/Result callbacks emit through closure environments; frees a wave-3 slot for traits on the critical path
---

# Self-hosted emitter: Option, Result and std/test builtins and the test main

## Goal

Emit the std/option, std/result and std/test builtins (callbacks through `callFunctionValue`) and the C main of a test build, so test-mode inputs compare.

## Context

- Go to port: the emit halves of `std_option.go` (205-340, about 140 lines), `std_result.go` (97-164, about 70) and `std_testing.go` (39-121, about 70), plus `emitTestMain` in `test_runner.go` (about 10). `testRunnerC` is in runtimetext.nra.
- Test-mode inputs: tests.nra, std_test*.nra, runner/mixed.nra and packages/runner.

## Acceptance

- [ ] The skip file is deleted. The skip list shrinks by exactly the inputs that listed only this node (the test-mode inputs, options.nra, std_option.nra, std_result.nra and others).
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 380 Narra lines in one file.
