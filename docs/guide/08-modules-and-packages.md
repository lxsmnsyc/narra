# 8. Modules and packages

Every file you have written so far is a module. This chapter starts with imports from the standard library. Then it splits a word counter into a package of three files, and shows how the files share names.

## Importing from the standard library

The word counter splits a text on spaces and counts each word. It needs string, array and number functions, which live in the standard modules `std/string`, `std/array` and `std/int`:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"
import * as Str from "std/string"

struct Tally {
  word: string
  count: int
}

function addWord(mutable tallies: Array<Tally>, word: string) {
  for (mutable tally of tallies) {
    if (tally.word == word) {
      tally.count = tally.count + 1
      return
    }
  }
  Arrays.push(mutable tallies, Tally { word: word, count: 1 })
}

export function run(args: Array<string>): int {
  let tallies: Array<Tally> = []
  const split = Str.split("the cat saw the dog", " ")
  if (split is None) {
    return 1
  }
  for (const word of split.value) {
    addWord(mutable tallies, word)
  }
  for (const tally of tallies) {
    print(`${tally.word} ${Int.toString(tally.count)}`)
  }
  return 0
}
```

```text
the 2
cat 1
saw 1
dog 1
```

`addWord` declares its parameter as `mutable tallies`, so it can change the caller's array, as `Arrays.push` does. The call writes `mutable` too, as in chapter 2. Chapter 9 explains `mutable` parameters.

`import * as Str from "std/string"` makes every exported name of `std/string` reachable through the alias `Str`, as in `Str.split`. You pick the alias. This form is the usual one, because the alias shows at each call which module a function comes from.

When you only need one or two functions, you can import them by name. `as` gives an imported name a different one in your file:

```narra
import { split } from "std/string"
import { toString as intToString } from "std/int"

export function run(args: Array<string>): int {
  const words = split("the cat saw the dog", " ")
  print(intToString(5))
  if (words is Some(pieces)) {
    print(pieces[1])
  }
  return 0
}
```

```text
5
cat
```

Without `as`, `toString` would be a vague name in this file, so the import renames it.

### What needs no import

`print`, `assert`, `panic` and `copy` work in every file. So do the types `Array`, `Map`, `Range`, `Box`, `Option` and `Result`, and the constructors `Some`, `None`, `Ok` and `Err`. Every other name comes from an import.

The functions that work on those types are in their own modules. Writing an array type needs no import, but pushing to an array does. If you forget it, the error gives the line to add:

```narra error
// error: 4:3: Array.push needs an import; add `import * as Arrays from "std/array"`
export function run(args: Array<string>): int {
  let words: Array<string> = []
  Array.push(mutable words, "cat")
  return 0
}
```

The alias `Arrays` is plural because an alias cannot have the same name as a type. `Array` already names the type, so `Array.length` would be unclear:

```narra error
// error: 2:1: the import alias Array is also a type name; rename it, for example `import * as Arrays from "std/array"`
import * as Array from "std/array"

