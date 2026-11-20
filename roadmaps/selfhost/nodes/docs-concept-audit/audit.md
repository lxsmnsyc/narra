# Audit: concepts the docs miss

For the maintainer to approve or change. Audited on `feat/batch-21` at `501f19f9` (2026-11-20). Nothing in the docs is changed by this node. `docs-structure-review/proposal.md` covered pages, order and generated references. This audit covers content.

## Decision needed

1. Approve the inventory ranks below, mainly the 11 crucial gaps and the 3 corrections.
2. Approve the 16 new writing nodes in "Proposed writing nodes" for batch-21, crucial first.
3. Approve that 6 of the 9 outline nodes join batch-21 (`site-reference-section`, `std-interface-query`, `std-reference`, `cli-reference`, `manifest-reference`, `error-index`), that `guide-concurrency` is replaced by `guide-async` and `guide-threads`, and that `std-doc-comments` and `spec-decision-links` stay after the release.
4. Approve the chapter order changes at the end.

## Summary

- 185 concepts. Every subsection of `SPEC.md` sections 1 to 15 has at least one row. Section 16 (Deferred) is left out.
- By rank: 51 crucial, 107 important, 27 niche.
- By status: 92 taught, 25 mentioned, 68 missing.

| Rank | Taught | Mentioned | Missing |
| --- | --- | --- | --- |
| crucial | 40 | 3 | 8 |
| important | 51 | 17 | 39 |
| niche | 1 | 5 | 21 |

- Crucial concepts not taught: 11 (3 mentioned, 8 missing). Async and C interop are both missing: the guide has no sample with `async`, `await`, `extern`, `unsafe` or `Ptr`.
- The guide states 2 rules wrong and 3 pages give an incomplete prelude, all checked with `narrac` built from `seed/narrac.c`. They are listed under crucial gaps as corrections.

## Method

- Sources read: `SPEC.md` in full, section by section; every guide chapter (`docs/guide/README.md`, `01` to `13`) and `docs/install.md` in full; `skills/narra/SKILL.md`; the "Supported", "Not supported yet" and "Differences from the spec" sections of `compiler/README.md`; the importable module list that `narrac check` prints for an unknown module (54 modules); `narrac builtins` for `std/fs`, `std/io`, `std/env`, `std/path` and `std/process`; `narrac --help`; the `testdata/` file names grouped by prefix; `selfhost/src/driver/diagnostics.nra` (479 codes).
- `DECISIONS.md` was used through the D numbers `SPEC.md` cites. It holds reasons for rules, not rules of its own, so it adds no rows.
- The site home page (`website/src/routes/index.brm`) shows claims with short snippets, not lessons, so it teaches nothing in the sense of this audit.
- Taught: the guide explains the concept and has a runnable sample that uses it. Mentioned: the guide names it, or uses it without explaining it, or teaches it only in chapter 13 (optional, written for TypeScript readers). Missing: no chapter or `docs/install.md` uses it in a sample or says what it does. The guide index (README) may still list it among the features that are built but not taught.
- Ranks follow the brief. Crucial: a first real program hits it, the compiler rejects code without it, or it sets Narra apart. Important: a larger program needs it. Niche: rare, or a detail a reference page can hold.
- Spot checks, all with the seed build: tuple index `pair[0]` compiles and prints; a local `const` array passed as `mutable` compiles; `total / Arrays.length(items)` fails with E0979; a `Cell` counter, `Fs.writeText` and `Fs.readText`, `Json.decode<User>` and `Json.encode`, an `async` entry with `Async.all` and `AsyncTime.sleep`, `Thread.scope`, and an `extern { function cbrt(...) } from "m"` call in `unsafe` inside a `@trusted` function all compile and run.

## Inventory

Ch is a guide chapter (`docs/guide/NN-*.md`). "ch13 only" means the concept appears only in the optional TypeScript chapter.

