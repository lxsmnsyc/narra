---
id: prelude-cnames
title: Keep prelude C names apart from module names
status: todo
priority: must
type: fix
outline: false
milestone: m3-core
target:
from:
size: S
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-28
owns:
  - compiler/testdata/packages/errors/prelude_module*/**
shared:
  - compiler/internal/narra/modules.go
reserves: []
provides: []
depends_on:
  - on: modules
    kind: hard
    why: the C name rule it fixes was added by modules
---

# Keep prelude C names apart from module names

## Goal

Prelude types have no module prefix (`E_Option__int`), so a module whose mapped path equals a prelude type name produces the same C names. Example: a single file `Option.nra` with `enum A_int`, plus a use of `Option<Array<int>>`, gives `redefinition of 'E_Option__A_int'`. Make the module-graph v2 mapping injective again.

## Context

Found by the review of PR #6. The fix should not change any existing generated C. Rejecting a module (or single file) whose mapped path equals a prelude type name (`Option`, `Result`, `Ordering`, `Array`, `Range` and any other prelude type) with a clear error is enough.

- Source: `reviews/wave-2b.md`, `contracts/module-graph.md` (v2)

## Acceptance

- [ ] An error package (or single-file error test) named after a prelude type gives a clear error instead of invalid C.
- [ ] Generated C for every existing test is unchanged.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

S: one check in the loader.

## Notes for the worker

- Record the rule in a change request against module-graph (v3), since it narrows valid module names.