export function run(args: Array<string>): int {
  return Array.length([1, 2])
}
```

The same goes for the other prelude types. Their modules are imported as `Arrays` (`std/array`), `Maps` (`std/map`), `Ranges` (`std/range`), `Boxes` (`std/box`), `Options` (`std/option`) and `Results` (`std/result`).

The compiler accepts 54 standard modules. Importing a path that does not exist is an error that lists them all. Here they are in groups, each with one line. Later chapters teach most of them.

Numbers and text:

- `std/int`: functions for `int` (`i64`), such as `toString`, `parse` and checked arithmetic.
- `std/i8`, `std/i16`, `std/i32`, `std/u8`, `std/u16`, `std/u32` and `std/u64`: the same for each sized integer type, including conversions such as `I32.from`.
- `std/float`: functions for `f64`, such as `fromInt` and `toString`.
- `std/f32`: the same for the 32-bit float.
- `std/math`: the constants `PI` and `E`, and `min`, `max`, `clamp`, `pow` and the other math functions.
- `std/random`: a seeded random number generator that gives the same numbers for the same seed.
- `std/string`: functions for strings, such as `split`, `trim` and `replace`.
- `std/char`: functions for single characters.
- `std/unicode`: grapheme clusters, case mapping, normalization and display width.
- `std/regex`: regular expressions.
- `std/template`: the `Raw` type and the `raw` tag for tagged template literals.

Collections and values:

- `std/array`, `std/map`, `std/range`: functions for the `Array`, `Map` and `Range` types.
- `std/fixedarray`: arrays whose length is part of the type.
- `std/bytes`: a block of raw bytes.
- `std/option`, `std/result`: functions for `Option` and `Result`.
- `std/box`, `std/cell`: a value on the heap, and a shared value that changes on one thread.
- `std/set`: a collection of distinct values.
- `std/iterator`: lazy `map`, `filter`, `take` and other operators over generators.
- `std/arena`: a pool of values that refer to each other by handle.
- `std/hash`: hashing of any value, and a fixed FNV-1a hash of a string.
- `std/shape`: the `Value` tree that serialization reads and writes.
- `std/json`, `std/toml`: parse and write JSON and TOML as a `Value`.
- `std/simd`: vector types that work on several numbers in one instruction.

Files, the system and tools:

- `std/fs`: read and write files and folders.
- `std/path`: join and split file paths.
- `std/io`: read standard input and write standard output and error.
- `std/env`: the program's arguments and environment variables.
- `std/process`: end the program with `Process.exit`, and run another program with `Process.run`.
- `std/time`: instants, durations, the clock and an async sleep.
- `std/args`: a command-line parser.
- `std/terminal`: whether a stream is a terminal, its width and color.
- `std/test`: checks for `test` blocks.
- `std/build`: constants for the target system, such as `Build.os`.

Generators, async and threads:

- `std/coroutine`: `resume` and `send` to run a generator or coroutine.
- `std/async`: wait for several async tasks, or cancel one.
- `std/net`: TCP connections for async code.
- `std/thread`, `std/parallel`: start threads, and run a function over an array on all cores.
- `std/channel`, `std/shared`, `std/atomic`: send values between threads, share a value with a lock, and use atomic numbers.

Unsafe code and C:

- `std/ptr`: raw pointers.
- `std/cstring`: strings for C calls.

`Process.exit(code)` in `std/process` ends the program at once with that exit code. It flushes the output printed so far, has the `io` effect, and does not run the code after it.

## Splitting a program into files

When the counter grows, you want the counting in one file and the command-line code in another. Chapter 1 split a program into a package: a folder with a `narra.toml` and a `src/` folder. Do the same here, with the files `src/cli.nra` and `src/count.nra`.

A file's module name comes from its path in `src/`: `src/count.nra` is the module `"words/count"`, and a file `src/text/split.nra` would be `"words/text/split"`. An import always writes this full path, starting with the package name. A path relative to the current file, such as `"./count"`, is not accepted.

Its `narra.toml` names the package `words` and gives it a `[[binary]]` with `name = "words"` and `entry = "cli.run"`:

```narra file words narra.toml
[package]
name = "words"

[[binary]]
name = "words"
entry = "cli.run"
```

Move the counting into `src/count.nra`:

```narra file words src/count.nra
import * as Arrays from "std/array"
import * as Str from "std/string"

export struct Tally {
  word: string
  count: int
}

export function countWords(text: string): Array<Tally> uses alloc {
  let tallies: Array<Tally> = []
  const split = Str.split(text, " ")
  if (split is None) {
    return tallies
  }
  for (const word of split.value) {
    addWord(mutable tallies, word)
  }
  return tallies
}

function addWord(mutable tallies: Array<Tally>, word: string) {
  for (mutable tally of tallies) {
    if (tally.word == word) {
      tally.count = tally.count + 1
      return
    }
  }
  Arrays.push(mutable tallies, Tally { word: word, count: 1 })
}
```

A name in a module is private to that file unless its declaration starts with `export`. Here `Tally` and `countWords` are exported, and `addWord` stays private, so `count.nra` can change it later without breaking any other file. `countWords` builds an array, so, as in chapter 4, the exported function writes `uses alloc` after its result type. Leave it out and the compiler names the array that allocates. Chapter 10 explains `uses`.

`src/cli.nra` imports what it needs from `"words/count"`. From the `words` folder, `narrac run .` builds and runs the package:

```narra file words src/cli.nra
import * as Int from "std/int"
import { countWords } from "words/count"

