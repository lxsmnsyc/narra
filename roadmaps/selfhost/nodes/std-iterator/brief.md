---
id: std-iterator
title: std/iterator
status: todo
priority: should
type: feature
outline: true
milestone: m10-std
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: generators
    kind: hard
    why: every operator is a generator function over a Generator<T>
  - on: std-source
    kind: hard
    why: the operators should be Narra source, which the spike decides
---

# std/iterator

## Goal

Lazy operators over `Generator<T>` (D31: iterators are generators): the ES2025 iterator helpers (map, filter, take, drop, flatMap, reduce, toArray, some, every, find, forEach) plus zip, enumerate, chunk and window from IterableX and Rust. The eager `Arrays.*` functions stay; `Iterators.*` is for streams, large inputs and early exit.

## Notes

Outline. Open points for refinement: operator callbacks must be escaping because the returned generator stores them; each stage costs an indirect call per element (measure, and see whether `inline` helps); document whether an operator consumes its source, since `for ... of` iterates its own copy; nested calls read badly without method syntax, so note whether a pipe operator is worth a language decision.
