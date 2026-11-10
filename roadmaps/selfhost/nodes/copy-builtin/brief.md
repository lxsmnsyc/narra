---
id: copy-builtin
title: "copy(x) and the debug copy check"
status: done
priority: should
type: feature
story: batch-5
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent: 538127
owns:
  - testdata/copy_*
  - testdata/errors/copy_*
  - testdata/fmt/*copy*
  - roadmaps/selfhost/nodes/copy-builtin/**
shared:
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - docs/guide/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E0975]
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the batch-2 runtime and Cell
---

# copy(x) and the debug copy check

## Goal

Build `copy(x)` and the debug copy check from D94 (debate `roadmaps/selfhost/debate/value-semantics/debate.md`, settled structure items 2 and 4, claims P11, P12, P18, O18).

- `copy(x)` duplicates a value at any depth. At the call it is one retain. The checker treats the result as unshared, and copy-on-write later copies only the nodes on each written path, once each. It works on every tracked type: structs, arrays, maps, `FixedArray`, `Box`, `StringBuilder`, generators and coroutines.
- In debug builds, `copy` grants a credit, and a copy-on-write uses one. Debug-only hooks at every reference-count decrement drop the credit once the sharing it created ends. At exit, like the leak check (D80), the program reports every copy-on-write that no credit explains. Release builds emit none of this.
- Rewrite the error hints in `selfhost/src/diagnostics.nra` that suggest `const copy = %s` (ten of them, around lines 223 to 737) to suggest `copy(%s)` instead, so no hint names a local `copy` next to the builtin. Decide whether a local may take the name `copy`: follow what the checker already does for prelude names such as `panic`, and record the answer in SPEC.md next to `copy`.
- Remove `@unique` from SPEC.md (lines near 65 and 895) and its skipped sample. Remove the unbuilt `--report-copies` and `NARRA_TRACE_COPIES` text wherever SPEC or the guide mention them. No deprecation path.

The shared-write rule itself is not part of this node (shared-write-check). Until it lands, the debug report lists every hidden copy in existing programs, so it stays off by default in the test suites until shared-write-error. Turn it on in the new samples only.

## Acceptance

- [x] `copy(x)` parses, checks, formats, dumps and emits; E0975 reports a misuse (wrong argument count, or an untracked type where a copy means nothing, with a fix that drops the call).
- [x] Samples: a nested write after `copy` changes only the copy; a debug run reports a copy-on-write without `copy`, and one with `copy` reports nothing.
- [x] A release build of a `copy` sample emits no credit or hook code.
- [x] `@unique` is gone from SPEC.md and testdata.
- [x] No diagnostic hint suggests `const copy = ...`; the error samples that print those hints are updated.
- [x] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.
