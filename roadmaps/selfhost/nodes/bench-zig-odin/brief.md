---
id: bench-zig-odin
title: "Zig and Odin in the benchmarks"
status: todo
priority: could
type: feature
story: batch-22
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
verify: sh bench/run.sh -n 1
exclusive: [bench]
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/bench-zig-odin/**
  - bench/**
shared:
  - website/src/**
  - README.md
reserves: []
provides: []
depends_on: []
---
# Zig and Odin in the benchmarks

## Goal

Maintainer request, 2026-11-20: add Zig and Odin to `bench/`. Each program (`fib`, `sieve`, `matmul`, `strings`, `interp`, `particles`, `dot`) gets a `.zig` and an `.odin` version that mirrors the Narra algorithm the way the Go and Rust versions do: plain idiomatic code, the same answer, release builds (`zig build-exe -O ReleaseFast`, `odin build -o:speed`; say which flags and why, matching how Go and Rust are built). `run.sh` adds two columns and skips a language whose toolchain is missing (shown as `-`), as it does for Go and Rust. `bench/README.md` lists the five languages.

Zig is installed on this machine (`/opt/homebrew/bin/zig`). Odin is not: write the Odin files, check them with `odin check` if a toolchain can be found without installing anything, and otherwise leave them unmeasured and say so in the outcome. Do not download or install toolchains; the maintainer installs Odin.

Measure only on an idle machine (`bench/CLAUDE.md`: 1-minute load average under 2). If the machine is busy, land the programs and the `run.sh` columns, and record the numbers in `BASELINE.md` later. Wherever the site or README shows benchmark numbers, keep them sourced from `BASELINE.md`.

## Acceptance

- [ ] Seven programs in Zig (each prints the same answer as the Narra version) and seven in Odin; `run.sh` has the two columns and skips a missing toolchain.
- [ ] A `BASELINE.md` section with Zig numbers measured on an idle machine (load written next to them), or a note that the run is pending.
