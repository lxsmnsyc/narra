---
id: std-toml
title: "std/toml"
status: active
priority: should
type: feature
story: batch-14
early: false
outline: false
milestone: p2-std
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
added: 2026-11-13
spent:
owns:
  - std/src/toml.nra
  - selfhost/src/stdtoml.nra
  - testdata/toml*.nra
  - testdata/toml*.out
  - testdata/errors/toml_*
  - roadmaps/selfhost/nodes/std-toml/**
shared:
  - selfhost/src/stdlib.nra
  - selfhost/src/stdtext.nra
  - testdata/errors/unknown_module.err
  - testdata/declsdump/imports/err_unknown_module.dump
  - SPEC.md
  - compiler/README.md
  - docs/guide/
reserves: [diag E1340-E1349]
provides: []
depends_on:
  - on: std-design
    kind: hard
    why: D118 item 5 fixes the functions, the Value mapping and the error type

---

# std/toml

## Goal

Build `std/toml` (alias `Toml`) as D118 item 5 says: TOML 1.0 `parse` into the `Value` of `std/shape`, `stringify`, `decode<T: Shape>`, `encode<T: Shape>` and `formatError`, with `TomlError { line, column, reason }`. Pure Narra source that the seed compiles, since narrac will read `narra.toml` with it.

## Context

- Source: D118 item 5; D83 and D84; `std/src/json.nra` as the model (parser struct, depth limit, error positions).
- Today's reader: `selfhost/src/manifest.nra` (a TOML subset, codes E0405 to E0409). Moving it is not this node.

## Steps

- [ ] Parser: bare, quoted and dotted keys; basic, literal and multi-line strings with every escape; integers (decimal, `0x`, `0o`, `0b`, underscores, 64-bit range check); floats (`inf`, `nan`, exponents); booleans; dates and times kept as `Value.Text` of their text, validated; arrays (mixed types allowed); inline tables; `[table]` and `[[array]]` headers; the redefinition rules of TOML 1.0 (a table defined twice, a key set twice, an inline table extended). Nesting deeper than 512 is an error, as in JSON.
- [ ] `stringify`: plain keys before sub-tables, `[[name]]` for a `List` of `Record`s, quoted keys where needed, `Err` for `Null` and a non-`Record` root. `encode` and `decode` through `Shape`, errors as D84 paths.
- [ ] Registry row, `stdlib.nra`, regenerated `stdtext.nra`, module-list goldens; SPEC, README and guide text.

## Acceptance

- [ ] `testdata/toml.nra` parses a document with every value kind and prints `Json.stringify` of the result, then round-trips it through `stringify` and `parse` to the same `Value`.
- [ ] `testdata/toml_errors.nra`: 15 or more invalid documents (duplicate key, table redefined, bad escape, integer out of range, bad date, unterminated string), each printed with `formatError`.
- [ ] `testdata/toml_manifest.nra` reads this repo's `selfhost/narra.toml` form (package, link, `[[binary]]`) into a struct with `Toml.decode` and a derived `Shape`.
- [ ] Error test: `testdata/errors/toml_decode_no_shape.nra` (`Toml.decode<T>` with a `T` that has no `Shape`).
- [ ] The module uses only features of the committed seed.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: about 900 lines of Narra over the JSON parser's pattern; three samples.

## Notes for the worker

- Do not change `manifest.nra` in this node. narrac can import `std/toml` only after the batch's seed update. The move (E1340 for a TOML syntax error in `narra.toml`, E0405, E0406 and E0408 retired as rows that say so, E0409 kept) is a later node the orchestrator adds (proposed name manifest-toml); E1340 stays reserved for it.
- The official `toml-test` suite is not committed; pick cases from it by hand, with a line naming the source.
- Codes E1340-E1349 are reserved; this node itself needs none.
