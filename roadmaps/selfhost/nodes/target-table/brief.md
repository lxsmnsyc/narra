---
id: target-table
title: "Target table and narrac build --target"
status: done
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
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent: 235171
owns:
  - roadmaps/selfhost/nodes/target-table/**
  - selfhost/src/buildtarget.nra
  - testdata/build-targets/cross_*
shared:
  - selfhost/src/driver.nra
  - selfhost/src/buildconditions.nra
  - selfhost/src/diagnostics.nra
  - std/src/build.nra
  - selfhost/src/stdtext.nra
  - selfhost/tests/src/*.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1600-E1604]
provides: [targets]
depends_on: []
---
# Target table and narrac build --target

## Goal

Freeze contract `targets` and let one host build for another: `narrac build --target <name>`, `narrac check --target <name>` and `narrac targets`. Today `TARGETS` (`selfhost/src/buildtarget.nra`) has four rows chosen by the host, and `cCompilerArguments` (`driver.nra`) calls `$CC` with fixed flags.

## Context

- Probe 2026-11-14 on macOS arm64 with zig 0.17.1: the emitted C of a small program builds with `zig cc -target` for `x86_64-linux-musl`, `aarch64-linux-gnu` and `x86_64-windows-gnu` with no runtime change. `seed/narrac.c` builds for `x86_64-linux-musl`. Windows and wasm need runtime work (nodes windows-runtime, wasi-runtime).
- `--all-targets` (node build-targets) already checks every row; it gains the new rows, so `@target` and `Build.os` branches for Windows and Wasi are checked on every build.

## Work

1. Rows, columns and enum variants as the contract says. `Os.Wasi` and `Arch.Wasm32` go into `std/src/build.nra` and `stdtext.nra` through the generated suite.
2. `--target <name>`: selects the row for build conditions and passes `-target <triple>` to `$CC` before the other flags. An unknown name is a new code that lists the names. `-sanitize` with a target other than the host is a new code (sanitizers do not cross-link).
3. `narrac targets` prints one line per row: name, triple.
4. `targetCompilerArguments(row)` holds the per-target flags, empty for now.
5. SPEC section 12 (build conditions) and section 15 (commands), compiler/README.md.

## Acceptance

- [ ] `narrac build --target Linux-X86_64-musl` with `CC=zig-cc-wrapper` builds a sample that runs in `docker run --platform linux/amd64 alpine`. The test harness covers only what needs no zig: the flag reaches `$CC` (a logging `CC` script, as c-interop did for `[link] search`), names and errors.
- [ ] `check --all-targets` covers the eight rows; goldens updated.
- [ ] Contract `targets` set to frozen, version 1.
- [ ] `sh scripts/test.sh` passes.

## Sizing

M: one table, one flag, one command, golden updates.