| # | Concept | Source | Rank | Where the docs teach it | Status |
| --- | --- | --- | --- | --- | --- |
| 1 | Entry function `run`, exit code | SPEC 1, 15 | crucial | ch1 | taught |
| 2 | Compiles to C, needs a C compiler | SPEC 1, 11 | important | ch1, install | taught |
| 3 | Install, `PATH`, update, build from source | install.md | crucial | install, ch1 | taught |
| 4 | No semicolons, line break ends a statement | SPEC 2 Statements | crucial | ch1, ch12 | taught |
| 5 | Line continuation rule (operator, `,`, `(` at line end; `?` line starts a statement) | SPEC 2 Statements | important | none | missing |
| 6 | Comments `//`, `/* */`, `///` | SPEC 2 Statements | important | ch1 (`//` only) | mentioned |
| 7 | `const` and `let`, `const` fixes the name not the contents | SPEC 2 Bindings, D81 | crucial | ch2, ch9 | taught |
| 8 | Every binding holds its own value; pattern and loop bindings are read-only copies | SPEC 2 Bindings | crucial | ch2, ch9 | taught |
| 9 | Compound assignment, `x++`, statements without value | SPEC 2 Bindings, D76 | important | ch2 | taught |
| 10 | String `+=` joins strings | SPEC 2 Bindings | important | none | missing |
| 11 | No shadowing | SPEC 2 Bindings | crucial | ch2 | taught |
| 12 | Unused names are errors, `_` prefix | SPEC 2 Bindings | crucial | ch3 | taught |
| 13 | A `let` never reassigned is an error (E1510) | SPEC 2 Bindings | crucial | ch2 | taught |
| 14 | Unreachable statements (E1500) | SPEC 2 Bindings | important | ch3 | taught |
| 15 | Struct and tuple destructuring | SPEC 2 Bindings | important | ch5 | taught |
| 16 | Array and FixedArray destructuring (`[...rest]`, E0715) | SPEC 2 Bindings, D103 | niche | none | missing |
| 17 | Function declarations, return on every path (E0969) | SPEC 2 Functions | crucial | ch3 | taught |
| 18 | No function inside a function (E0930) | SPEC 2 Functions, D74 | important | ch3 | taught |
| 19 | Arrow functions and function types | SPEC 2 Functions | crucial | ch3 | taught |
| 20 | Std functions as values, E1013 for generic ones | SPEC 2 Functions | important | none | missing |
| 21 | Named arguments | SPEC 2 Functions, D87 | important | ch3 (to skip a default) | mentioned |
| 22 | Parameter defaults (D115) | SPEC 2 Functions | important | ch3 | taught |
| 23 | Arguments evaluated left to right | SPEC 2 Functions | niche | none | missing |
| 24 | No method calls, varargs, overloading, currying | SPEC 2 Functions, Literals | crucial | ch2, ch13 | taught |
| 25 | Recursion and tail calls compiled to loops | SPEC 2 Functions | important | ch3 | taught |
| 26 | Closure capture: by value, or by reference when non-escaping (D77) | SPEC 2 Closures | crucial | ch3, ch9 | taught |
| 27 | `escaping` parameters | SPEC 2 Closures | crucial | ch3 | taught |
| 28 | A variable a closure assigns is borrowed for the call (E0614) | SPEC 2 Closures | important | ch3 | taught |
| 29 | `Cell` for state an escaping closure changes (D90) | SPEC 2 Closures, 8 Cell | crucial | ch11 (one line, in an error message) | mentioned |
| 30 | `if`, `while`, `for ... of`, `break`, `continue`, no truthiness | SPEC 2 Control flow | crucial | ch2 | taught |
| 31 | `for (mutable item of items)` (D85) | SPEC 2 Control flow | important | ch8, ch9 | taught |
| 32 | `Ranges.upTo`, `between`, `stepped` | SPEC 2 Control flow | important | ch2 | taught |
| 33 | C-style `for (init; cond; step)` | SPEC 2 Control flow, D76 | important | ch13 only | mentioned |
| 34 | Labeled `break` and `continue` | SPEC 2 Control flow | niche | none | missing |
| 35 | Block expressions, `if` as a value, `? :` | SPEC 2 Block expressions, D75 | important | ch4, ch5 | taught |
| 36 | A `? :` over several lines needs parentheses (E0934) | SPEC 2 Block expressions | niche | none | missing |
| 37 | Template strings and escapes | SPEC 2 Literals | crucial | ch2, ch11 | taught |
| 38 | Tagged templates, `Template.raw` (D114) | SPEC 2 Literals | niche | none | missing |
| 39 | `char` literals | SPEC 2 Literals | important | ch2 | taught |
| 40 | Integer literal forms (`0x`, `0o`, `0b`, `_`), floats, literal must fit | SPEC 2 Literals | important | ch2 (`0o` missing) | taught |
| 41 | Map literals, type arguments on literals, `[key]: value` keys | SPEC 2 Literals | important | ch2 (computed keys missing) | taught |
| 42 | Tuples, including `pair[0]` | SPEC 2 Literals, 4 Composite | important | ch2 (says tuples cannot be indexed) | taught |
| 43 | Operands of one type, no implicit conversion | SPEC 2 Operators | crucial | ch2 | taught |
| 44 | Bitwise operators | SPEC 2 Operators | niche | none | missing |
| 45 | No power or length operator: `Int.pow`, `Float.pow`, `Arrays.length` | SPEC 2 Operators | important | ch2 (length only) | mentioned |
| 46 | `Int.floorDiv`, `Int.divmod`, `Int.compare` | SPEC 2 Operators | niche | none | missing |
| 47 | TypeScript forms that are errors with a fix | SPEC 2 Differences | important | ch13 | taught |
| 48 | One file is one module, package-rooted paths | SPEC 3 | crucial | ch8 | taught |
| 49 | `export`, private by default | SPEC 3 | crucial | ch8 | taught |
| 50 | Prelude contents (`copy`, `Dynamic`, `Async`, traits included) | SPEC 3 | important | ch1, ch8 (both omit `copy`) | mentioned |
| 51 | Std aliases are plural, alias cannot be a type name | SPEC 3, D72 | important | ch8 | taught |
| 52 | Named imports and `as` | SPEC 3 | important | ch8 | taught |
| 53 | No re-export; `export const f = M.f`, `export type T = M.T` | SPEC 3 | important | none | missing |
| 54 | No import cycles | SPEC 3 | important | ch8 | taught |
| 55 | Constructors are qualified (D71) | SPEC 3 | crucial | ch4, ch8 | taught |
| 56 | No top-level `let`, no import side effects | SPEC 3 | important | ch8 | taught |
| 57 | `narrac move` | SPEC 3 | niche | none | missing |
| 58 | Catalog of std modules (54 importable) | SPEC 3 | crucial | ch8 (lists 14) | mentioned |
| 59 | Files: `std/fs` | SPEC 3, 4 Bytes | crucial | ch8 (name only) | missing |
| 60 | Standard input and output: `std/io` | SPEC 3 | crucial | none | missing |
| 61 | Arguments and environment: `std/env` | SPEC 3 | crucial | none | missing |
| 62 | Paths: `std/path` | SPEC 3 | important | none | missing |
| 63 | `Process.run` and `Process.exit` | SPEC 3 | important | ch8 (`exit` only) | mentioned |
| 64 | Windows behavior of `std/path`, `std/process`, `std/test`, `std/async` | SPEC 3 | niche | none | missing |
| 65 | Narra parts of builtin modules (`std/ext`) | SPEC 3 | niche | ch8 (not importable) | mentioned |
| 66 | `std/math` | SPEC 3 | important | none | missing |
| 67 | `std/time` | SPEC 3 | important | none | missing |
| 68 | `std/random` | SPEC 3 | important | none | missing |
| 69 | `std/hash`, `Hash.of` | SPEC 3 | niche | none | missing |
| 70 | `std/iterator` | SPEC 3, D118 | important | ch11 | taught |
| 71 | `std/arena` | SPEC 3, 6 Cycles | important | ch11 | taught |
| 72 | `std/unicode` | SPEC 3 | important | ch11 | taught |
| 73 | `std/args` | SPEC 3 | important | ch11 | taught |
| 74 | `std/toml` | SPEC 3 | important | ch11 | taught |
| 75 | `std/regex`, `Regexes.checked` at compile time | SPEC 3 | important | ch11 (`checked` missing) | taught |
| 76 | `std/terminal` | SPEC 3 | important | ch11 | taught |
| 77 | `std/internal` is not importable (E0474) | SPEC 3, 15 | niche | ch8 | taught |
| 78 | `std/string` functions (split, slice, trim, indexOf, join, replace) | SPEC 4 Strings | crucial | ch2, ch6, ch11 | taught |
| 79 | `std/array` and `std/map` functions beyond push and keys | SPEC 3, 4 | important | ch2 (a few) | mentioned |
| 80 | Primitive types table | SPEC 4 Primitive | important | ch2 (no sized types) | mentioned |
| 81 | Sized integers and conversions (`I32.from`, `tryFrom`, `wrap`, E0295) | SPEC 4 Numbers, D82 | important | none | missing |
| 82 | `f32` | SPEC 4 Numbers | niche | none | missing |
| 83 | `FixedArray<T, N>` and proven indexes (E0994) | SPEC 4 Composite | important | none | missing |
| 84 | Vector types `std/simd` (D112) | SPEC 4 Vector types | niche | none | missing |
| 85 | Struct fields, defaults, literals, shorthand, inferred type name | SPEC 4 Structs | crucial | ch4 | taught |
| 86 | Spread update `{ ...s, f: v }` | SPEC 4 Structs | crucial | ch4 | taught |
| 87 | Generic structs and enums, inferred type arguments | SPEC 4 Structs, Generics | important | none | missing |
| 88 | Struct field patterns in `match` | SPEC 4 Structs | important | ch5 | taught |
| 89 | A struct cannot contain itself; `Box` | SPEC 4 Structs | important | ch4 | taught |
| 90 | Opaque structs | SPEC 4 Opaque | important | ch4 | taught |
| 91 | Type aliases, generic aliases | SPEC 4 Type aliases | important | none | missing |
| 92 | `typeof` in types (D96) | SPEC 4 typeof | niche | none | missing |
| 93 | Distinct types, `from` and `toInt` | SPEC 4 Distinct | important | ch4 | taught |
| 94 | Trait lists on distinct types, `opaque distinct` | SPEC 4 Distinct | niche | none | missing |
| 95 | Refinement types (`requires` on a distinct type, `T.check`, D116) | SPEC 4 Refinements | important | none | missing |
| 96 | `Option`, `is` narrowing, `.value` | SPEC 4 Optional values | crucial | ch5 | taught |
| 97 | `std/option` helpers | SPEC 4 Optional values | important | ch6 | taught |
| 98 | `match`, guards, literal and tuple patterns, exhaustiveness | SPEC 4 Pattern matching | crucial | ch4, ch5 | taught |
| 99 | Pattern alternatives with `\|` (D98) | SPEC 4 Pattern matching | important | ch5 | taught |
| 100 | Array patterns `[first, ...rest]` (D103) | SPEC 4 Pattern matching | important | none | missing |
| 101 | Unreachable arms (E0983) | SPEC 4 Pattern matching | niche | none | missing |
| 102 | Generic functions, constraints, type arguments at the call | SPEC 4 Generics | crucial | ch3, ch7 | taught |
| 103 | Constraints on struct and enum type parameters | SPEC 4 Generics | niche | none | missing |
| 104 | Constant type parameters `const N: int` | SPEC 4 Generics | niche | none | missing |
| 105 | Traits: declare, `implements`, `implement`, call through the trait | SPEC 4 Traits | crucial | ch7 | taught |
| 106 | Only the declaring module implements (orphan rule) | SPEC 4 Traits | important | ch7 | taught |
| 107 | `implement Show for Pair<T>` on a generic type | SPEC 4 Traits | important | none | missing |
| 108 | Trait functions with `Self` in the result (`Shape.fromValue<User>`) | SPEC 4 Traits | niche | none | missing |
| 109 | `Dynamic<Trait>` values | SPEC 4 Traits | important | README (name only, as built but not taught) | missing |
| 110 | `Show`, `Ordered`, `Equatable`, `derives` | SPEC 4 Traits | crucial | ch7 | taught |
| 111 | `Hashable` and map keys | SPEC 4 Traits | important | ch7 (keys compare by contents) | mentioned |
| 112 | `Shape`, `derives Shape`, `std/json`, `std/shape` (D83, D84) | SPEC 3, 4 Traits | important | ch11 (used for TOML, not explained) | mentioned |
| 113 | Inference: named functions write their types | SPEC 4 Inference | important | ch3 | taught |
| 114 | Overflow wraps, checked arithmetic (D92) | SPEC 4 Numbers | important | ch2 | taught |
| 115 | Divisors must be proven (E0979) | SPEC 4 Numbers, D92 | crucial | ch6 (one sentence) | mentioned |
| 116 | Shift counts must be proven (E0980) | SPEC 4 Numbers | niche | none | missing |
| 117 | What a check proves: narrowing of integer names | SPEC 4 Numbers | important | ch3 (with `requires`) | mentioned |
| 118 | Float text form | SPEC 4 Numbers | niche | ch2 (output only) | mentioned |
| 119 | UTF-8 strings, byte offsets, `Str.slice` returns `Option`, `charCount` | SPEC 4 Strings | crucial | ch2 | taught |
| 120 | String builder (`Str.builder`, `Str.append`, `Str.build`) | SPEC 2 Literals, 6 | important | ch9 (type name only) | mentioned |
| 121 | `bytes` and `std/bytes` | SPEC 4 Bytes | important | none | missing |
| 122 | `Result`, `Ok`, `Err`, early return and narrowing | SPEC 5 | crucial | ch6 | taught |
| 123 | `std/result` helpers | SPEC 5 | important | ch6 | taught |
| 124 | An unused `Result` is an error | SPEC 5 | crucial | ch6 | taught |
| 125 | `panic`, `assert`, exit code 101 | SPEC 5 | important | ch4, ch6 | taught |
| 126 | Std functions return `Option` instead of panicking | SPEC 5, D92 | important | ch2 | taught |
| 127 | `Maps.update` for counters (`counts[key] += 1` is an error) | SPEC 2 Bindings, 5 | crucial | none | missing |
| 128 | Values, owners, heap and inline storage, reference counts | SPEC 6 Values | crucial | ch9 | taught |
| 129 | In-place updates and reuse | SPEC 6 In-place | important | ch9 | taught |
| 130 | `copy(x)` and what it takes (E0975) | SPEC 6 Copies, D94 | crucial | ch2, ch4, ch9 | taught |
| 131 | A named rest pattern copies | SPEC 6 Copies | niche | none | missing |
| 132 | Debug copy check `NARRA_CHECK_COPIES` | SPEC 6 Copies, D80 | important | none | missing |
| 133 | Renames are errors (E0971, D124) | SPEC 6 Renames | crucial | ch9 | taught |
| 134 | Shared writes (E0978) | SPEC 6 Shared writes | crucial | ch9 | taught |
| 135 | `from` on results, in arrow functions and function types (E0976, E0977) | SPEC 6 Shared results | crucial | ch7, ch9 (function types missing) | taught |
| 136 | `mutable` parameters and their rules | SPEC 6 Mutable parameters | crucial | ch9 | taught |
| 137 | Enums and tuples are replaced, not changed (a struct in an `Option` field) | SPEC 6 Mutable parameters | important | none | missing |
| 138 | Cycles: identity types leak, break them, use an arena | SPEC 6 Cycles, D80 | important | ch9, ch11 (arena) | mentioned |
| 139 | Leak check `NARRA_CHECK_LEAKS` | SPEC 6, 13 | important | ch1, ch9, ch10 | taught |
| 140 | `Sendable`, values across threads | SPEC 6 Threads | important | none | missing |
| 141 | `uses alloc` and `io`, exported functions declare, private infer | SPEC 7 | crucial | ch10 | taught |
| 142 | The `unsafe` effect | SPEC 7, 9 | important | none | missing |
| 143 | Function types carry effects; a top-level function constant writes its type | SPEC 7, D86 | important | ch10 (constant missing) | taught |
| 144 | Generator types carry effects (D78) | SPEC 7 | important | ch11 | mentioned |
| 145 | Effect variables | SPEC 7 Effect variables | important | ch10 | taught |
| 146 | Generators: `function*`, `yield`, `yield*`, `Generator<T>` | SPEC 8 Generators | important | ch11 (used, not explained) | mentioned |
| 147 | Coroutines, `Co.resume`, `Co.send`, `Step` (D79) | SPEC 8 Generators | niche | ch11 (`resume` only) | mentioned |
| 148 | Scoped threads: `Thread.scope`, `start`, `join` | SPEC 8 Threads | important | README (name only, as built but not taught) | missing |
| 149 | `Parallel.map` | SPEC 8 Threads | important | none | missing |
| 150 | Channels | SPEC 8 Channels | important | README (name only, as built but not taught) | missing |
| 151 | `Shared<T>`, `Atomic<T>`, lock rules, self-lock stop (D111) | SPEC 8 Shared state | important | none | missing |
| 152 | Async functions, `await`, `Async<T>`, unused handle (E1103), `Async.detach` | SPEC 8 Async, D93, D95, D109 | crucial | README (name only, as built but not taught) | missing |
| 153 | `async` arrows and `async { }` blocks | SPEC 8 Async | important | none | missing |
| 154 | `Async.then`, `all`, `race`, `peek`, `idle`, `Async<void>` | SPEC 8 Async | important | none | missing |
| 155 | The event loop, program exit, cooperative cancel, `AsyncError` | SPEC 8 Async | important | none | missing |
| 156 | Async entry `export async function run` | SPEC 8, 15 | crucial | none | missing |
| 157 | `std/async/time`, `std/async/file`, `std/async/net` | SPEC 8 Async | important | none | missing |
| 158 | Async generators, `for await`, `Async.next` (D110) | SPEC 8 Async generators | important | none | missing |
| 159 | `unsafe { }` blocks and `@trusted` (E0996, E0997) | SPEC 9 | crucial | ch13 (the word, in the Rust comparison) | missing |
| 160 | `Ptr<T>` and `std/ptr` | SPEC 9 | important | none | missing |
| 161 | `extern { } from "lib"`, C type mapping, `@extern`, linking | SPEC 9 | crucial | README (name only, as built but not taught) | missing |
| 162 | `std/cstring` | SPEC 9 | important | none | missing |
| 163 | `narrac audit` | SPEC 9 | important | none | missing |
| 164 | `@repr`, C-callable functions (not built yet) | SPEC 9 | niche | README (planned list) | mentioned |
| 165 | Top-level constant values | SPEC 10 | important | ch8, ch9 | taught |
| 166 | `@const` functions and their limits (D105) | SPEC 10 | important | README (name only, as built but not taught) | missing |
| 167 | Code generation cost model (monomorphization, no boxing, `mutable` as pointers, closure environments) | SPEC 11 | niche | ch3 (escaping cost) | mentioned |
| 168 | Build conditions `Build.os`, `arch`, `debug`, `@target`, `--all-targets` | SPEC 12 Build conditions | important | ch12 (one sentence) | mentioned |
| 169 | Targets, `--target`, cross `CC`, `narrac targets`, WASI, `narrac emit`, `-emit-c` | SPEC 12 | important | README (name only, as built but not taught) | missing |
| 170 | `test` blocks, `std/test`, `narrac test` flags, timeout | SPEC 13 | crucial | ch10, ch12 | taught |
| 171 | `Test.golden` and `NARRA_UPDATE_GOLDEN` | SPEC 13 | important | none | missing |
| 172 | `requires` contracts (D95, D102) | SPEC 13 | important | ch3 (relations and `\|\|` clauses missing) | taught |
| 173 | `ensures` lines | SPEC 13 | important | README (name only, as built but not taught) | missing |
| 174 | `check --json`, `explain`, `apply-fix` | SPEC 14 Diagnostics | crucial | ch12 | taught |
| 175 | Index of error codes (479 codes, 96 explained) | SPEC 14, diagnostics.nra | important | none | missing |
| 176 | `narrac query` | SPEC 14 Queries | important | ch12 | taught |
| 177 | `narrac lsp` and editor setup | SPEC 14 Language server | important | ch12 | taught |
| 178 | `narrac fmt`, `[fmt]` | SPEC 2, 14 | important | ch12 | taught |
| 179 | Package layout, `[package]`, `[[binary]]` | SPEC 15 | crucial | ch1, ch8 | taught |
| 180 | Git dependencies, version selection, `narra.lock`, cache trust | SPEC 15 | important | ch8 | taught |
| 181 | `[exports]` and `[overrides]` | SPEC 15 | important | ch8 | taught |
| 182 | `[link] search` | SPEC 15 | important | none | missing |
| 183 | Workspaces and member commands | SPEC 15 | important | ch8 | taught |
| 184 | Manifest and lock error codes (E0430 to E0490) | SPEC 15 | niche | none | missing |
| 185 | Every `narrac` command and flag in one place | `narrac --help`, compiler/README.md | important | ch1, ch12 (tables miss `build`, `run`, `emit`, `move`, `audit`, `targets`) | mentioned |

