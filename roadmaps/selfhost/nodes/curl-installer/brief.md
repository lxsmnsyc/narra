---
id: curl-installer
title: "curl | sh installer from GitHub releases"
status: active
priority: could
type: feature
story: batch-17
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
verify: sh scripts/install-check.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/curl-installer/**
  - install.sh
  - install.ps1
  - scripts/install-check.sh
  - scripts/release-build.sh
  - .github/workflows/release.yml
  - docs/install.md
  - roadmaps/selfhost/contracts/install-command.md
shared:
  - README.md
reserves: []
provides: [install-command]
depends_on:

---
# curl | sh installer

## Goal

Install is one line, as pnpm.io and gojo.egoist.dev do it:

    curl -fsSL https://github.com/lxsmnsyc/narra/releases/latest/download/install.sh | sh

and on Windows PowerShell:

    irm https://github.com/lxsmnsyc/narra/releases/latest/download/install.ps1 | iex

The commands are frozen in contract install-command; the home page shows them.

Today `install.sh` exists at the repo root but only for macOS arm64 and Linux x86_64, and release.yml does not attach it. Make release.yml build and attach archives for every row the target table can build (macOS arm64 and x86_64, Linux x86_64 and arm64 glibc and musl, Windows x64), plus `install.sh`, `install.ps1` and checksums. install.sh picks musl on musl systems. install.ps1 installs to `%LOCALAPPDATA%\narra\bin` and adds it to the user PATH. Both print the next step (open a new terminal, `narrac --version`).

`docs/install.md` leads with the one-liner per OS; building from source moves to its own section below.

`scripts/install-check.sh` runs install.sh against archives built locally (`NARRA_BASE_URL`), and install.ps1 under pwsh when present.

No tag is pushed: that is node first-release.

## Acceptance

- [ ] release.yml attaches every archive, install.sh, install.ps1 and checksums (checked with `act` or by reading; no workflow run).
- [ ] install-check passes on macOS here and on Linux under docker.
- [ ] docs/install.md starts with the one-liner.
