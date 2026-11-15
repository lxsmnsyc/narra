---
id: release-binaries
title: "Release workflow with macOS and Linux binaries"
status: active
priority: could
type: chore
story: batch-16
early: false
outline: false
milestone: p5-distribution
target:
ship: pr
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/release-build.sh --dry-run
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-08
spent:
owns:
  - roadmaps/selfhost/nodes/release-binaries/**
  - .github/workflows/release.yml
  - scripts/release-build.sh
shared:
  - compiler/README.md
  - README.md
reserves: []
provides: []
depends_on: []
---
# Release workflow with macOS and Linux binaries

## Goal

On a pushed version tag (`v*`), a GitHub Actions workflow builds `narrac` from `seed/narrac.c` on macOS arm64 and Linux x86_64, runs the bootstrap fixed point, and uploads one archive per platform (binary plus `std/` sources and a checksum) to a GitHub Release. The repository is public (checked 2026-11-14), so the minutes are free. There is no `.github/` directory yet.

## Work

1. `scripts/release-build.sh [--dry-run] <version>`: builds from the seed with `cc`, runs `sh selfhost/bootstrap.sh`, `sh scripts/test.sh`, and writes `narra-<version>-<os>-<arch>.tar.gz` and `.sha256` into `dist/`. `--dry-run` skips nothing but the upload step, which the script does not do anyway.
2. `.github/workflows/release.yml`: on `push: tags: ['v*']`, a matrix of `macos-14` and `ubuntu-24.04` runs the script, then one job runs `gh release create` with the archives. No other triggers, so no workflow runs before a tag.
3. Users install with `curl -fsSL <raw url>/install.sh | sh` (maintainer, 2026-11-14). This node writes `install.sh`: it detects the OS and CPU, downloads the matching archive from the latest release (or `NARRA_VERSION`), checks its SHA-256 against the release checksum file, and puts `narrac` in `~/.local/bin` (or `NARRA_INSTALL`). The workflow uploads the checksum file. Test the script against archives served from a local folder.

## Testing without a tag

The workflow cannot be proven without a tag push, and no tag is pushed by this node. The script is the whole build: run it on this Mac and in `docker run --platform linux/amd64 ubuntu:24.04`, which is what the two runners do. `act` is not installed; installing it is optional and only checks the YAML. The YAML is checked with `actionlint` if available. The first real run is the maintainer's first tag.

## Acceptance

- [ ] `sh scripts/release-build.sh --dry-run 0.0.0-test` produces both archives, one on the host and one in docker, and each unpacked binary compiles a sample.
- [ ] `release.yml` passes `actionlint` (or the outcome says it was not available).
- [ ] The maintainer reviews the workflow before merge (gate); no tag is pushed.

## Sizing

M: one script, one workflow.
