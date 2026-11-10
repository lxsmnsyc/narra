---
id: workspaces
title: "Workspaces: several packages in one repository"
status: todo
priority: should
type: feature
story:
early: false
outline: true
milestone: m13-distribution
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

## Notes

Outline until workspace-design writes D101; refine the children from it. Likely related to package-dependencies (git dependencies, lock file) and package-exports (D97).
