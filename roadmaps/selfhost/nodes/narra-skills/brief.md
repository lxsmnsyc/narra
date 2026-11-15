---
id: narra-skills
title: Agent skills for Narra
status: done
priority: could
type: docs
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
gate: true
added: 2026-11-01
spent:
owns:
  - roadmaps/selfhost/nodes/narra-skills/**
  - skills/**
  - .claude-plugin/**
shared:
  - README.md
reserves: []
provides: []
depends_on:
  - on: language-guide
    kind: hard
    why: the skills point agents at the guide chapters instead of restating them
  - on: query-cli
    kind: hard
    why: the skills teach agents to use narra query, so its commands must exist
  - on: fmt
    kind: hard
    why: the skills tell agents to run narra fmt
  - on: guide-tools
    kind: hard
    why: the tool loop the skill teaches is documented in that chapter, and the skill links to it
  - on: dead-code-check
    kind: hard
    why: the skill lists the checker rules agents break, and this node adds one
  - on: let-never-reassigned
    kind: hard
    why: same
---
# Agent skills for Narra

## Goal

Claude Code skills that let an agent write, check and debug Narra code, and an eval that shows they help.

## Where they live (maintainer, 2026-11-14)

In this repository, under `skills/narra/` with a `.claude-plugin/marketplace.json` at the root, so the skill is versioned with the language it describes and a change to the language can update it in the same commit. Users add it with `/plugin marketplace add lxsmnsyc/narra`. The alternative, the agent-skills marketplace, would let the skill drift from the compiler.

## Content

- `SKILL.md`: the rules agents get wrong, each with a short wrong and right pair (no shadowing, no nested function declarations, unused names and dead code are errors, `let` only when reassigned, effects in `uses`, `mutable` parameters, `Result` instead of exceptions, only `print`, `assert` and `panic` as global functions). Mine the list from the compiler's own diagnostics table and from the inbox, not from memory.
- The loop: `narrac check --json`, read the code and range, `narrac explain <code>`, `narrac apply-fix`, `narrac fmt`, `narrac test`.
- `narrac query` for finding definitions and callers before editing.
- Links to guide chapters for everything else. The skill stays under about 2,000 tokens.

## Eval

Size decided on 2026-11-14: the smallest eval that still tells whether the skill helps, because each run holds the 8 GB machine for minutes and blocks batch workers.

- 10 fixed tasks in `skills/narra/evals/`: 4 write-a-function tasks with a `narrac test` file, 3 fix-the-compile-error tasks, 3 change-an-existing-program tasks. The tests are written before any run.
- Run each task with and without the skill through `claude plugin eval`, 2 runs each, same model: 40 runs, at most 2 at a time.
- Run it alone, at the end of the batch, after every code child has landed, with no other worker building.
- Measure: share of runs whose final program passes `narrac check` and the task test, and the number of `narrac check` calls before the first pass.
- Success threshold, fixed now: the skill raises the pass share by at least 15 points, or cuts the median check calls by a third with an equal pass share. With 20 runs a side, report the counts, not only the shares. Report the result in `outcome.md` also when it misses.
- Cost cap: estimate from 2 trial runs first; stop if the eval would pass 1.5M tokens or 2 hours.

## Acceptance

- [ ] The skill and marketplace file load in Claude Code (`/plugin marketplace add` from a local path).
- [ ] The eval ran, with the numbers and the threshold verdict in `outcome.md`.
- [ ] README links to the skill.

## Sizing

M: one skill file, 20 small tasks, one eval run.
