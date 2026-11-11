---
id: test-runner-fixes
title: "Test runner: empty suite filters fail, unique temp folders, dead code"
status: todo
priority: should
type: fix
story: batch-9
early: false
outline: false
milestone: 
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
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/test-runner-fixes/**
  - selfhost/tests/**
  - scripts/test.sh
shared:
  - testdata/**
  - std/src/json.nra
  - selfhost/src/stdtext.nra
reserves: []
provides: []
depends_on: []
---

# Test runner: empty suite filters fail, unique temp folders, dead code

## Goal

Three inbox items about the test runner (`selfhost/tests/src/`, `scripts/test.sh`):

1. From batch-7 landing: `NARRAC=<built narrac> sh scripts/test.sh --suite <name>` ran 0 cases and printed PASS for `bench`, `errors`, `packages` and `emit` from a worktree (seen twice). Find the cause (runner build, path, or suite name), fix it, and make a run whose filters match no case fail with a message that names the filters.
2. From batch-6: `testdata/io_files.nra` writes to a fixed folder (`narra_io_files_test`), so two runs at once (parallel worktrees) delete each other's files. Use a folder unique to the run.
3. From the batch-6 review: `selfhost/tests/src/textops.nra` has a dead `return` after each `panic`; `cutText` in `std/src/json.nra` has a function-level comment and returns "" on a `None` the callers rule out. Remove the dead code and make `cutText` panic with a message like the other callers. Regenerate the std text (`sh scripts/test.sh --suite generated --update`).

## Acceptance

- Item 1 is a fix: show the 0-case PASS on main first (record the command and output in `log.md`), then show it gone on the branch. A test covers a filter that matches nothing.
- Two `io_files` runs at once both pass.
- `sh scripts/test.sh` passes.
