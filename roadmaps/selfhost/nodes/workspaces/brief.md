---
id: workspaces
title: "Workspaces: several packages in one repository"
status: todo
priority: could
type: feature
story:
early: false
outline: false
milestone: p4-packages
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: workspace-design
    kind: hard
    why: D101 decides the design and splits this node into children
---

# Workspaces: several packages in one repository

## Goal

Build the workspace design from D101: a root manifest that lists member packages, dependencies between members, commands that run across members, and a language service that knows the workspace. Requested by the maintainer on 2026-11-10 ("monorepo/workspace support, if it's possible").

## Children

D101 decided the design (gate passed 2026-11-15). The children are not in batch-16. They build on the batch-16 dependency nodes and on contract `dependencies` version 2.

- workspace-manifest (M): `[workspace]`, `workspace = true`, `findWorkspace`, `listMembers`.
- workspace-loader (M): member dependencies, resolve from every member, one lock at the root.
- git-member-deps (M): git dependencies on a member of another repository.
- workspace-commands (M): driver commands at the workspace root.
- workspace-lsp (M): the language service over members.
- repo-workspace (S): this repository becomes a workspace; it lands after workspace-commands and workspace-lsp.

## Notes

Not in batch-16 (the batch brief says why). They start in the next batch, after deps-fetch, deps-resolve and package-dependencies land with the changes D101 asked for. repo-workspace and selfhost-layout both edit `selfhost/CLAUDE.md`, and either order works.
