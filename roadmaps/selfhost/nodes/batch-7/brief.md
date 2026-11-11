---
id: batch-7
title: "Batch 7: compiler cost, smaller C output, contracts and build conditions"
status: active
priority: should
type: story
story: 
early: false
outline: false
milestone: m6-performance
target:
ship: pr
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/batch-7/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-6
    kind: hard
    why: builds on the shared-write analysis, the proven-operands prover and the seed regenerated at the end of batch-6
---

# Batch 7: compiler cost, smaller C output, contracts and build conditions

## Goal

Seventh one-shot batch, scope set by the maintainer on 2026-11-11. Batch-6 raised `narrac emit selfhost` from 1.74G to 2.47G instructions and peak memory from 125 MB to 135 MB, and the seed grew 8.3% (171.9k to 186.2k lines) against 6.4% more source. This batch takes back the cost, shrinks the C the compiler writes, and builds two D95 syntax features. Children land on `feat/batch-7` as direct commits after targeted tests. At the end the batch gets one full check, one Linux check, one seed update, one fresh review of the whole branch and one pull request.

Children:

- Compiler cost: place-ids (the shared-write analysis), shared-results-walk (the `from` check).
- C output: return-trailer, copy-check-macros, tree-shake, tree-shake-types.
- Syntax (D95): contracts, then contracts-ensures; build-conditions, then build-targets.
- Inbox: bench-check-suite.

## Evidence

Profile and C sizes from the batch-7 planning session (2026-11-11, commit 94360a39, Apple M2): `narrac emit selfhost` is 2.49G instructions and 138 MB peak.

- Sampling (`sample`, 8 runs, 1172 samples): `reportSharedWrites` holds 272 samples (about 23%). Self time under it: `findRest` 40, `memcmp` 43, `hasPair` 35, `evaluate` 20. Place keys are strings (`"L3.items[=0]"`) matched by slicing; `hasPair` scans every pair, and `mergeStates` calls it per pair.
- Seed C: 487 three-line `#ifndef NARRA_RELEASE` blocks, 76 `credit_`/`uncredit_` functions and 1036 `NARRA_COPY_GRANT` lines (about 2.5k lines, 250 KB; `unifdef -DNARRA_RELEASE` gives 183.6k lines). 1878 functions end with an unreachable `narra_panic(... "function ended without returning a value"); return (T){0};`. Std Narra-source modules are emitted whole: `bench/fib` carries `compare`, `divmod` and `floorDiv` from `std/src/ext__int.nra` without calling them.

## Order and shared files

- The output nodes all edit the emitter. return-trailer and copy-check-macros touch different files and land in any order. tree-shake waits for copy-check-macros (both edit `emitdecls.nra`) and for build-conditions (it must skip folded branches); tree-shake-types comes last. Each later branch rebases first. build-conditions also edits `emitIf` (`emitstmts.nra`) and `emitexprs.nra`.
- contracts and build-conditions both edit the parser, checker and `fmt`; they may run at once and land in the order they finish, the second rebasing. To keep the conflicts small, each syntax feature puts its checking and folding in a module of its own (for example `selfhost/src/contracts.nra`, `selfhost/src/buildconditions.nra`), and edits the shared files (`lexer`, `parser`, `ast`, `astdump`, `check`, `fmt`, `diagnostics`, `SPEC.md`, `DECISIONS.md`) only for dispatch cases, codes and text. No split node first: the shared files are switch tables that every syntax batch has edited this way.
- The syntax children were the outline nodes contracts (m12-spec-gaps) and build-conditions (m4-systems); they ship with this story's milestone.
- place-ids and shared-results-walk run at once (`sharedwrites.nra` imports two helpers from `sharedresults.nra`; keep their signatures).

## Start order (one-shot, orchestrator's call)

- First: place-ids, contracts, build-conditions, copy-check-macros. place-ids is the largest measured cost; build-conditions and copy-check-macros start the longest chain (tree-shake, then tree-shake-types).
- As slots free: tree-shake (when both of its inputs land), contracts-ensures, build-targets, return-trailer, shared-results-walk, bench-check-suite, then tree-shake-types.
- Four workers at a time (8 GB machine, each runs the test suites).

## Seed

copy-check-macros edits runtime text in `selfhost/src/runtime*.nra`, and contracts and build-conditions add syntax. The compiler source must not use the new syntax in this batch. Regenerate the seed once at the end; if copy-check-macros needs the seed to know new runtime macros before the compiler can emit them, the end of the batch is seed update, then the switch, then seed update (GOAL.md constraint).

## Acceptance

- [x] Every child is done, or moved out of the batch with a reason in this brief.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch. (test.sh passes; the Linux check did not run, Docker was not running.)
- [x] The seed is regenerated at the end and `sh selfhost/bootstrap.sh` reports it current.
- [ ] `narrac emit selfhost` instructions on the end-of-batch source are at or below 2.1G (missed: 2.18G; 2.05G on the batch-6 source, see wave-20), and the release-stripped seed (`unifdef -DNARRA_RELEASE seed/narrac.c | wc -l`) is smaller than at the start of the batch even with the new features.
- [x] A fresh review of the whole story branch, before the pull request.
- [x] `bench/BASELINE.md` has an end-of-batch row, recorded on an idle machine.
