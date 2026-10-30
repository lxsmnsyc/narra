---
id: test-runner
title: Golden files and narra test runner features
status: todo
priority: should
type: feature
outline: true
milestone: m7-tooling
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: std-test
    kind: hard
    why: golden comparison reports failures like Test.eq, and the runner counts std/test failures
---

# Golden files and narra test runner features

## Goal

`Test.golden(name, actual)` compares a string with a file under the package's `testdata/` and rewrites it when `NARRA_UPDATE_GOLDEN=1` is set. `narra test` gains `--filter <text>`, per-test timing and a summary of passed, failed and total tests.

## Notes

- Requested by the user on 2026-10-30, with std-test. Golden files would let the self-hosted compiler test its own dumps in Narra.
- Not in `SPEC.md`. Propose the spec text in the pull request and get the user's approval.
- Golden files need `uses io` in tests; decide whether test blocks may use `io` freely.
- Outline. Break into a full brief when `plan.py refine` lists it.
