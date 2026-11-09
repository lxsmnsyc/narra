# narra

`narrac` is the Narra compiler, written in Narra (`selfhost/`). It compiles a `.nra` file or a package to C and builds it with the system C compiler. A checked-in C file, `seed/narrac.c`, builds it with `cc` alone, so no other compiler is needed. The Go bootstrap compiler that came first is gone.

## Usage

```sh
sh selfhost/bootstrap.sh out    # writes out/narrac

narrac run testdata/numbers.nra          # build and run
narrac build -o app file.nra             # build a debug executable
narrac build -release -o app file.nra    # optimized, integer overflow wraps
narrac test testdata/tests.nra           # run `test` blocks
narrac emit file.nra                     # print the generated C
narrac check [--json] file.nra          # report errors; --json prints one object per line
narrac explain E0003                    # explain a diagnostic code
narrac build testdata/packages/app       # build every [[binary]] of a package
narrac run -bin app testdata/packages/app
```

A single file's entry function is `export function run(args: Array<string>): int`.

A package is a directory with `narra.toml` and `src/`. The file `src/text.nra` of package `app` is module `"app/text"`. `narra.toml` sets the package name and the binaries:

```toml
[package]
name = "app"

[[binary]]
name = "app"
entry = "cli.run"     # module inside the package, then the function; "app/cli.run" also works
```

`[link]` (`libraries`, `search`) is read but not used yet. Other tables, such as `[build]`, are skipped. `[dependencies]` is an error.

Set `NARRA_CHECK_LEAKS=1` when running a program to report allocations that are still live at exit. The program exits with code 102 when any are found.

## Bootstrapping

`selfhost/bootstrap.sh` needs only `cc`. It builds the compiler from the checked-in C seed, `seed/narrac.c`:

1. Stage 0: `cc` builds the seed into `narrac0`.
2. Stage 1: `narrac0 emit selfhost > stage1.c`, then `cc` builds it into `narrac1`.
3. Stage 2: `narrac1 emit selfhost > stage2.c`. The script exits 1 when `stage1.c` and `stage2.c` differ.

`sh scripts/seed-update.sh` regenerates the seed from the stage 2 C. `seed/README.md` explains the seed rule and when to regenerate. The sanitized self-build is the `bootstrap` suite of the test runner.

## Commands of narrac

`narrac build`, `run`, `test` and `check` (`selfhost/src/driver.nra`) call `$CC` (else `cc`) to build the C. They add `-sanitize` (AddressSanitizer and UBSan) and, for `run` and `test`, `-stdin file`. `run` and `test` capture the program's output and write it after the program ends. `narrac check --json` prints one JSON line per diagnostic and `narrac explain CODE` prints the explanation, fixes included (`selfhost/src/fixes.nra` and the report sites). `narrac apply-fix [--diagnostic N] [--fix M] path` prints the file a fix edits with the fix applied: N defaults to the first diagnostic that has a fix, M to 0; exit 1 when there is no such fix. Every report site passes its diagnostic code; the table is `selfhost/src/diagnostics.nra`, and the `codes` suite of the test runner checks the sites.

## Formatting

`narrac fmt` (the self-hosted compiler, `selfhost/src/fmt.nra`) formats Narra source.

- `narrac fmt [paths]` rewrites the files in place. A directory means every `.nra` file under it (hidden directories are skipped); no path means the current directory.
- `narrac fmt --check [paths]` prints each file that would change and exits 1, or exits 0.
- `narrac fmt --stdin` (or `-`) reads one file from stdin and writes the result to stdout.
- A file that does not parse prints `error <path>:<line>:<col>: <message>`, stays unchanged, and the exit code is 1. A bad `[fmt]` key or value prints `error narra.toml:<line>: <message>` and the exit code is 1.
- Options come from the `[fmt]` table of the nearest `narra.toml` above each file (`printWidth`, `tabWidth`, `useTabs`, `trailingComma`; see `SPEC.md` section 2). Without one the defaults apply.
- Declarations and statements are printed by `fmt.nra`; expressions, types and patterns by `fmtexprs.nra`.
- The `fmt` and `fmt-stable` suites of the test runner run the goldens in `testdata/fmt/` and, for every `.nra` file under `testdata/` and `selfhost/src`, check that formatting twice gives the same text and that the tokens, the comment texts and the syntax tree are unchanged.
- The `sources` suite runs `narrac fmt --check` over `selfhost/src`, `selfhost/tests` and every sample under `testdata/` except `testdata/fmt/`, `lex`, `parse` and `habits`, whose inputs are unformatted or invalid on purpose, and the layout samples, which test a layout the formatter would change. A sample that does not parse is skipped only if its path contains `error` or `/diagnostics/`. `bench/` and `std/` are not checked. After editing Narra source, run `narrac fmt` on it. A sample whose `// error:` line or dump golden holds positions must be updated after formatting.

