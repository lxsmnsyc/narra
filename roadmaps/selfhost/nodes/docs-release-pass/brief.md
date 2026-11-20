---
id: docs-release-pass
title: "Docs fixes before the first release"
status: done
priority: could
type: docs
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
verify: pnpm --dir website build
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/docs-release-pass/**
  - docs/guide/**
  - docs/install.md
  - README.md
  - LICENSE
  - seed/README.md
  - skills/narra/evals/**
  - website/src/**
shared:
  - SPEC.md
reserves: []
provides: []
depends_on: []
---
# Docs fixes before the first release

## Goal

Changes 1 to 7 of the approved proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`, maintainer approval 2026-11-20), plus an MIT `LICENSE`:

1. Chapter 1 (`docs/guide/01-getting-started.md`) installs with the one-line command and links to `/install` for other systems and source builds, then checks with `narrac --version` (added in this batch by narrac-version). The guide index line for chapter 1 says "install the compiler".
2. Drop "Workspaces of several packages" from the guide's "Planned" list.
3. Rewrite the three comparisons with TypeScript and JavaScript outside chapter 13 (02:66, 11:360, 11:432 on the proposal's commit) without the comparison.
4. Bootstrap stages, exit codes and `linux-check.sh` move from `docs/install.md` to `seed/README.md`; `docs/install.md` keeps the commands.
5. `SPEC.md`: the command is `narrac`, not `narra` (25 lines); section 15 drops "(sketch)".
6. `README.md`: replace "draft design" and "Status: specification draft" with the release status (a first prerelease is coming); install command and site first; `bootstrap.sh` and `test.sh` move to a "Working on the compiler" section; a License line.
7. Home page: link the agent claim (19 of 20 against 15 of 20) to a committed result under `skills/narra/evals/`. If the result is not committed, commit the numbers from the narra-skills outcome as a small results file there and link it.
8. `LICENSE`: MIT, `Copyright (c) 2026 Alexis H. Munsayac` (maintainer decision, 2026-11-20).

The guide stays standalone: no SPEC or decision-record citations, no assumed TypeScript outside chapter 13. Prose follows agent-skills:plain-writing.

## Acceptance

- [ ] Each change above made, or a note saying why not.
- [ ] `pnpm --dir website build` passes; `sh scripts/test.sh --suite sources` (or the docs suites that check guide samples) passes.
