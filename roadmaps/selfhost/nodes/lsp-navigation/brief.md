---
id: lsp-navigation
title: "narrac lsp: hover, definition, references and symbols"
status: active
priority: should
type: feature
story: batch-1
early: false
outline: false
milestone: m9-adoption
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
spent:
owns:
  - selfhost/src/lspnavigation.nra
  - testdata/lsp/nav_*/**
  - roadmaps/selfhost/nodes/lsp-navigation/**
shared:
  - selfhost/src/lsp.nra
  - selfhost/src/query.nra
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: language-service
    kind: hard
    why: adds request handlers to the server, its document store and its check results
---

# narrac lsp: hover, definition, references and symbols

## Goal

`textDocument/hover`, `textDocument/definition`, `textDocument/references` and `textDocument/documentSymbol` in `narrac lsp`, answered from the same site index and checker results as `narrac query` (`type-of`, `def`, `refs`, `interface`).

## Context

- Read `nodes/language-service/outcome.md` and `nodes/query-cli/outcome.md` (shapes, limits and gotchas of the site index).
- `query.nra` prints its answers. Move the lookups into functions that return data, keep `narrac query` output byte-identical (`TestSelfhostQuery`), and call them from the server.

## Acceptance

- [ ] Hover shows the `type-of` answer as a Markdown code block (type, plus `uses` effects when there are any). Nothing under the cursor gives `null`.
- [ ] Definition finds the declaration of a function, constant, struct, enum, constructor, trait or distinct type, and of a local or parameter (query-cli has no `def` for locals; add it here from the checker's local ids).
- [ ] References lists uses of functions, constants and constructors across the package, honouring `includeDeclaration`. Other kinds return an empty list (query-cli limit).
- [ ] Document symbols list the module's declarations with their kind and range, and enum constructors and struct fields as children.
- [ ] Positions use the encoding negotiated by language-service. Answers reflect unsaved edits.
- [ ] Scripted sessions in `testdata/lsp/nav_*/` cover each request, a cross-module definition, a request on a stale document after a change, and a position inside a template string (whole template, as in `query`).
- [ ] Fast check passes.

## Sizing

M. Wiring existing query functions into handlers, plus definition for locals; uncertainty low.
