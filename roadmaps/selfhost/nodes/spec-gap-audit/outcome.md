# Outcome: spec-gap-audit

## Built

- An audit of `SPEC.md` against the Go compiler at origin/main 8409e3b. 241 minimal programs were run with `narra check|run|test` in a scratch directory, plus one scratch package for cross-module rules. No code changed.
- 21 confirmed gaps (table below). F1 to F21 were all run; the "Actual" column is the output seen. Gaps the README already lists under "Not supported yet" are still gaps. They are marked "(README)". Items under "Differences from the spec" are not listed.
- A proposed grouping into 18 fix nodes, and a list of every SPEC.md section with its state.

### Findings

Programs are the body of `export function run(args: Array<string>): int` unless they show their own declarations. `⏎` is a line break. Imports are left out where the spec leaves them out.

| # | Spec line | Construct | Minimal program | Actual | Size |
| --- | --- | --- | --- | --- | --- |
| F1 | 73, 926 | Self tail calls compile to loops | `function count(n: int, acc: int): int { if (n == 0) { return acc } ⏎ return count(n - 1, acc + 1) }`, call `count(10000000, 0)` | Debug build segfaults (exit 139; `narra run` exits 255). The C is a plain recursive call. Release passes only because clang does TCO | M |
| F2 | 62 | Function declared inside a function | `function twice(x: int): int { return x * 2 }` inside `run` | `expected an expression, found "function"` | M |
| F3 | 120 | Labeled `break`/`continue` | `outer: for (const i of Ranges.upTo(3)) { ... break outer }`; also `outer: while (true)` | `expected the end of the statement, found ":"` | M |
| F4 | 64 | Block-body arrow as a call argument | `apply((x) => { ⏎ print("io") ⏎ return x ⏎ })` | `expected the end of the statement, found "return"`. One statement in the block works. The same arrow bound with `const` works | S |
| F5 | 57 | Unused local or parameter is an error | `const x = 1 ⏎ return 0`; `function f(a: int, b: int): int { return a }` | Compiles clean. README records only unused imports and private functions as a difference | M |
| F6 | 110-111 | `for ... of` over any iterator value | `const r = Ranges.upTo(3) ⏎ for (const i of r) {}`; `for (const item of ["a"]) {}`; `for (const c of Str.chars("ab")) {}` | `for ... of needs Range.upTo, Range.between, Array.values, Map.keys or Map.entries, found Range` (also `Array<string>`, `Array<char>`) | M |
| F7 | 131, 231-238 | Sized numbers and `bytes` | `const x: u8 = 200`; `i8`..`u64`, `i32`, `f32`, `bytes` | `unknown type "u8"` for every sized type, `f32` and `bytes`. "A literal must fit its type" holds only for `int` | L |
| F8 | 344 | `b.value = v` on a `Box` | `let b = Boxes.new(5) ⏎ b.value = 6` | Invalid C: `member reference base type 'B_Box_d2ab61dd *' ... is not a structure or union` (int, struct and string payloads) | S |
| F9 | 407, 569 | Narrowing after `if (x is None) { return }` | `let x: Option<int> = Some(1) ⏎ if (x is None) { return 0 } ⏎ print(Int.toString(x.value))` | `this Option<int> may hold any of its variants`. Works on `const` and parameters only | M |
| F10 | 667 | A `mutable` parameter cannot be returned or stored | `function keep(mutable p: P): P { return p }`; `function keep(mutable p: P): H { return H { p: p } }` | Both compile clean. Capture is rejected correctly | S |
| F11 | 1007-1016 | `requires` / `ensures` contracts | `export function sqrt(x: f64): f64 ⏎ requires x >= 0.0 ⏎ ensures result >= 0.0 ⏎ { ... }` | `expected "{", found line break` | M |
| F12 | 910-915 | `lazy` values and `@const` functions (README) | `const table = lazy build()`; `@const function build(): int` | `expected the end of the statement, found "build"`; `unexpected character '@'` | M |
| F13 | 209, 219 | `type` / `export type` aliases (README) | `type Tok = Token`, then `Tok.Word("ok")` | `expected a declaration, found "type"` | M |
| F14 | 501-503 | `Dynamic<Trait>` (README) | `const items: Array<Dynamic<Drawable>> = [Dynamic.of(Circle { r: 1 })]` | `expected a name, found "of"` | L |
| F15 | 623-624 | `@unique` and `NARRA_TRACE_COPIES=1` | `@unique let buffer: Array<int> = []`; run a copy-on-write program with `NARRA_TRACE_COPIES=1` | `unexpected character '@'`; the trace prints nothing | M |
| F16 | 165-183, 1039 | Each TypeScript habit has a diagnostic with a fix | `class A {}`, `interface A {}`, `new Thing()`, `try {} catch (e) {}`, `throw "x"`, `for (const k in xs)`, `const x: number = 1`, `'abc'`, `s.length` | `class`/`interface`: E0019 `expected a declaration`. `new`/`try`/`throw`: E0013 `expected the end of the statement`. `for in` E0031, `number` E0113, `'abc'` E0002, `.length` E0208 have messages but `fixes=[]`. Rows that do have a fix: `===`, `;`, `null`, `undefined`, `T[]`, `a + b`, default import, `xs.push` | M |
| F17 | 191 | `Arrays.push` without the import gives E0908 with an add-import fix | `let xs = [1] ⏎ Arrays.push(mutable xs, 2)` with no import | E0308 `unknown module "Arrays"`, `fixes=[]`. Only `Array.push` gets E0908 with the fix | S |
| F18 | 29, 58, 66, 98, 110, 159, 479, 534, 540-541, 627 | Std functions and constants the spec uses | `Float.PI`, `Int.divmod(7, 2)`, `Arrays.map(xs, f)`, `Arrays.entries(xs)`, `Int.floorDiv(-7, 2)`, `Int.compare(1, 2)`, `Str.find`, `Str.copy`, `Arrays.slice`, `Arrays.copy`, `Arrays.withCapacity`, `Int.truncate` | `module "std/..." has no function "..."` each. `Float.PI` is `a module is not a value`: std modules have no constants | M |
| F19 | 223 | `narra move old new` | `narra move a b` | Usage text, exit 2. No subcommand | S |
| F20 | 702 | Calls through function values carry their effects (README) | `struct H { f: () => void } ⏎ export function call(h: H) { h.f() }`, called with `H { f: () => print("io") }` | Compiles clean: the exported `call` does `io` without `uses io` | M |
| F21 | 791, 805 | Explicit type arguments in a call (README) | `function id<T>(x: T): T`, then `id<int>(1)` | Parsed as comparisons: `unknown name "int"`, `cannot order void and void` | M |