export function run(args: Array<string>): int {
  for (const tally of countWords("the cat saw the dog")) {
    print(`${tally.word} ${Int.toString(tally.count)}`)
  }
  return 0
}
```

```text
the 2
cat 1
saw 1
dog 1
```

`narrac build .` builds the program instead of running it. The executable is named `words`, after the `name` of the `[[binary]]`.

### Importing a private name

If `cli.nra` tries to import `addWord`, the compiler points at the import and says what would make it work:

```text
src/cli.nra:2:1: "addWord" is not exported by module "words/count"; add `export` to its declaration
```

Before you add `export`, decide whether other files should depend on that function. Every exported name is something you have to keep working for them.

## What a module can hold

Besides functions and structs, a module can export enums, traits, type aliases and constants. To pass on a name that another module exports, import it and export it again, as [chapter 4](04-structs-and-enums.md#export-a-name-from-another-module) shows. A constant at the top of a file is a `const` whose value the compiler can work out. That can be a number, a string, or an array or struct of them:

```narra
import * as Str from "std/string"

const SEPARATOR = " "
const STOP_WORDS = ["the", "a"]

export function run(args: Array<string>): int {
  const split = Str.split("the cat saw a dog", SEPARATOR)
  if (split is None) {
    return 1
  }
  for (const word of split.value) {
    let skip = false
    for (const stop of STOP_WORDS) {
      if (word == stop) {
        skip = true
      }
    }
    if (!skip) {
      print(word)
    }
  }
  return 0
}
```

```text
cat
saw
dog
```

A variable at the top of a file is not allowed:

```narra error
// error: 2:1: expected a declaration, found "let"
let wordsSeen = 0

export function run(args: Array<string>): int {
  return 0
}
```

A module holds declarations only, so importing it never runs any code and no file can change state that another file reads. To keep a running total, declare the `let` inside a function. Pass it to the functions that update it, as `run` does with `tallies` above. Chapter 9 covers how that works.

## Using a module through an alias

The counter gets a third file, `src/report.nra`, that formats a tally in one of two styles. The styles are an enum:

```narra file words src/report.nra
import * as Int from "std/int"
import * as Str from "std/string"
import { Tally } from "words/count"

export enum Style {
  Numbers
  Bars
}

export function render(tally: Tally, style: Style): string uses alloc {
  return match (style) {
    Style.Numbers => `${tally.word} ${Int.toString(tally.count)}`
    Style.Bars => match (Str.repeat("#", tally.count)) {
      Some(bars) => `${tally.word} ${bars}`
      None => tally.word
    }
  }
}
```

`cli.nra` can import your own modules with an alias, like the standard ones. Through the alias you reach functions, types, and the constructors of an enum:

```narra file words src/cli.nra
import * as Count from "words/count"
import * as Report from "words/report"

export function run(args: Array<string>): int {
  const tallies: Array<Count.Tally> = Count.countWords("the cat saw the dog")
  for (const tally of tallies) {
    print(Report.render(tally, Report.Style.Bars))
  }
  return 0
}
```

```text
the ##
cat #
saw #
dog #
```

`Count.Tally` is the type `Tally` from `"words/count"`, and `Report.Style.Bars` is the constructor `Bars` of the enum `Style`. A constructor is always written with its enum, also inside the module that declares it (`Style.Bars` in `report.nra`). Two enums in different modules can then use the same constructor names without a clash.

### Import cycles

`report.nra` imports `count.nra`. If `count.nra` also imported `report.nra`, neither module could be checked first, and the compiler reports the loop:

```text
src/report.nra:3:1: import cycle: words/count -> words/report -> words/count
```

To break a cycle, find the declarations both files need and move them into a new module that both import.

## Using a package from git

A package can use code that someone else published. A dependency is a git repository with a `narra.toml` of its own. It is pinned to a version tag, so the same tag gives the same code on every machine. There is no central registry.

Suppose the repository `https://example.com/narra-json` holds a package named `json` with the file `src/parse.nra`, and a tag `v1.2.0` marks version 1.2.0. List it in the `[dependencies]` table of your `narra.toml`:

```toml
[package]
name = "words"
version = "0.1.0"

[dependencies]
json = { git = "https://example.com/narra-json", version = "1.2.0" }
```

Each entry is one line. The key is the package name, and it is also the first part of the import path: `src/parse.nra` of `json` is the module `"json/parse"`. The version is the lowest version your package accepts, and its tag is `v` followed by the version. A repository that holds several packages tags each one as `<name>/v<version>`, for example `json/v1.2.0`.

```narra fragment
import { parse } from "json/parse"
```

The first `narrac build`, `narrac run` or `narrac test` does three things:

