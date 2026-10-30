---
id: strings-std
title: String positions, slicing and characters
status: active
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
  - compiler/internal/narra/std_string.go
  - compiler/internal/narra/runtime_string.h
  - compiler/testdata/strings2*
shared:
  - compiler/internal/narra/runtime.h
  - compiler/internal/narra/emit_types.go
  - compiler/internal/narra/emit_call.go
reserves: []
provides: []
depends_on:
  - on: split-compiler
    kind: hard
    why: edits std_string.go, which split-compiler creates and fills with the existing Str functions
---

# String positions, slicing and characters

## Goal

Add `ByteIndex` positions and the string functions a lexer needs: `Str.indexOf`, `Str.slice`, `Str.sliceFrom`, `Str.byteAt`, `Str.charAt`, `Str.nextChar`, `Str.chars`, `Str.trim`, `Str.split`, `Str.startsWith`, and a `StringBuilder` for building output.

## Context

See `SPEC.md` section 4, "Strings", and decisions D15 and D33. Slices share memory with the original string. `ByteIndex` can be a compiler-provided type until node opaque-distinct generalizes it.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

- From split-compiler: put new runtime C in the module's `Module.Runtime` (in its `std_*.go` file), not in `runtime.h`. The emitter writes it once per imported module.

- From io-std: `Str.split` and similar functions that return `Array<string>` must call `c.arrayOf(...)` in `Check` so the array type is emitted. A `Result`-returning function builds it with `ioResult` from `std_fs.go`. `testdata/errors/unknown_module.nra` matches a prefix, so adding modules needs no edit there.

- From hashed-cnames: C names for any new type (a `StringBuilder`) go through `cnames.go` with their own kind letter.

## Acceptance

- [ ] A sample program tokenizes a short source string into words and numbers using these functions.
- [ ] Panic tests: slicing inside a character, out-of-range slice.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: one module, mostly runtime C and registry entries.
