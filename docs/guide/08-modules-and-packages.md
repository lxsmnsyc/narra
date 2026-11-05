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
  for (const #[index, tally] of Arrays.entries(tallies)) {
    if (tally.word == word) {
      tallies[index].count = tally.count + 1
      return
    }
  }
  Arrays.push(mutable tallies, Tally { word: word, count: 1 })
}

export function run(args: Array<string>): int {
  let tallies: Array<Tally> = []
  for (const word of Str.split("the cat saw the dog", " ")) {
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

`import * as Str from "std/string"` makes every exported name of `std/string` reachable through the alias `Str`, as in `Str.split`. You pick the alias. This form is the usual one, because the alias shows at each call which module a function comes from.

When you only need one or two functions, you can import them by name. `as` gives an imported name a different one in your file:

```narra
import { split } from "std/string"
import { toString as intToString } from "std/int"

export function run(args: Array<string>): int {
  const words = split("the cat saw the dog", " ")
  print(intToString(5))
  print(words[1])
  return 0
}
```

```text
5
cat
```

Without `as`, `toString` would be a vague name in this file, so the import renames it.

### What needs no import

`print`, `assert` and `panic` work in every file. So do the types `Array`, `Map`, `Range`, `Box`, `Option` and `Result`, and the constructors `Some`, `None`, `Ok` and `Err`. Every other name comes from an import.

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

The other standard modules are `std/char`, `std/env`, `std/float`, `std/fs`, `std/int`, `std/process`, `std/string` and `std/test`. Importing a path that does not exist is an error that lists them all.

## Splitting a program into files

When the counter grows, you want the counting in one file and the command-line code in another. Several files that build together are a package. A package is a folder with a `narra.toml` file and a `src/` folder:

```text
words/
  narra.toml
  src/
    cli.nra
    count.nra
```

`narra.toml` gives the package a name and says where its program starts:

```toml
[package]
name = "words"

[[binary]]
name = "words"
entry = "cli.run"
```

`entry = "cli.run"` means the function `run` in `src/cli.nra`. Each file's name inside the package comes from its path: `src/count.nra` is the module `"words/count"`, and a file `src/text/split.nra` would be `"words/text/split"`. An import always writes this full path, starting with the package name. A path relative to the current file, such as `"./count"`, is not accepted.

Move the counting into `src/count.nra`:

```narra fragment
import * as Arrays from "std/array"
import * as Str from "std/string"

export struct Tally {
  word: string
  count: int
}

export function countWords(text: string): Array<Tally> {
  let tallies: Array<Tally> = []
  for (const word of Str.split(text, " ")) {
    addWord(mutable tallies, word)
  }
  return tallies
}

function addWord(mutable tallies: Array<Tally>, word: string) {
  for (const #[index, tally] of Arrays.entries(tallies)) {
    if (tally.word == word) {
      tallies[index].count = tally.count + 1
      return
    }
  }
  Arrays.push(mutable tallies, Tally { word: word, count: 1 })
}
```

A name in a module is private to that file unless its declaration starts with `export`. Here `Tally` and `countWords` are exported, and `addWord` stays private, so `count.nra` can change it later without breaking any other file.

`src/cli.nra` imports what it needs from `"words/count"`:

```narra fragment
import * as Int from "std/int"
import { countWords } from "words/count"

export function run(args: Array<string>): int {
  for (const tally of countWords("the cat saw the dog")) {
    print(`${tally.word} ${Int.toString(tally.count)}`)
  }
  return 0
}
```

Run the package from its folder. `.` names the package in the current folder:

```sh
cd words
narra run .
```

The compiler stops before it runs anything:

```text
src/count.nra:10:31: Function `countWords` uses `alloc` (through `array literal`) but its signature does not declare it; write `uses alloc`
```

An exported function must say in its signature what it does besides computing a value. `countWords` builds an array, which allocates memory, so its signature needs `uses alloc` after the result type:

```narra fragment
export function countWords(text: string): Array<Tally> uses alloc {
```

Chapter 10 explains these effects. With the change, the program runs:

```text
the 2
cat 1
saw 1
dog 1
```

`narra build .` builds the program as the file `words` instead of running it.

### Importing a private name

If `cli.nra` tries to import `addWord`, the compiler points at the import and says what would make it work:

```text
src/cli.nra:2:1: "addWord" is not exported by module "words/count"; add `export` to its declaration
```

Before you add `export`, decide whether other files should depend on that function. Every exported name is something you have to keep working for them.

## What a module can hold

Besides functions and structs, a module can export enums, traits and constants. A constant at the top of a file is a `const` whose value the compiler can work out. That can be a number, a string, or an array or struct of them:

```narra
import * as Str from "std/string"

const SEPARATOR = " "
const STOP_WORDS = ["the", "a"]

export function run(args: Array<string>): int {
  for (const word of Str.split("the cat saw a dog", SEPARATOR)) {
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

```narra fragment
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
    Style.Bars => `${tally.word} ${Str.repeat("#", tally.count)}`
  }
}
```

`cli.nra` can import your own modules with an alias, like the standard ones. Through the alias you reach functions, types, and the constructors of an enum:

```narra fragment
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

Next: [9. Memory and `mutable`](09-memory-and-mutable.md).
