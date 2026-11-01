---
id: std-io-more
title: stdin, directories and paths
status: todo
priority: should
type: feature
outline: true
milestone: m10-std
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: std grows on the self-hosted compiler, so nothing is built in Go and ported
---

# stdin, directories and paths

## Goal

Fill the I/O gaps real tools hit: read stdin (lines and exact byte counts, for LSP framing), write to stderr, flush stdout, create and remove files and directories, and a `std/path` with join, parent, base name and extension. Same `Result<T, string>` error convention as `std/fs`.

## Notes

Outline. These are primitives, so they stay builtins whatever std-source decides.
