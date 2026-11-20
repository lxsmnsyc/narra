# 18. Tools

The earlier chapters used `narrac run`, `narrac build`, `narrac check` and `narrac test`. `narrac` has more commands for the work around writing code. They read and fix errors, keep the layout tidy, answer questions about a package, and talk to an editor. This chapter goes through each one with a small package.

## The package

The package converts temperatures, as in chapter 10. It has the usual `narra.toml`:

```narra file weather narra.toml
[package]
name = "weather"

[[binary]]
name = "weather"
entry = "main.run"
```

`src/convert.nra` holds the conversions and two tests:

```narra file weather src/convert.nra
import * as Int from "std/int"
import * as Test from "std/test"

export function toFahrenheit(celsius: int): int {
  return celsius * 9 / 5 + 32
}

export function describe(celsius: int): string uses alloc {
  return `${Int.toString(celsius)} C is ${Int.toString(toFahrenheit(celsius))} F`
}

export function hottest(readings: Array<int>): int {
  let best = readings[0]
  for (const reading of readings) {
    if (reading > best) {
      best = reading
    }
  }
  return best
}

test "water freezes at 32 F" {
  Test.eq(toFahrenheit(0), 32)
}

test "hottest picks the largest reading" {
  Test.eq(hottest([18, 25, 21]), 25)
}
```

`src/main.nra` prints two readings:

```narra file weather src/main.nra
import { describe, hottest } from "weather/convert"

export function run(args: Array<string>): int {
  print(describe(21))
  print(describe(hottest([18, 25, 21])))
  return 0
}
```

```text
21 C is 69 F
25 C is 77 F
```

Every command below runs from inside the `weather` folder, and `.` names the package. A command that takes a path also takes a single `.nra` file, except `narrac query`, which needs a package.

## Building, running and looking at the C

Chapter 1 introduced `narrac build` and `narrac run`. A few options matter once a package has more than one binary or you want to see what the compiler wrote. `-bin NAME` picks one `[[binary]]`: without it, `build` makes every binary of the package, and `run` needs the package to have exactly one. `-o PATH` names the executable `build` writes. Words after the path go to the program as its arguments, so `narrac run -bin weather . celsius` starts the `weather` binary with `args[1]` set to `celsius`.

`narrac emit` prints the C code of a debug build on the standard output and builds nothing:

```sh
narrac emit . > weather.c
```

The file starts with the runtime that every program shares, followed by the code of the package. `narrac emit --tests .` adds the test blocks, which `build` leaves out, and `--bin NAME` picks one binary. For the C of an optimized build, build with `-emit-c`: `narrac build -release -emit-c weather.c .` writes the file and the executable. To read the C of one function, search the file for its name: `toFahrenheit` is `f_` followed by a hash of its module, then `_toFahrenheit`.

`narrac targets` lists the systems that `-target` and `check --target` accept, each as a name and the triple of the C compiler that builds for it:

```sh
narrac targets
```

```text
MacOS-Arm64 aarch64-macos
MacOS-X86_64 x86_64-macos
Linux-X86_64 x86_64-linux-gnu
Linux-Arm64 aarch64-linux-gnu
Linux-X86_64-musl x86_64-linux-musl
Linux-Arm64-musl aarch64-linux-musl
Windows-X86_64 x86_64-windows-gnu
Wasi-Wasm32 wasm32-wasi
```

Chapter 19 builds for them. `narrac audit PATH` lists the `unsafe` blocks and `@trusted` functions of a package, as chapter 16 shows. The `weather` package has none, so it prints nothing.

## Checking without building

`narrac check` loads and checks a program and builds nothing, so it is the fastest way to find mistakes:

```sh
narrac check .
```

It prints nothing and exits with code 0 when the program is correct. When it finds errors, it prints one line for each on the standard error, as `file:line:column: message`, and exits with code 1. A wrong command line, such as a missing path, exits with code 2. A script can stop on the code without reading the text.

Here is a file with a mistake. Save it as `hello.nra`, first comment line included, so that the line numbers below match:

```narra error
// error: 3:17: semicolons are not used; end the statement with a line break
export function run(args: Array<string>): int {
  print("hello");
  return 0
}
```

