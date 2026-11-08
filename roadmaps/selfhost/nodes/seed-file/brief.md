---
id: seed-file
title: Checked-in C seed and cc-only bootstrap
status: todo
priority:
type: feature
story: c-seed
early: false
outline: false
milestone:
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh selfhost/bootstrap.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-08
spent:
owns:
  - seed/**
  - selfhost/bootstrap.sh
  - scripts/seed-update.sh
  - .gitattributes
  - roadmaps/selfhost/nodes/seed-file/**
shared:
  - compiler/README.md
  - README.md
reserves: []
provides: []
depends_on:
  - on: selfhost-driver
    kind: hard
    why: the seed should contain the driver commands, so a compiler built from it is usable on its own
---

# Checked-in C seed and cc-only bootstrap

## Goal

Check in `seed/narrac.c`, the C the self-hosted compiler emits for itself, and make it the starting point of every build. `selfhost/bootstrap.sh` builds the compiler from the seed with `cc` alone and checks the fixed point. `scripts/seed-update.sh` regenerates the seed from the current source.

## Steps

- [ ] `seed/narrac.c`: generated, never edited. A header comment says how it was made and from which commit.
- [ ] `selfhost/bootstrap.sh`: stage 0 is `cc seed/narrac.c`; stage 1 builds `selfhost/` with it; stage 2 builds `selfhost/` with stage 1; stage 1 and stage 2 C must be identical. No `go`.
- [ ] `scripts/seed-update.sh`: runs the bootstrap from the current seed, then replaces `seed/narrac.c` with the stage 2 output, and fails when the fixed point does not hold.
- [ ] `.gitattributes`: mark `seed/narrac.c` as generated (`linguist-generated`, `-diff`), so diffs and GitHub stats skip it.
- [ ] `seed/README.md`: the seed rule in plain words (the compiler source may only use features the seed compiles; land a feature, regenerate the seed, then use it), and when to regenerate (when the compiler wants a new feature, and at each release, not on every commit).

## Acceptance

- [ ] `sh selfhost/bootstrap.sh` passes on macOS and in the Linux container (`scripts/linux-check.sh`), with no `go` on the path.
- [ ] `sh scripts/seed-update.sh` on an unchanged tree leaves `seed/narrac.c` byte-identical.

## Sizing

M: two scripts and a generated file; the fixed point already holds today.

## Notes for the worker

- The emitted C must not depend on the host platform beyond the existing `#ifdef __APPLE__` in the runtime. Check that the seed built on macOS also builds and passes the fixed point on Linux.