### Planned or parked (not tested in depth)

Each was run once to confirm it is rejected today.

- `Ranges.stepped`, `Int.pow`, `Float.pow`, `Hash.of` -> std-basics (all `has no function`; `Hash.of` fails to parse).
- `for (const #[i, item] of ...)` -> for-destructure (`expected a name, found "#["`).
- `x += 1`, `x++` -> compound-assign (E0034 / parse error).
- `@inline` -> inline (`unexpected character '@'`).
- `function*`, `yield`, `Coroutine` -> generators (`generators are not supported`).
- `async function`, `await` -> async (`expected a declaration, found "async"`).
- `unsafe`, `extern`, `@repr`, `Red = 1` with its fix, `narra audit` -> c-interop (`unsafe {` and `Red = 1` are parse errors).
- `derives Shape` -> std-json (`derives Shape is not supported by this compiler yet`).
- `--report-copies` -> perceus (`narra check` usage error).
- `narra query` -> query-cli; `narra fmt` -> fmt; `check --all-targets`, `Build.os`, `@target` -> build-conditions (all usage errors).
- Threads, channels, `Shared`, `Atomic`, `Parallel.map` -> threads. Regions -> regions. `FixedArray` -> fixed-arrays. `std/simd` -> simd. `Arena` -> std-arena. `std/unicode` -> std-unicode. Non-escaping closures capturing `let` -> closure-borrow.

### Proposed fix nodes

