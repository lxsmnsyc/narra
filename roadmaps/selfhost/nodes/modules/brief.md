---
id: modules
title: Multiple modules and narra.toml
status: active
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
  - compiler/internal/narra/modules*.go
  - compiler/internal/narra/driver.go
  - compiler/cmd/narra/**
  - compiler/testdata/packages/**
shared:
  - compiler/internal/narra/check.go
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/emit.go
  - compiler/internal/narra/emit_types.go
  - compiler/internal/narra/emit_call.go
  - compiler/internal/narra/types.go
  - compiler/internal/narra/builtins.go
  - compiler/internal/narra/check_types.go
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/ast.go
reserves: []
provides: [module-graph]
depends_on:
  - on: split-compiler
    kind: hard
    why: rewrites the checker entry point and driver that split-compiler moves
---

# Multiple modules and narra.toml

## Goal

Compile packages made of several files: package-rooted imports between modules, `export` visibility, import cycle errors, and `narra.toml` with `[package]` and `[[binary]]` entries. Freeze `contracts/module-graph.md`.

## Context

Today one file is one program and only `std/*` modules can be imported. Section 3 and section 15 of `SPEC.md` define modules and `narra.toml`. Imports are static and never run code. Use a minimal TOML reader for the keys the contract lists; no third-party modules.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Acceptance

- [ ] A sample package in `compiler/testdata/packages/` with three modules and one binary builds with `narra build <dir>` and runs.
- [ ] Error tests: importing an unexported name, an import cycle, an unknown module.
- [ ] `narra run file.nra` still works for single files.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L: size medium, new to this repo (module resolution and a TOML subset), and it changes the driver.

## Notes for the worker

- Keep type checking per module against exported signatures only, as the spec says. Interface files (`.nrai`) and incremental builds are not part of this node.