## Crucial gaps

The 11 crucial rows that are not taught, in 9 items, then 3 corrections. Each gives why a user needs it, with evidence.

1. **Async functions and the async entry** (rows 152, 156). Missing. The maintainer named async as missing. SPEC section 8 spends 120 lines on it, `testdata/` has 45 `async_*` programs, and a call of an async function without `await` is a compile error (E1103) that no guide page explains.
2. **C interop: `extern`, `unsafe`, `@trusted`** (rows 159, 161). Missing. The maintainer named FFI as missing. SPEC section 9 is the only description, and every call of an extern member outside an `unsafe` block is E0996.
3. **Files: `std/fs`** (row 59). Missing. A first real tool reads or writes a file. No guide sample calls `Fs.readText`, and the only mention is the module name in the chapter 8 list.
4. **Standard input and output: `std/io`** (row 60). Missing. `narrac run -stdin FILE` is taught in chapter 12, but no guide program reads standard input, and `Io.readLine` and `Io.readAll` appear on no page.
5. **Arguments and environment: `std/env`** (row 61). Missing. `Env.args()` is the only way to read non-ASCII arguments correctly on Windows (SPEC 3), and `Env.get` is how a program reads `HOME`. The SPEC 12 sample relies on both.
6. **Catalog of std modules** (row 58). Mentioned. Chapter 8 lists 14 modules. The compiler's unknown-module error lists 54, so a learner has no way to find `std/json`, `std/path` or `std/time`.
7. **Divisors must be proven** (row 115). Mentioned in one sentence of chapter 6. `return total / Arrays.length(items)`, the average every learner writes, fails with E0979. Chapter 2 divides by a constant without saying why that one compiles.
8. **`Maps.update` for counters** (row 127). Missing. `counts[key] += 1` is a compile error (SPEC 2 Bindings, D92), and counting words is a first program. The chapter 5 till works around it with `basket[name] is Some(count) ? count : 0`.
9. **`Cell` for changing state in an escaping closure** (row 29). Mentioned. The E0612 message tells the user to keep the state in a `Cell` from `std/cell` (the chapter 11 error sample prints it), yet no page shows one, and chapter 13 answers the counter question without it.

