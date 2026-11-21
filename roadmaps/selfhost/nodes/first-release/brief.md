---
id: first-release
title: "First release tag"
status: done
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
fulfilled_by: "v0.1.0"
gate: true
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/first-release/**
shared:

reserves: []
provides: []
depends_on:
  - on: batch-23
    kind: hard
    why: the compiler uses the new std and std headers are clean before the tag (maintainer, 2026-11-20)
  - on: curl-installer
    kind: hard
    why: the release must carry the installers
  - on: batch-22
    kind: hard
    why: one form per std API, clean samples and std gaps before the tag (maintainer, 2026-11-20)
  - on: batch-21
    kind: hard
    why: the docs must teach the crucial concepts before the release (maintainer, 2026-11-20)
  - on: batch-20
    kind: hard
    why: pre-release fixes and the docs review land before the tag (maintainer, 2026-11-19)
---
# First release tag

## Goal

Push the first `v*` tag (proposed `v0.1.0-alpha.1`, a prerelease) so release.yml publishes binaries and the one-liner works. Publishing is visible to everyone: the maintainer decides and pushes, or approves the push.

Note: `releases/latest` skips prereleases, so with only a prerelease the one-liner finds nothing. Either the first tag is a full release, or install.sh falls back to the newest prerelease. curl-installer decides and records it.

## Acceptance

- [x] The maintainer's decision recorded here: the first tag is `v0.1.0`, a full release, so `releases/latest` finds it. Not before the docs review (batch-21) lands (maintainer, 2026-11-20).
- [ ] After the tag: the one-liner installs narrac on macOS and Linux.
