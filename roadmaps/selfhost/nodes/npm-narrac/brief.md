---
id: npm-narrac
title: "narrac on npm, command form"
status: todo
priority: could
type: feature
story: batch-16
early: false
outline: false
milestone: p5-distribution
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/npm-package.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-15
spent:
owns:
  - roadmaps/selfhost/nodes/npm-narrac/**
  - npm/**
  - scripts/npm-package.sh
shared:
  - README.md
reserves: []
provides: []
depends_on:
  - on: narrac-wasm
    kind: hard
    why: the wasm package comes from `dist/narrac.wasm`
  - on: cross-targets
    kind: behavior
    via: targets
---
# narrac on npm, command form

## Goal

`npm/` in this repository holds narrac's npm packages in the esbuild layout (D123 item 6): `narrac-<os>-<arch>[-<libc>]` per row cross-targets builds, `narrac-wasm`, and a main `narrac` package with a `bin` entry, a small loader that picks the platform package or falls back to wasm, and the WASI preview1 calls narrac imports (no `node:wasi`). `scripts/npm-package.sh` fills them from the cross builds and `dist/narrac.wasm`. A JS API (`check(files)`, `format(source)`, `emit(files)`, diagnostics as data) with TypeScript types runs the command.

## Context

- This is a script of the package that ships. The compiler knows nothing about npm (D123 item 6): no `[npm]` table and no `narrac package` command.
- A fresh wasm instance per call costs 93 to 325 ms (T1), so the API is for tools that call it a few times, not in a hot loop.

## Acceptance

- [ ] The tarballs from `npm pack` install into a scratch project. `npx narrac check` and each API function match native narrac, once with the host package and once with only the wasm package.
- [ ] Nothing in `selfhost/src` or `narra.toml` changes.
- [ ] No publish: the maintainer publishes and confirms the package names.

## Sizing

M: packaging script, loader, WASI shim and a typed API.