Corrections. These rows count as taught, but the guide states the rule wrong:

- **Tuples can be indexed** (row 42). Chapter 2 says "That is the only way to read a tuple: it has no positions to index", but `pair[0]` is valid (SPEC 4 Composite types) and compiles with the seed build.
- **A `const` can be passed as `mutable`** (row 7). Chapter 2, "Lists of values", says "Only a `let` name can be passed that way". SPEC 2 Bindings allows a `const`, the seed build compiles it, and chapter 2 itself does it two sections later.
- **Prelude contents** (row 50). Chapter 1 says "Only `print`, `assert` and `panic` are there without an import", chapter 8 repeats it, and `skills/narra/SKILL.md` says the same. `copy` is a prelude function too (SPEC 3), and every guide chapter from 2 on calls it.

## Proposed writing nodes

Crucial first. "Ch" is the current file name; new chapters get numbers when the order below is approved.

| id | Chapter or page | Concepts covered (rows) | Rank | Size |
| --- | --- | --- | --- | --- |
| `guide-corrections` | ch1, ch2, ch8, `skills/narra/SKILL.md` | Fix the tuple index rule (42) and the `const` as `mutable` rule (7); add `copy` to the prelude lists (50); list every std module in ch8 or link the std reference (58) | crucial | S |
| `guide-basics-gaps` | ch2 | Divisor proofs and what a check proves (115, 117); `Maps.update` (127); string `+=` and the string builder (10, 120); C-style `for` and labels (33, 34); line continuation and multi-line `? :` (5, 36); `Int.pow`, `floorDiv`, `divmod` (45, 46); more `Arrays` and `Maps` functions (79) | crucial | M |
| `guide-system-io` | new chapter after ch10: "Files and the system" | `std/fs` (59), `std/io` (60), `std/env` (61), `std/path` (62), `Process.run` and `exit` (63), `bytes` and `std/bytes` (121), `std/time` (67) | crucial | M |
| `guide-cell-and-state` | ch9 ("Closures and `let`"), ch13 counter | `Cell` (29); enums and tuples are replaced, not changed (137); identity-type cycles and how to break them (138); `NARRA_CHECK_COPIES` (132) | crucial | M |
| `guide-async` | new chapter: "Async" | async functions, `await`, `Async<T>`, E1103, `detach` (152); async arrows and blocks (153); `then`, `all`, `race`, `peek`, `idle` (154); loop, exit, cancel, `AsyncError` (155); async entry (156); `std/async/*` (157); async generators and `for await` (158) | crucial | L |
| `guide-c-interop` | new chapter: "Unsafe code and C" | `unsafe` effect and blocks, `@trusted` (142, 159); `Ptr<T>` (160); `extern` blocks, C types, `@extern`, linking, variadic limit (161); `[link] search` (182); `std/cstring` (162); `narrac audit` (163); `@target` on an extern block; `@repr` as planned (164) | crucial | M |
| `std-interface-query` (outline) | `narrac query interface std/<module>` | Signatures of every public std module, builtin and source (58, 66 to 69, 79) | crucial (for 58) | M |
| `site-reference-section` (outline) | `/reference` index, header link, routes, search | Home for rows 58, 175, 182, 184, 185 | crucial (for 58) | S |
| `std-reference` (outline) | `/reference/std`, generated | Every std export with its signature and effects (58, 79) | crucial (for 58) | M |
| `guide-generators` | new chapter before ch11: "Generators" | `function*`, `yield`, `yield*`, generator effects and `escaping` parameters (144, 146); coroutines (147); array and map loops as generators | important | M |
| `guide-threads` | new chapter after `guide-async`: "Threads and shared state" | `Thread.scope` (148); `Parallel.map` (149); channels (150); `Shared`, `Atomic`, locks (151); `Sendable` (140) | important | M |
| `guide-types-more` | ch4, ch7 | Generic structs and enums, their constraints (87, 103); generic `implement` (107); type aliases and re-exporting (53, 91); `typeof` (92); `Dynamic<Trait>` (109); `Hashable` (111); `Self` in a trait result (108); distinct trait lists (94) | important | M |
| `guide-serialization` | ch7, new section "Saving and loading values" | `Shape`, `derives Shape`, `std/shape`, `std/json` (112); needed before the TOML section of ch11 | important | S |
| `guide-numbers` | ch2 or a new section after it | Sized integers and conversions (81); `f32` (82); bitwise operators and shift proofs (44, 116); float text (118); `std/math`, `std/random`, `std/hash` (66, 68, 69) | important | M |
| `guide-patterns-more` | ch5 | Array patterns and array destructuring (100, 16, 131); unreachable arms (101); named arguments in full (21); std functions as values (20); argument order (23) | important | S |
| `guide-proofs` | ch3 section, or new chapter "Checks the compiler proves" | `requires` relations and `\|\|` clauses (172); `ensures` (173); refinement types and `T.check` (95); `FixedArray` and proven indexes (83); constant type parameters (104) | important | M |
| `guide-compile-time` | ch8 section "Constants", or new chapter | `@const` functions and limits (166); `Regexes.checked` (75); tagged templates and `Template.raw` (38) | important | S |
| `guide-targets` | new chapter after ch12: "Building for other systems" | Build conditions and `@target` (168); `--target`, cross `CC`, `narrac targets`, WASI, Windows notes (169, 64); `narrac emit` and `-emit-c`; debug and release builds and the cost model (167); `std/simd` (84) | important | M |
| `guide-tools-gaps` | ch10, ch12 | `Test.golden` (171); `narrac move` (57); commands missing from the ch12 table | important | S |
| `cli-reference` (outline) | `/reference/commands` | Every command and flag (185) | important | M |
| `manifest-reference` (outline) | `/reference/narra-toml` | Every table and key; `[link]` (182); manifest and lock errors (184) | important | S |
| `error-index` (outline) | `/reference/errors`, generated | Every diagnostic code (175) | important | M |

