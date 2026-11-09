---
id: template-nested-strings
title: "Template interpolation scan skips nested strings"
status: done
priority: should
type: fix
story: batch-2
early: false
outline: false
milestone: m12-spec-gaps
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
added: 2026-11-09
spent: 80465
owns:
  - testdata/template_nested*
  - testdata/errors/template_nested*
  - roadmaps/selfhost/nodes/template-nested-strings/**
shared:
  - selfhost/src/**
  - std/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes go into the Narra compiler only, after the bootstrap fixed point
---

# Template interpolation scan skips nested strings

## Goal

The `${...}` scan in `parseTemplate` (`selfhost/src/parser.nra`) counts `{` and `}` inside string literals in the expression, so `` `${"}"}` `` ends the interpolation early. The raw scan in `lexTemplate` (`selfhost/src/lexer.nra`) likewise ends the template at a backtick inside a nested string or nested template. Both scans must skip string literals, nested templates and their escapes.

## Acceptance

- [ ] Proof on the base: a sample with `` `${"}"}` ``, `` `${"`"}` `` and a nested template fails before the change.
- [ ] Sample `testdata/template_nested*.nra` with `.out` passes, clean under AddressSanitizer.
- [ ] An unterminated string inside `${...}` reports an error at the string, not at the backtick.
- [ ] `sh scripts/test.sh` passes.
