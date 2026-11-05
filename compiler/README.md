# narra

`narra` is the bootstrap compiler for Narra, written in Go. It compiles a `.nra` file or a package to C and builds it with the system C compiler. Once the language is stable, the compiler will be rewritten in Narra. The rewrite is in `selfhost/`: its checker and C emitter are done, and `narrac emit` prints the same C as `narra emit` for every test program and for the selfhost package itself. `TestSelfhostEmit` checks this byte for byte.

## Usage

```sh
cd compiler
go build -o narra ./cmd/narra

./narra run testdata/numbers.nra          # build and run
./narra build -o app file.nra             # build a debug executable
./narra build -release -o app file.nra    # optimized, integer overflow wraps
./narra test testdata/tests.nra           # run `test` blocks
./narra emit file.nra                     # print the generated C
./narra check [--json] file.nra          # report errors; --json prints one object per line
./narra explain E0003                    # explain a diagnostic code
./narra build testdata/packages/app       # build every [[binary]] of a package
./narra run -bin app testdata/packages/app
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

`selfhost/bootstrap.sh` runs these steps by hand and needs only `go` and `cc`:

1. Stage 1: `narra build -release -o narrac1 selfhost` (the Go compiler builds the Narra compiler).
2. `narrac1 emit selfhost > stage1.c`, then `cc -O2 -DNARRA_RELEASE -std=c11 -w -o narrac2 stage1.c -lm` (stage 2: the Narra compiler built from its own C).
3. `narrac2 emit selfhost > stage2.c`, then `cmp` Go's `narra emit selfhost`, `stage1.c` and `stage2.c`. The script exits non-zero on any difference.

The fixed point shows that the Narra compiler, compiled by itself, translates the source of the compiler into the same C as the Go compiler does. Go is no longer needed to produce a working compiler from that C. `TestSelfhostBootstrap` does the same under AddressSanitizer, UBSan and the leak check, and also compares stage 2 with Go on every emit input.

## Formatting

`narrac fmt` (the self-hosted compiler, `selfhost/src/fmt.nra`) formats Narra source. The Go compiler has no formatter.

- `narrac fmt [paths]` rewrites the files in place. A directory means every `.nra` file under it (hidden directories are skipped); no path means the current directory.
- `narrac fmt --check [paths]` prints each file that would change and exits 1, or exits 0.
- `narrac fmt --stdin` (or `-`) reads one file from stdin and writes the result to stdout.
- A file that does not parse prints `error <path>:<line>:<col>: <message>`, stays unchanged, and the exit code is 1. A bad `[fmt]` key or value prints `error narra.toml:<line>: <message>` and the exit code is 1.
- Options come from the `[fmt]` table of the nearest `narra.toml` above each file (`printWidth`, `tabWidth`, `useTabs`, `trailingComma`; see `SPEC.md` section 2). Without one the defaults apply.
- Declarations and statements are printed by `fmt.nra`; expressions, types and patterns by `fmtexprs.nra`.
- `TestSelfhostFmt` runs the goldens in `testdata/fmt/` and, for every `.nra` file under `testdata/` and `selfhost/src`, checks that formatting twice gives the same text and that the tokens, the comment texts and the syntax tree are unchanged.
- `TestSourcesFormatted` runs `narrac fmt --check` over `selfhost/src` and every sample under `testdata/` except `testdata/fmt/`, whose inputs are unformatted on purpose, and the samples in `layoutSamples`, which test a layout the formatter would change. A sample that does not parse is skipped only if its path contains `error` or `/diagnostics/`. `bench/` and `std/` are not checked. After editing Narra source, run `narrac fmt` on it. A sample whose `// error:` line or dump golden holds positions must be updated after formatting.

## Tests

```sh
go test ./...
```

- `testdata/*.nra` run and compare their output with the matching `.out` file. Each runs twice: a debug build with AddressSanitizer, UndefinedBehaviorSanitizer and the leak check, and a release build.
- `tests.nra` and `testdata/std_test*.nra` run their `test` blocks (`narra test`) instead of an entry function. `std_test.go` checks failing `std/test` assertions.
- `testdata/panics/` must panic with a given message and exit code 101.
- `testdata/errors/` must fail to compile. The first line, `// error: text`, gives text the error must contain.
- `testdata/packages/*/` with a `run.out` build the package's binary, run it in both modes like `testdata/*.nra`, and run its tests.
- `testdata/packages/errors/*/` must fail to compile. The first line of `narra.toml`, `# error: text`, gives the text.
- Standard modules written in Narra: the Go compiler loads `std/option` (and any std source module that `selfhost/src` imports), because the bootstrap needs it. Every newer std source module is selfhost-only: the Go compiler does not load it, and the Go tests skip the samples that use it. Those samples live in `testdata/stage2/*.nra` (with `.out`) and `testdata/stage2/errors/*.nra` (first line `// error: text`). `TestStage2Programs` and `TestStage2Errors` emit them with the cached stage 2 compiler of `TestSelfhostBootstrap`, compile the C with the sanitizer flags, run it with the leak check and compare the output or the error.
- `std/internal/<name>` modules hold the primitives behind std source modules. Only `std/...` modules import them (E0426). `std/internal/probe` is a test-only primitive (the same code as `Char.code`).

## Supported