1. It reads the `[dependencies]` of your package and of each dependency, and picks one version of every package. When two packages ask for different versions of the same one, the higher version wins. Your package never gets two versions of one package. A new major version is a new package name, such as `json2`.
2. It downloads each chosen tag with `git` into a cache folder (`$NARRA_CACHE`, or `~/.cache/narra`) and makes it read-only. A version that is already in the cache is not downloaded again.
3. It writes `narra.lock` next to `narra.toml`.

`narra.lock` records, for each package, the repository, the chosen version, the tag it was fetched from, the commit of the tag and a hash of its files. Commit this file. When the lock and the repository disagree, for example because someone moved a tag, the build stops with an error instead of using different code. A build that changes nothing leaves the file as it was.

A package may import only itself, `std` and the packages in its own `[dependencies]`. If `json` itself depends on `util`, your package still needs its own entry for `util` before it can import `"util/text"`. Without it the build reports that the package has no `[dependencies]` entry for `util`.

A package can choose which modules other packages may import. Add an `[exports]` table to its `narra.toml`. Each key is an import path and each value is the file it stands for:

```toml
[exports]
"json" = "src/index.nra"
"json/stream" = "src/streaming/reader.nra"
```

A key is the package name or starts with `json/`. The bare name `"json"` stands for one module, so `import * as Json from "json"` works. A package that uses `json` can import `"json"` and `"json/stream"`, and nothing else: importing `"json/streaming/reader"` is an error that lists the two paths. Inside `json`, its own files still import each other by their real paths, such as `"json/streaming/reader"`. Without an `[exports]` table, every file is open to importers.

The standard library works the same way. Modules under `std/internal/` and `std/ext/` are not open to your code.

The tests of a dependency belong to its author. `narrac test` runs the tests of your package only.

To use a fork or to pin a version, add an `[overrides]` table to the `narra.toml` of the package you build. An override has the same form as a dependency and replaces the repository and the version wherever any package asks for that name:

```toml
[overrides]
json = { git = "https://example.com/my-fork/narra-json", version = "1.2.1" }
```

Narra trusts the folder where it keeps downloaded packages. The default is `.cache/narra` in your home folder, and `NARRA_CACHE` names another one. When Narra downloads a package, it computes the hash of the files once and writes it next to the copy. Each build compares that hash with `narra.lock`. It does not read the files again, so a file that you or another program changes inside the cache afterwards is built as it is now. If a package contains a symbolic link, Narra follows it. Keep the cache folder writable only by you, and make a package's folder writable with `chmod -R u+w` and delete it if you suspect a change. The next build downloads it again and compares it with the lock.

The editor support (`narrac lsp`) never downloads anything and never changes `narra.lock`. It reads the lock and the cache. Run `narrac build` once after adding a dependency so the editor can find it.

## Workspaces

A workspace is a folder with several packages that you build together. Its `narra.toml` has a `[workspace]` table that lists the member folders. A `*` stands for one whole folder name, and an entry that starts with `!` leaves folders out:

```toml
[workspace]
members = ["packages/*", "!packages/old"]
```

Each member is a package with its own `narra.toml` and a unique name. A member uses another member by name, and `workspace = true` stands in place of a repository and a version:

```toml
[dependencies]
json = { workspace = true }
```

All members share one `narra.lock`, at the workspace root, and one version of each package that comes from git. A member folder holds no lock of its own.

The commands take a folder, and the folder decides how much work they do:

- A member folder (`narrac test packages/json`) works on that member.
- The workspace root (`narrac test .`) works on every member. `build` makes every binary of every member, `test` runs the tests of each member in the order of the `members` entries (the folders that one `*` entry matches are sorted by name), `check` reports the errors of every member, `fmt` formats the files under every member and `query` looks in every member.
- `narrac run -bin tool .` finds the binary named `tool` in whichever member has it. Two members may not have binaries with one name, because a build writes all binaries to one folder.
- With a member folder, `-dependents` adds the members that depend on it, directly or not, and `-dependencies` adds the members it depends on: `narrac test packages/json -dependents` also runs the tests of the members that use `json`. These flags work with `build`, `test` and `check`, and they need a member folder. A folder that is not a member, or the workspace root, is an error.

`narrac fmt` gives a member without a `[fmt]` table the `[fmt]` table of the workspace root, and the defaults when the root has none. A `narra.toml` that has neither a `[package]` nor a `[fmt]` table is skipped when `narrac fmt` looks for its settings.

Next: [9. Memory and `mutable`](09-memory-and-mutable.md).
