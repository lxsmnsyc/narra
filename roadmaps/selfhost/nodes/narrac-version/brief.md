---
id: narrac-version
title: "narrac --version"
status: active
priority: could
type: feature
story: batch-20
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
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-19
spent:
owns:
  - roadmaps/selfhost/nodes/narrac-version/**
  - install.sh
  - install.ps1
  - docs/install.md
  - scripts/install-check.sh
  - scripts/release-build.sh
  - .github/workflows/release.yml
shared:
  - selfhost/src/main.nra
  - selfhost/tests/src/**
  - compiler/README.md
  - seed/README.md
reserves: []
provides: []
depends_on: []
---
# `narrac --version`

## Goal

`narrac --version` prints the version: the release tag when the binary was built by the release (`scripts/release-build.sh`, `.github/workflows/release.yml`), `dev` otherwise. Builds from `seed/narrac.c` and `selfhost/bootstrap.sh` print `dev`. Pick the mechanism (for example a C define the runtime reads, or a value the build passes in); `selfhost/src` may use only what the committed seed compiles, and the bootstrap fixed point must still hold.

Today `narrac --version` fails with `--version: No such file or directory` (main, 2026-11-19). Also make `narrac --help` and `-h` print the usage instead of a file error, if the usage text already exists.

`install.sh`, `install.ps1` and `docs/install.md` print `narrac targets` as the check after installing; point them at `narrac --version`. `scripts/install-check.sh` checks the output.

## Acceptance

- [ ] `narrac --version` prints `dev` for a local build; a release build prints its tag (show how you checked it).
- [ ] Installers and `docs/install.md` use `narrac --version`.
- [ ] A test case covers `--version`; `sh scripts/test.sh` passes.
