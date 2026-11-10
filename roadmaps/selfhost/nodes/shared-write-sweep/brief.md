---
id: shared-write-sweep
title: "Fix the shared writes in the checker"
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
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns:
  - roadmaps/selfhost/nodes/shared-write-sweep/**
shared:
  - selfhost/src/**
  - DECISIONS.md
reserves: []
provides: []
depends_on:
  - on: shared-write-check
    kind: hard
    why: needs the warnings and their fixes
  - on: proven-operands
    kind: hard
    why: both rewrite call sites across selfhost/src; the debate orders D94 after the D92 nodes
  - on: std-no-panic
    kind: hard
    why: both rewrite call sites across selfhost/src and std/src; the debate orders D94 after the D92 nodes
---

# Fix the shared writes in the checker

## Goal

D94 step 4, first half (debate `roadmaps/selfhost/debate/value-semantics/debate.md`, P19, O9, O10). Split from the original sweep on 2026-11-11 after shared-write-check measured 198 warnings in `selfhost/src` plus 185 shared returns (`nodes/shared-write-check/outcome.md`, Warning counts). This node covers the checker modules: every file in `selfhost/src` except `emit*.nra`, `cnames.nra`, `fmt*.nra`, `query*.nra`, `lsp*.nra`, `trivia*.nra`, `parser.nra`, `lexer.nra`, `modules.nra`, `manifest.nra` and `driver.nra`, which shared-write-sweep-rest owns. About 140 of the 198 warnings and most of the 185 returns are here.

- Shared writes (E0978): apply the `copy(...)` fixes with `narrac apply-fix`. Edit by hand the loop-held writes (mostly `for (const x of Arrays.values(checker.<list>)) { reportError(mutable checker, ...) }`): loop over a `copy(...)`, or collect and write after the loop, whichever copies less. Edit by hand the save/restore sites in `resolve.nra:510`, `closures.nra:162` and `stmts.nra:1125`, `:1132` (O9).
- Shared results (E0976): add `from <parameter>` to the checker functions that share a parameter, and `copy(...)` where one shares two. Run the check with the `shared-results = "unchecked"` line removed from `selfhost/narra.toml` locally to list them, but keep the line in the commit: shared-write-sweep-rest removes it.

No behavior change except removed hidden copies.

## Notes

- The seed reads `copy(...)` and `from` (seed updated at the end of batch-5).
- Before choosing between a copy and moving a write, prefer the edit that copies less: a `copy(...)` of a whole `Checker` field on a hot path costs instructions every call.

## Acceptance

- [ ] `narrac check selfhost` reports no E0978 in the checker modules, and, with the switch removed locally, no E0976 in them.
- [ ] `sh selfhost/bootstrap.sh` reaches the fixed point.
- [ ] Instructions retired for `narrac emit selfhost` before and after, in the outcome.
- [ ] `sh scripts/test.sh` passes.
