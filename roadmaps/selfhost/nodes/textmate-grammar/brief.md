---
id: textmate-grammar
title: "TextMate grammar for Narra, and narra fences in SPEC.md"
status: done
priority: could
type: feature
story: batch-16
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
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent: 165871
owns:
  - roadmaps/selfhost/nodes/textmate-grammar/**
  - editors/textmate/**
shared:
  - SPEC.md
reserves: []
provides: []
depends_on: []
---
# TextMate grammar for Narra, and narra fences in SPEC.md

## Goal

One TextMate grammar, `editors/textmate/narra.tmLanguage.json`, that the website's highlighter and editors can both load. It covers keywords, `uses` effect lists, annotations (`@const`, `@export(C)`), string templates and tagged templates, numeric literals with suffixes, comments, types after `:`, and `function*` and `async`. Then switch the 63 `ts` fences in `SPEC.md` to `narra`.

## Context

- Keywords and literal forms: `selfhost/src/lexer.nra` is the source of truth. Derive the keyword list from it, not from memory.
- Tests: building and testing the compiler need only `cc` and `sh`, so the grammar's test must not join `scripts/test.sh`. Put a small check under `editors/textmate/` (for example a Node script with `vscode-textmate` and `vscode-oniguruma`) that tokenizes every guide sample and fails on an `invalid` scope or an unclosed string or template. The website build runs it.
- Switching SPEC fences: a `ts` fence that holds TypeScript on purpose (a comparison) stays `ts`. Check each.

## Acceptance

- [ ] The grammar tokenizes every `narra` fence in `docs/guide` with no `invalid` scopes, checked by the script.
- [ ] `SPEC.md` uses `narra` for Narra code; the guide suite still passes.
- [ ] A README in `editors/textmate/` says how to load the grammar in VS Code and in Shiki.

## Sizing

M: a grammar of a few hundred lines and a test script.
