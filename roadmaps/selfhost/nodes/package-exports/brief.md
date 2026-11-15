---
id: package-exports
title: "Package exports table"
status: done
priority: could
type: feature
story: batch-16
early: false
outline: false
milestone: p4-packages
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
added: 2026-11-09
spent:
owns:
  - roadmaps/selfhost/nodes/package-exports/**
  - testdata/packages/exports_*/**
  - testdata/errors/exports_*
shared:
  - selfhost/src/manifest.nra
  - selfhost/src/modules.nra
  - selfhost/src/scopes.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/stdtext.nra
  - std/src/*.nra
  - SPEC.md
  - DECISIONS.md
  - docs/guide/*.md
reserves: [diag E0470-E0479]
provides: []
depends_on:
  - on: package-dependencies
    kind: behavior
    via: dependencies
---

# Package exports table

## Goal

Build the decision on package exports (DECISIONS.md, "A package lists the module paths it exposes"): an `[exports]` table in `narra.toml` maps import paths to module files. Other packages may import only listed paths; without the table every module is exposed, as today. Inside the package, modules import each other by their real paths.

## Context

- Keys must be the package name or start with `<name>/`. The bare name maps to one module (`import * as Json from "json"`). Values are paths under `src/` ending in `.nra`.
- Importing an unlisted path from another package is an error that lists the exposed paths.
- std has no `narra.toml`: its sources are embedded through `stdtext.nra`, and `scopes.nra` line 103 rejects `std/internal/*` with E0426. Give std the same rule through an exports list built into the compiler (every std module except `std/internal/*` and `std/ext/*`), and fold E0426 into the new code. Remove E0426 with no deprecation path.
- The decision says `narra move` updates the table. There is no `narra move` command, so that part is out of scope; record it in the outcome.
- Cross-package imports need package-dependencies. Until it lands, test against a fixture that the loader treats as a second package, then switch the cases to real git fixtures.

## Acceptance

- [ ] Errors in E0470-E0479: a key that does not start with the package name, a value that names no file, two keys for one bare name, an import of an unlisted path (the message lists the exposed paths).
- [ ] A fixture dependency with `[exports]` mapping `"lib"` to `src/index.nra` and `"lib/stream"` to `src/streaming/reader.nra` imports by both paths; a direct import of `lib/streaming/reader` from the app is an error.
- [ ] `std/internal/unicode` imported from a user package fails with the new code; E0426 is gone.
- [ ] SPEC section 15 documents `[exports]`; the decision record drops "Not built yet" for this part.

## Sizing

M: manifest table, one scope check, and the std list.
