---
id: selfhost-emit
title: Self-hosted C emitter: no skips left, and the selfhost package compares
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
size: S
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostEmit'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns:
  - roadmaps/selfhost/nodes/selfhost-emit/**
shared:
  - compiler/internal/narra/selfhost_emit_test.go
  - compiler/testdata/emit/skips/**
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: selfhost-emit-program
    kind: hard
    why: closes the skip ratchet; needs every emitter node merged
  - on: selfhost-emit-decls
    kind: hard
    why: closes the skip ratchet; needs every emitter node merged
  - on: selfhost-emit-bodies
    kind: hard
    why: closes the skip ratchet; needs every emitter node merged
  - on: selfhost-emit-match
    kind: hard
    why: closes the skip ratchet; needs every emitter node merged
  - on: selfhost-emit-closures
    kind: hard
    why: closes the skip ratchet; needs every emitter node merged
  - on: selfhost-emit-traits
    kind: hard
    why: closes the skip ratchet; needs every emitter node merged
  - on: selfhost-emit-maps
    kind: hard
    why: closes the skip ratchet; needs every emitter node merged
  - on: selfhost-emit-std
    kind: hard
    why: closes the skip ratchet; needs every emitter node merged
  - on: selfhost-emit-options
    kind: hard
    why: closes the skip ratchet; needs every emitter node merged
---

# Self-hosted C emitter: no skips left, and the selfhost package compares

## Goal

Remove the last skip files and make any `unsupported:` line a test failure. Add the selfhost package itself as an input: `narrac emit selfhost` equals `narra emit selfhost` (58,893 lines of C). This is the input selfhost-bootstrap starts from.

## Context

- narrac runs under AddressSanitizer, UBSan and the leak check in every harness (`buildSelfhost`), so the emitter is already sanitizer-checked on all inputs. The generated C is byte-identical to Go's, which `TestPrograms` already runs under ASan, so running it again is not needed.

## Acceptance

- [ ] `compiler/testdata/emit/skips/` is gone; the harness fails on any `unsupported:` and on exit code 3.
- [ ] The selfhost package is an input and compares byte-for-byte. The outcome gives the time of `narrac emit selfhost` against Go.
- [ ] `compiler/README.md` says the self-hosted compiler emits C.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

S: a few harness lines plus fixes for divergences that only the selfhost package shows. If those fixes grow past S, split them into fix nodes.