| Node | Findings | Size | Reach | Files |
| --- | --- | --- | --- | --- |
| block-arrow-args | F4 | S | Parser only. Statements inside a `{ }` that sits inside `( )` must end at line breaks | `parser.go` (or `lexer.go`), `selfhost/src/parser.nra` / lexer |
| box-assign | F8 | S | Emitter only: write through the box pointer, copy first when the count is above 1 | `emit_stmt.go` or `emit_box.go`, selfhost emitter |
| mutable-escape | F10 | S | Checker only: reject `return p` and storing `p` in a literal for a `mutable` parameter | `check_stmt.go`, `check_expr.go`, selfhost checker |
| tail-loops | F1 | M | Emitter only: a self call in tail position becomes parameter reassignment and `goto` | `emit_stmt.go`, `emit_call.go`, selfhost emitter |
| labels | F3 | M | Parser, checker (label scope, unknown label) and emitter (`goto` past the loop) | `parser.go`, `ast.go`, `clone.go`, `check_stmt.go`, `emit_stmt.go`, selfhost |
| local-functions | F2 | M | Parser, checker and emitter. One option: desugar a non-capturing one to a private top-level function with a hidden name. A capturing one is a closure, so it needs the closure path | `parser.go`, `ast.go`, `clone.go`, `check.go`, `emit.go`, selfhost |
| let-narrowing | F9 | M | Checker only: narrow a `let` until its next assignment | `check_patterns.go`, `check_stmt.go`, selfhost checker |
| for-of-sources | F6, `Arrays.entries` from F18 | M | Checker and emitter: iterate any `Range`, `Array<T>` or `Str.chars` value, and add `Arrays.entries` | `check_stmt.go`, `emit_stmt.go`, `std_array.go`, `std_range.go`, selfhost |
| std-spec-functions | F18 except `Arrays.entries` | M | Checker and emitter builtins. `Float.PI` also needs module constants in the checker | `std_int.go`, `std_float.go`, `std_array.go`, `std_string.go`, `builtins.go`, selfhost builtins |
| sized-numbers | F7, plus `I32.from` / `F64.fromInt` conversions (line 534) | L | Lexer, checker and emitter: types, literal range checks, conversions. Split by integer types, `f32` and `bytes` | `types.go`, `check_expr.go`, `check_types.go`, `emit_types.go`, `runtime.h`, new std modules, selfhost |
| unused-names | F5 | M | Checker only. Every existing sample and `selfhost/src` must be cleaned first | `check.go`, `check_stmt.go`, `testdata/`, `selfhost/src` |
| contracts | F11 | M | Parser, checker (`result` name, bool type) and emitter (debug-only checks) | `parser.go`, `ast.go`, `clone.go`, `check.go`, `emit.go`, selfhost |
| const-lazy | F12 | M | Parser, checker and emitter (`Lazy.force`, a once guard; `@const` evaluation) | `parser.go`, `check_const.go`, `emit_const.go`, selfhost |
| type-aliases | F13 | M | Parser and checker (aliases resolve to the target type, enum constructors through the alias). Emitter unchanged | `parser.go`, `ast.go`, `check_types.go`, `modules_check.go`, selfhost |
| dynamic-traits | F14 | L | Parser, checker and emitter (boxed value and vtable) | `check_traits.go`, `emit_traits.go`, selfhost |
| copy-tools | F15 | M | Parser (`@unique`), checker and emitter/runtime (debug panic on a shared update, trace on copy). Could fold into perceus with `--report-copies` | `parser.go`, `emit_*.go`, `runtime.h`, selfhost |
| ts-habit-fixes | F16, F17 | M | Parser and checker diagnostics only, with `testdata/diagnostics/` cases | `parser.go`, `diagnostic_fixes.go`, `import_names.go`, selfhost |
| typed-calls-and-effects | F20, F21 | M | Parser (`f<T>(x)` with the same balanced `<...>` rule as literal type args), checker (effects on function types) | `parser.go`, `check_call.go`, `check_effects.go`, `types.go`, selfhost. May split in two |

F19 (`narra move`) needs no node of its own. It can go to query-cli or the m9 tooling.

### SPEC.md sections

