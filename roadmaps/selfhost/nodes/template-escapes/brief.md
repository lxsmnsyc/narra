---
id: template-escapes
title: "Template strings decode every escape plain strings do"
status: active
priority: should
type: fix
story: batch-4
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - testdata/template_escapes*
  - roadmaps/selfhost/nodes/template-escapes/**
shared:
  - selfhost/src/parser.nra
  - selfhost/src/lexer.nra
  - selfhost/src/lsp.nra
  - selfhost/tests/src/lsp.nra
  - roadmaps/selfhost/inbox.md
reserves: []
provides: []
depends_on: []
---

# Template strings decode every escape plain strings do

## Goal

`parseTemplate` (`selfhost/src/parser.nra:3082`) decodes only `\n` and `\t`; `\r`, `\0` and `\u{...}` come out as the bare letter (`` `\r\n` `` gives `r` and a newline). Plain strings decode them in `lexEscape` (`selfhost/src/lexer.nra:541`). Make templates share the plain-string decoding, then drop the `CRLF` constant workaround in `selfhost/src/lsp.nra:43` and `selfhost/tests/src/lsp.nra:29`.

## Context

- Inbox entry 2026-11-09, from language-service.
- Proof first: a sample printing the bytes of `` `\r\0\u{41}` `` fails on the base.
- Seed rule: `selfhost/src` may drop the workaround only after the seed decodes the escapes. Land the fix, let the orchestrator regenerate the seed, then remove `CRLF` (or leave that edit for the end of the batch and say so in the outcome).

## Acceptance

- [ ] `testdata/template_escapes.nra` covers every escape `lexEscape` accepts, plus an invalid escape error test if plain strings have one.
- [ ] `sh scripts/test.sh` passes; inbox entry removed.

## Sizing

S: one decoding path shared, one sample.
