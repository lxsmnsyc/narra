---
id: std-hex-base64
title: "Hex and base64 text of bytes in std/bytes"
status: todo
priority: should
type: feature
story: batch-24
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: S
risk: low
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/std-hex-base64/**
  - testdata/bytes_encoding*
shared:
  - selfhost/src/std/stdtext.nra
  - docs/reference/std/README.md
  - docs/reference/std/*.md
  - std/src/*.nra
reserves: []
provides: []
depends_on: []
---

# Hex and base64 text of bytes in std/bytes

## Goal

`std/bytes` gets `toHex`, `fromHex`, `toBase64`, `fromBase64` and the URL-safe base64 alphabet (with and without padding). Decoding bad text gives `None`, not a panic.

## Context

- Source: std-gap-survey rank 13: 16 hand-written hex functions (`driver/hash.nra`, `lexer.nra`, `json.nra`, `toml.nra`, `scripts/unicode/generate.nra`). `GOAL.md` Decisions, 2026-11-21, keeps hex and base64 in std.
- `std/bytes` today has `at`, `concat`, `fromArray`, `fromString`, `length`, `slice`, `toArray`, `toString`. Check where `std/bytes` is defined before adding to it (runtime builtins or `std/src`).
- Replacing the 16 copies in the compiler is not part of this node. It waits for a seed that knows the new functions.

## Acceptance

- [ ] The RFC 4648 section 10 test vectors pass for base64 and hex (base16), both directions.
- [ ] `fromHex("zz")`, odd-length hex and base64 with bad padding give `None`.
- [ ] `sh scripts/test.sh` passes.

## Sizing

S: four functions and an alphabet flag.