## Queries

`narrac query` (the self-hosted compiler, `selfhost/src/query.nra`) answers the questions of `SPEC.md` section 14 about a package. The command line and the `--json` shapes are the frozen contract `roadmaps/selfhost/contracts/query-output.md`.

- `narrac query [--package DIR] [--json] <command> <operand>`, where the command is `type-of FILE:LINE:COL`, `def NAME`, `refs NAME`, `callers FUNCTION`, `interface MODULE` or `recursion MODULE`. `DIR` defaults to `.`.
- The package is loaded and checked as `narrac types` does it. Load or check errors print as there (with `--json`: `{"errors": [...]}`, the diagnostic's code; `E0000` for a problem that is not a diagnostic) and the exit code is 1. Any other query error prints `error <message>` (with `--json`: `{"error": ...}` on stdout) and exits 1. A usage error exits 2.
- `type-of` reports the innermost expression, local name or function or constant name at the position. `type` is the type of the value there, function types with their `uses` list (from the declaration or the annotation; an unannotated local holding a function shows none). `effects` is what evaluating the expression there does: a name has none, a call has the call's. A builtin or constructor name in a call answers for the call, because the checker records nothing for the name. A generic function shows its type parameter names. A position inside a template string with `${...}` parts is an error for now: the parser gives the nodes there the template's position.
- `def` lists every declaration a qualified (`app/text.helper`, `app/text.Shape.Circle`) or bare (`helper`, `Shape.Circle`, `Circle`) name finds. `refs` and `callers` need a name that finds one declaration.
- `refs` finds functions, constants and enum variants (calls, values, constructors and patterns). Uses of a type name in annotations or struct literals are not tracked. A use inside a template string is reported at the whole template, with `"inTemplate": true`; `callers` does the same.
- `callers` and `recursion` read the call graph of the effect pass (`EffectState.calls`). Calls through function values are not edges. A call inside an arrow function belongs to the enclosing function.
- Only the package's own modules are searched, not std source modules.
- `selfhost/src/querysites.nra` builds the position index, `querysignature.nra` prints `interface` signatures, `querygraph.nra` finds cycles and `queryjson.nra` writes JSON.
- The `query` suite runs `testdata/query/<case>/queries.txt` (`> command` lines, each followed by `# exit N` when the exit code is not 0, the expected stdout, and `# stderr` with the expected stderr when there is any) against the package in that folder.

## Tests

```sh
sh scripts/test.sh                      # build narrac from the seed, run every suite
NARRAC=/path/to/narrac sh scripts/test.sh --suite errors
sh scripts/linux-check.sh               # the same in a Linux container
```

The test runner is written in Narra (`selfhost/tests`). `scripts/test.sh` builds it with narrac and runs it in shards. Every sample runs under AddressSanitizer, UndefinedBehaviorSanitizer and the leak check. `--suite NAME`, `--filter TEXT` and `--update` (rewrite goldens) go to the runner. Each sample in `testdata/errors`, `testdata/stage2/errors` and `testdata/packages/errors` has a golden with narrac's full `check` output: `<name>.err`, or `expected.err` inside a package folder. The `// error:` lines stay as documentation.

`scripts/linux-check.sh` runs `sh selfhost/bootstrap.sh` and `scripts/test.sh` in a Linux container (gcc with sanitizers, repo mounted read-only). Docker runs through colima on macOS: `colima start --cpu 4 --memory 8` first, and work from a path under your home folder.

- `testdata/*.nra` run and compare their output with the matching `.out` file. Each runs twice: a debug build with AddressSanitizer, UndefinedBehaviorSanitizer and the leak check, and a release build.
- `tests.nra` and `testdata/std_test*.nra` run their `test` blocks (`narrac test`) instead of an entry function. The `std-tests` suite checks failing `std/test` assertions.
- `testdata/panics/` must panic with a given message and exit code 101.
- `testdata/errors/` must fail to compile. The first line, `// error: text`, gives text the error must contain.
- `testdata/packages/*/` with a `run.out` build the package's binary, run it in both modes like `testdata/*.nra`, and run its tests.
- `testdata/packages/errors/*/` must fail to compile. The first line of `narra.toml`, `# error: text`, gives the text.
- Standard modules written in Narra live in `std/src/`. `selfhost/src/stdtext.nra` embeds their text and is generated from them; the `generated` suite fails when it is stale (`sh scripts/test.sh --suite generated --update` rewrites it, then `sh scripts/seed-update.sh`). The C runtime of the compiler lives in `selfhost/runtime/` and `selfhost/src/runtimetext.nra` is generated from it the same way. Samples that use the newest std modules are in `testdata/stage2/*.nra` (with `.out`) and `testdata/stage2/errors/*.nra`.
- `std/internal/<name>` modules hold the primitives behind std source modules. Only `std/...` modules import them (E0426). `std/internal/probe` is a test-only primitive in `selfhost/src/stdlib.nra` (the same code as `Char.code`).

## Supported

- The prelude has the types `Array`, `Map`, `Range`, `Box`, `Option`, `Result` and the functions `print`, `assert`, `panic`. The functions `Array.*`, `Map.*`, `Range.*` and `Box.*` come from `std/array`, `std/map`, `std/range` and `std/box` and need the import (error E0908, with an add-import fix; D67). By convention the aliases are plural: `import * as Arrays from "std/array"`, `Maps`, `Boxes`, `Ranges`, `Options`, `Results`. An alias may not equal a type name, two imports may not share an alias, and `import { map as mapOption }` renames a name (D72).
- Integer conversions (D82): modules `std/i8`, `std/i16`, `std/i32`, `std/u8`, `std/u16`, `std/u32` and `std/u64` (aliases `I8` ... `U64`) have `from`, `tryFrom`, `wrap`, `toString` and `parse`; `std/int` has `from`, `tryFrom` and `wrap`. They are builtins in `selfhost/src/stdsized.nra`: `from` and `wrap` emit a C cast, `tryFrom` a range check. `from` on a type that can lose values is E0940, a non-integer argument E0941. `parse` reports a value outside the range with the range in the message.
- `std/option` and `std/result` (`import * as Options from "std/option"`) have the helpers `Option.map`, `andThen`, `withDefault`, `orElse`, `isSome`, `isNone`, `okOr` and `Result.map`, `mapErr`, `andThen`, `withDefault`, `ok`, `err` (D69). The callbacks are non-escaping, their effects flow to the call (`Builtin.Callbacks`), and a capturing arrow function gets a stack environment. There is no `unwrap` or `expect`.
- Packages with several modules and `narra.toml`, or one source file. `import * as X from` and `import { a } from` for `std/int`, `std/float`, `std/string` and the modules of the same package. `export` on functions, structs and enums; other names are private. Import cycles are errors that list the cycle.
- `std/fs` (`readText`, `writeText`, `exists`, `listDir`), `std/process` (`run`) and `std/env` (`get`, `args`). Fallible functions return `Result<T, string>`; the string is `<subject>: <reason>`. `writeText` returns the byte count, `listDir` the sorted names (Err for the whole listing if any entry name is not valid UTF-8), `Process.run(command, args)` returns `#[exit code, stdout, stderr]` (stdin closed; killed by signal gives 128 + signal), `Env.get` returns `Option<string>`, and `Env.args()` equals the entry function's `args`.
- `std/fs` also has `isDir`, and `createDir`, `createDirAll`, `removeFile` and `removeDir`, which return `Result<bool, string>`: `Ok(true)` when the call changed the filesystem, `Ok(false)` when `createDirAll` found the directory already there, Err otherwise (with the path in the message). `std/io` reads stdin through C `FILE*` only: `Io.readLine()` returns `Result<Option<string>, string>` (the line without `\n`; a `\r` of a `\r\n` line stays, so the caller trims it; `None` at the end of input), `Io.readBytes(count)` exactly `count` bytes (Err at an early end, a negative count or invalid UTF-8), `Io.readAll()` the rest; `Io.write` and `Io.writeError` write to stdout and stderr without a newline and `Io.flush` flushes stdout. `std/path` follows Node.js `path.posix`: `join(left, right)`, `dirname`, `basename` (no suffix argument) and `extname`; `/` only, pure string work.
- `std/string` also has `slice`, `sliceFrom`, `byteAt`, `charAt`, `nextChar`, `charCount`, `indexOf` (returns `Option<int>`), `startsWith`, `trim` (ASCII whitespace), `split` (panics on an empty separator), `chars` (`Array<char>`) and a `StringBuilder` (`Str.builder()`, `Str.append(mutable sb, s)`, `Str.appendChar(mutable sb, c)`, `Str.build(sb)`). Slicing, `charAt` and `nextChar` panic on a position out of range or inside a character.
- `std/char` has `Char.code` and `Char.fromCode` (`Option<char>`). `std/int` also has `Int.parse(s, base)` (`Result<int, string>`, bases 2 to 36; no prefix, and a leading `+` or whitespace is rejected) and `Int.toHex`. `std/string` also has `endsWith`, `contains`, `replace` (an empty `from` returns `s` unchanged) and `join`.
- Names from a namespace import: calls `X.f(...)` and types `X.T`. Named imports bring functions, and structs and enums. Constructors come with their enum and are always written qualified: `Shape.Circle(1)`, `Geo.Shape.Circle(1)`, in expressions and in patterns (D71). `import { Circle }` of a constructor is E0425.
- Functions, `const`, `let`, `if`, `while`, `for (const x of Range.upTo(n))`, `Range.between(a, b)` and `Array.values(items)`, `Array.entries(items)` (yields `#[index, item]`), `for ... of` over any `Range` value, any `Array<T>` value and `Str.chars(s)` (a `Range` is the type `Range`: two ints, usable in a `const`, a field, a parameter or a result; `Array.values` and `Array.entries` outside a loop head are error E0920), `for (const #[i, x] of ...)` and `for (const { x, y } of ...)` heads (the parser rewrites them to a hidden loop name `__for_<line>_<column>` and a `const`/`let` declaration at the top of the body), `break`, `continue`.
- Block expressions and `if` as a value (D75, SPEC.md section 2): a `{ ... }` in value position gives its value with `return`, which ends the block and not the function; `break` and `continue` cannot leave it (E0932), every path must end in `return value` (E0931) and a block with no `return`, such as `const o = { a }`, is E0933. `if (c) { ... } else { ... }` in value position needs the `else` and is parsed into a block expression holding an `if` statement whose branches each return a block expression. `{}`, `{ ...s }` and `{ name: value }` stay struct literals typed from context.
- Conditional expression `c ? a : b` (D75, SPEC.md section 2): parsed into a block expression holding an `if` whose branches each return one value; `return c ? a : b` is parsed into an `if` statement so a self call in a branch is a tail call. A missing `:` is E0934.
- `int`, `f64`, `bool`, `char` and `string`.
- Structs with constant field defaults, typed literals (`Point { x: 1 }`) and literals typed from context.
- Type arguments on map and generic struct literals: `Map<string, int> {}`, `Pair<int, string> { a: 1, b: "x" }` (D68). The checker resolves them like an annotation and reports E0237 when they differ from the expected type. Function types are not accepted inside them.
- Enums with fields, and `match` as a statement or a value, with guards and `_`.
- Patterns in `match`, `is` and `const`: nested constructors (`Some(Node(a, 0))`), struct patterns (`{ x: 0.0, y }`), tuple patterns (`#[a, Some(b)]`), literals (`0`, `-1`, `1.5`, `"a"`, `'c'`, `true`), bindings and `_`. `match` works on any value and checks exhaustiveness through nested enums, bools, structs and tuples.
- Destructuring: `const { host, port: serverPort } = config` and `const #[q, r] = divmod(a, b)`, nested too.
- Spread updates: `{ ...s, field: v }` and `Point { ...s, x: 1.0 }`.
- Tuples: the type `#[int, string]` (in every type position), the literal `#[a, b]`, and `const #[a, b] = t`. Tuples implement the standard traits when their elements do.
- `Array<T>`: literals, `Array.push`, `Array.length`, `Array.filled`, indexing and element assignment, with copy on write.
- `mutable` parameters, with the exclusive access check.
- Generic functions, structs and enums. Type arguments are inferred from arguments, the expected type and struct fields.
- `Option<T>` and `Result<T, E>`, declared in a prelude written in Narra.
- `is` patterns in conditions. `if (x is Some(v))` binds `v` in the branch. After `if (x is None) { return ... }`, `x.value` reads the value. Narrowing works on constants and parameters.
- Calls whose `Result` is not checked are compile errors.
- Template strings.
- Integer overflow checks in debug builds, and division by zero and bounds checks in every build.
- `test` blocks and `assert`. `std/test` (inside test blocks only): `Test.eq`, `Test.ne`, `Test.panics(f)` and `Test.panicsWith(f, message)` and `Test.golden(name, actual)`. A failed assertion panics with the call's position and both values written with `Show`. `narra test` runs each test in its own process, prints `ok - name` or `FAIL - name: message (file:line:col)`, then `N passed, M failed`, and exits 1 on a failure. `-filter text` runs the tests whose name contains the text and `-time` adds the milliseconds of each test. A test that runs past 10 seconds (or `NARRA_TEST_TIMEOUT`) fails with `timed out`. A filter with no match exits 1 (D66).
- Reference counting for strings, arrays, and structs and enums that hold them.
- Arrow functions `(x: int): int => x + 1` and `(x) => { ... }`, function types `(x: int) => int`, named functions as values (`apply(double, 1)`, `Alias.f`), calls through function values (`f(x)`, `make()(x)`, `button.onClick(x)`), and captures of constants. A closure passed straight to a non-escaping parameter also captures `let` variables and `mutable` parameters by reference (D77). Parameter types of an arrow function come from the expected type, also through a generic call such as `map(items, (x) => x * 2)`.
- `escaping` function parameters. A non-escaping closure keeps its environment on the caller's stack; an escaping one has a reference-counted heap environment.
- Traits: `trait`, `implements`, `implement Trait for Type`, `derives`, constraints such as `<T: Show & Ordered>`, and calls through the trait name such as `Show.show(x)`.
- The standard traits `Equatable`, `Hashable`, `Ordered` and `Show`, and the `Ordering` enum. `==` and `!=` work on structs, enums and arrays.
- Effects: exported functions must declare `uses alloc, io, unsafe` for everything they do; private functions and arrows are inferred; `<effect E>` and `uses E` on function types work. Function types carry their effects (D86): a function with fewer effects fits a slot that allows more, a store of one with more is a type error, and a call through a value charges the effects of its type. A function type inferred from an arrow or a name lists no effects, so a type argument inferred from such a value is pure; write the type to store an effectful one in a generic. On an exported function a function-type parameter without `uses` accepts only pure arguments; on a private function it acts as an implicit effect variable. `alloc` also comes from array literals, template strings, string `+` and constructors with fields.
- Opaque structs and distinct types: `export opaque struct`, `distinct type A = B`, `export opaque distinct type`. Another module cannot build, read, assign, spread or match the fields of an opaque struct. `A.from(x)` and `A.toInt(x)` (`toString` for a string base, named after the base type) are generated and emit no code. Distinct types have no `Show`/`Ordered`. `opaque`, `distinct` and `type` are contextual words, not keywords.
- Generators (SPEC.md section 8): `function*`, `yield`, `yield*`, `Generator<T> uses E`, generic generators and `for ... of` over a generator. Each `function*` is a state machine over a heap frame with a counted header (`narra_gen` in `runtime.h`): every C variable of the body is a frame field, and the drop and clone functions switch on the state to release or retain what is live at that yield. Every `Generator<T>` is the C type `narra_gen*`. A `for ... of` and a `yield*` resume their own copy, so iterating a generator variable twice runs it twice.
- Coroutines (D79): `Coroutine<In, Out>`, `const input = yield value` and `Co.resume` from `std/coroutine`. A coroutine is the same state machine as a generator; the resume input reaches the body through `narra_gen.in`, set by `narra_co_next` (`runtime.h`). `std/coroutine` is Narra source (`std/src/coroutine.nra`: the `Step` enum and `resumeGenerator`/`resumeCoroutine`) over the primitives of `std/internal/coroutine` (`selfhost/src/stdcalls.nra`, `selfhost/src/emitstd.nra`). The compiler maps `Co.resume` to one of the two by argument count (`moduleCall` in `selfhost/src/calls.nra`). In a test build `narra_co_next` catches a panic with `setjmp`.

- Top-level `const` (`export const` too): literals, operators on constants, constructors, structs, arrays, maps, function names and other constants, in any order; cycles are errors. Scalar constants are emitted at each use. Other constants are built on first use (see Differences).

## Not supported yet

- `@const` functions and `lazy` values. A top-level constant cannot call a function.
- Dependencies on other packages, `narra.lock`, interface files (`.nrai`) and incremental builds.
- `export type` aliases (so constructors through an alias are not available either).
- `Dynamic<Trait>`, `derives Shape` and `Hash.of`.
- Defaults for generic struct fields whose type uses a type parameter.
- Effects: trait function calls to user implementations are not followed; `unsafe` has no source yet.
- `@inline` (node inline), `async` arrow functions, and standard functions (`Int.toString`) as values.
- Reading a tuple element by index.
- Async functions, threads and channels.
- `unsafe` and `extern`.
- `std/test` on Windows: `Test.panics` and `Test.panicsWith` use `fork` and `alarm` (D65), so they are POSIX only. A closure that runs past the timeout (10 seconds, or `NARRA_TEST_TIMEOUT`) fails. The test runner (D66) forks per test, so it is POSIX only too. Report formats other than the text lines are not there yet.

## Differences from the spec

- The `mutable` and `const` rules of D81 are not built yet (node mutable-bindings). The compiler still follows D27: a local `const` cannot be changed at any depth, a `mutable` argument must be a `let`, a `mutable` parameter or a part of one, a `mutable` parameter may have any type and may be assigned whole, and a non-escaping closure captures a `const` by value.
- `for (mutable item of items)` and `for (const #[index, mutable item] of Arrays.entries(items))` (D85, SPEC section 2, Control flow) are not built yet (node mutable-bindings). The parser rejects `mutable` in a `for` head.
- `Co.resume` is not an importable name: `import { resume } from "std/coroutine"` is an error with a hint, and `Co.resume` works only through a module alias. A resume charges `alloc` for the `Step` payload besides the effects of the generator type. A panicked coroutine does not free what it held.
- A function that is not a `function*` and returns a generator type writes its own `uses` list after the type's (`(): Generator<T> uses io uses alloc`). A private function's `Generator<T>` parameter accepts only pure generators (D78).
- `==` and `!=` work on a distinct type when its base type supports them (and it can be a map key). SPEC.md says no operators carry over.
- A distinct type cannot list traits with `implements` or `derives`; SPEC.md allows it.
- Top-level constants that are not scalars (strings built by templates, structs, arrays, enums with fields, function values) are built on first use, behind a getter, and kept until the program ends. The spec says modules have no initialization code; this keeps that, but the first-use build is not thread safe yet (the threads node must add a once guard). Struct and array constants are heap values, not static C data.

- An unused local or parameter is an error (E0921, E0922), and `_` or a `_` prefix marks it as unused on purpose. The `args` parameter of the entry function of a binary (the exported function that `narra.toml`, or the single-file build, names as entry) is exempt, because its signature is fixed; a `run` in another module is not. An assignment `x = v` does not read `x`, so a `let` that is only assigned is unused (`x = x + 1` reads `x`, and so do `p.x = 1` and `e[0] = 3` for `p` and `e`); a `mutable` parameter that is only assigned counts as used, because the caller sees the write (an output parameter). Unused imports and private functions are not errors yet. The error is reported once per body, at the declaration (the `const`, `let` or `for` keyword, or the name of a parameter or pattern binding), and only when the body has no other error.
- `narrac check --json` ranges are positions only: `range.end` equals `range.start` for a diagnostic (fix edits have real ranges), and `file` is the base name for a single file.
- Enums are stored inline as tagged unions, so an enum cannot contain itself.
- `Str.length` returns `int` instead of `ByteIndex`, and so do all string positions (`Str.indexOf`, `Str.slice`, `Str.nextChar` and the rest). Slices are copies, not views of the original buffer.
- Reference counts are updated on every copy. The Perceus-style removal of unneeded updates is not implemented.
- Generic functions and types are compiled once per set of type arguments in every build. The spec's layout tables for debug builds are not implemented.
- Generic function bodies are checked once with their type parameters left open, then again for each instance.
- One C file per program, not one per module. Declaration C names are `<kind>_<module hash>_<name>`, such as `f_3fa9c1e7_show`: the hash is xxHash32 (seed 0) of the module path relative to the package. Composite types (generic instances, tuples, function types, arrays) are interned as `<kind>_<head>_<hash>`, and a comment table in the C maps each to its Narra type. A single file `hello.nra` is module `"hello"`. A package cannot be named `std`.
- Modules are checked against the declarations of the modules they import, in memory. There are no `.nrai` files.
- Unused imports and private functions are not errors yet.
- Every line of a compile error starts with its file, as in `src/text.nra:3:4: ...`.
- Trait constraints do not compile to a hidden table parameter. Generic functions are compiled once per set of type arguments, so every trait call is a direct call.
- Constraints are allowed on the type parameters of functions only, not of structs and enums.
- `implement` blocks work for non-generic structs and enums only. A generic type gets traits through `derives`, and an instance implements a derived trait when its contents do.
- An `implement` block must be in the module that declares the type. The prelude (`Option`, `Result`, `Ordering`) and primitive and array types belong to no user module, so they cannot get `implement` blocks.
- Traits are imported by name (`import { Area } from "app/shapes"`). `Alias.Trait.f(x)` and `<T: Alias.Trait>` are not supported.
- A bare constructor is E0323 with a fix that writes `Enum.Name`, except `Some`, `None`, `Ok` and `Err`, which stay bare as prelude names. The fix picks the expected type's enum first, then the enum the module declares, imports or the prelude, in that order; when the module cannot name the enum in a pattern, the fix also adds `import { Enum } from "module"`. A bare name shared by two enums of one rank is E0324, with no fix. Two enums may share a constructor name; a duplicate name in one enum is E0194.
- `trait`, `implement`, `implements` and `derives` are keywords, so they cannot be used as names. `Ordering` (with `Less`, `Equal`, `Greater`) and the traits `Equatable`, `Hashable`, `Ordered` and `Show` are prelude names, like `Option`, so a program cannot declare its own type or trait with those names.
- `Hashable.hash(x)` returns `int`. `Hash.of` does not exist yet.
- A function type carries no `escaping` marker, so the parameters of a function value are escaping: an argument passed through a function value must be escaping, and a parameter of an arrow function that has a function type may be stored.
- Only a parameter written with a function type can be non-escaping. A parameter whose type is a type parameter is escaping even when the type argument is a function type, so passing a non-escaping parameter to `id<T>(v: T)` is an error.
- An arrow function with a block body and no result type infers it from its `return` statements (void when none returns a value).
- Derived `Show` writes `Point { x: 1, y: 2 }`, `Some(1)` and `[1, 2]`, and quotes strings and chars inside: `"text"`, `'c'`. `Show.show` on a string quotes it too. A tuple shows as `#[1, "a"]`.
- A name in a pattern is a constructor when the matched value is an enum with that variant. Otherwise it binds a new name, except that a name starting with an uppercase letter is an error. The spec does not say how a pattern tells the two apart.
- `match` accepts any value, not only enums, and so does `is`. A match on `int` or `string` needs a `_` arm or a binding arm.
- `let { a } = s` is allowed and declares `let` variables. The spec shows only `const` for struct destructuring.
- A destructuring pattern must match every value (`const { x: 0 } = p` and `const Some(x) = o` are errors). A constructor works when its enum has one variant: `const Only(s) = u`.
- `Map.keys(m)` and `Map.entries(m)` are arrays (of keys, or of `#[key, value]` tuples) outside a loop, not generators. `for ... of` over them reads the map directly and sees the entries as they were when the loop started. `Map.remove(mutable m, key)` returns `bool`. A `Map` has no `Equatable`, `Hashable` or `Show` yet, so it cannot be compared, used as a key or put in a `derives` struct. A key type must implement `Equatable` and `Hashable` together or neither, and the check looks at the key type and the fields inside it.
