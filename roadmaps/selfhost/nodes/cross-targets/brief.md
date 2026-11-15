---
id: cross-targets
title: "Build and run every target from one machine"
status: todo
priority: could
type: feature
story: batch-16
early: false
outline: false
milestone: p5-distribution
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/cross-check.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-06
spent:
owns:
  - .github/workflows/windows.yml
  - roadmaps/selfhost/nodes/cross-targets/**
  - scripts/cross-check.sh
shared:
  - scripts/test.sh
  - compiler/README.md
  - README.md
reserves: []
provides: []
depends_on:
  - on: target-table
    kind: hard
    why: the script drives `narrac build --target` for every row
  - on: wasi-runtime
    kind: behavior
    via: targets
  - on: windows-runtime
    kind: behavior
    via: targets
---
# Build and run every target from one machine

## Goal

`sh scripts/cross-check.sh` builds the sample programs for every row of `narrac targets` from one host and runs each where it can, comparing output with the `.out` goldens. This is the check the portability nodes use, and the source of release binaries and of any per-platform npm builds js-interop-design picks.

## Context

- Source: the 2026-11-06 discussion on shipping Narra tools to npm as per-platform packages, the way esbuild, swc and oxc do.
- C compiler: `zig cc` (in `CC` through a one-line wrapper, since `$CC` is one program name, `driver.nra`). Other platforms run on: Linux through `docker run --platform linux/amd64` and `linux/arm64` (`alpine` for musl, `debian` for gnu), wasm through `wasmtime -W exceptions=y`, Windows through wine in a `linux/amd64` docker container locally (and natively in the CI job below), macOS x86_64 through Rosetta on an arm64 Mac.
- CI: `.github/workflows/windows.yml` runs the Windows rows of the script natively on `windows-latest`, on pull requests to `main` only (the repository is public, so minutes are free). It is the proof for the Windows nodes; wine is only the fast local loop. Check the YAML with `actionlint` when installed. No agent pushes or triggers it; the batch pull request runs it.
- A runner that is missing is a skip with a message, not a failure, so the script works on a plain Linux box with only docker absent.

## Acceptance

- [ ] On this machine (macOS arm64, zig, docker, wasmtime) the script builds every row and runs every row that has a runner; the samples a target leaves out (wasi-runtime's list) are skipped by name.
- [ ] `seed/narrac.c` is built for every row and, where it runs, compiles `testdata/c_interop.nra` to the same C as the host.
- [ ] Exit status 0 only when every run matches its golden. Not part of `scripts/test.sh` (needs zig and docker); `compiler/README.md` says when to run it.

## Sizing

M: one script, a sample list, a runner table.