- The prelude has the types `Array`, `Map`, `Range`, `Box`, `Option`, `Result` and the functions `print`, `assert`, `panic`. The functions `Array.*`, `Map.*`, `Range.*` and `Box.*` come from `std/array`, `std/map`, `std/range` and `std/box` and need the import (error E0908, with an add-import fix; D67). By convention the aliases are plural: `import * as Arrays from "std/array"`, `Maps`, `Boxes`, `Ranges`, `Options`, `Results`. An alias may not equal a type name, two imports may not share an alias, and `import { map as mapOption }` renames a name (D72).
- `std/option` and `std/result` (`import * as Options from "std/option"`) have the helpers `Option.map`, `andThen`, `withDefault`, `orElse`, `isSome`, `isNone`, `okOr` and `Result.map`, `mapErr`, `andThen`, `withDefault`, `ok`, `err` (D69). The callbacks are non-escaping, their effects flow to the call (`Builtin.Callbacks`), and a capturing arrow function gets a stack environment. There is no `unwrap` or `expect`.
- Packages with several modules and `narra.toml`, or one source file. `import * as X from` and `import { a } from` for `std/int`, `std/float`, `std/string` and the modules of the same package. `export` on functions, structs and enums; other names are private. Import cycles are errors that list the cycle.
- `std/fs` (`readText`, `writeText`, `exists`, `listDir`), `std/process` (`run`) and `std/env` (`get`, `args`). Fallible functions return `Result<T, string>`; the string is `<subject>: <reason>`. `writeText` returns the byte count, `listDir` the sorted names (Err for the whole listing if any entry name is not valid UTF-8), `Process.run(command, args)` returns `#[exit code, stdout, stderr]` (stdin closed; killed by signal gives 128 + signal), `Env.get` returns `Option<string>`, and `Env.args()` equals the entry function's `args`.
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
- Arrow functions `(x: int): int => x + 1` and `(x) => { ... }`, function types `(x: int) => int`, named functions as values (`apply(double, 1)`, `Alias.f`), calls through function values (`f(x)`, `make()(x)`, `button.onClick(x)`), and captures of constants. Parameter types of an arrow function come from the expected type, also through a generic call such as `map(items, (x) => x * 2)`.
- `escaping` function parameters. A non-escaping closure keeps its environment on the caller's stack; an escaping one has a reference-counted heap environment.
- Traits: `trait`, `implements`, `implement Trait for Type`, `derives`, constraints such as `<T: Show & Ordered>`, and calls through the trait name such as `Show.show(x)`.
- The standard traits `Equatable`, `Hashable`, `Ordered` and `Show`, and the `Ordering` enum. `==` and `!=` work on structs, enums and arrays.
- Effects (`check_effects.go`): exported functions must declare `uses alloc, io, unsafe` for everything they do; private functions and arrows are inferred; `<effect E>` and `uses E` on function types work. On an exported function a function-type parameter without `uses` accepts only pure arguments; on a private function it acts as an implicit effect variable. `alloc` also comes from array literals, template strings, string `+` and constructors with fields.
- Opaque structs and distinct types (`check_opaque.go`): `export opaque struct`, `distinct type A = B`, `export opaque distinct type`. Another module cannot build, read, assign, spread or match the fields of an opaque struct. `A.from(x)` and `A.toInt(x)` (`toString` for a string base, named after the base type) are generated and emit no code. Distinct types have no `Show`/`Ordered`. `opaque`, `distinct` and `type` are contextual words, not keywords.
- Generators (`emit_generator.go`, SPEC.md section 8): `function*`, `yield`, `yield*`, `Generator<T> uses E`, generic generators and `for ... of` over a generator. Each `function*` is a state machine over a heap frame with a counted header (`narra_gen` in `runtime.h`): every C variable of the body is a frame field, and the drop and clone functions switch on the state to release or retain what is live at that yield. Every `Generator<T>` is the C type `narra_gen*`. A `for ... of` and a `yield*` resume their own copy, so iterating a generator variable twice runs it twice.

- Top-level `const` (`export const` too): literals, operators on constants, constructors, structs, arrays, maps, function names and other constants, in any order; cycles are errors. Scalar constants are emitted at each use. Other constants are built on first use (see Differences).

## Not supported yet

- `@const` functions and `lazy` values. A top-level constant cannot call a function.
- Dependencies on other packages, `narra.lock`, interface files (`.nrai`) and incremental builds.
- `export type` aliases (so constructors through an alias are not available either).
- `Dynamic<Trait>`, `derives Shape` and `Hash.of`.
- Explicit type arguments in calls, such as `f<int>(x)`.
- Defaults for generic struct fields whose type uses a type parameter.
- Effects: function types carry no effects in the type checker, so calls through struct fields or call results count as pure; trait function calls to user implementations are not followed; `unsafe` has no source yet.
- `@inline` (node inline), `async` arrow functions, and standard functions (`Int.toString`) as values.
- Reading a tuple element by index.
- `Coroutine<In, Out>` and `Co.resume`.
- Async functions, threads and channels.
- `unsafe` and `extern`.
- `std/test` on Windows: `Test.panics` and `Test.panicsWith` use `fork` and `alarm` (D65), so they are POSIX only. A closure that runs past the timeout (10 seconds, or `NARRA_TEST_TIMEOUT`) fails. The test runner (D66) forks per test, so it is POSIX only too. Report formats other than the text lines are not there yet.

## Differences from the spec

- A function that is not a `function*` and returns a generator type writes its own `uses` list after the type's (`(): Generator<T> uses io uses alloc`). A private function's `Generator<T>` parameter accepts only pure generators (D78).
- `==` and `!=` work on a distinct type when its base type supports them (and it can be a map key). SPEC.md says no operators carry over.
- A distinct type cannot list traits with `implements` or `derives`; SPEC.md allows it.
- Top-level constants that are not scalars (strings built by templates, structs, arrays, enums with fields, function values) are built on first use, behind a getter, and kept until the program ends. The spec says modules have no initialization code; this keeps that, but the first-use build is not thread safe yet (the threads node must add a once guard). Struct and array constants are heap values, not static C data.

- `narra check --json` ranges are positions only: `range.end` equals `range.start` for a diagnostic (fix edits have real ranges), and `file` is the base name for a single file.
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
