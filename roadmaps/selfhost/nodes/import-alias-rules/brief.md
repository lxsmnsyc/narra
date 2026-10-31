---
id: import-alias-rules
title: Import aliases never equal type names, and named imports can rename
status: done
spent: 622000
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
  - compiler/testdata/errors/import_*.nra
  - compiler/testdata/diagnostics/import_*.nra
  - roadmaps/selfhost/nodes/import-alias-rules/**
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/astdump.go
  - compiler/internal/narra/modules_check.go
  - compiler/internal/narra/check.go
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/internal/narra/diagnostic_fixes.go
  - compiler/testdata/**
  - selfhost/src/**
  - bench/**
  - compiler/README.md
  - SPEC.md
reserves: []
provides: []
depends_on: []
---

# Import aliases never equal type names, and named imports can rename

## Goal

Implement D72. An import alias that equals a type name visible in the module (prelude, declared or imported) is an error, and so is a second import with the same alias. The fix suggests the plural for the six std modules of capitalized prelude types (`Arrays`, `Maps`, `Boxes`, `Ranges`, `Options`, `Results`). Named imports accept `{ a as b }`. Migrate `compiler/testdata`, `selfhost` and `bench` to the plural aliases.

## Context

- Decision: `DECISIONS.md` D72, debate `roadmaps/debates/qualified-enum-constructors/debate.md`. Spec: section 3, "An import alias may not equal a type name" and "Named imports can take a new name".
- Today `import * as Array from "std/array"` is the convention (D67), and the checker accepts duplicate aliases and `import * as Option from "std/array"` (inbox, 2026-10-30, from #29). This node closes that inbox item.
- `import { push as add } from "std/array"` fails today with `expected "}", found "as"` (`parser.go` `importDecl`). `ImportDecl.Names` becomes a list of name and local name. Port the parser change to `selfhost/src/parser.nra`, `ast.nra` and `astdump.nra` so `TestSelfhostParser` stays byte-identical, and the import check to `selfhost/src/modules.nra` so `TestSelfhostImports` matches.
- Remove D67's special case where `import * as Box` wins over a user type `Box` (`builtins.go` comment near line 66, `check_call.go` `missingImport`). The E0908 fix for a missing import now inserts the plural alias.
- Migration: about 107 imports and 700 calls. Use a script in this node's folder, like `prelude-imports/migrate.py`: rename the alias in the import and every `Alias.` call in the file. Skip a file that declares its own type with the plural name.

## Acceptance

- [x] Error tests: an alias equal to a prelude type, to a declared type, to an imported type; a duplicate alias; a renamed import that clashes with a declaration. Each new format has a code in `diagnosticCodes`.
- [x] A `testdata/diagnostics/` case: `import * as Array from "std/array"` with a call, fixed to `Arrays` in the import and the call, compiles.
- [x] `import { map as mapArray } from "std/array"` and `import { map as mapOption } from "std/option"` in one file run (sample with `.out`).
- [x] `compiler/testdata`, `selfhost` and `bench` use the plural aliases; the self-hosted parser and import harnesses pass.
- [x] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: one parser form in two parsers, two checks with fixes, and a scripted migration.

## Notes for the worker

- Do not change module paths (`std/array` stays). Plural is only the convention: any alias that follows the rules must work.
- Land this before any selfhost-check-* node starts, because the migration rewrites `selfhost/src`.

Shipped in PR #35 (merge bb3e306).
