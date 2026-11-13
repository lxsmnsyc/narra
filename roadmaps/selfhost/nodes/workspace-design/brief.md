---
id: workspace-design
title: "Spike: workspaces for several packages in one repository"
status: todo
priority: could
type: spike
story:
early: false
outline: false
milestone: p4-packages
target:
ship: auto
from:
size: M
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/workspace-design/**
shared:
  - DECISIONS.md
reserves: [decision D101]
provides: []
depends_on: []
---

# Spike: workspaces for several packages in one repository

## Goal

Decide whether and how one repository can hold several Narra packages that depend on each other and build, check and test together, like Cargo workspaces or pnpm workspaces. Write the answer as decision D101 and propose the child briefs for node workspaces, or say why it should wait.

## Context

- Today a package is a directory with `narra.toml` and `src/` (SPEC section 15, `selfhost/src/manifest.nra`, `selfhost/src/modules.nra`). `[dependencies]` is rejected with E0407, and git dependencies (node package-dependencies) are not built. Module visibility across packages is D97 (node package-exports, not built).
- Cargo: a root `Cargo.toml` with `[workspace] members = [...]`, path dependencies between members, one lock file and one target directory. pnpm: `pnpm-workspace.yaml` globs and `workspace:*` versions.
- This repository could be the first user: `selfhost/`, `std/` and the `bench/` programs are separate trees today, and only `selfhost/` has a `narra.toml`.

## Questions

1. **Manifest form.** A `[workspace]` table with `members` globs in a root `narra.toml`, and whether the root can also be a package.
2. **Dependencies between members.** A path form (`json = { path = "../json" }`) or a workspace form (`json = { workspace = true }`), how it relates to the planned git dependencies and minimal version selection, and whether path dependencies can land before git dependencies (and so before package-dependencies).
3. **Commands.** What `narra build`, `check`, `test`, `fmt` and `query` do at the root and inside a member (all members, the current one, `--package <name>`), and the build output layout.
4. **One lock file** at the root once git dependencies exist; one version per package across the workspace (SPEC section 15).
5. **Language service.** How `narra lsp` finds the workspace root from a file (today it uses the nearest `narra.toml`) and checks only the members a change can affect.
6. **Exports.** How member-to-member imports follow D97, and whether package-exports must land first.
7. **Cost.** What changes in the loader (`modules.nra`), the manifest parser, the driver and the language service, and how many nodes that is.

## Acceptance

- [ ] D101 written in the style of D100, with Decision, Reason, Cost and Rejected, marked "Not built yet", or a recommendation to wait with the reason.
- [ ] Proposed child briefs in `outcome.md` (id, title, goal, size, dependencies, acceptance), including whether this repository moves to a workspace.
- [ ] The maintainer reads D101 before any child starts (gate).

## Sizing

M: a design over the manifest, loader, driver and language service, with no shipped code.
