---
id: selfhost-names-exported
title: Readable exported names across the selfhost modules
status: done
spent: 144000
priority: must
type: refactor
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestSelfhost|TestASTDump|TestTypesDump'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-31
owns:
  - selfhost/src/**
  - selfhost/CLAUDE.md
  - roadmaps/selfhost/nodes/selfhost-names-exported/**
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-check-effects
    kind: hard
    why: renames fields and functions every checker module uses; it must run when no other node edits selfhost/src
  - on: typesdump-targs-spacing
    kind: hard
    why: both edit typesdump.nra; run after it so the rename does not conflict
---

# Readable exported names across the selfhost modules

## Goal

Finish the readable-names pass. selfhost-names-frontend and selfhost-names-checker renamed only names private to a file, because the two ran in parallel. The exported short names are left, and their outcomes list them. Rename them in one pass across every module, before selfhost-emit starts reading them.

## Context

- Read `roadmaps/selfhost/nodes/selfhost-names-frontend/outcome.md` and `roadmaps/selfhost/nodes/selfhost-names-checker/outcome.md` for the full lists and the method (a token-aware renamer that skips string text, then a literal diff against main).
- Known names: AST fields `id`, `op`, `col`, `params`, `args`, `decls`, `stmts`, `funcs` and the other short ones; idcheck `funcIds`, `blockIds`, `exprIds`; `Diag`, `diagText`; `Checker.cur`, `Checker.errs`, `BodyInfo.targs`; any other exported name of one to five letters that is an abbreviation.
- Follow the naming rule in `roadmaps/selfhost/CLAUDE.md`. No behaviour change.

## Rules

- Do not change anything printed: dump keywords and attribute names (`targs=`, `ctor=`, kind names such as `field-expr`), error codes and messages. A struct field may be renamed even when its dump keyword keeps the short spelling.
- Do not rename Go code; only `selfhost/`.
- `function`, `from` and `to` are keywords; shadowing is an error.

## Acceptance

- [ ] No exported struct field, enum variant field or exported function in `selfhost/src` is an abbreviation (`id` may stay: it is a word in the domain; say so in the outcome if you keep others).
- [ ] Every string, char and template literal in `selfhost/src` is byte-identical to main (literal diff).
- [ ] Every selfhost, AST dump and types dump test keeps its pass and skip counts.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`
- [ ] The outcome lists every renamed exported name, old -> new, for selfhost-emit.

## Sizing

M: mechanical, scripted, wide but low risk; the parity tests catch slips.
