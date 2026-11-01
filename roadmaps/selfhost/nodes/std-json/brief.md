---
id: std-json
title: std/json and derives Shape
status: todo
priority: should
type: feature
outline: true
milestone: m10-std
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
  - on: std-source
    kind: hard
    why: the spike decides whether this module is Narra source or compiler builtins
---

# std/json and derives Shape

## Goal

`derives Shape` (SPEC section 5) and `std/json` with `Json.encode(value)` and `Json.decode(text, Type)` for any type that derives `Shape`, plus a dynamic `Json.Value` enum for untyped input. The language service needs it for JSON-RPC.

## Notes

Outline. Probably splits into a Shape node and a json node when refined.
