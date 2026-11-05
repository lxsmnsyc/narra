# 1. Getting started

Narra is a statically typed language with enums that carry data, pattern matching, no `null` and no exceptions. It compiles to C, and a C compiler builds the program. This chapter builds and runs a first program, then a package with a `narra.toml`.

You need Go (to build the bootstrap compiler `narra`) and a C compiler (`cc`, `clang` or `gcc`).

```sh
cd compiler
go build -o narra ./cmd/narra
```

## A first program

A single file is a program when it exports a function named `run`. The function takes the command-line arguments and returns the exit code.

```narra
export function run(args: Array<string>): int {
  print("hello, narra")
  return 0
}
```

```text
hello, narra
```

Save it as `hello.nra` and run it:

```sh
./narra run hello.nra
```

`print`, `assert` and `panic` are the only functions you can call without an import. Everything else comes from a module you import (chapter 8).

What to notice:

- There are no semicolons. A line break ends a statement, and a `;` is a compile error.
- Every named function writes its parameter and return types. A function with no return type returns nothing.
- `narra fmt` defines the one layout, so there is nothing to configure or argue about.

## Commands

| Command | What it does |
| --- | --- |
| `narra run file.nra` | build and run |
| `narra build -o app file.nra` | build a debug executable |
| `narra build -release -o app file.nra` | build an optimized executable |
| `narra check file.nra` | report errors without building |
| `narra test file.nra` | run the `test` blocks of the file |
| `narra emit file.nra` | print the generated C |

A debug build checks integer overflow, and a program run with `NARRA_CHECK_LEAKS=1` reports memory that was never freed. A release build wraps on integer overflow. Division by zero and out-of-range indexes panic in both.

## Errors say what to write instead

A compile error names the line and column, and says what to write instead when it can.

```narra error
// error: 3:11: semicolons are not used
export function run(args: Array<string>): int {
  return 0;
}
```

Run `narra explain E0003` to read about a diagnostic code, or `narra check --json file.nra` to get one JSON object per diagnostic, with the fix as an edit.

## A package

A program with several files is a package: a directory with a `narra.toml` and a `src/` folder.

```text
app/
  narra.toml
  src/
    cli.nra
    text.nra
```

```toml
[package]
name = "app"

[[binary]]
name = "app"
entry = "cli.run"
```

The file `src/text.nra` of package `app` is the module `"app/text"`. The `entry` names a module inside the package, then its function. Build every binary of a package, or run one, by naming the directory:

```sh
./narra build app
./narra run -bin app app
```

Chapter 8 covers imports and exports. `narra.toml` also has a `[fmt]` table for the formatter.

## Tests

A `test` block lives next to the code, and `narra test` runs each block in its own process. Chapter 10 covers `std/test` in full.

```sh
./narra test file.nra
```

Next: [2. Values and types](02-values-and-types.md).
