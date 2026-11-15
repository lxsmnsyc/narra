---
id: docs-highlighting
title: "Consistent highlighting in the docs"
status: active
priority: could
type: fix
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
verify: pnpm --dir website build
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/docs-highlighting/**
  - website/src/highlight/**
  - website/plugins/markdown.ts
  - website/scripts/check-grammar.ts
  - website/package.json
  - website/pnpm-lock.yaml
shared:
  - docs/guide/*.md
  - docs/install.md
  - SPEC.md
reserves: []
provides: []
depends_on:

---
# Consistent highlighting in the docs

## Goal

The maintainer reports that some parts of the docs are highlighted inconsistently. Find every case first and list it in log.md with the page and the fence: start the site (`pnpm --dir website build && pnpm --dir website preview`) and compare its HTML with the source.

Known leads, not proof: fences use `narra`, `narra error`, `narra file`, `narra fragment`, `text`, `sh`, `toml`, `ts`, `json`, `lua` and a few bare fences. `toml` and `lua` have no twinkleplop language and render plain. The Narra language in `website/src/highlight/narra.ts` may miss tokens that the TextMate grammar (`editors/textmate/narra.tmLanguage.json`) has; compare class by class. Inline code is not highlighted.

Fix at the cause: a missing language gets one, a grammar gap gets a rule, a fence with the wrong tag gets the right tag. Extend `scripts/check-grammar.ts` so the build fails on a fence with no language, an unknown language, or a Narra token the grammar does not classify.

## Acceptance

- [ ] log.md lists every inconsistency found, each with its fix.
- [ ] The build fails on an untagged or unknown fence; it passes after the fixes.
- [ ] The narra tokens match the TextMate grammar's classes on every guide and spec sample.
