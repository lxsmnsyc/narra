---
id: effect-error-impl-name
title: Name implement functions by their source name in effect errors
status: todo
priority: should
type: fix
outline: false
milestone: m8-selfhost
target:
from:
size: S
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestTypesDump|TestSelfhostEffects|TestCheck'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-31
owns:
  - roadmaps/selfhost/nodes/effect-error-impl-name/**
shared:
  - compiler/internal/narra/check_effects.go
  - compiler/internal/narra/check_traits.go
  - selfhost/src/effects.nra
  - compiler/testdata/typesdump/effects/**
  - compiler/testdata/errors/**
reserves: []
provides: []
depends_on:
  - on: selfhost-check-effects
    kind: hard
    why: the Narra effect pass lands there; this fix changes the same error text in both checkers
---

# Name implement functions by their source name in effect errors

## Goal

An effect error that goes through an implement block function shows its C name with a hash. Example: an exported `greet` calls `Named.name(cat)`, the impl prints, and `greet` does not declare `uses io`. `narra check` reports ``(through `Cat_Named_name_5508671d`)``. The cause is `check_traits.go` (around line 425): `fd.Name = implName(...)` replaces the source name, and `check_effects.go:167` prints `cause.via` from it. The Narra checker (selfhost-check-effects, #44) prints ``(through `name`)``, so the two disagree, and no compared input covers it.

## Acceptance

- [ ] Go prints the implement function as `<Type>.<Trait>.<function>` (for example ``(through `Cat.Named.name`)``), the same form the types dump uses in `func=` (`implSourceName` in typesdump.go). `narra check`, `narra check --json` and `narra types` agree.
- [ ] The Narra checker prints the same text.
- [ ] A new error case covers it in `testdata/typesdump/effects/` (an `err_` golden), compared by `TestSelfhostEffects`, and an `errors/` case for `narra check` if the error tests have one per message.
- [ ] Fails on main, passes on the branch (show both runs).
- [ ] `narra emit` output unchanged on every sample; only this error text changes.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

S: one name lookup on each side, one or two goldens.
