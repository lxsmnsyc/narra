---
id: batch-9
title: "Batch 9: debt the fresh seed unblocks"
status: active
priority: should
type: story
story: 
early: false
outline: false
milestone: m9-adoption
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/batch-9/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-8
    kind: hard
    why: the inbox items below waited for the seed regenerated at the end of batch-8, and some are batch-8 review findings
---

# Batch 9: debt the fresh seed unblocks

## Goal

Ninth one-shot batch, scope set by the maintainer on 2026-11-11: inbox items that waited for a seed update, plus two test-infrastructure bugs that slow every worker. No new syntax. Children land on `feat/batch-9` as direct commits after a worker check and a fresh review. At the end: the full check and the Linux check, one fresh review of the whole branch, a bench row, and one pull request.

Children:

- test-runner-fixes: `--suite` filters that run 0 cases and print PASS, the fixed `io_files` temp folder, and the dead code in the test runner's text helpers and in `std/src/json.nra` `cutText`.
- process-signals: `narrac run` and `narrac test` forward termination signals to the program they start.
- process-exit: a std function that ends the process with an exit code.
- host-target-builtin: a builtin that returns the host target from the C compiler's predefined macros.
- ptr-cast: `Ptr.cast` and `Ptr.element` in `std/ptr`, and the `from /* x */ "c"` fmt comment.
- diagnostics-polish: the mutual `typeof` cycle message, a code of its own for "type not known yet", the narrowing cascade in `narrowing_shadowed_panic`, and source positions past the 20-bit packing.
- hidden-copy-fix: the copy in `loadFile` that a debug-built compiler reports (exit 103), and the unguarded `NARRA_COPY_GRANT` in `takeOwnCopy`.
- fmt-block-hug: `narrac fmt` hugs a block expression in the last call argument and as a binary operand, as oxfmt does.
- guide-multifile-samples: the guide suite compiles multi-file package samples (chapters 1, 4, 8).
- seed-switch, after every other child: after a mid-batch seed update, the compiler uses process-exit (in place of `skipTeardown`) and host-target-builtin (in place of running `uname`).

Scope changes during the batch:

- hidden-copy-fix (2026-11-12): narrowed to the copies in `loadFile` and `loadPackage`. A debug-built compiler still exits 103 because of hidden copies in the checker and emitter; that work moved to the inbox. The `NARRA_COPY_GRANT` item was stale and needed no change.

## Order and shared files

- Wave 1, four workers at a time (8 GB machine): test-runner-fixes, process-exit, host-target-builtin, ptr-cast. Then, as slots free: process-signals, hidden-copy-fix, diagnostics-polish, guide-multifile-samples, and fmt-block-hug last (it reformats sources the others edit, so it rebases onto them and reruns `narrac fmt`).
- process-exit, host-target-builtin and ptr-cast each add a std function and touch `stdtext.nra`, `builtins.nra`, `SPEC.md` and the diagnostics table only for dispatch cases and text; the logic goes in each node's own module. The second to land rebases and keeps both sides.
- process-signals and process-exit both edit the process runtime (`selfhost/runtime/std_process.c` and its generated `runtimestdprocess.nra`). Land process-exit first if both are ready; the other rebases and regenerates.
- Lesson from batch-7 and batch-8: until test-runner-fixes lands, a targeted suite that prints PASS in seconds may have run no cases, so trust only the full run. Check a rebase's exit code before testing. Each worker uses a scratch folder named after its node under /private/tmp/claude-501/.

## Seed

Runtime text changes (process-signals, process-exit) need two seed updates (GOAL.md constraint), and seed-switch needs the seed to know the new std functions. The end of the batch is: seed update, seed-switch, seed update. The compiler source uses no new function before the first update.

## Acceptance

- [x] Every child is done, or moved out of the batch with a reason in this brief.
- [x] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [x] The seed is regenerated and `sh selfhost/bootstrap.sh` reports it current.
- [x] A fresh review of the whole story branch, before the pull request.
- [x] `bench/BASELINE.md` has an end-of-batch row; `narrac emit selfhost` instructions do not rise (process-exit should lower them a little).
- [x] The inbox entries this batch closes are removed from `inbox.md`.
