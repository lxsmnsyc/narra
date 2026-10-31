---
id: selfhost-names-frontend
title: Readable names in the selfhost front end
status: active
priority: must
type: refactor
outline: false
milestone: m8-selfhost
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-31
owns:
  - selfhost/src/lexer.nra
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/idcheck.nra
  - selfhost/src/manifest.nra
  - roadmaps/selfhost/nodes/selfhost-names-frontend/**
  - roadmaps/selfhost/CLAUDE.md
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-check-bodies
    kind: hard
    why: renames the files that node edits, and the four body modules it adds; running before it merges would conflict on every file
---

# Readable names in the selfhost front end

## Goal

Rename the selfhost code in the owned files to readable, descriptive names, as the maintainer writes JS/TS. Runs beside selfhost-names-checker, which owns the checker files. Also add the naming rule to `roadmaps/selfhost/CLAUDE.md` (one short paragraph) so every later worker loads it.

## Context

- The maintainer finds the selfhost code hard to read: it carries Go's short names over (`c`, `m`, `t`, `e`, `s`, `ok`, `idx`, `lt`, `rt`, `n`). Selfhost code must read like the maintainer's JS/TS projects. Load the agent-skills:ts-style skill and apply its rules to Narra.
- This is a rename and readability pass only. No behaviour change, no new features, no change to any dump format or error message.

## Rules

- Descriptive names for locals, parameters and fields: `checker`, `module`, `expression`, `statement`, `leftType`, `rightType`, `index`, `parameters`. A mutable checker parameter is `checker`, never `c`.
- Function names are actions (`checkBinary`, `resolveLocal`, `collectIds`), never `<noun>Of`. Rename functions only inside the files this node owns; when a renamed function is exported and used by the other names node's files, keep the old name and leave a note in the outcome instead.
- for-of style loops where Narra allows them.
- Comments describe the steps inside a function, not the function above it. Keep existing comments that explain why.
- Do not rename AST node kinds, dump keywords, error codes or anything printed.

## Acceptance

- [ ] No one- or two-letter locals, parameters or fields remain in the owned files, except loop counters `i`/`j` over plain index ranges where a name adds nothing.
- [ ] `TestSelfhostDecls`, `TestSelfhostTraits`, `TestSelfhostImports`, `TestSelfhostBodies` and the AST and id dump tests pass with the same compared and skipped counts as on main.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`
- [ ] The outcome lists every renamed exported name, so later nodes use the new names.

## Sizing

L by line count, but mechanical: low uncertainty, low coupling, the parity tests catch any slip. Rename one file at a time and run the fast check after each.