That is 16 new nodes (5 S, 10 M, 1 L) and 6 outline nodes. Rows 5, 33 and 120 are important, not crucial. They sit in `guide-basics-gaps` because they belong in chapter 2. The SIMD material sits in `guide-targets` because it depends on the target's vector registers. It is niche and can be cut from the node without loss for a first user.

### Reconciling the nine outline nodes

| Outline node | Decision | Reason |
| --- | --- | --- |
| `site-reference-section` | Join batch-21 | The std, command, manifest and error pages need a home on the site. |
| `std-interface-query` | Join batch-21 | `std-reference` depends on it; the 54-module catalog (row 58) is crucial. |
| `std-reference` | Join batch-21 | A learner cannot find `std/path` or `std/time` today. Signatures without descriptions are still a list a learner can read. |
| `std-doc-comments` | Stay after release | L. The reference is usable with signatures; the guide nodes above teach the modules a user needs first. |
| `cli-reference` | Join batch-21 | Chapter 12's table misses `build`, `run`, `emit`, `move`, `audit` and `targets`. |
| `manifest-reference` | Join batch-21 | `[link]` is taught on no page; it is needed by `guide-c-interop`. |
| `error-index` | Join batch-21 | 479 codes, and `narrac explain` shows one at a time. A learner meets codes from chapter 12 on. |
| `spec-decision-links` | Stay after release | It links reasons, it teaches no concept. |
| `guide-concurrency` | Replace with `guide-async` and `guide-threads` | Async is crucial and threads are important; one L chapter would mix two different models (one loop per thread against scoped threads). |

