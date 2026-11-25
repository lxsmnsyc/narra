---
id: pkg-yaml
title: "yaml package"
status: todo
priority: could
type: feature
story:
early: false
outline: true
milestone: p4-packages
target:
ship: auto
land: auto
from:
size: L
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/pkg-yaml/**
  - packages/yaml/**
shared:
  - narra.toml
  - scripts/test.sh
reserves: []
provides: []
depends_on: []
---

# yaml package

## Goal

A YAML 1.2 package that reads block and flow collections, scalars of the core schema, multi-line strings, anchors and aliases, and comments; and writes block-style YAML. Mapped onto the same value shape `std/json` and `std/toml` use, so a program can read any of the three.

## Context

- Source: `GOAL.md` Decisions, 2026-11-21 (YAML is a package); std-gap-survey rank 29.
- Motivating user: game data and mod files in YAML. The yaml-test-suite is the reference; record which cases are skipped and why.
- Tags beyond the core schema, and complex keys, can be left out.
