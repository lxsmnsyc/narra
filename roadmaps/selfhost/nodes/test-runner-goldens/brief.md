---
id: test-runner-goldens
title: Golden-file suites in the Narra test runner
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
  - on: test-runner-core
    kind: hard
    why: adds suites to the runner core builds
---

# Golden-file suites in the Narra test runner

## Goal

Port the remaining language-level Go tests to the Narra runner: dumps (astdump, declsdump, typesdump, pkgdump), fmt and fmt-doc goldens, query, diagnostics JSON, packages, runner, the guide samples (TestGuide), the std tests and `TestSourcesFormatted`. Tests that only compare Go with the self-hosted compiler are not ported. Outline until the spike reports.
