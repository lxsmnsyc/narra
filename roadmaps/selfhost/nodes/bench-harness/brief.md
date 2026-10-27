---
id: bench-harness
title: Benchmark harness against Go and Rust
status: todo
priority: should
type: feature
outline: false
milestone: m6-performance
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns:
  - bench/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Benchmark harness against Go and Rust

## Goal

Build `bench/` with the same small programs written in Narra, Go and Rust, a script that builds and times them, and a recorded baseline. Later performance nodes report their before and after numbers with it.

## Context

Use only features the bootstrap compiler supports today (see `compiler/README.md`). Good programs: recursive fibonacci, an array-heavy loop (sieve or matrix multiply), string building, and a struct and enum workload (a small interpreter over an enum AST without recursion, or a token classifier).

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Acceptance

- [ ] `bench/run.sh` builds every program in release mode, runs each several times, and prints a table of median times per language.
- [ ] `bench/BASELINE.md` records the first run with the machine description.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: new directory, no compiler changes, but three languages and a script.

## Notes for the worker

- Skip Rust or Go programs cleanly when the toolchain is missing, and say so in the output.
- Benchmarks need a quiet machine; do not run them in parallel with other work.
