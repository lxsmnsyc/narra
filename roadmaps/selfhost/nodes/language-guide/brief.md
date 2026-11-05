---
id: language-guide
title: Narra language guide
status: active
priority: should
type: story
story:
early: false
outline: false
milestone: m9-adoption
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run TestGuide
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: the user chose to keep slots free for the must line; the guide starts after bootstrap
---

# Narra language guide

## Goal

A guide in `docs/guide/` that teaches Narra in chapters that build on each other, to a reader who has programmed before but needs no TypeScript or Rust. `SPEC.md` stays the reference for language designers; the guide does not cite it, its sections or the decision records. A separate chapter compares Narra with TypeScript for readers who know it. (Changed 2026-11-05 by the maintainer: the first drafts leaned on SPEC and D-number references and on TypeScript knowledge, so they did not read standalone.) Every code sample is a program that a test compiles and runs, so the guide cannot drift from the compiler. The maintainer reviews the guide as a whole in the story pull request.

## Children

- `guide-basics`: the sample test harness, the index, and chapters 1 to 5.
- `guide-advanced`: chapters 6 to 10.
- `guide-typescript`: chapter 11, Narra for TypeScript programmers, built from the comparisons chapters 1 to 10 leave out.

Both follow `contracts/guide-samples.md` (draft, frozen by guide-basics) on the story branch `feat/language-guide`. guide-advanced starts when guide-basics freezes the contract.

## Chapters

1. Getting started: a package, `narra.toml`, `run`, `narra build` and `narra run`.
2. Values and types: numbers, strings and templates, bool, `const` and `let`, arrays, maps, tuples.
3. Functions and closures: declarations, arrow functions, function types, non-escaping parameters, tail calls.
4. Structs and enums: declarations, literals, spread updates, Box and recursive enums, opaque and distinct types.
5. Pattern matching: `is`, `match`, destructuring, narrowing, block expressions and the ternary.
6. Option and Result: `std/option`, `std/result`, error handling without exceptions.
7. Traits: declarations, `implement`, constraints, `derives`, equality and ordering.
8. Modules and packages: imports, aliases, exports, std imports.
9. Memory and `mutable`: reference counting, copies, `mutable` parameters and their rules.
10. Effects and testing: `uses`, effect checking, `std/test` and `narra test`.
11. Narra for TypeScript programmers: what carries over, what differs and why, each habit with its Narra form and a link to the chapter that teaches it.

## Acceptance

- [ ] `docs/guide/README.md` lists the chapters in order with links.
- [ ] Every chapter is written, and every code sample in it is checked by `TestGuide`.
- [ ] Chapters 1 to 10 read standalone: no `SPEC.md`, section or D-number references, no "the author decided" framing, no TypeScript or Rust knowledge assumed. TypeScript comparisons live only in chapter 11.
- [ ] Chapters teach in a guided, tutorial voice: each section starts from something the reader wants to do, builds a program, shows its output, then explains. Rules come out of the examples (often a `narra error` sample), not bullet lists of restrictions. No mechanism is explained before the concept under it is taught. (Maintainer review of #87, 2026-11-05.)
- [ ] Only implemented features are taught. Parked spec features (m4 to m7: generators, async, threads, regions, C interop) appear only in a short "Planned" list in the index.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L as a whole; split into two M children by chapter group so they run in parallel and each stays under the context budget.

## Notes

- `spec-samples-test` (m12) later checks `SPEC.md` samples. It should reuse the harness from `guide-basics` rather than write a second one.
