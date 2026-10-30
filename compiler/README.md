# narra

`narra` is the bootstrap compiler for Narra, written in Go. It compiles a `.nra` file or a package to C and builds it with the system C compiler. Once the language is stable, the compiler will be rewritten in Narra.

## Usage

```sh
cd compiler
go build -o narra ./cmd/narra

./narra run testdata/numbers.nra          # build and run
./narra build -o app file.nra             # build a debug executable
./narra build -release -o app file.nra    # optimized, integer overflow wraps
./narra test testdata/tests.nra           # run `test` blocks
./narra emit file.nra                     # print the generated C
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

## Tests

```sh
go test ./...
```

- `testdata/*.nra` run and compare their output with the matching `.out` file. Each runs twice: a debug build with AddressSanitizer, UndefinedBehaviorSanitizer and the leak check, and a release build.
- `testdata/panics/` must panic with a given message and exit code 101.
- `testdata/errors/` must fail to compile. The first line, `// error: text`, gives text the error must contain.
- `testdata/packages/*/` with a `run.out` build the package's binary, run it in both modes like `testdata/*.nra`, and run its tests.
- `testdata/packages/errors/*/` must fail to compile. The first line of `narra.toml`, `# error: text`, gives the text.

## Supported

- Packages with several modules and `narra.toml`, or one source file. `import * as X from` and `import { a } from` for `std/int`, `std/float`, `std/string` and the modules of the same package. `export` on functions, structs and enums; other names are private. Import cycles are errors that list the cycle.
- `std/fs` (`readText`, `writeText`, `exists`, `listDir`), `std/process` (`run`) and `std/env` (`get`, `args`). Fallible functions return `Result<T, string>`; the string is `<subject>: <reason>`. `writeText` returns the byte count, `listDir` the sorted names (Err for the whole listing if any entry name is not valid UTF-8), `Process.run(command, args)` returns `#[exit code, stdout, stderr]` (stdin closed; killed by signal gives 128 + signal), `Env.get` returns `Option<string>`, and `Env.args()` equals the entry function's `args`.
- `std/string` also has `slice`, `sliceFrom`, `byteAt`, `charAt`, `nextChar`, `charCount`, `indexOf` (returns `Option<int>`), `startsWith`, `trim` (ASCII whitespace), `split` (panics on an empty separator), `chars` (`Array<char>`) and a `StringBuilder` (`Str.builder()`, `Str.append(mutable sb, s)`, `Str.appendChar(mutable sb, c)`, `Str.build(sb)`). Slicing, `charAt` and `nextChar` panic on a position out of range or inside a character.
- Names from a namespace import: calls `X.f(...)` and types `X.T`. Named imports bring functions, and structs and enums with their constructors.
- Functions, `const`, `let`, `if`, `while`, `for (const x of Range.upTo(n))`, `Range.between(a, b)` and `Array.values(items)`, `break`, `continue`.
- `int`, `f64`, `bool`, `char` and `string`.
- Structs with constant field defaults, typed literals (`Point { x: 1 }`) and literals typed from context.
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
- `test` blocks and `assert`.
- Reference counting for strings, arrays, and structs and enums that hold them.
- Arrow functions `(x: int): int => x + 1` and `(x) => { ... }`, function types `(x: int) => int`, named functions as values (`apply(double, 1)`, `Alias.f`), calls through function values (`f(x)`, `make()(x)`, `button.onClick(x)`), and captures of constants. Parameter types of an arrow function come from the expected type, also through a generic call such as `map(items, (x) => x * 2)`.
- `escaping` function parameters. A non-escaping closure keeps its environment on the caller's stack; an escaping one has a reference-counted heap environment.
- Traits: `trait`, `implements`, `implement Trait for Type`, `derives`, constraints such as `<T: Show & Ordered>`, and calls through the trait name such as `Show.show(x)`.
- The standard traits `Equatable`, `Hashable`, `Ordered` and `Show`, and the `Ordering` enum. `==` and `!=` work on structs, enums and arrays.
- Effects (`check_effects.go`): exported functions must declare `uses alloc, io, unsafe` for everything they do; private functions and arrows are inferred; `<effect E>` and `uses E` on function types work. On an exported function a function-type parameter without `uses` accepts only pure arguments; on a private function it acts as an implicit effect variable. `alloc` also comes from array literals, template strings, string `+` and constructors with fields.

## Not supported yet

- Dependencies on other packages, `narra.lock`, interface files (`.nrai`) and incremental builds.
- `export const`, `export type` aliases, and enum constructors through a namespace (`X.Token.Word`).
- `Dynamic<Trait>`, `derives Shape` and `Hash.of`.
- Explicit type arguments in calls, such as `f<int>(x)`.
- Defaults for generic struct fields whose type uses a type parameter.
- Effects: function types carry no effects in the type checker, so calls through struct fields or call results count as pure; trait function calls to user implementations are not followed; `unsafe` has no source yet.
- `@inline` (node inline), `async` arrow functions, and standard functions (`Int.toString`) as values.
- Destructuring in `for (const #[i, x] of ...)` loops, and reading a tuple element by index.
- Generators, async functions, threads and channels.
- Opaque and distinct types, `unsafe` and `extern`.

## Differences from the spec

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
- A bare constructor name resolves from the expected type first. Otherwise the enums visible in the module are searched in three ranks: those the module declares, then those it imports by name, then the prelude. The first rank with a match wins, so a user `Less` is not `Ordering.Less`. Two enums of the same rank sharing the name are an ambiguity error; write `Enum.Name`.
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
