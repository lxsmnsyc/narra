---
id: release-binaries
title: Release workflow with macOS and Linux binaries
status: todo
priority: could
type: chore
story:
early: false
outline: true
milestone: m9-adoption
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-08
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: const-lazy
    kind: hard
    why: user decision 2026-11-09, binaries ship after the remaining language features
  - on: type-aliases
    kind: hard
    why: same
  - on: typed-calls-and-effects
    kind: hard
    why: same
  - on: c-seed
    kind: hard
    why: releases build from the seed, which the story adds
---

# Release workflow with macOS and Linux binaries

## Goal

A GitHub Actions workflow that, on a version tag, builds `narrac` from `seed/narrac.c` on macOS arm64 and Linux x86_64, runs the bootstrap fixed point, and uploads the binaries (with the std sources) to a GitHub Release. How users install it (a shell installer, Homebrew, npm) is decided with install-guide, which can then point at these releases. The repository is public, so the workflow's minutes are free.