- 1 Overview: checked. The sample fails only on `Float.PI` (F18).
- 2 Syntax, statements and layout: checked. Semicolons, line continuation and all comment forms work. `narra fmt` is planned (fmt).
- 2 Bindings: checked. F5. Compound assignment is planned (compound-assign).
- 2 Functions: checked. F1, F2, F4. Arity, method calls, no varargs and no optional parameters are correct.
- 2 Closures: checked. Escaping rules work. `@inline` is planned (inline).
- 2 Control flow: checked. F3, F6. Destructuring heads are planned (for-destructure). `Ranges.stepped` is planned (std-basics).
- 2 Literals: checked. F7. All other forms work, including type arguments on literals and tuples.
- 2 Operators: checked. All rows work, including wrapping, shifts, NaN and division panics. `pow` is planned (std-basics). `Int.floorDiv` is F18.
- 2 Differences from TypeScript: checked. F16. The async row is planned (async).
- 3 Modules and imports: checked, including a scratch package (re-export through `const`, named imports, cycles, opaque reads). F13, F17, F19.
- 4 Primitive and composite types: checked. F7. `FixedArray` is planned (fixed-arrays) and SIMD is planned (simd).
- 4 Structs: checked. Defaults, building, spread, assignment, destructuring and struct patterns all work.
- 4 Boxes: checked. F8.
- 4 Opaque and distinct types: checked. They work. `==` on distinct types and no traits on distinct types are recorded differences.
- 4 Optional values: checked. F9. The `std/option` helpers work.
- 4 Pattern matching: checked. Guards, `match` as a value, exhaustiveness and block arms work.
- 4 Generics: checked. F21. `Hash.of` is planned (std-basics).
- 4 Traits: checked. F14. `derives Shape` is planned (std-json). Constraints and implement-blocks have recorded differences.
- 4 Inference: checked. It works.
- 4 Numbers: checked. Overflow panics in debug and wraps with `run -release`. F7. Conversions are in F18 and sized-numbers.
- 4 Strings: checked. The `parsePair` sample works with `uses alloc`. `ByteIndex` and slice views are recorded differences. `Str.find` and `Str.copy` are F18.
- 5 Errors: checked. `is Err` narrowing, an unchecked Result, `const _ =`, panic and `Results` all work.
- 6 Values and ownership, and in-place updates: checked. Copy-on-write and map index assignment work. Perceus is planned (perceus). Batched large frees were not tested (the spec says "may").
- 6 Copies on shared values: checked. F15. `--report-copies` is planned (perceus).
- 6 Regions: planned (regions).
- 6 Mutable parameters: checked. F10. Exclusive access and argument checks work.
- 6 Cycles: `let` capture is rejected. `Arena` is planned (std-arena).
- 6 Values across threads: planned (threads).
- 7 Effects: checked. F20. Effect variables, unbound variables and `unsafe` propagation work.
- 8 Generators: planned (generators). Threads, channels and shared state: planned (threads). Async: planned (async).
- 9 Unsafe and C interop: planned (c-interop).
- 10 Top-level declarations: checked. F12. Constants work.
- 11 Code generation: partly a recorded difference (one C file, no layout tables, no hidden trait tables). F1 (tail calls). `@specialize`, `generics = "compact"` and TCC debug builds are planned (layout-tables) or not testable from programs. `[build]` is skipped, as the README says.
- 12 Build model: interface files and incremental builds are a recorded difference. Determinism checked: two `emit` runs are byte-identical. Build conditions are planned (build-conditions).
- 13 Tests and contracts: checked. `std/test`, the runner, `-filter` with no match (exit 1) and the outside-test error work. F11.
- 14 Tooling: checked. `--json` and `explain` work. F16. Queries are planned (query-cli). The 100 ms speed target was not measured.
- 15 Packages: checked. `[[binary]]` and `[link]` are read, `[build]` is skipped and `[dependencies]` is an error. All are recorded in the README Usage section. Dependencies and `narra.lock` are in the README "Not supported yet" list, with no node.
- 16 Deferred: nothing to check.

### Suspicions (not confirmed, not in the table)

- A non-capturing arrow function may still get an environment rather than a plain C function pointer (line 87). The C was not inspected.
- Dependencies, `narra.lock` and minimal version selection (lines 1062-1089) have no node. They were not run beyond the README's "`[dependencies]` is an error".

### Spec text to fix (not compiler gaps)

- Line 329 uses `F64.toString` and line 534 uses `I32.from`, `F64.fromInt` and `Int.truncate`. The std list (line 193) has no `std/f64` or `std/i32`, and the compiler has `Float.toString`, `Float.fromInt` and `Float.truncate`.
- Line 578 writes the bare constructor `IoFailed(e)`, which line 222 (D71) makes an error.
- Line 552: the exported `parsePair` needs `uses alloc` (through `Str.indexOf` and `Some(...)`). The sample leaves it out.
- Line 625 `@unique let buffer: Array<u8>` also depends on F7.

## Deviations

- None from the brief. Findings that the README already lists as "Not supported yet" (F12, F13, F14, F20, F21) are kept in the table, because that list is not a recorded difference.

## Stubs left

- None.

## Gotchas

- for-destructure's acceptance uses `Arrays.entries`, which does not exist (F18). Either for-destructure adds it or it waits for for-of-sources.
- unused-names (F5) conflicts with the lint brief, which treats unused bindings as something "the checker accepts on purpose". SPEC.md says they are errors.
- F4 and F8 are bugs in features the README lists as supported. They may hit selfhost code after bootstrap.

## Proposed learnings

- Run every SPEC.md sample when a feature node closes | F8 (Box assignment) and F4 (block arrows in calls) are in features marked done | `compiler/testdata/` | yes: a test that extracts the fenced `ts` blocks and checks them

## Verification

- `go build ./cmd/narra` at 8409e3b. Each finding has a program run with `narra check`, `run` or `test` (F1 also `build` and `build -release`). The outputs are in the table.
- `git diff origin/main --stat` lists only `roadmaps/selfhost/nodes/spec-gap-audit/`.

## Size

- Estimated M, actual M.
