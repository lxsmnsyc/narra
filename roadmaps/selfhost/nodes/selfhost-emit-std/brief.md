---
id: selfhost-emit-std
title: Self-hosted emitter: string, char, int, float, env, fs and process builtins
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
  - selfhost/src/emitstd.nra
  - compiler/testdata/emit/skips/selfhost-emit-std.txt
  - roadmaps/selfhost/nodes/selfhost-emit-std/**
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
---

# Self-hosted emitter: string, char, int, float, env, fs and process builtins

## Goal

Emit every std builtin of std/string, std/char, std/int, std/float, std/env, std/fs and std/process, including `ioResult` and `ioEither` for the fallible ones. Dispatch is by `StdFunc.name`.

## Context

- Go to port: the `Emit` functions in `std_string.go` (156 lines, `strFn`), `std_char.go` (32), `std_int.go` (83), `std_float.go` (35), `std_env.go` (40), `std_fs.go` (162: `ioResult`, `ioEither`, `stringOut`) and `std_process.go` (37). About 300 Go lines are emit code; the checks are already ported.
- The runtime C of these modules is already in runtimetext.nra (core).

## Acceptance

- [ ] Every builtin of these modules has an emit case; an unknown name is an internal error, not a stub.
- [ ] The skip file is deleted. The skip list shrinks by exactly the inputs that listed only this node (strings*.nra, convert.nra, io_files.nra and others).
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 400 Narra lines in one file of many small cases.
