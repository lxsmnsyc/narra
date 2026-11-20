# Outcome: std-registry-signatures

## Built

- `StdFunc` (`selfhost/src/std/builtins.nra`) has three display-only fields: `parameterNames` (one name per entry of `parameters`), `signature` (the written signature after the name, for a builtin with a check function) and `note` (a call-site rule, printed as a `//` comment after the signature). Each row in `selfhost/src/std/std<module>.nra` carries its own, so a new row sits beside the fields it needs. The sized integer rows (`stdsized.nra`) write theirs from the row type.
- Every public builtin has them: 109 rows with fixed types get `parameterNames`, and the 84 public check-function rows (and the prelude `copy`) get a `signature`. `narrac query interface std/<module>` prints no `(...)` and no unnamed parameter:
  - `export function push<T>(mutable items: Array<T>, item: T) uses alloc`
  - `export function keys<K, V>(map: Map<K, V>): Array<K> uses alloc`
  - `export function new<T>(value: T): Box<T> uses alloc`
  - `export function update<K, V>(mutable map: Map<K, V>, key: K, initial: V, change: (V) => V) uses alloc`
  - `export function element<T, const N: int>(pointer: Ptr<FixedArray<T, N>>, index: int): Ptr<T> uses unsafe // index: a constant or a name proven to be 0 to N - 1`
  - `export function select<M, V>(mask: M, ifTrue: V, ifFalse: V): V // V: a vector type such as I32x4 (listed above), M: the signed integer vector of the same shape (its mask), L: the lane type of V, N: its lane count` (every std/simd row carries this note; shifts, `shuffle` and `withLane` add their constant-argument rule)
  - `export function from<T: i8 | i16 | i32 | u8 | u16>(value: T): i32`: a type parameter bound by a list of builtin types stands for one of them. `from` lists the types that convert without loss, `tryFrom` and `wrap` every integer type (sized modules and std/int; the lists come from the module rows in `stdsized.nra`).
  - `export function values<T>(items: Array<T>): Array<T> // only as the source of a for ... of loop` (also `entries`)
  - std/test rows end with `// only inside a test block`.
  - `export function join(first: string, second: string): string uses alloc`
- Builtin types: `StdType { module, name, typeParameters, functions }` in `builtins.nra`, held per module in the registry (`StdModule.types`, added by `registerType` after the module's functions). `listStringTypes()` (`StringBuilder`: the std/string rows that take or return it), `listPointerTypes()` (`Ptr<T>`: every std/ptr row) and `listSimdTypes()` (the 20 vector types, each with the std/simd functions its check accepts: no `abs` for unsigned lanes, no `any`, `all` or shifts for float lanes) derive the function lists from the rows. `querystd.nra` prints each first as `export opaque struct <name><typeParameters>` (kind `struct`, `location` null) and in JSON adds `"functions": [...]`.
- `InterfaceEntry` (`queryjson.nra`) has `functions: Array<string> = []`; the field is written only when not empty, so package modules print as before.
- Test: case `query/builtin_signatures` (`selfhost/tests/src/querybuiltins.nra`) checks the rows, not the printed text.
  - `narrac builtins` prints a problem line for each public row with a check function and no `signature`, and each row whose `parameterNames` count differs from `parameters` (`listDisplayProblems` in `stdlib.nra`); it now also prints each row's `signature` and `note`. The case fails on any problem line.
  - Each public row with a `signature` becomes a probe in one generated file: a function whose header is the signature with its type parameters replaced by concrete types (`int`, `string` for `K`, `bool` for `U`, `I32x4`/`i32` in std/simd, `4` for `const N`; a type-list bound gives one probe per listed type) and whose body forwards the parameters to the builtin and returns its result. Notes pick the call form: a `for ... of` loop for `values`/`entries`, a closure in a `test` block for std/test, a constant argument for the rows whose note asks for one. `narrac check` on the file must pass, so a wrong arity, parameter type, result or sized bound fails the case. With four signatures changed in the `narrac builtins` output it fails with `expected int, found string` (push item), `Ptr.offset takes 2 arguments, found 1`, `expected Array<int>, found Array<string>` (Maps.keys result) and `I32.from cannot take u32 without loss`.
- `testdata/query/std_modules` now also covers std/array, std/map, std/box, std/string (with `StringBuilder`) and `--json` std/ptr.
- LSP use is deferred: the brief named LSP hover and signature help, and neither reads the display data. Hover still shows the checker's type of the expression at the cursor (`(int) => string uses alloc`), and the server has no signature help. `roadmaps/selfhost/inbox.md` has a 2026-11-20 line for signature help built on the display data.
- A function without a result prints no result type (`push<T>(...) uses alloc`), as Narra source writes it; `void` is not a type name a program can write, so `: void` made the display invalid Narra. This also applies to fixed rows whose `result` is `"void"`.

## Deviations

- The registry rows live in `selfhost/src/std/std*.nra`, not `selfhost/src/check/registry*.nra` as the brief's `shared` list says; the data went beside each row there, as the orchestrator asked.
- The brief's `135` counts more than the public rows: 105 rows have a check function, 21 of them in `std/internal/*`. Internal rows got no display data; `query interface` answers `unknown module` for them. Prelude rows (`print`, `assert`, `panic`, `copy`) got names though no query prints them.
- Parameter names follow the module's Narra part and SPEC calls where they exist: `items` for arrays (as `Arrays.map(items, change)` in std/ext/array and `Arrays.push(mutable items, v)` in SPEC), not the brief's `array`.
- Result types follow the checker, not SPEC prose: `Arrays.values`, `Arrays.entries`, `Maps.keys` and `Maps.entries` return arrays (`Array<T>`, `Array<#[K, V]>`), though SPEC section 2 calls them generators. `Arrays.values` and `Arrays.entries` carry the note `only as the source of a for ... of loop`.
- The type-list bound (`T: i8 | u8`) is display text only; Narra has no union types or bounds of this form.
- Builtin types print as `opaque struct`; Narra has no declaration form for a type without a body, and none of these exposes fields.

## Stubs left

- None.

## Gotchas

- The display `signature` is free text: the checker does not read it. A change to a check function's arity, types or result needs the row's `signature` changed by hand; `builtin_signatures` then fails until it matches.
- A new note that states a call-site rule needs a matching call form in `querybuiltins.nra` (`CONSTANT_ARGUMENTS`, or the note text that picks the loop or test-block probe).
- `Test.panics` and `Test.panicsWith` treat a function parameter passed as `body` as escaping (a probe function would need `escaping body`); the display does not say so, and the closure probe in a `test` block does not test it.
- The probe tries one concrete instance per unbounded type parameter, so it proves the shape of a signature, not every type the check accepts.

## Proposed learnings

- A new public std row needs `parameterNames`, or `signature` when it has a check function | query interface and the std reference print them | selfhost/src/std | yes (query/builtin_signatures)

## Verification

- `sh scripts/test.sh --suite query`: PASS, 18 cases (with builtin_signatures).
- `sh scripts/test.sh --suite lsp`: PASS, 25 cases.
- `sh scripts/test.sh --suite sources`: PASS, 1 case.
- Full `sh scripts/test.sh`: `PASS, 10790 cases in 237s` (after the review fixes).

## Size

- Estimated M, actual M.
