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

Besides functions and structs, a module can export enums, traits and constants. A constant at the top of a file is a `const` whose value the compiler can work out. That can be a number, a string, or an array or struct of them:

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

Next: [9. Memory and `mutable`](09-memory-and-mutable.md).