## Chapter order changes

A concept used before it is taught:

- **Generators.** `Arrays.values`, `Arrays.entries` and `Maps.keys` return generators from chapter 2 on, chapter 9 names `Generator` as a `mutable` type, and chapter 11 opens with `function*` and `yield` with one sentence of explanation. Add `guide-generators` as a chapter before the standard library chapter.
- **`derives Shape`.** Chapter 11's TOML section decodes into structs that derive `Shape`, which no earlier chapter explains. Teach `Shape` and `std/json` (`guide-serialization`) in chapter 7 or before the TOML section.
- **`Cell`.** Chapter 11's error sample prints a message that recommends a `Cell`. Teach it in chapter 9 (`guide-cell-and-state`).
- **Divisor proofs.** Chapter 2 divides by constants without saying why, chapter 3 divides inside a `requires` body, chapter 10 divides after a `panic` check, and the rule itself first appears in chapter 6. Teach it in chapter 2.
- **C-style `for` and labels.** Only chapter 13 shows the C-style loop. Teach both in chapter 2, next to `while`.
- **Build conditions.** Chapter 12 mentions `std/build` and `--all-targets` before any page explains them. Either move that paragraph to `guide-targets` or link forward to it.
- **Forward references that stay.** Chapter 2 uses `mutable` (taught in chapter 9), chapter 4 uses `Result` and `uses alloc` (chapters 6 and 10), and chapter 7 uses `from items` (chapter 9). Each says where it is taught, and each is needed where it appears, so no move is proposed.

Proposed order: 1 Getting started, 2 Values and types, 3 Functions and closures, 4 Structs and enums, 5 Pattern matching, 6 Option and Result, 7 Traits (with serialization), 8 Modules and packages, 9 Memory and `mutable`, 10 Effects and testing, 11 Files and the system, 12 Generators, 13 The standard library, 14 Async, 15 Threads and shared state, 16 Unsafe code and C, 17 Checks the compiler proves and compile-time constants, 18 Tools, 19 Building for other systems, 20 Narra for TypeScript programmers. Chapters 17 and 19 can also be sections of existing chapters if the maintainer prefers fewer files.

## Not done

- No doc or code is changed by this node.
- `DECISIONS.md` was not read in full (see Method).
- The audit does not check every guide sample's output against the compiler. The guide suite does that.
