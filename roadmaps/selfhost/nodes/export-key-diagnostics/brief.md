---
id: export-key-diagnostics
title: "Diagnostics name the export key the importer wrote"
status: done
priority: could
type: fix
story: batch-20
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
added: 2026-11-19
spent:
owns:
  - roadmaps/selfhost/nodes/export-key-diagnostics/**
shared:
  - selfhost/src/driver/diagnostics.nra
  - selfhost/src/check/scopes.nra
  - selfhost/src/check/unused.nra
  - selfhost/src/driver/**
  - testdata/packages/**
  - selfhost/tests/src/**
  - compiler/README.md
reserves: []
provides: []
depends_on: []
---
# Diagnostics name the export key the importer wrote

## Goal

From package-exports (inbox, 2026-11-16):

- Diagnostics and `narrac pkg` dumps show the real module path (`lib/streaming/reader`), not the export key the importer wrote (`lib/stream`). Show the key the importer wrote, with the real path where it helps (for example in a note).
- The std E0474 message does not list the exposed std paths. List them, as the "unknown module" message already does.
- The remove-unused-import fix (`selfhost/src/check/unused.nra`) stops at the first quoted line for a rewritten import, and no case covers it. Add the case; fix it if it fails.

## Acceptance

- [ ] A case per item that fails on the batch base and passes on the branch.
- [ ] `sh scripts/test.sh` passes.
