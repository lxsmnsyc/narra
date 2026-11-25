---
id: std-digest
title: "std/digest: SHA-256, SHA-1, HMAC and PBKDF2"
status: active
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
size: M
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/std-digest/**
  - std/src/digest.nra
  - testdata/digest*
shared:
  - selfhost/src/std/stdtext.nra
  - docs/reference/std/README.md
  - docs/reference/std/*.md
reserves: []
provides: []
depends_on:
  - on: std-hex-base64
    kind: hard
    why: tests and `Digest.hex` print digests through `Bytes.toHex`
---

# std/digest: SHA-256, SHA-1, HMAC and PBKDF2

## Goal

A new `std/digest` module computes SHA-256 and SHA-1 over bytes, both in one call and in parts (`create`, `update`, `finish`), plus HMAC over either hash and PBKDF2-HMAC-SHA-256. Written in Narra, with no C library.

## Context

- Source: std-gap-survey rank 14; `GOAL.md` Decisions, 2026-11-21 (digests stay in std). `scripts/unicode/generate.nra` has its own `sha256`; start from it.
- HMAC and PBKDF2 are here because the `pkg-postgres` node needs SCRAM-SHA-256 login, and `pkg-websocket` needs SHA-1 for the handshake key. They are small on top of the hashes.
- This is not a crypto library. Ciphers, signatures and constant-time comparison beyond `Digest.equal` belong to a later `crypto` package.

## Acceptance

- [ ] NIST test vectors for SHA-256 and SHA-1 (empty input, `abc`, the 448-bit message, one million `a`) pass, in one call and fed in parts of uneven sizes.
- [ ] RFC 4231 HMAC-SHA-256 vectors and RFC 7914 section 11 PBKDF2-HMAC-SHA-256 vectors pass.
- [ ] A benchmark hashes 64 MiB at -O2 and its time is recorded in the outcome next to `shasum -a 256`.
- [ ] `sh scripts/test.sh` passes.

## Sizing

M: two compression functions, the streaming state, HMAC and PBKDF2.

## Notes for the worker

- Use the wrapping integer operations Narra has for `u32`; check SPEC for overflow rules before writing the rounds.
