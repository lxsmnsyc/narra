---
id: literal-type-args
title: Type arguments on map and generic struct literals
status: todo
priority: should
type: feature
outline: false
milestone: m3-core
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestRun|TestErrors|TestASTDump|TestSelfhostParser'
exclusive: []
fulfilled_by:
gate: true
added: 2026-10-30
owns:
  - compiler/testdata/literal_type_args*.nra
  - compiler/testdata/literal_type_args*.out
  - compiler/testdata/errors/literal_type_args_*.nra
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/astdump.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/check_types.go
  - compiler/internal/narra/std_map.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/testdata/astdump/**
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - compiler/README.md
  - SPEC.md
  - DECISIONS.md
reserves: [decision D68]
provides: []
depends_on:
  - on: prelude-imports
    kind: hard
    why: both edit selfhost/src files and SPEC.md; running after it avoids a rebase of a repo-wide migration
---

# Type arguments on map and generic struct literals

## Goal

Allow type arguments on a literal: `Map<string, int> {}`, `Map<string, int> { "a": 1 }` and `Pair<int, string> { a: 1, b: "x" }`. An empty map then needs no variable annotation and can be passed straight to a generic function.

## Context

- User decision 2026-10-30: no new map syntax (no `{K: V}` type, no `#{...}` literal; D23 keeps names over symbols). Add type arguments on the literal instead.
- Today both forms fail to parse: `const m = Map<string, int> {}` gives `expected the end of the statement, found ","`, and so does `Pair<int, string> { ... }`. `const m = Map {}` gives `the type of Map {} is unknown; add a type annotation`. The parser is `mapLit` and the struct literal path in `parser.go`; `Name<` in expression position is currently a comparison, so decide the lookahead rule (for example: `UpperName<` followed by a type list, `>` and `{`) and make sure `a < b` and generic calls still parse.
- Check the type arguments: arity, each a valid type, map key rules (no float keys, Equatable and Hashable), and every key and value against them. With type arguments the literal's type no longer depends on the expected type; when both exist they must agree.
- AST: add the type arguments to `MapLit` and the struct literal node, with cases in `clone.go`, `astdump.go` (update the format doc comment and goldens) and the checker and emitter switches. Port the parser change to `selfhost/src/parser.nra`, `ast.nra` and `astdump.nra` so `TestSelfhostParser` stays byte-identical.
- Empty arrays: `const xs = []` has the same problem. Do not invent an array form here; keep the annotation (`let xs: Array<int> = []`) and say so in the spec text, unless the user asks otherwise in review.
- Spec: describe the form in the literal sections of `SPEC.md` and record D68. Gate: the user approves the spec text.

## Acceptance

- [ ] `Map<K, V> {}`, `Map<K, V> { ... }` and `Pair<A, B> { ... }` parse, check and run (sample with `.out`, debug and release); passing `Map<string, int> {}` to a generic function works without annotation.
- [ ] Error tests: wrong arity, float key type, a value that does not match, type arguments that disagree with the expected type.
- [ ] `a < b`, `f<int>(x)` and existing samples still parse; `TestASTDump` goldens updated; `TestSelfhostParser` passes with the Narra parser change.
- [ ] `SPEC.md`, D68 and `compiler/README.md` describe it.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: a parser lookahead rule in two parsers, one AST field, checker rules and spec text.

## Notes for the worker

- The selfhost checker port (selfhost-check-bodies) will port literal checking later; leave a note in the outcome so it picks up the type arguments.
