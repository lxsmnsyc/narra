---
id: language-guide
title: Narra language guide
status: todo
priority: should
type: docs
outline: true
milestone: m9-adoption
target:
from:
size: L
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
  - on: selfhost-bootstrap
    kind: hard
    why: the user chose to keep slots free for the must line; the guide starts after bootstrap
---

# Narra language guide

## Goal

A guide that teaches Narra to a programmer who knows TypeScript or Rust, in chapters that build on each other: values and types, functions and closures, structs and enums, pattern matching, Option and Result, traits, modules, memory and `mutable`, effects, testing with `std/test`. `SPEC.md` stays the reference; the guide links to it. Every code sample is a program under test, so the guide cannot drift from the compiler.

## Notes

Outline. Probably splits into two or three nodes by chapter group when refined. Only document what the compiler implements; mark parked spec features (m4-m7) as planned. Waits for bootstrap by user decision (2026-11-01), though it needs no compiler work.
