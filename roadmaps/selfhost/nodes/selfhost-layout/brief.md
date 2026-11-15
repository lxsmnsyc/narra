---
id: selfhost-layout
title: Group the self-hosted compiler's modules into directories
status: todo
priority: could
type: refactor
story: batch-16
early: false
outline: false
milestone: p4-packages
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent:
owns:
  - roadmaps/selfhost/nodes/selfhost-layout/**
  - selfhost/src/**
  - scripts/layout-move.sh
shared:
  - selfhost/tests/src/generated.nra
  - selfhost/tests/src/slotstores.nra
  - selfhost/tests/src/sources.nra
  - scripts/*.sh
  - selfhost/CLAUDE.md
  - seed/narrac.c
reserves: []
provides: []
depends_on:
  - on: package-dependencies
    kind: hard
    why: the move rewrites every import in selfhost/src, so it lands after every other batch-16 node that edits selfhost/src
  - on: package-exports
    kind: hard
    why: same
  - on: deps-manifest-lock
    kind: hard
    why: same
  - on: deps-fetch
    kind: hard
    why: same
  - on: deps-resolve
    kind: hard
    why: same
  - on: target-table
    kind: hard
    why: same
  - on: wasi-runtime
    kind: hard
    why: same
  - on: windows-runtime
    kind: hard
    why: same
  - on: windows-process-async
    kind: hard
    why: same
  - on: cross-targets
    kind: hard
    why: same
  - on: dead-code-check
    kind: hard
    why: same
  - on: let-never-reassigned
    kind: hard
    why: same
---

# Group the self-hosted compiler's modules into directories

## Goal

`selfhost/src/` holds 187 modules in one flat directory. Group them by area so a reader finds a feature's code in one place. Behavior does not change: every test program's emitted C, every dump and the bootstrap fixed point stay as they are. Requested by the maintainer on 2026-11-05.

## What the code already answers

- **Nested paths work.** `collect` in `selfhost/src/modules.nra` walks `src/` recursively, and `src/check/body.nra` loads as `narrac/check/body` (fixture `testdata/pkgdump/nested_order`). No loader change is needed.
- **The compiler's own C names change.** `moduleHash` in `hash.nra` hashes the module path inside the package, so every symbol of narrac gets a new hashed name. Test program goldens do not change, because they are other packages. `seed/narrac.c` changes in full: regenerate it with `sh scripts/seed-update.sh` in this node, twice if runtime text moves.
- **The Go compiler is gone** (c-seed), so `selfhost/` is the compiler's home. Renaming the top directory is out of scope.

## Groups (approved by the maintainer, 2026-11-14)

`syntax/` (lexer, parser, ast, literals, trivia and their dumps), `check/` (checker passes, types, traits, effects, patterns, generics, async and thread checks), `emit/` (emit*, cnames), `runtime/` (runtime* text modules), `std/` (std* tables and checks), `tools/` (fmt*, lsp*, query*), and `driver/` (main, driver, manifest, modules, lockfile, dependency modules, hash, diagnostics). Write the final table in `log.md` first.

## Approach

Do the move with a script (`scripts/layout-move.sh`, kept in the node folder or deleted after) that runs `git mv` from the table and rewrites every `"narrac/<name>"` import. Then fix paths in the test runner (`generated.nra` writes `stdtext.nra`, `slotstores.nra` reads `selfhost/src/slotstores.nra`) and in `scripts/`. Run it in a quiet window: no other open node edits `selfhost/src`.

## Acceptance

- [ ] No `.nra` file is left directly in `selfhost/src/` except `main.nra`.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass; no golden file outside `seed/` changes.
- [ ] `sh selfhost/bootstrap.sh` reaches the fixed point and reports the seed current.
- [ ] `selfhost/CLAUDE.md` names the groups.

## Sizing

M: mechanical, scripted, with one seed update.
