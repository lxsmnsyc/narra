---
id: selfhost-check-patterns
title: Self-hosted checker: patterns, match and aggregate literals
status: active
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostPatterns'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - selfhost/src/patterns.nra
  - selfhost/src/match.nra
  - selfhost/src/literals.nra
  - compiler/internal/narra/selfhost_patterns_test.go
  - compiler/testdata/typesdump/patterns/**
shared:
  - selfhost/src/main.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/body.nra
  - selfhost/src/exprs.nra
  - compiler/internal/narra/typesdump.go
  - compiler/internal/narra/selfhost_test.go
  - compiler/testdata/typesdump/**
  - compiler/README.md
  - selfhost/src/stmts.nra
  - selfhost/src/check.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-check-bodies
    kind: hard
    why: defines the body context, locals, per-body tables, the dispatch arms to replace and the dump; no frozen interface exists yet
  - on: selfhost-names-frontend
    kind: hard
    why: renames the files this node edits; building first would conflict on every line it touches
  - on: selfhost-names-checker
    kind: hard
    why: renames the files this node edits; building first would conflict on every line it touches
---

# Self-hosted checker: patterns, match and aggregate literals

## Goal

Port to Narra the checks of `is` and `match`, pattern resolution and bindings, exhaustiveness and unreachable-arm errors, local narrowing, destructuring declarations, and the map, tuple and spread literals. Compare through the `narra types` dump.

## Context

- Read the selfhost-check-bodies outcome. Go source (about 740 lines, 1,000 Narra): `check_match.go` (`resolvePattern`, `resolvePat`, `bindRest`, `irrefutable`, `isExpr`, `match`, `isWild`, `ctorCount`, `ctorArgTypes`, `specialize`, `missingCases`, `ctorText`), `check_patterns.go` (`destructure`, `spreadLit`, `tupleType`/`tupleOf` if still needed, `tupleLit`; in Narra, Go's `tupleOf` and `fieldsOf` are now `findTupleType` and `listFields` after selfhost-names-checker), the narrowing helpers of `check.go` and `check_stmt.go` (`narrowed`, `setNarrow`, `narrowable`, `otherVariant`, the narrowing in `ifStmt` and `diverges`), and `mapLit`, `mapIndex` from `std_map.go`.
- Dump: you own `Pattern` lines (kind as `astdump.go` prints it) with the resolved variant or binding, `Local` lines for pattern bindings, `narrow=<Variant>` on identifiers that are narrowed, and the `IsExpr`, `MatchExpr`, `MapLit`, tuple and spread literal lines. Add them to both printers and the Owners list.
- Skipped until later nodes land: arrow functions in arms or scrutinees (closures), calls to generic functions and constructors in a scrutinee (calls). They show as `unsupported:`.

## Acceptance

- [ ] `TestSelfhostPatterns` compares `narrac types` with `narra types` on the earlier inputs plus `testdata/typesdump/patterns/`: nested and spread patterns, tuple patterns, enum payload patterns, literal and range patterns, shorthand fields, `is` with narrowing in `if`, `&&`, early return, `match` as value and as statement, map and tuple literals with and without type arguments, spread update, destructuring declarations, and the errors (non-exhaustive with the missing-case text, unreachable arm, unknown variant, arity, wrong type, duplicate binding, refutable pattern in a declaration, float map key, map type arguments disagreeing with the expected type).
- [ ] The missing-case text equals Go's (the list order and the `limit`).
- [ ] Earlier harnesses pass; `narra emit` output unchanged.
- [ ] The outcome lists the narrowing representation and the stubs left.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 1,000 Narra lines, a known algorithm (matrix exhaustiveness) already in Go. Uncertainty medium: narrowing as a side table instead of a mutable `Local` field.

## Notes for the worker

- Narrowing: Go writes `Local.narrow`. In Narra keep narrowing in the scope stack (a map from local to variant that is saved and restored with `push`/`pop`), not in the AST. An unrestored narrowing after a branch is the likely bug; test `if (x is A) { ... } else { ... }` and an early return.
- Pattern bindings retain at bind and release at scope pop in the emitter; the checker only declares the locals. Declare them in Go's order so the `Local` lines match.
- From literal-type-args: port `mapLit` with `typeArgs`: resolve with `resolveType`, E0237 against the expected type, the written type is the literal's type; map key rules (no float keys, `Equatable` and `Hashable` together) go through `checkMapKey`.
- From selfhost-check-traits: `checkMapKey` queues; `flushMapKeys` runs once after all bodies.
- `match` arms cannot be `=> return x` in Narra; write blocks. Large `match` on `Pattern` kinds: keep each arm to a call.
- Exhaustiveness uses recursion on pattern depth only; loop over rows and arms.
