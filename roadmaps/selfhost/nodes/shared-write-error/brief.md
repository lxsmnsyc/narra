---
id: shared-write-error
title: "Shared writes become errors"
status: done
priority: should
type: feature
story: batch-6
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent: 447610
owns:
  - roadmaps/selfhost/nodes/shared-write-error/**
shared:
  - selfhost/src/**
  - selfhost/tests/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: shared-write-sweep-rest
    kind: hard
    why: the compiler must be clean first
---

# Shared writes become errors

## Goal

D94 step 5 (debate settled structure items 1 and 4, P18, O18):

- The shared-write warning (E0978) becomes an error. Keep the fix.
- Turn the debug copy check on for every program sample: `isCopySample` true for every program, or `COPY_CHECK_ENVIRONMENT` in `Shell.leakCheckEnvironment()` (see `nodes/copy-builtin/outcome.md`). Any report that remains is a hidden copy to fix or explain in the outcome.
- Run the analysis in every command that checks: `build`, `run`, `test`, `emit`, `check`, `apply-fix` and the language server (`lspcache.nra`). Today it runs only in `check` and `apply-fix`, after checking, and returns early when there are other errors (`sharedwrites.nra`, `driver.nra`; see `nodes/shared-write-check/outcome.md`). Drop that early return so E0978 is reported next to other errors. Accept the cost: about +0.33G instructions per selfhost compile, measured by shared-write-check. Record the before and after cost in the outcome.
- Fix a gap the shared-write-sweep review found in the debug copy check (`selfhost/src/emitdecls.nra`, `creditFunction`, called from `stdcopy.nra`): the credits `credit_<enum>` gives to arrays and maps inside a heap enum payload are never removed, because those arrays keep count 1. After a `copy(tree)` ends, its credit still explains a later hidden copy of the original's array. Remove them when the payload box is released (an `uncredit_<enum>` walk calling `narra_copy_forget`). Probe: `p2.nra` in the batch-6 review scratchpad; add it as a sample that expects exit 103. Also rename `creditFunction` and `creditCode` to action names and move their function-level comments inside.
- `testdata/copy_report.nra` hides its two copies behind a `mutable` argument to a user function, which the static rule does not follow (shared-write-sweep-rest outcome, Gotchas). If this node changes that, give the sample a new hidden copy the static rule cannot see, or move it to `testdata/errors/`.
- Error samples for each D2 form with all-scalar structs, which the runtime check cannot see (copy-builtin outcome, Gotchas: borrowed values and inline all-scalar structs).

## Acceptance

- [ ] Every shared-write warning sample from shared-write-check is an error sample.
- [ ] Every program sample runs with the copy check on and reports nothing, or carries a `.copies` file whose line the outcome explains.
- [ ] One error sample per D2 form with an all-scalar struct.
- [ ] `sh scripts/test.sh` passes.
