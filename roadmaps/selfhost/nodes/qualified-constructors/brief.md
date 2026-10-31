---
id: qualified-constructors
title: Enum constructors are always qualified
status: active
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestRun|TestErrors|TestDiagnostics|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-31
owns:
  - compiler/testdata/errors/ctor_*.nra
  - compiler/testdata/diagnostics/ctor_*.nra
  - roadmaps/selfhost/nodes/qualified-constructors/**
shared:
  - compiler/internal/narra/check.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/check_match.go
  - compiler/internal/narra/check_patterns.go
  - compiler/internal/narra/modules_check.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/internal/narra/diagnostic_fixes.go
  - compiler/testdata/**
  - selfhost/src/**
  - bench/**
  - compiler/README.md
  - SPEC.md
reserves: []
provides: []
depends_on:
  - on: import-alias-rules
    kind: hard
    why: both migrate every file in compiler/testdata and selfhost/src; running after it avoids a rebase of a repo-wide rewrite, and `Shape.Circle` is only unambiguous once an alias cannot be `Shape`
---

# Enum constructors are always qualified

## Goal

Implement D71. A user enum's constructors are written `Enum.Variant`, in expressions and in patterns. A bare constructor is an error whose fix qualifies it. `import { Teacher }` of a constructor is an error whose fix imports its enum. `Some`, `None`, `Ok` and `Err` stay bare. Migrate `compiler/testdata`, `selfhost` and `bench`.

## Context

- Decision: `DECISIONS.md` D71, debate `roadmaps/debates/qualified-enum-constructors/debate.md` (evidence there: the ambiguity and shadowing tests). Spec: section 3, "Constructors come with their enum and are always qualified".
- Today a bare constructor resolves from the expected type or the matched type, and two enums with the same constructor give `constructor Low is ambiguous: A, B all have it; write A.Low`. That error and the expected-type lookup go away. Constructors move out of module scope: two enums may share a name, and a struct `Circle` may sit next to `Shape.Circle`.
- Qualified access also works through a type alias (`export type Token = Text.Token`, then `Token.Word`) and for imported enums (`import { Shape }`, then `Shape.Circle`). Through a namespace import it is `Geo.Shape.Circle`; check that the parser and checker accept the three-part form in expressions and patterns.
- The selfhost parser does not resolve constructors, so `TestSelfhostParser` only needs the migrated sources. selfhost-check-patterns and selfhost-check-calls port the new rule.
- Migration: about 1,100 sites. Apply the new fix with a script in this node's folder, or through `narra check --json` fixes, until every file compiles.

## Acceptance

- [ ] Error tests: a bare constructor in an expression, in a match arm and in an `is` pattern; `import { Teacher }`; a duplicate constructor in one enum. Each new format has a code.
- [ ] `testdata/diagnostics/` cases for the bare constructor and the constructor import; the fixed files compile.
- [ ] Samples run: two enums with the same constructor name and different fields in one module; `Maybe.None` next to the prelude `None`; a struct and a constructor with the same name; qualified access through an alias and a namespace import.
- [ ] `compiler/testdata`, `selfhost` and `bench` are migrated; every selfhost harness passes.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: the lookup moves from scope to the enum, two errors with fixes, and a scripted migration.

## Notes for the worker

- Do not rename prefixed constructors (`Type.TStruct`, `Token.TokIdent`). That is a separate optional change.
- Land this before any selfhost-check-* node starts.
