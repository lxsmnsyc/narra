---
id: mutable-reference
title: "Rethink mutable parameters as mutable references"
status: todo
priority: should
type: experiment
outline: true
milestone: m12-spec-gaps
target:
from:
size: L
spike: true
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-04
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: mutable-read-conflict
    kind: hard
    why: the soundness fix ships first; the redesign builds on what it learns
---

# Rethink mutable parameters as mutable references

## Goal

The maintainer (2026-11-04): "mutable variable shouldn't be a thing anyways, it should be a mutable reference." Today `mutable` marks a parameter and is repeated at the call (`damage(mutable hero, 10)`); the compiler passes a pointer, access is exclusive during the call, and the parameter cannot escape (SPEC section 6, D50). Decide whether this becomes an explicit mutable reference: a type or a reference form at the call (for example `&mut`-like), what it may point at (locals, fields, elements), whether it can be stored or returned, and how it interacts with narrowing, closures, copy-on-write and the exclusive-access rule.

## Notes

- Outline. Start with a goal debate (agent-skills:goal-debate) before refining: it is a language-model change that touches the spec, both compilers and every `mutable` use in `selfhost/src`.
- Inputs: SPEC section 6 (mutable parameters, cycles), DECISIONS.md D46, D50, D51, the mutable-escape outcome (what can and cannot be stored), the let-narrowing outcome, mutable-read-conflict.
- Prior art to survey: Rust `&mut` with borrow checking, Swift `inout` with exclusivity enforcement (SE-0176), C# `ref`, Val/Hylo mutable projections, Koka/Lean in-place update by uniqueness.
