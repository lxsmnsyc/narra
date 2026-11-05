# 1. Getting started

In this chapter you build the compiler, run a first program and fix your first compile error. Then you turn the program into a project with two files.

## Build the compiler

Narra programs compile to C, and a C compiler turns that C into an executable. The compiler, `narra`, is written in Go. You need Go and a C compiler (`cc`, `clang` or `gcc`) installed.

```sh
git clone https://github.com/lxsmnsyc/narra
cd narra/compiler
go build -o narra ./cmd/narra
```

This leaves a `narra` executable in `narra/compiler`. Put that folder on your `PATH` so you can type `narra` from anywhere, as the rest of the guide does.

## Say hello

Create a file named `hello.nra` and run it with `narra run hello.nra`. Each program in this guide is followed by what it prints:

```narra
export function run(args: Array<string>): int {
  print("hello, narra")
  return 0
}
```

```text
hello, narra
```

`narra run` compiled the file, built an executable, and started it. The executable called the function named `run`, so every program has one. Here is what each part of it does:

- `export` makes `run` visible outside the file, which is how the compiler finds it.
- `args: Array<string>` is the command-line arguments, as a list of strings. The first one is the path of the program.
- `: int` after the parentheses is the type of value the function returns. `run` returns the exit code, and `0` means success.
- `print` writes a line of text to standard output.

A statement ends at the end of its line, so you write no `;` after it.

## Print a number

Next, make the program print the result of a sum. The obvious version does not compile:

```narra error
// error: 3:11: expected string, found int
export function run(args: Array<string>): int {
  print(1 + 2)
  return 0
}
```

The first line is a comment. `//` starts a note for the reader that runs to the end of the line, and the compiler skips it. In this guide, such a comment on the first line of a sample shows the message the compiler prints for that file. `3:11` is the line and column of the problem, with the comment counted as line 1: `print` wants a `string`, and `1 + 2` is an `int`. Narra never turns a number into text for you, so you convert it yourself with `Int.toString`:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  // Int.toString turns a number into text.
  print(Int.toString(1 + 2))
  return 0
}
```

```text
3
```

`Int.toString` lives in the standard module `std/int`. The `import` line makes the functions of that module available under the name `Int`. Only `print`, `assert` and `panic` are there without an import. Everything else, including the functions for strings, arrays and maps, comes from a module you import.

When you are only looking for mistakes, `narra check hello.nra` reports the errors without building anything.

## Build an executable

`narra run` builds into a temporary place. To keep the executable, use `narra build`:

```sh
narra build -o hello hello.nra
./hello
```

This is a debug build. It stops the program with a message when an integer calculation overflows, and running it with `NARRA_CHECK_LEAKS=1` set reports memory that was never freed. Add `-release` for an optimized build, in which integer overflow wraps around instead. In both builds, dividing by zero or reading past the end of an array stops the program with a message.

## Split the program into files

Once a program grows, you want it in several files. A project of several files is a package: a folder with a `narra.toml` file and a `src` folder for the code.

```text
shop/
  narra.toml
  src/
    main.nra
    prices.nra
```

`narra.toml` names the package and says which function starts the program:

```toml
[package]
name = "shop"

[[binary]]
name = "shop"
entry = "main.run"
```

`entry = "main.run"` means the function `run` in `src/main.nra`. Each file in `src` is a module, named after the package and the file: `src/prices.nra` is the module `"shop/prices"`. Put a function in it and export it:

```narra fragment
export function withTax(cents: int): int {
  return cents + cents / 10
}
```

Then `src/main.nra` imports that function by name:

```narra fragment
import * as Int from "std/int"
import { withTax } from "shop/prices"

export function run(args: Array<string>): int {
  print(Int.toString(withTax(200)))
  return 0
}
```

From inside the `shop` folder, `narra run .` builds and runs the package and prints `220`. `narra build .` writes the executable `shop`, named after the `[[binary]]`. Chapter 8 covers modules and imports in full.

## The commands

| Command | What it does |
| --- | --- |
| `narra run file.nra` | build and run a program |
| `narra build -o app file.nra` | build a debug executable |
| `narra build -release -o app file.nra` | build an optimized executable |
| `narra check file.nra` | report errors without building |
| `narra test file.nra` | run the tests in the file (chapter 10) |

Each command also takes a package folder in place of a file.

Next: [2. Values and types](02-values-and-types.md).
