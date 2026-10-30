# Outcome: maps

## Built

All in `compiler/internal/narra/` unless noted.

- `Map<K, V>` as a prelude type: `MapType{Key, Val}` (`types.go`), C name `M_Map_<hash>` from `mapCName` (`cnames.go`, kind letter `M`, listed in the C name comment table through `Program.Maps`).
- `std_map.go` holds the checker and emitter code: `Map` namespace builtins `Map.size`, `Map.remove(mutable m, k)` (returns `bool`), `Map.keys`, `Map.entries`; `mapOf`, `mapKeyProblem`/`checkMapKey`, `mapIndex`, `mapLit`; emitter `mapFuncs`, `mapGet`, `mapSet`, `mapLitCode`, `mapFor`, `mapToArray`.
- `runtime_map.h` is the C template for one map type (embedded in `std_map.go`, `@...@` placeholders). Layout: entries in insertion order (`keys`, `vals`, `hashes`, `alive`) plus an open-addressing `idx` table. Removal leaves a dead entry; re-insert appends, so the key moves to the end. Entries are compacted or doubled when full. `_unique` copies when `rc > 1` (copy on write). Keys use `e.useEq` and `e.useHash`. Values and keys are retained and released through `retain`/`release`, so boxed enums work.
- Syntax: `Map { "a": 1, [key]: 2 }` (`MapLit`, `parser.go` `mapLit`, `clone.go`). A non-literal key needs brackets. `Map {}` needs an expected type. `Map<K, V>` resolves in `resolveType`.
- `m[k]` is `Option<V>`; `m[k] = v` stores (`IndexExpr.MapSet`, set through `Checker.assigning`). `for (const k of Map.keys(m))` and `Map.entries(m)` iterate directly (`forMapKeys`, `forMapEntries`); the loop holds a reference, so writes to the variable in the body copy and the loop sees the old entries. Outside a loop `Map.keys`/`Map.entries` return `Array<K>` / `Array<#[K, V]>`.
- Key check: float keys, functions, maps, and a key type with a user `implement Equatable` but no `implement Hashable` (or the reverse) are compile errors.
- Map C functions are written after the trait functions (they call the key's eq and hash), in `Emit` (`emit.go`).
- Tests: `testdata/maps.nra` (+ `.out`: word count in insertion order, literals, copy on write, remove and re-insert order, struct keys with enum values, 1000-entry growth and compaction, write inside a loop, array values), `testdata/errors/maps_float_key.nra`, `maps_float_literal.nra`, `maps_partial_traits.nra`, `maps_partial_traits_signature.nra`.

## Deviations

- Files touched outside the brief's `owns` (all `shared`): `types.go`, `cnames.go`, `check.go`, `check_types.go`, `check_call.go`, `check_expr.go`, `check_stmt.go`, `check_closure.go`, `emit.go`, `emit_types.go`, `emit_expr.go`, `emit_stmt.go`, `ast.go`, `clone.go`, `parser.go`. Not in the brief's lists: `modules_check.go` (one line, `maps:` table in `CheckPackage`) and `compiler/README.md` (one bullet under "Differences from the spec").
- `Map.keys` and `Map.entries` are arrays outside a loop, not generators (generators do not exist yet). `Map.remove` returns `bool`; the spec does not say.
- The error text is `<key type> cannot be a map key: float keys are not allowed`; the type is the key, not the whole map.

## Stubs left

- `Map` does not implement `Equatable`, `Hashable`, `Ordered` or `Show`. So `==` on maps, a map as a key, and a struct with a map field using `derives` or `==` are compile errors. The spec says maps implement the standard traits when their contents do; a later node can add `eq`/`show` in `emit_traits.go`.
- No `Map.values`, `Map.has` or `Map.get` (not in the brief).
- Generators: when they land, `Map.keys`/`Map.entries` should return them and the `for ... of` special case can go.

## Gotchas

- Key checks are queued (`checkMapKey`) and run once after all modules are declared and checked (`flushMapKeys`), so trait facts of later-declared types are known. Their errors come after the other checker errors.
- Map functions are emitted after `traitFuncs` and before the function bodies, so any new emitter code that needs a map function prototype must use `mapProtos`.
- A statement call like `Map.remove(...)` returns `bool`; the emitter emits it through `e.temp` so the call runs even when the result is dropped (`ExprStmt` only prints void calls).
- `maps` and `Map.keys` snapshots: a `for` over `Map.keys(m)` holds a reference to the map, so each write to `m` inside the loop copies the whole map once (then `m` is unique again).

## Proposed learnings

- Emit a statement-position call that returns a value through `e.temp`, or the call is dropped | `emit_stmt.go` `ExprStmt` only prints void expressions | compiler/internal/narra | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass. `maps.nra` runs in debug (AddressSanitizer, UBSan, leak check) and release.
- Acceptance: word count with `Map<string, int>` in insertion order (`testdata/maps.nra`): yes. Float key error test (`testdata/errors/maps_float_key.nra`): yes. Fast check: yes.

## Size

- Estimated M, actual M.
