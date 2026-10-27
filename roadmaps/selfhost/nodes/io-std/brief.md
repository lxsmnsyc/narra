---
id: io-std
title: Files, processes and environment
status: todo
priority: must
type: feature
outline: false
milestone: m3-core
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns:
  - compiler/internal/langc/std_fs.go
  - compiler/internal/langc/std_process.go
  - compiler/internal/langc/std_env.go
  - compiler/internal/langc/runtime_io.h
  - compiler/testdata/io*
shared: []
reserves: []
provides: []
depends_on:
  - on: split-compiler
    kind: contract
    via: builtins
---

# Files, processes and environment

## Goal

Add `std/fs` (`readText`, `writeText`, `exists`, `listDir`), `std/process` (`run` a command and get its exit code and output), and `std/env` (`get`, `args`). Fallible functions return `Result`.

## Context

The self-hosted compiler must read source files, write C and run the C compiler. These functions have the `io` effect; effect checking itself is node effects.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Acceptance

- [ ] A sample program writes a file, reads it back, and runs `echo` through `std/process`.
- [ ] Errors come back as `Err` values, never panics.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: runtime C wrappers around POSIX calls.
