---
id: query-hovers
title: "narrac query hovers: every type and diagnostic of a file"
status: done
priority: could
type: feature
story: batch-17
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/query-hovers/**
  - selfhost/src/tools/query*.nra
  - testdata/query/**
shared:
  - selfhost/src/tools/lsp*.nra
  - seed/narrac.c
  - SPEC.md
  - docs/guide/12-tools.md
reserves: []
provides: [query-hovers]
depends_on:

---
# narrac query hovers

## Goal

A build-time twoslash for Narra (narra-twoslash) needs, for one source file, the hover text of every identifier and every diagnostic, from one process. Today `narrac query type-of FILE:LINE:COL` answers one position, and the LSP answers hover per request.

Add `narrac query hovers FILE` (with `--json`): one JSON object with `hovers` (each with line, column, length and text, the same text the LSP hover shows) and `diagnostics` (line, column, length if known, code, message). It works on a file that has errors: diagnostics are listed and hovers cover what checked. Reuse the LSP hover code; do not write a second type printer.

Freeze the JSON shape in `roadmaps/selfhost/contracts/query-hovers.md` first (status frozen, version 1), so narra-twoslash can build against a stub.

## Acceptance

- [ ] Contract query-hovers frozen.
- [ ] Cases in testdata for a clean file, a file with an error, and generics.
- [ ] `sh scripts/test.sh` passes; seed regenerated until "seed unchanged".
- [ ] The tools chapter and SPEC list the command.
