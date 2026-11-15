---
id: install-guide
title: Installation guide
status: todo
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
size: S
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns:
  - roadmaps/selfhost/nodes/install-guide/**
  - docs/install.md
shared:
  - README.md
  - docs/guide/01-getting-started.md
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: the install steps change once the self-hosted compiler replaces the Go build
---
# Installation guide

## Goal

`docs/install.md` tells someone who wants to use Narra how to get `narrac` on their machine and check that it works. The README links to it in one line, and chapter 1 of the guide links to it instead of repeating the details.

## Content

- Prerequisites: a C11 compiler (`cc`, `clang` or `gcc`) and `sh`. No Go, no other tools. The std sources are built into `narrac` (`selfhost/src/stdtext.nra`), so the binary is the whole install.
- Build: `sh selfhost/bootstrap.sh out`, what the fixed point check means, `CC=...` to pick a compiler, and the exit codes.
- Install: copy `out/narrac` to a folder on `PATH` (for example `~/.local/bin`), and how to update (pull, rebuild) and uninstall (delete the file).
- First check: `narrac run` on a hello program, then link to chapter 1.
- Platforms: macOS and Linux on arm64 and x64 are tested (`scripts/linux-check.sh` covers Linux). Name Windows and others as not supported yet.
- Prebuilt binaries: write the section header and one line saying they are not published yet. Node release-binaries (group p5-distribution) fills this section with the `curl -fsSL <raw url>/install.sh | sh` installer when its workflow ships; it owns that edit.

## Acceptance

- [ ] Every command in the guide runs as written on this machine, and on Linux through `sh scripts/linux-check.sh` or the container it uses.
- [ ] README and chapter 1 link to it; chapter 1 keeps only the three-line build.
- [ ] No SPEC or decision-number citations.

## Sizing

S: one page, checked by hand.
