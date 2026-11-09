---
id: bytes-type
title: "The bytes type and std/bytes"
status: done
priority: should
type: feature
story: batch-1
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent: 165601
owns:
  - selfhost/src/stdbytes.nra
  - testdata/bytes*
  - testdata/errors/bytes_*
  - roadmaps/selfhost/nodes/bytes-type/**
shared:
  - selfhost/runtime/**
  - selfhost/src/types.nra
  - selfhost/src/resolve.nra
  - selfhost/src/traits.nra
  - selfhost/src/emittraits.nra
  - selfhost/src/stdfs.nra
  - selfhost/src/stdlib.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E0942]
provides: []
depends_on:
  - on: sized-numbers
    kind: hard
    why: byte values are u8
  - on: std-tables-split
    kind: hard
    why: adds a std module and std/fs functions in the per-module files the split creates
---

# The bytes type and std/bytes

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

Spec-gap F7, `bytes` part: the immutable byte sequence of SPEC section 4 in both compilers, with a `std/bytes` module and binary file I/O.

## Context

- Read `nodes/sized-numbers/outcome.md` and `nodes/std-tables-split/outcome.md`.
- `string` is `lstr*` (length plus bytes, reference counted). `bytes` reuses that representation without the UTF-8 guarantee, so conversions from `string` share the buffer and conversions to `string` validate UTF-8.

## API

| Function | Result | Meaning |
| --- | --- | --- |
| `Bytes.fromString(text)` | `bytes` | the UTF-8 bytes of `text`, shared buffer |
| `Bytes.toString(data)` | `Result<string, string>` | Err when not valid UTF-8, with the byte offset |
| `Bytes.length(data)` | `int` | number of bytes |
| `Bytes.at(data, index)` | `u8` | panics out of range |
| `Bytes.slice(data, start, end)` | `bytes` | copies, as string slices do (D60); panics out of range |
| `Bytes.concat(left, right)` | `bytes` | |
| `Bytes.fromArray(items)` / `Bytes.toArray(data)` | `bytes` / `Array<u8>` | |
| `Fs.readBytes(path)` / `Fs.writeBytes(path, data)` | `Result<bytes, string>` / `Result<bool, string>` | as `readText` and `writeText`, without the UTF-8 check |

`bytes` implements `Equatable`, `Hashable` and `Show` (`bytes[68 69 0a]`, hex pairs), and works as a map key. `Ordered` compares bytewise.

## Acceptance

- [ ] Sample `bytes.nra` and `.out`: each function, invalid UTF-8 in `toString`, a binary file written and read back in a temp directory, `bytes` as a map key, `Show`. ASan and leak check clean.
- [ ] Error tests: `Bytes.at` with a non-`int` index, a `string` passed where `bytes` is expected (E0119), and E0942 if a new format is needed.
- [ ] SPEC "Standard modules" lists `std/bytes`; `@unique let buffer: Array<u8>` in SPEC section 6 compiles once std-spec-functions adds `Arrays.withCapacity`. README updated. Go and `selfhost/` emit the same C; fast check passes.

## Sizing

M. A primitive type that reuses the string runtime plus one builtin module; uncertainty low.

Landed on feat/batch-1 (47b1e5f) after a clean rebase onto mutable-for-of; touched suites passed. Gate decided by the orchestrator: bytes is a primitive sharing the string representation; Fs.writeBytes returns Result<int, string> like writeText (the table said bool); < <= > >= allowed. Left for later: Shape for bytes, Io.readBytes still returns string, no Bytes or sized-integer entries in STD_ALIAS_PATHS.
