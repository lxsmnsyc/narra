# 8. Modules and packages

One file is one module. Names are private until exported, and a module sees only what it imports. If you know ES modules, the syntax is the same subset. If you know Rust, think of a file as a `mod` with `pub` replaced by `export` (`SPEC.md` section 3; D3, D4, D25).

## Imports

An import names a module by a package-rooted string path and gives it an alias. There are two forms:

```narra
import * as Str from "std/string"
import * as Arrays from "std/array"
import { toString as intToString } from "std/int"

export function run(args: Array<string>): int {
  const words = [" narra ", " guide "]
  print(Str.trim(words[0]))
  print(Str.repeat("ab", Arrays.length(words)))
  print(intToString(Arrays.length(words)))
  return 0
}
```

```text
narra
abab
2
```

- `import * as Name from "path"` binds a namespace. You use `Name.function(...)`. This is the main style, and it is how almost all std modules are used.
- `import { name } from "path"` binds specific names, and `name as other` renames one. A renamed import follows the same rules as any other name: it may not clash with a declaration, another import or a type (D72).

Imports sit at the top of the file with a string literal path. They never run code, and `import(...)` as an expression is a compile error: imports are static (D4).

### The prelude

Some names need no import: `Option`, `Some`, `None`, `Result`, `Ok`, `Err`, `Array`, `Map`, `Range`, `Box`, the primitive types, and the three functions `print`, `assert` and `panic`. Nothing else is global.

A type in the prelude does not bring its functions. `Array` is always there, but `Arrays.push` needs `import * as Arrays from "std/array"` (D67). Forget it, and the error tells you the line to add:

```narra error
// error: Array.push needs an import; add `import * as Arrays from "std/array"`
export function run(args: Array<string>): int {
  let numbers: Array<int> = []
  Array.push(mutable numbers, 1)
  return 0
}
```

### Alias names

By convention the std module of a capitalized prelude type gets the plural: `Arrays`, `Maps`, `Boxes`, `Ranges`, `Options`, `Results`. The rule behind it is that an alias may not equal a type name, so `Shape.Circle` always names a constructor of the type `Shape` and never a function of a module `Shape`:

```narra error
// error: the import alias Array is also a type name; rename it, for example `import * as Arrays from "std/array"`
import * as Array from "std/array"

export function run(args: Array<string>): int {
  return Array.length([1, 2])
}
```

### The standard modules

`std/array`, `std/box`, `std/char`, `std/env`, `std/float`, `std/fs`, `std/int`, `std/map`, `std/option`, `std/process`, `std/range`, `std/result`, `std/string` and `std/test`. An unknown path is an error that lists the modules the compiler supports. Modules under `std/internal/` are for the standard library only; importing one from your code is an error (E0426).

## Exports

`export` makes a top-level declaration visible to other modules: `export function`, `export const`, `export type`, `export struct`, `export enum`, `export trait`. Everything else is private to the file.

Some rules differ from what TypeScript allows (D4):

- No `export { x } from "m"`, no `export *` and no `export default`. An exported name is always declared in the file that exports it.
- To expose an imported name, declare a new one from it: `export const classify = Text.classify` for a function, `export type Token = Text.Token` for a type.
- No top-level `let`. A module level holds functions, types, traits, `const` and `lazy` values, so importing a module never runs code.

## Constructors are always qualified

Importing an enum brings its constructors, but you always write them with the enum name, in expressions and in patterns (D71). `Some`, `None`, `Ok` and `Err` are prelude names and stay bare.

```narra
enum Shape {
  Circle(radius: int)
  Dot
}

enum Maybe {
  None
  Just(value: int)
}

function describe(shape: Shape): string {
  return match (shape) {
    Shape.Circle(radius) => "circle"
    Shape.Dot => "dot"
  }
}

export function run(args: Array<string>): int {
  print(describe(Shape.Circle(2)))
  print(describe(Shape.Dot))
  const none: Option<int> = None
  const maybe = Maybe.None
  if (none is None && maybe is Maybe.None) {
    print("two different Nones")
  }
  return 0
}
```

```text
circle
dot
two different Nones
```

Because constructors never go bare, two enums can use the same constructor name without a clash, and a user enum can even have its own `None`. A bare `Circle(2)` is a compile error, and the message gives the fix:

```narra error
// error: constructor Circle needs its enum; write Shape.Circle
enum Shape {
  Circle(radius: int)
  Dot
}

export function run(args: Array<string>): int {
  const shape = Circle(2)
  return 0
}
```

## A package with several modules

The samples in this guide are one file each, because a file is the unit the tests run. Real programs have several files, in a package. A package is a directory with `narra.toml` and a `src/` folder, and the path of a file under `src/` is its module name:

```narra fragment
app/
  narra.toml
  src/
    cli.nra          module "app/cli"
    text.nra         module "app/text"
    text/
      words.nra      module "app/text/words"
```

`narra.toml` names the package and its binaries. `entry` is a module inside the package, then the function:

```narra fragment
[package]
name = "app"

[[binary]]
name = "app"
entry = "cli.run"
```

Paths are package rooted. Your own modules start with the package name, the standard library with `std`, and a dependency with its package name. Relative paths such as `"./text"` are not allowed (D3).

```narra fragment
// src/text.nra
export struct Token {
  word: string
}

export function classify(word: string): Token {
  return Token { word: word }
}

function secret(): int {   // private: no other module can import it
  return 1
}
```

```narra fragment
// src/cli.nra
import * as Str from "std/string"
import { classify, Token } from "app/text"

export function run(args: Array<string>): int {
  const token: Token = classify(Str.trim(" hello "))
  print(token.word)
  return 0
}
```

Build and run the whole package from its directory:

```narra fragment
narra build app             # every [[binary]] of the package in app/
narra run -bin app app      # run one of them
```

The import graph must have no cycles. If `app/a` imports `app/b` and `app/b` imports `app/a`, the compiler reports the loop as a list of modules (D4). To break a cycle, move the shared declarations into a third module that both import.

Next: [Memory and `mutable`](09-memory-and-mutable.md).
