---
id: pkg-protobuf
title: "protobuf package: runtime and code generator"
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
  - roadmaps/selfhost/nodes/pkg-protobuf/**
  - packages/protobuf/**
shared:
  - narra.toml
  - scripts/test.sh
reserves: []
provides: []
depends_on: []
---

# protobuf package: runtime and code generator

## Goal

A protobuf (proto3) package: a runtime that encodes and decodes the binary wire format, and a generator that reads `.proto` files and writes Narra types with encode and decode functions. Its output must interoperate byte for byte with other implementations such as `@bufbuild/protobuf` and Go's `protobuf`.

## Context

- Decide in refinement whether the generator parses `.proto` itself or reads the `FileDescriptorSet` that `protoc` or `buf` writes (as a `protoc` plugin). The descriptor route avoids a `.proto` parser.
- Well-known types needed first: `Timestamp`. JSON mapping can wait.
- Split into runtime and generator nodes when refined if it stays L.
