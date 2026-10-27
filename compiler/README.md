# narra

`narra` is the bootstrap compiler for Narra, written in Go. It compiles one `.nra` file to C and builds it with the system C compiler. Once the language is stable, the compiler will be rewritten in Narra.

## Usage

```sh
cd compiler
go build -o narra ./cmd/narra

./narra run testdata/numbers.nra          # build and run
./narra build -o app file.nra             # build a debug executable
./narra build -release -o app file.nra    # optimized, integer overflow wraps
./narra test testdata/tests.nra           # run `test` blocks
./narra emit file.nra                     # print the generated C
```

The program's entry function is `export function run(args: Array<string>): int`. `narra.toml` is not read yet.

Set `NARRA_CHECK_LEAKS=1` when running a program to report allocations that are still live at exit. The program exits with code 102 when any are found.

## Tests

```sh
go test ./...
```

- `testdata/*.nra` run and compare their output with the matching `.out` file. Each runs twice: a debug build with AddressSanitizer, UndefinedBehaviorSanitizer and the leak check, and a release build.
- `testdata/panics/` must panic with a given message and exit code 101.
- `testdata/errors/` must fail to compile. The first line, `// error: text`, gives text the error must contain.

## Supported

- One source file per program, with `import * as X from` and `import { a } from` for `std/int`, `std/float` and `std/string`.
- Functions, `const`, `let`, `if`, `while`, `for (const x of Range.upTo(n))`, `Range.between(a, b)` and `Array.values(items)`, `break`, `continue`.
- `int`, `f64`, `bool`, `char` and `string`.
- Structs with constant field defaults, typed literals (`Point { x: 1 }`) and literals typed from context.
- Enums with fields, and `match` as a statement or a value, with guards and `_`.
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

## Not supported yet

- Several modules, and `narra.toml`.
- Traits and trait constraints such as `<T: Equatable>`.
- Nested patterns such as `Some(Node(a, b))`, and struct patterns.
- Explicit type arguments in calls, such as `f<int>(x)`.
- Defaults for generic struct fields whose type uses a type parameter.
- Closures and arrow functions.
- Effects are parsed but not checked.
- Struct spread updates, destructuring, `==` on structs and enums.
- Generators, async functions, threads and channels.
- Opaque and distinct types, `unsafe` and `extern`.

## Differences from the spec

- Enums are stored inline as tagged unions, so an enum cannot contain itself.
- `Str.length` returns `int` instead of `ByteIndex`.
- Reference counts are updated on every copy. The Perceus-style removal of unneeded updates is not implemented.
- Generic functions and types are compiled once per set of type arguments in every build. The spec's layout tables for debug builds are not implemented.
- Generic function bodies are checked once with their type parameters left open, then again for each instance.
