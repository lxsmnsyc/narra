---
id: website
title: Narra website
status: todo
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
added: 2026-11-05
spent:
owns:
  - roadmaps/selfhost/nodes/website/**
  - website/**
shared:
  - README.md
  - .gitignore
reserves: []
provides: []
depends_on:
  - on: language-guide
    kind: hard
    why: the site publishes the guide chapters and their samples
  - on: textmate-grammar
    kind: hard
    why: the site highlights narra fences with that grammar
  - on: install-guide
    kind: hard
    why: the Install page renders docs/install.md
---
# Narra website

## Goal

A static website that publishes the language guide, the installation guide and `SPEC.md`, with `narra` fences highlighted. Requested by the maintainer on 2026-11-05. Pages are read from `docs/` and `SPEC.md` at build time, so the docs stay the source and the site holds no copies.

## Choices (maintainer, 2026-11-14)

- Generator: brumate, in `website/`.
- Hosting: Vercel, with `website` as the root directory and only `main` deployed (`website/vercel.json`, from the repo-setup template). The maintainer imports the repo on Vercel; agents do not.
- Highlighter: twinkleplop at build time with `editors/textmate/narra.tmLanguage.json`, so pages ship no highlighting script. If twinkleplop cannot read a TextMate grammar, say so in the log and stop before writing a second grammar.

## Acceptance

- [ ] `pnpm --dir website build` builds every guide chapter, `docs/install.md` and `SPEC.md` with working links between them and highlighted `narra` fences.
- [ ] The grammar check from textmate-grammar runs in the website build.
- [ ] Keyboard navigation and a readable layout on a phone; no layout shift when highlighting loads.
- [ ] No deploy from an agent. The pull request says what the maintainer must do on Vercel.

## Sizing

M: a small static site over existing markdown.
