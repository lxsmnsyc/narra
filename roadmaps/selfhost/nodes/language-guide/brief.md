---
id: language-guide
title: Narra language guide
status: todo
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

A guide in `docs/guide/` that teaches Narra to a programmer who knows TypeScript or Rust, in chapters that build on each other. `SPEC.md` stays the reference; the guide links to it. Every code sample is a program that a test compiles and runs, so the guide cannot drift from the compiler. The maintainer reviews the guide as a whole in the story pull request.

## Children

- `guide-basics`: the sample test harness, the index, and chapters 1 to 5.
- `guide-advanced`: chapters 6 to 10.

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

## Acceptance

- [ ] `docs/guide/README.md` lists the chapters in order with links.
- [ ] Every chapter is written, and every code sample in it is checked by `TestGuide`.
- [ ] Only implemented features are taught. Parked spec features (m4 to m7: generators, async, threads, regions, C interop) appear only in a short "Planned" list in the index, linked to `SPEC.md`.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L as a whole; split into two M children by chapter group so they run in parallel and each stays under the context budget.

## Notes

- `spec-samples-test` (m12) later checks `SPEC.md` samples. It should reuse the harness from `guide-basics` rather than write a second one.
