---
id: selfhost-emit-program
title: Self-hosted emitter: C names, xxHash32 and the checker facts the emitter reads
status: done
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
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostEmit|TestSelfhostXXH32|TestSelfhostModules|TestSelfhostBodies|TestSelfhostEffects|TestXXH32'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent: 268404
owns:
  - selfhost/src/hash.nra
  - selfhost/src/cnames.nra
  - compiler/testdata/emit/skips/selfhost-emit-program.txt
  - compiler/testdata/packages/errors/module_hash_clash/**
  - roadmaps/selfhost/nodes/selfhost-emit-program/**
shared:
  - selfhost/src/check.nra
  - selfhost/src/declare.nra
  - selfhost/src/traits.nra
  - selfhost/src/instances.nra
  - selfhost/src/calls.nra
  - selfhost/src/funcvalues.nra
  - selfhost/src/resolve.nra
  - selfhost/src/generics.nra
  - selfhost/src/stdcalls.nra
  - selfhost/src/closures.nra
  - selfhost/src/literals.nra
  - selfhost/src/match.nra
  - selfhost/src/exprs.nra
  - selfhost/src/body.nra
  - selfhost/src/modules.nra
  - selfhost/src/main.nra
  - compiler/internal/narra/selfhost_emit_test.go   # adds TestSelfhostXXH32 only
reserves: []
provides: []
depends_on:
  - on: selfhost-emit-core
    kind: contract
    via: emitter
---

# Self-hosted emitter: C names, xxHash32 and the checker facts the emitter reads

## Goal

Port `cnames.go`, and make the Narra checker record what Go hands to the emitter. That covers function order, constant order, composite types, local numbers, call-to-instance links, the Self type of trait calls, and body lookup. Also port the module hash clash check that selfhost-modules deferred. After this node the C name comment table, the array and type typedef lines and the `// Functions` prototypes match Go on every input.

## Context

- Go to port: `cnames.go` (290 lines). The checker sites are in the contract's "Checker facts" list: `check.go:279`, `check_traits.go:429`, `check_call.go:475`, `check_const.go:88`, `check.go:416` and `check_match.go:44`, and the composite sites (`check_types.go:300-309, 471-503`, `check_expr.go:180`, `check_closure.go:184, 213, 476`, `std_env.go:29`, `std_map.go:196, 245, 247`, `std_array.go:83`, `std_box.go:31`, `std_fs.go:139`, `std_process.go:18`, `std_string.go:57, 83, 100`, `std_option.go:97, 100`).
- The module clash check: `modules.go:189` in `LoadPackage`. It needs xxh32, so `hash.nra` imports only std, and `modules.nra` imports it.

## Rules

- Record each fact at the Narra point that matches Go's call site, not by walking types afterwards. Go creates composites in unused signatures and std checks too.
- Do not change dump output. `TestSelfhostBodies` and `TestSelfhostEffects` keep their counts.
- Set `programRecorded` when done.

## Acceptance

- [ ] `TestSelfhostXXH32` passes the `TestXXH32` vectors through `narrac hash`.
- [ ] `packages/errors/module_hash_clash/` holds two modules whose relative paths collide (find a pair with a Go scratch search). Go and `narrac pkg` report the same error.
- [ ] In `TestSelfhostEmit`, the comment table, the typedef lines before the first marker, and the whole `// Functions` section match Go on every input.
- [ ] The skip file `selfhost-emit-program.txt` is deleted. The skip list shrinks by exactly the inputs that listed only this node.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 300 Narra lines for names and hashing, and about 200 lines of small edits across about 12 checker files. Blast radius medium (append-only fields and recording calls). Verified by the comment table and the prototypes.

## Notes for the worker

- xxh32 on int64: use `*%` and `+%` (wrapping), then mask with `& 0xFFFFFFFF`. Check SPEC.md on `<<` overflow before relying on it.
- Sort the composites by Go's key string: arrays and boxes by `typeId(element)`, maps by `typeId(key)|typeId(value)`, function types by `typeId(type)`. Use a byte-order merge sort, not `sortStrings` in modules.nra (insertion sort).
- Every local declaration point needs `number`, including pattern bindings, arrow parameters and per-instance rechecks. One off-by-one renames every later local; the prototypes show it first.
- The comment table prints Go's `Type.String()`. Check that `typeString` agrees for function types, tuples and distinct types.
