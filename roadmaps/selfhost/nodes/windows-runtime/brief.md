---
id: windows-runtime
title: "Runtime and file std on Windows x64"
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
spent:
owns:
  - roadmaps/selfhost/nodes/windows-runtime/**
  - testdata/windows_*
shared:
  - selfhost/runtime/*
  - selfhost/src/runtime*.nra
  - std/src/*.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/driver.nra
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: target-table
    kind: behavior
    via: targets
---
# Runtime and file std on Windows x64

## Goal

Programs that use the core runtime, `std/io`, `std/fs`, `std/path`, `std/env`, `std/clock`, `std/entropy` and `std/thread` build with `--target Windows-X86_64` (`zig cc -target x86_64-windows-gnu`, MinGW) and give the same output on Windows. Processes and async I/O are node windows-process-async.

## Context

Probe 2026-11-14: a small program builds for `x86_64-windows-gnu` unchanged. `seed/narrac.c` fails on `mkdir` with two arguments (MinGW takes one) and `poll.h`. `std_env.c` includes `crt_externs.h` only on macOS; check the Linux `environ` path on MinGW. Paths: `std/path` assumes `/`; Windows needs `\` and drive letters accepted on input, and SPEC must say what `Path.join` writes. Text files: open in binary mode so `\n` is not rewritten. Threads: MinGW ships winpthreads, so `-pthread` may work as is.

How the output is checked (maintainer, 2026-11-14): wine locally for the working loop. The Homebrew wine cask is disabled and this Mac has no Rosetta, so run wine inside `docker run --platform linux/amd64` (a Debian image with `wine64`); if that is too slow or fails, skip the local run and say so in the log, and real Windows in CI through the `windows-latest` job that cross-targets adds. Wine passing is not proof; the CI job is.

## Acceptance

- [ ] The samples that use only the modules above give the same output on Windows, run by node cross-targets' script.
- [ ] `std/path` behavior on Windows written in SPEC section 3, with samples.
- [ ] `sh scripts/test.sh` passes on the host, native output unchanged.

## Sizing

M: a handful of `#ifdef _WIN32` branches in the runtime C and a path rule.
