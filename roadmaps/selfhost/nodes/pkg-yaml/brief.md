---
id: pkg-yaml
title: "yaml package"
status: active
priority: could
type: feature
story: batch-26
early: false
outline: false
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
reserves: []
provides: []
depends_on: []
---

# yaml package

## Goal

A YAML 1.2 package that reads block and flow collections, scalars of the core schema, multi-line strings, anchors, aliases and merge keys, and comments; and writes block-style YAML. Mapped onto the `std/shape` `Value` that `std/json` and `std/toml` use, so a program decodes the same `derives Shape` type from any of the three.

## Context

- Source: `GOAL.md` Decisions, 2026-11-21 (YAML is a package); std-gap-survey rank 29. Motivating user: game data and mod files in YAML.
- Follow the package pattern of `packages/http` (read `roadmaps/selfhost/nodes/pkg-http/outcome.md`): a workspace member, `[exports]`, a test binary run by `scripts/test.sh` for every `packages/*`, and an example under `examples/` that `scripts/test.sh` builds as an outside program. Exported functions, including the test entry `run`, declare their effects.
- `std/toml` (`TomlError`, `parse`, `stringify`, `decode`, `encode`) is the model for names and error style.

## API (maintainer approved, 2026-11-26)

One module, `"yaml"`:

- `YamlError { line, column, reason }` (both from 1); `formatError(error)` gives `"line:column: reason"`.
- `ReadOptions { maxDepth (512), maxAliasNodes (100000), mergeKeys (true) }`; `defaultOptions()`.
- `parse(text): Result<Value, YamlError>` (exactly one document), `parseAll(text): Result<Array<Value>, YamlError>` (a `---` stream), `parseWith(text, options)`, `stringify(value): string`, `decode<T: Shape>(text): Result<T, string>`, `encode<T: Shape>(value): string`.

Reading:

- Block and flow mappings and sequences; plain, single- and double-quoted scalars (all escapes); `|` and `>` block scalars with chomping and indentation indicators; comments; `%YAML` directives and `---`/`...` markers.
- Core schema: `null`, `~` and empty give `Null`; `true`/`false` give `Bool`; decimal, `0x`, `0o` give `Int`; floats, `.inf`, `.nan` give `Float`; else `Text`. The tags `!!str`, `!!int`, `!!float`, `!!bool`, `!!null`, `!!seq`, `!!map` set the type; any other tag is an error.
- Keys: a scalar key becomes its text (`1: a` gives key `"1"`); a collection key is an error; a duplicate key is an error; keys keep their order.
- Anchors and aliases: an alias gives a copy of the anchored node. `maxAliasNodes` caps the nodes copies make in total, so an exponential alias file fails instead of exhausting memory. An alias to a node that contains it is an error.
- Merge keys (YAML 1.1, on by default, off with `mergeKeys: false` where `<<` is then an ordinary key): `<<: *a` or `<<: [*a, *b]` copies the fields in; the mapping's own keys win, and earlier entries of a list win over later ones. A merge value that is not a mapping or a list of mappings is an error.

Writing: block style, fields in order; a string is quoted only when it would read back as another type or break the syntax; a string with line breaks is a `|` block; empty list `[]`, empty record `{}`; no anchors.

## Acceptance

- [ ] The root `narra.toml` lists `packages/yaml`, and `scripts/test.sh` runs its checks.
- [ ] The yaml-test-suite data, vendored under `packages/yaml/` (no network), runs in the checks; every skipped case is listed in the outcome with its reason (complex keys and non-core tags are expected skips).
- [ ] Round trip: `parse(stringify(value))` gives `value` back for a property-style check over random `Value` trees.
- [ ] Anchors, aliases and merge keys give the documented results; an alias bomb hits `maxAliasNodes`, deep nesting hits `maxDepth`, and a self-containing alias is an error, none of them a panic; a fuzz check feeds random text.
- [ ] `decode` reads one `derives Shape` struct from equal YAML, JSON and TOML texts.
- [ ] An example under `examples/` reads a YAML file and builds as an outside program.
- [ ] `packages/yaml/README.md` documents the package and its limits.
- [ ] `sh scripts/test.sh` passes.

## Sizing

L: the scanner and block structure (high), flow collections and scalars (medium), aliases and merge keys (low), writer (low), the test-suite harness (medium).
