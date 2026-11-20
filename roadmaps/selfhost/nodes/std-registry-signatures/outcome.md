# Outcome: std-registry-signatures

## Built

- `StdFunc` (`selfhost/src/std/builtins.nra`) has two display-only fields: `parameterNames` (one name per entry of `parameters`) and `signature` (the written signature after the name, for a builtin with a check function). Each row in `selfhost/src/std/std<module>.nra` carries its own, so a new row sits beside the fields it needs. The sized integer rows (`stdsized.nra`) write theirs from the row type.
- Every public builtin has them: 109 rows with fixed types get `parameterNames`, and the 84 public check-function rows (and the prelude `copy`) get a `signature`. `narrac query interface std/<module>` prints no `(...)` and no unnamed parameter:
  - `export function push<T>(mutable items: Array<T>, item: T): void uses alloc`
  - `export function keys<K, V>(map: Map<K, V>): Array<K> uses alloc`
  - `export function new<T>(value: T): Box<T> uses alloc`
  - `export function update<K, V>(mutable map: Map<K, V>, key: K, initial: V, change: (V) => V): void uses alloc`
  - `export function element<T, const N: int>(pointer: Ptr<FixedArray<T, N>>, index: int): Ptr<T> uses unsafe`
  - `export function select<M, V>(mask: M, ifTrue: V, ifFalse: V): V` (std/simd: `V` a vector type, `M` its mask, `L` its lane type)
  - `export function from<T>(value: T): i32` (and `tryFrom`, `wrap` for every sized module and std/int)
  - `export function join(first: string, second: string): string uses alloc`
- Builtin types: `StdType { module, name, signature, functions }` in `builtins.nra`; rows `STRING_TYPES` (`StringBuilder`), `POINTER_TYPES` (`Ptr<T>`) and `listSimdTypes()` (the 20 vector types, each with the std/simd functions its check accepts: no `abs` for unsigned lanes, no `any`, `all` or shifts for float lanes). `listStdTypes()` in `stdlib.nra` collects them. Query prints each first as `export opaque struct StringBuilder` (kind `struct`, `location` null) and in JSON adds `"functions": ["builder", "append", "appendChar", "build"]`.
- `InterfaceEntry` (`queryjson.nra`) has `functions: Array<string> = []`; the field is written only when not empty, so package modules print as before.
- Test: case `query/builtin_signatures` (`selfhost/tests/src/querybuiltins.nra`). It takes the public modules from `narrac builtins` (no `std/internal/`), runs `narrac query --json interface` on each, and fails for a builtin function (`location` null) whose signature prints `(...)` or has a parameter without `name: `, and for a builtin type that lists no function or one its module lacks. On the base compiler it fails with `std/int abs: has an unnamed parameter \`int\`` and `std/int from: prints (...)`.
- `testdata/query/std_modules` now also covers std/array, std/map, std/box, std/string (with `StringBuilder`) and `--json` std/ptr.
- LSP: unchanged. Hover shows the checker's type of the expression at the cursor (`(int) => string uses alloc`), not the registry rows, and the server has no signature help. Neither reads the display data.

## Deviations

- The registry rows live in `selfhost/src/std/std*.nra`, not `selfhost/src/check/registry*.nra` as the brief's `shared` list says; the data went beside each row there, as the orchestrator asked.
- The brief's `135` counts more than the public rows: 105 rows have a check function, 21 of them in `std/internal/*`. Internal rows got no display data; `query interface` answers `unknown module` for them. Prelude rows (`print`, `assert`, `panic`, `copy`) got names though no query prints them.
- Parameter names follow the module's Narra part and SPEC calls where they exist: `items` for arrays (as `Arrays.map(items, change)` in std/ext/array and `Arrays.push(mutable items, v)` in SPEC), not the brief's `array`.
- Result types follow the checker, not SPEC prose: `Arrays.values`, `Arrays.entries`, `Maps.keys` and `Maps.entries` return arrays (`Array<T>`, `Array<#[K, V]>`), though SPEC section 2 calls them generators. `Arrays.values` and `Arrays.entries` are still only valid as a `for ... of` source.
- Builtin types print as `opaque struct`; Narra has no declaration form for a type without a body, and none of these exposes fields.

## Stubs left

- None.

## Gotchas

- The display `signature` is free text: the checker does not read it. A change to a check function's arity or result needs the row's `signature` changed by hand; the `builtin_signatures` case checks only names and `(...)`, not types.
- std/simd type parameters (`V`, `M`, `L`) are not bounded in the text; which vector types a function takes is in its check and in `listSimdTypes`.
- A row without `parameterNames` still prints its bare types; only the test catches it.

## Proposed learnings

- A new public std row needs `parameterNames`, or `signature` when it has a check function | query interface and the std reference print them | selfhost/src/std | yes (query/builtin_signatures)

## Verification

- `sh scripts/test.sh --suite query`: PASS, 18 cases (with builtin_signatures).
- `sh scripts/test.sh --suite lsp`: PASS, 25 cases.
- Full `sh scripts/test.sh`: `PASS, 10628 cases`.

## Size

- Estimated M, actual M.
