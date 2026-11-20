---
id: first-release
title: "First release tag"
status: todo
priority: could
type: chore
story:
early: false
outline: false
milestone: p6-docs
target:
ship: pr
land: auto
from:
size: S
spike: false
hypothesis:
verify: 
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/first-release/**
shared:

reserves: []
provides: []
depends_on:
  - on: curl-installer
    kind: hard
    why: the release must carry the installers
  - on: batch-20
    kind: hard
    why: pre-release fixes and the docs review land before the tag (maintainer, 2026-11-19)
---
# First release tag

## Goal

Push the first `v*` tag (proposed `v0.1.0-alpha.1`, a prerelease) so release.yml publishes binaries and the one-liner works. Publishing is visible to everyone: the maintainer decides and pushes, or approves the push.

Note: `releases/latest` skips prereleases, so with only a prerelease the one-liner finds nothing. Either the first tag is a full release, or install.sh falls back to the newest prerelease. curl-installer decides and records it.

## Acceptance

- [ ] The maintainer's decision recorded here.
- [ ] After the tag: the one-liner installs narrac on macOS and Linux.