```sh
narrac check hello.nra
```

```text
hello.nra:3:17: semicolons are not used; end the statement with a line break
```

A program that reads which system it is built for can differ from one system to another. `narrac check --all-targets` checks it for every system, as [chapter 19](19-building-for-other-systems.md#functions-that-exist-on-one-system) explains.

### Errors as JSON

The text form is for people. A program that reads the errors, such as an editor or a script, should ask for JSON with `--json`:

```sh
narrac check --json hello.nra
```

```json
{"code":"E0003","severity":"error","message":"semicolons are not used; end the statement with a line break","file":"hello.nra","range":{"start":[3,17],"end":[3,17]},"fixes":[{"title":"remove the semicolon","edits":[{"file":"hello.nra","range":{"start":[3,17],"end":[3,18]},"text":""}]}]}
```

Each error is one JSON object on a line of its own, on the standard output. The fields are:

| Field | What it holds |
| --- | --- |
| `code` | the error's code, such as `E0003`; the same mistake always has the same code |
| `severity` | `error` |
| `message` | the text the text form prints |
| `file` | the file, relative to the package folder |
| `range` | where the problem is, as `[line, column]` pairs, both counted from 1 |
| `fixes` | the changes that would fix the error, often none |

A fix has a `title` and a list of `edits`. Each edit replaces the text from `start` up to, but not including, `end` with `text`. Here the one edit replaces the `;` at line 3, column 17 with nothing.

## Explaining an error

The code from the JSON form names the error. `narrac explain` prints what it means:

```sh
narrac explain E0003
```

```text
E0003: semicolons are not used; end the statement with a line break

Narra has no semicolons. A line break ends a statement.
Fix: delete the `;`.
Example: `return 0;` becomes `return 0`.
```

The first line is the message. Where the message names something from your code, it has a placeholder such as `%s`. Some codes have a longer explanation below it, with a fix and an example. Others have only the message line. The code may be written in small letters, `narrac explain e0003`. An unknown code prints `unknown diagnostic code "E9999"` and exits with code 1.

## Applying a fix

`narrac apply-fix` applies one of those fixes and prints the whole file with the change made:

```sh
narrac apply-fix hello.nra
```

```narra fragment
// error: 3:17: semicolons are not used; end the statement with a line break
export function run(args: Array<string>): int {
  print("hello")
  return 0
}
```

The file on disk does not change. To keep the result, write it to a new file and move it over the old one:

```sh
narrac apply-fix hello.nra > fixed.nra
mv fixed.nra hello.nra
```

Do not redirect the output straight to `hello.nra`. The shell empties the file before `narrac` reads it.

For a package, it prints the file of the package that the fix changes. Without options, `narrac apply-fix` takes the first error that has a fix, and that error's first fix. To pick another, give the error's place in the list with `--diagnostic N` and the fix's place with `--fix M`. Both count from 0, in the order `narrac check --json` prints them. When there is no such error or fix, the command prints `no such diagnostic` or `no such fix` and exits with code 1. A file with no errors prints `no diagnostics`.

## Formatting

`narrac fmt` rewrites files in one standard layout. It sets the indentation, line breaks and spaces, and by default puts a comma after the last item of a list split over lines. When every file uses the same layout, a change to the code shows only what changed.

```sh
narrac fmt .
```

A folder means every `.nra` file inside it, at any depth. With no path at all, `narrac fmt` formats the current folder. It writes the files in place and prints nothing.

The layout follows the code, not how it was typed. Here the block of the `if` is written on one line, which is valid Narra:

```narra fragment
export function hottest(readings: Array<int>): int {
  let best = readings[0]
  for (const reading of readings) {
    if (reading > best) { best = reading }
  }
  return best
}
```

After `narrac fmt`, that block is on lines of its own, as in `src/convert.nra` above.

Two options are for scripts:

- `narrac fmt --check .` changes nothing. It prints the path of each file that is not formatted and exits with code 1 when there is one. A check before a commit can stop on that.
- `narrac fmt --stdin` reads one file from the standard input and writes the formatted text to the standard output. `-` works the same. An editor can format the text it holds without saving it first.

A file that does not parse is left as it is. `narrac fmt` prints `error`, the path, the line and column and the message, goes on with the other files, and exits with code 1.

### Choosing the layout

A `[fmt]` table in `narra.toml` changes the options for every file of the package:

```toml
[fmt]
printWidth = 80
tabWidth = 4
useTabs = false
trailingComma = "none"
```

| Key | Default | What it sets |
| --- | --- | --- |
| `printWidth` | `100` | the line length `narrac fmt` tries to stay within |
| `tabWidth` | `2` | the spaces per level of indentation |
| `useTabs` | `false` | indent with tabs instead of spaces |
| `trailingComma` | `"all"` | `"all"` adds a comma after the last item of a split list, `"none"` leaves it out |

Each file takes the `narra.toml` nearest to it: in its own folder, or in the closest folder above it. A file with no `narra.toml` above it gets the defaults. Any other key in `[fmt]`, or a value of the wrong type, is an error that names the line, such as `error narra.toml:10: unknown key indent in [fmt]`. The `[fmt]` table changes only formatting: `narrac build` and `narrac run` skip it.

Some lines stay long, whatever the width. A template string is never split, so line 9 of `src/convert.nra` stays on one line even with `printWidth = 60`.

## Asking questions about a package

`narrac query` answers questions about the code of a package, from the results of checking it. It is for finding your way in a package you do not know, and for tools. Each question is a word and what it is about:

| Question | What it answers |
| --- | --- |
| `type-of FILE:LINE:COL` | the type at a place in a file, and the effects of evaluating it |
| `def NAME` | where a name is declared |
| `refs NAME` | every use of a function, constant or enum variant |
| `callers FUNCTION` | every call of a function, and the function each call is in |
| `interface MODULE` | the exported declarations of a module, as signatures |
| `recursion MODULE` | the groups of functions in a module that call each other in a cycle |
| `hovers FILE` | the hover text of every name in a file, and its errors, in one answer |

`FILE` is the path from the package folder, as errors print it. Line and column count from 1. The local `best` starts at line 13, column 7 of `src/convert.nra`:

```sh
narrac query type-of src/convert.nra:13:7
```

```text
int
```

A place inside a call gives the call's result and its effects. Line 5, column 17 of `src/main.nra` is the `(` of `describe(...)`:

```sh
narrac query type-of src/main.nra:5:17
```

```text
string
effects: alloc
```

A name can be bare, `hottest`, or qualified with its module, `weather/convert.hottest`. When a bare name belongs to more than one module, `def` lists each match, and `refs` and `callers` ask for the qualified name.

```sh
narrac query def hottest
```

```text
weather/convert.hottest function src/convert.nra:12:17
```

`callers` lists each call on its own line, with the place and the calling function. A call from a test block names the test:

```sh
narrac query callers toFahrenheit
```

```text
src/convert.nra:9:56 weather/convert.describe
src/convert.nra:23:11 weather/convert.test "water freezes at 32 F"
note: calls through function values are not counted
```

The note means what it says: a function passed as a value and called later is not in the list.

`interface` shows what a module offers to other modules, without the bodies:

```sh
narrac query interface weather/convert
```

```narra fragment
export function toFahrenheit(celsius: int): int
export function describe(celsius: int): string uses alloc
export function hottest(readings: Array<int>): int
```

`recursion weather/convert` prints `(none)`, because no function there calls itself, directly or through other functions.

By default the query runs on the package in the current folder. `--package DIR` names another one. `--json` prints the answer as one JSON object, for tools:

```sh
narrac query --json type-of src/main.nra:5:17
```

```json
{"query": "type-of", "location": {"file": "src/main.nra", "range": {"start": [5, 17], "end": [5, 18]}}, "type": "string", "effects": ["alloc"]}
```

`hovers` also works on a file with errors: it lists the errors and the names the checker could type. This holds for errors in that file. When another file of the package fails to load, `hovers` prints no hovers and exits with code 1. With `--json`, each hover has `line`, `column` and `length` (in characters) and `text`, and each error has the same three and a `code` and `message`.

A query exits with code 0 when it has an answer, an empty one included. An unknown name, a place with no expression, or a package with errors exits with code 1. With `--json`, the errors of a package that does not check come as `{"errors": [...]}`, each one in the form of `narrac check --json`.

## Running tests

Chapter 10 introduced `narrac test`. On a package, it runs the test blocks of every module, in order:

```sh
narrac test .
```

```text
ok - water freezes at 32 F
ok - hottest picks the largest reading
2 passed, 0 failed
```

The options are those of chapter 10. `-filter TEXT` runs the tests whose name contains `TEXT`, and `-time` adds how long each one took (the time differs from run to run):

```sh
narrac test -time -filter hottest .
```

```text
ok - hottest picks the largest reading (1 ms)
1 passed, 0 failed
```

A test that calls `Test.golden` (chapter 10) reads its expected text from the `testdata` folder of the package. `narrac test` finds the folder itself, and `NARRA_UPDATE_GOLDEN=1 narrac test .` rewrites the files.

`narrac test`, `run` and `build` also take `-release` for an optimized build. `-sanitize` builds with AddressSanitizer, which stops the program at the first bad memory access. `run` and `test` take `-stdin FILE` to give the program that file as its standard input. Without it, the standard input is closed.

## Using an editor

`narrac lsp` is a language server: a program that an editor starts in the background and talks to while you type. The editor and the server speak the Language Server Protocol over the server's standard input and output. Any editor with a client for that protocol can use it.

The server gives the editor:

- the errors of each open file, as you type, without saving
- formatting, the same as `narrac fmt`
- hover: the type and effects of the name or expression under the cursor
- go to definition, and find references
- the outline of a file: its functions, types and constants

A file inside a package is checked as part of that package. The text of each open file takes the place of what is on disk. A file with no `narra.toml` above it is checked alone.

The editor needs three things: the command, `narrac lsp`, the file extension, `.nra`, and the file that marks the top of a package, `narra.toml`. In Neovim 0.11 or later, put this in `init.lua`:

```lua
vim.filetype.add({ extension = { nra = "narra" } })
vim.lsp.config("narra", {
  cmd = { "narrac", "lsp" },
  filetypes = { "narra" },
  root_markers = { "narra.toml" },
})
vim.lsp.enable("narra")
```

In Helix, put this in `languages.toml`:

```toml
[language-server.narrac]
command = "narrac"
args = ["lsp"]

[[language]]
name = "narra"
scope = "source.narra"
file-types = ["nra"]
roots = ["narra.toml"]
language-servers = ["narrac"]
```

Editors that need an extension for each language have no Narra extension yet.

Two options help when the server seems wrong. By default it checks again only the modules a change can affect. `narrac lsp --full` checks the whole package on every change, which is slower but rules out a mistake in that shortcut. `narrac lsp --trace` writes a line to the standard error for each check, saying what it checked. The server writes nothing else there, and only protocol messages to the standard output.

## The commands

| Command | What it does |
| --- | --- |
| `narrac build [-release] [-sanitize] [-target NAME] [-bin NAME] [-o PATH] [-emit-c FILE] PATH` | build the executables of a package or file |
| `narrac run [-release] [-sanitize] [-bin NAME] [-stdin FILE] PATH [ARGS...]` | build and run, with the arguments for the program |
| `narrac emit [--tests] [--bin NAME] PATH` | print the C of a debug build |
| `narrac check [--json] [--all-targets] [--target NAME] PATH` | report errors without building |
| `narrac explain CODE` | explain an error code |
| `narrac apply-fix [--diagnostic N] [--fix M] PATH` | print a file with a fix applied |
| `narrac fmt [--check] [PATH...]` | format files in place, or list the ones that are not formatted |
| `narrac fmt --stdin` | format the standard input to the standard output |
| `narrac query [--package DIR] [--json] QUESTION` | answer a question about a package |
| `narrac test [-filter TEXT] [-time] [-release] [-sanitize] [-stdin FILE] PATH` | run the test blocks |
| `narrac audit [PATH]` | list the `unsafe` blocks and `@trusted` functions |
| `narrac targets` | list the systems `-target` accepts |
| `narrac lsp [--full] [--trace]` | start the language server for an editor |

Next: [19. Building for other systems](19-building-for-other-systems.md).
