---
id: narra-skills
title: Agent skills for Narra
status: todo
priority: should
type: docs
outline: true
milestone: m9-adoption
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
owns: []
shared: []
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
---

# Agent skills for Narra

## Goal

Claude Code skills that let an agent write, check and debug Narra code: the language rules agents get wrong (no shadowing, no global functions except print, assert and panic, effects), the compile and test loop, reading JSON diagnostics, and using `narra query` and `narra fmt`. Includes evals that show an agent with the skill writes code that compiles more often than one without.

## Notes

Outline. Decide when refining where the skills live (in this repo, or the agent-skills marketplace).
