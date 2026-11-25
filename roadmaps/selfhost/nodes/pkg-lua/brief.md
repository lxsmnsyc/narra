---
id: pkg-lua
title: "lua package: binding to the Lua C library"
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
size: M
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
  - roadmaps/selfhost/nodes/pkg-lua/**
  - packages/lua/**
shared:
  - narra.toml
  - scripts/test.sh
reserves: []
provides: []
depends_on:
  - on: package-c-sources
    kind: hard
    why: the package compiles the Lua C sources it ships
  - on: c-export-functions
    kind: hard
    why: Lua calls host functions through lua_CFunction callbacks into Narra
---

# lua package: binding to the Lua C library

## Goal

A package that ships the Lua 5.4 C sources and binds the C API: create a state, load and run chunks, move values between Narra and Lua, register Narra functions Lua can call, and limit a script with an instruction count hook. Scripts run sandboxed: only the libraries the host opens.

## Context

- Motivating user: a game server that runs mod scripts in Lua, today through a cgo binding.
- Variadic Lua API functions (`lua_pushfstring`) go through C wrappers in the package.
- `c-export-functions` is `later` priority; refinement must raise it or this node waits.
