---
id: batch-4
title: "Batch 4: the performance batch"
status: done
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
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/batch-4/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Batch 4: the performance batch

## Goal

Fourth one-shot batch, scope chosen by the maintainer on 2026-11-10: make the self-hosted compiler and the programs it builds faster by spending less on reference counts and allocation. Children land on `feat/batch-4` as direct commits after targeted tests; at the end one full check, one Linux check, one fresh review and one pull request.

Children:

- Wave 1, in parallel: bench-compiler (the baseline every perf child measures against), regions-design (spike, no code), and the small fixes from the inbox: closure-arg-temps, template-escapes, nested-or-guards, map-key-panic-sample, typed-calls-and-effects.
- Then profile-compiler (spike), and perceus after the two emitter fixes.
- Then option-inline, then perceus-reuse and const-params-pointer.
- fast-alloc (added 2026-11-10 from regions-design) runs beside perceus: it edits only the runtime.
- From profile-compiler (2026-11-10): int-to-string, effect-int-keys, cache-type-names, finish-facts-lines, emit-string-helpers, quote-on-error, template-one-alloc and lexer-slices. Tail-loop parameter borrowing (profile proposal 3) went to the inbox.
- profile-compiler may propose more children; the orchestrator adds the ones the maintainer accepts.

Benchmark children hold the `benchmarks` resource, so they never run at the same time.

## Seed

Every child changes the emitted C. The seed is regenerated after perceus, after option-inline and at the end, so later children and the compiler itself run on the new code.

## Acceptance

- [x] Every child is done, or moved out of the batch with a reason in this brief.
- [x] `bench/BASELINE.md` has a row for the start of the batch and one for the end, recorded on an idle machine, including the compiler row.
- [x] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [x] The seed is regenerated at the end and `sh selfhost/bootstrap.sh` reports it current.

All 22 children landed. Speed nodes: fast-alloc, perceus, const-params-pointer, effect-int-keys, lexer-slices, option-inline, int-to-string, emit-string-helpers, template-one-alloc, quote-on-error, finish-facts-lines and cache-type-names. Fixes: closure-arg-temps, nested-or-guards, or-pattern-crash and map-key-panic-sample. Spikes and measurement: bench-compiler, profile-compiler, regions-design (D100, regions parked at could) and perceus-reuse (measured, not built). template-escapes and typed-calls-and-effects were already fixed on main.

Start to end, same input, idle M2: `narrac emit selfhost` 0.61 s to 0.23 s (9.30G to 2.59G instructions, 129 MB to 121 MB peak). interp 0.41 s to 0.19 s, particles 0.18 s to 0.11 s, strings 0.31 s to 0.22 s; fib, sieve and matmul about the same. Cost: interp about 20% slower than right after perceus because of option-inline (D99).

One fresh review of the whole branch: no blocking findings; a D55 cost line, the D50 number and a SPEC sentence fixed. Final checks at afb1eac3: scripts/test.sh PASS, linux-check PASS, the seed current.

Merged into main as PR #109 on 2026-11-10. Follow-ups planned as nodes: profile-compiler-2, profile-programs, tail-loop-borrow, map-loop-borrow, cache-type-names-rest, string-concat-one-alloc and interp-selects.
