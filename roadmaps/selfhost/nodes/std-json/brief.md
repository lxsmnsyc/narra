---
id: std-json
title: std/json parse and stringify over a dynamic Value
status: todo
priority: should
type: feature
story: batch-1
early: false
outline: false
milestone: m10-std
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-01
spent:
owns:
  - std/src/json.nra
  - std/src/shape.nra
  - testdata/json*
  - testdata/errors/json_*
  - roadmaps/selfhost/nodes/std-json/**
shared:
  - selfhost/runtime/**
  - selfhost/src/stdtext.nra
  - selfhost/src/modules.nra
  - selfhost/src/modules.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/stdinternal.nra
  - selfhost/src/stdlib.nra
  - SPEC.md
  - compiler/README.md
  - DECISIONS.md
reserves: [decision D83]
provides: []
depends_on:
  - on: std-tables-split
    kind: hard
    why: the float primitive adds a std/internal module, whose rows go in the per-module files the split creates
---

# std/json parse and stringify over a dynamic Value

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

Two std source modules. `std/shape` declares `Value`, a dynamic tree for structured data. `std/json` reads and writes JSON text as a `Value`. Typed `Json.encode`/`Json.decode` and `derives Shape` come later (nodes shape-trait and shape-derive). The language service (`language-service`) uses this module for JSON-RPC.

## Context

- Pattern: `std/src/option.nra` and `nodes/std-source/outcome.md` (registering a source module, `NARRA_UPDATE=1 go test -run TestStdSources`), `nodes/std-foundation/outcome.md` (`std/internal/` primitives).
- Both compilers load these two modules: the language service imports `std/json` from `selfhost/src`, so the Go compiler must build them (README "Tests", decision 2026-11-05 option b). Write them in today's language and put their samples in `testdata/`, not `testdata/stage2/`.
- `std/json` importing `std/shape` is the first std source module that imports another. Sort std source modules with `orderModules` in single-file packages too (`compiler/internal/narra/modules.go`, `selfhost/src/modules.nra`); today they skip it (deferred from std-foundation).

## API

```ts
// std/shape
export enum Value {
  Null
  Bool(value: bool)
  Int(value: int)
  Float(value: f64)
  Text(value: string)
  List(items: Array<Value>)
  Record(fields: Map<string, Value>)
}

// std/json
export function parse(text: string): Result<Value, string> uses alloc
export function stringify(value: Value): string uses alloc
```

- `parse` follows RFC 8259. A number without fraction or exponent that fits `int` is `Value.Int`, every other number is `Value.Float`. Duplicate keys: the last wins, as in `JSON.parse`; `Record` keeps first-insertion order (D64). Errors read `<line>:<col>: <reason>` (`1:7: expected , or }`). Nesting deeper than 512 is an error, not a stack overflow. Strings decode every escape, including surrogate pairs; a lone surrogate is an error.
- `stringify` is compact (no spaces), escapes `"`, `\` and control characters, writes non-ASCII as UTF-8, writes `Record` fields in map order, and writes a non-finite `Float` as `null`, as `JSON.stringify` does. A `Float` with an integral value keeps a fraction or exponent (`1.0`), so it parses back as `Float`.
- Float text both ways goes through a new primitive module `std/internal/float` (`strtod`, and the shortest text that reads back to the same `f64`) unless `Float.toString` already round-trips; it replaces the test-only `std/internal/probe` if that is simpler.

## Acceptance

- [ ] Samples `testdata/json*.nra` with `.out`, clean under ASan and the leak check: round trips of every `Value` form; escapes and surrogate pairs; `Int` versus `Float` (`1`, `1.0`, `1e2`, `9223372036854775808`); float text that reads back exactly (`0.1`, `1e-300`, `5e-324`); a JSON-RPC request picked apart with `match`; error messages with positions for truncated input, a trailing comma, a bad escape, a lone surrogate and nesting over 512.
- [ ] The same samples emit identical C in both compilers (`TestSelfhostEmit`), and the bootstrap passes.
- [ ] SPEC.md "Standard modules" names `std/shape` and `std/json` and shows the API; section 4 "Serialization with `Shape`" says encode and decode are planned. `compiler/README.md` "Supported" lists them.
- [ ] DECISIONS.md D83 records the design below.
- [ ] Fast check passes.

## Sizing

M. Two source modules (about 500 lines of Narra), one small primitive in both compilers, a loader fix; uncertainty low, tests automatic.

## Notes for the worker

- A Narra import alias may not equal a type name (D72): users write `import * as Json from "std/json"` and `import { Value } from "std/shape"`.
- Keep the parser iterative or bounded: deep recursion aborts the program.

## Decision to confirm

Plan (D83): the dynamic tree is `Value` in `std/shape`, a module of its own, and `std/json` has `parse` and `stringify`, the names of TypeScript's `JSON`.

- `Value` sits in `std/shape`, not in `std/json`, because the `Shape` trait (node shape-trait) converts types to and from `Value` without knowing the format. A later TOML or binary format reuses it, and `std/shape` never imports `std/json`, so the two modules do not form an import cycle.
- Separate `Int` and `Float` constructors, because Narra has no single number type, and a 64-bit integer id must survive a round trip, which a `double` cannot hold.
- Rejected: `Json.Value` in `std/json` (the outline's name). It would tie `Shape` to JSON. Rejected: reading JSON straight into typed values without a tree. It is faster but needs `derives Shape` first, and JSON-RPC needs the tree for messages whose type depends on a field.
