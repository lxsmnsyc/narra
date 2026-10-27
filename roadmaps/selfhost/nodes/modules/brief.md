---
id: modules
title: Multiple modules and lang.toml
status: todo
priority: must
type: feature
outline: false
milestone: m3-core
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns:
  - compiler/internal/langc/modules*.go
  - compiler/internal/langc/driver.go
  - compiler/cmd/langc/**
  - compiler/testdata/packages/**
shared:
  - compiler/internal/langc/check.go
  - compiler/internal/langc/parser.go
  - compiler/internal/langc/emit.go
  - compiler/internal/langc/emit_types.go
  - compiler/internal/langc/emit_call.go
  - compiler/internal/langc/types.go
  - compiler/internal/langc/builtins.go
  - compiler/internal/langc/check_types.go
  - compiler/internal/langc/check_call.go
  - compiler/internal/langc/ast.go
reserves: []
provides: [module-graph]
depends_on:
  - on: split-compiler
    kind: hard
    why: rewrites the checker entry point and driver that split-compiler moves
---

# Multiple modules and lang.toml

## Goal

Compile packages made of several files: package-rooted imports between modules, `export` visibility, import cycle errors, and `lang.toml` with `[package]` and `[[binary]]` entries. Freeze `contracts/module-graph.md`.

## Context

Today one file is one program and only `std/*` modules can be imported. Section 3 and section 15 of `SPEC.md` define modules and `lang.toml`. Imports are static and never run code. Use a minimal TOML reader for the keys the contract lists; no third-party modules.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Acceptance

- [ ] A sample package in `compiler/testdata/packages/` with three modules and one binary builds with `langc build <dir>` and runs.
- [ ] Error tests: importing an unexported name, an import cycle, an unknown module.
- [ ] `langc run file.lang` still works for single files.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L: size medium, new to this repo (module resolution and a TOML subset), and it changes the driver.

## Notes for the worker

- Keep type checking per module against exported signatures only, as the spec says. Interface files (`.langi`) and incremental builds are not part of this node.
