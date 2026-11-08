---
id: test-runner-core
title: Test runner in Narra for programs, panics and errors
status: todo
priority:
type: feature
story: c-seed
early: false
outline: true
milestone:
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-08
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-driver
    kind: hard
    why: the runner builds samples through narrac build and run
  - on: seed-parity-spike
    kind: hard
    why: the spike's test-file table decides which suites the runner ports
---

# Test runner in Narra for programs, panics and errors

## Goal

A test runner written in Narra (for example `selfhost/tests/`) that replaces the Go harness for the language suites: run each program sample under the sanitizer and leak check and compare stdout with `.out` (TestPrograms, TestStage2Programs), check panic samples, and compare compile errors with their expected output (TestCompileErrors). One command runs it, and it runs in parallel. Outline until the spike reports.
