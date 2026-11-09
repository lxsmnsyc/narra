---
id: type-aliases
title: "type and export type aliases"
status: active
priority: should
type: feature
story: batch-3
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns:
  - testdata/type_alias*
  - testdata/errors/type_alias*
  - roadmaps/selfhost/nodes/type-aliases/**
shared:
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - docs/guide/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [code E0972, code E0973]
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the batch-2 compiler
---

# type and export type aliases

## Goal

SPEC.md sections 3 and 4 (audit F13): `type Tok = Token` fails with `expected a declaration, found "type"`. Build transparent aliases (maintainer, 2026-11-09):

- `type Name = T` and `export type Name = T`; the alias is the same type, and enum constructors work through it (`Tok.Word(x)`, in patterns too).
- Generic aliases: `type Pair<T> = #[T, T]`.
- An alias may not refer to itself, directly or through other aliases (error; recursion needs an enum).
- Diagnostics show the alias with the underlying type (`Tok (Token)`), and `narra query` prints the alias.
- `export type Token = Text.Token` re-exports a type from another module (the SPEC workaround for D4).
- Nothing reaches the emitted C.

Parser, checker, `fmt`, `astdump`, declsdump. Do not use aliases in `selfhost/src` in this batch.

## Acceptance

- [ ] Samples: an enum alias with constructors and patterns, a tuple alias, a generic alias, a re-export across modules in a package sample.
- [ ] Error tests (E0972, E0973): a self-referencing alias, a cycle through two aliases, a wrong number of type arguments.
- [ ] A diagnostic sample shows the alias with the underlying type.
- [ ] compiler/README.md no longer lists aliases as not supported; SPEC.md matches.
- [ ] `sh scripts/test.sh` passes.
