# langc

`langc` is the bootstrap compiler for Lang, written in Go. It compiles one `.lang` file to C and builds it with the system C compiler. Once the language is stable, the compiler will be rewritten in Lang.

## Usage

```sh
cd compiler
go build -o langc ./cmd/langc

./langc run testdata/numbers.lang          # build and run
./langc build -o app file.lang             # build a debug executable
./langc build -release -o app file.lang    # optimized, integer overflow wraps
./langc test testdata/tests.lang           # run `test` blocks
./langc emit file.lang                     # print the generated C
```

The program's entry function is `export function run(args: Array<string>): int`. `lang.toml` is not read yet.

Set `LANG_CHECK_LEAKS=1` when running a program to report allocations that are still live at exit. The program exits with code 102 when any are found.

## Tests

```sh
go test ./...
```

- `testdata/*.lang` run and compare their output with the matching `.out` file. Each runs twice: a debug build with AddressSanitizer, UndefinedBehaviorSanitizer and the leak check, and a release build.
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
- Template strings.
- Integer overflow checks in debug builds, and division by zero and bounds checks in every build.
- `test` blocks and `assert`.
- Reference counting for strings, arrays, and structs and enums that hold them.

## Not supported yet

- Several modules, and `lang.toml`.
- Generics, traits, `Option`, `Result` and `is`.
- Closures and arrow functions.
- Effects are parsed but not checked.
- Struct spread updates, destructuring, `==` on structs and enums.
- Generators, async functions, threads and channels.
- Opaque and distinct types, `unsafe` and `extern`.

## Differences from the spec

- Enums are stored inline as tagged unions, so an enum cannot contain itself.
- `Str.length` returns `int` instead of `ByteIndex`.
- Reference counts are updated on every copy. The Perceus-style removal of unneeded updates is not implemented.
