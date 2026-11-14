---
id: field-shorthand
title: "Field shorthand in struct literals"
status: todo
priority: must
type: feature
story: batch-15
early: false
outline: false
milestone: p3-review
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
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/field-shorthand/**
  - selfhost/src/fieldshorthand.nra
  - testdata/field_shorthand*
  - testdata/errors/field_shorthand*
  - testdata/astdump/field_shorthand*
  - testdata/fmt/field_shorthand*
shared:
  - selfhost/src/ast.nra
  - selfhost/src/parser.nra
  - selfhost/src/astdump.nra
  - selfhost/src/fmt.nra
  - selfhost/src/exprs.nra
  - selfhost/src/query.nra
  - selfhost/src/lsp*.nra
  - docs/guide/*.md
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1280-E1284]
provides: []
depends_on: []
---
# Field shorthand in struct literals

## Goal

Build D119's field shorthand: `Point { x, y }` means `Point { x: x, y: y }`, in struct literals, enum struct-payload constructors and generic struct literals with type arguments. Mixed forms (`Point { x, y: 2 }`) and spreads work as before.

## Context

- Source: D119 in `DECISIONS.md` and SPEC. They override this brief.
- fmt keeps the form the author wrote: `x` stays `x`, `x: x` stays `x: x`. Add a flag on the field node in `ast.nra` (a field default so other literals need no change).
- A shorthand name that is not in scope reports the usual unknown-name error at the field, with a hint naming the field. Use a code from E1280-E1284 only if the existing message cannot say it.
- Language service: hover and go-to-definition on a shorthand field reach both the field and the local; rename of the local rewrites `x` to `x: newName`.

## Acceptance

- [ ] `testdata/field_shorthand.nra` with output: plain, mixed, enum payload, generic, nested, with a spread, and a shorthand of a parameter, a `let` and a top-level constant. Clean under the copy check.
- [ ] Error tests for an unknown name and a duplicate field via shorthand.
- [ ] astdump and fmt goldens (both forms kept). SPEC and the guide's structs section show the form.
- [ ] `sh scripts/test.sh` passes.

## Notes

Do not use the shorthand in `selfhost/src` or std: the seed cannot compile it until the batch seed update. The sweep is field-shorthand-sweep.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
