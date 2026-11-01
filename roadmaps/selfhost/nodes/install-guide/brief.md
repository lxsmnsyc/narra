---
id: install-guide
title: Installation guide
status: todo
priority: should
type: docs
outline: true
milestone: m9-adoption
target:
from:
size: S
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: the install steps change once the self-hosted compiler replaces the Go build
---

# Installation guide

## Goal

A guide for people who want to use Narra: prerequisites (a C11 compiler, Go for the bootstrap), building and installing `narra`, putting it on `PATH`, and a first program from `narra new` or a hand-written `narra.toml` to a running binary. Covers macOS and Linux, and names the platforms that are not supported yet.

## Notes

Outline. Break into a full brief when `plan.py refine` lists it. Keep the README short and link to the guide.
