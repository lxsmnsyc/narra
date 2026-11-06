---
id: unused-imports
title: "Unused imports and private functions are errors"
status: todo
priority: should
type: feature
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestCompileErrors|TestPrograms|TestSelfhost|TestGuide'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - compiler/testdata/errors/unused_import*
  - compiler/testdata/errors/unused_function*
  - roadmaps/selfhost/nodes/unused-imports/**
shared:
  - compiler/internal/narra/check_unused.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/README.md
  - selfhost/src/**
  - std/src/**
  - docs/guide/**
  - compiler/testdata/**
reserves: [code E0946, code E0947]
provides: []
depends_on:
  - on: unused-names
    kind: hard
    why: extends the unused-name pass and its README entry
---

# Unused imports and private functions are errors

## Goal

Finish SPEC section 2 "Unused names are errors": an unused import and an unused private (not exported) function are compile errors, in Go and selfhost. unused-names built locals and parameters and left these out; the maintainer approved this follow-up on 2026-11-07.

## Context

- `nodes/unused-names/outcome.md` and the "Differences from the spec" entry in `compiler/README.md` say what is built and what is not.
- An import counts as used when any imported name (or the module alias) is referenced in the file. A private function counts as used when any function in its module calls or references it, other than itself (a function that only calls itself is unused).
- `_` prefix marks either as unused on purpose, as for locals.
- Test files: a private function used only by a `test` block counts as used.

## Acceptance

- [ ] E0946 (unused import name) and E0947 (unused private function) in both checkers, with the same messages, and an add-underscore or remove fix in `narra check --fix` if the unused-names fixes have one.
- [ ] Error tests `compiler/testdata/errors/unused_import*.nra` and `unused_function*.nra`: a named import, a module alias import, a private function, a self-recursive private function, and a clean sample where a test block is the only user.
- [ ] Every unused import and private function in `selfhost/src`, `std/src`, test programs and guide samples removed or renamed. The outcome counts them.
- [ ] `compiler/README.md` drops imports and private functions from the unused-names difference.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test -timeout 45m ./...`.

## Sizing

M. The rule is small; the cleanup across `selfhost/src` and the test programs is most of the work.
