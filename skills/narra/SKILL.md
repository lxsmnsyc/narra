---
name: narra
description: Write, check, fix and debug Narra code (.nra files, narra.toml) with the narrac compiler. Use it when a task creates or changes Narra code, fixes a narrac error, or asks where something is defined or used in a Narra package.
---

# Narra with narrac

Narra is a compiled language. `narrac` is its compiler and its tools. Do not guess a rule from another language: run `narrac check` and read the error.

## The loop

1. Before editing, find the code: `narrac query def NAME`, `narrac query callers NAME`, `narrac query refs NAME`, `narrac query interface PACKAGE/MODULE` (such as `task/calc`), `narrac query type-of FILE:LINE:COL` (the position must sit on an expression). Queries need a package (a folder with `narra.toml`). Options such as `--json` go before the question. `refs` and `callers` skip import lines and calls through function values, so also search the text for the name.
2. After editing, run `narrac check --json PATH`. It prints one JSON object per error with `code`, `message`, `range` (`start` and `end`, each `[line, column]`, from 1) and often `fixes`. Exit code 0 means no errors.
3. For an error you do not understand, run `narrac explain CODE`.
4. For an error with a fix, run `narrac apply-fix PATH`. It prints the whole file with the first fix applied and writes nothing. Read it, then save it through a new file (`narrac apply-fix a.nra > b.nra && mv b.nra a.nra`), never straight to `a.nra`.
5. Run `narrac fmt PATH` (it rewrites files in place), then `narrac test PATH`.

Fix the first error first. Later errors often come from it.

## Rules agents get wrong

Each pair shows the error code the compiler gives. The second block of a pair compiles.

A name is declared once per function. Shadowing is an error.

```narra error E0103
export function total(prices: Array<int>): int {
  let sum = 0
  for (const price of prices) {
    const sum = price
  }
  return sum
}
```

Pick another name: `const current = price`.

There are no function declarations inside a function. Use an arrow function in a `const`, or a private function at module level.

```narra error E0930
export function run(args: Array<string>): int {
  function double(value: int): int {
    return value * 2
  }
  return double(2)
}
```

```narra
export function run(args: Array<string>): int {
  const double = (value: int): int => value * 2
  return double(2)
}
```

A name that is never read is an error, a parameter too (except the entry function's, when `narra.toml` has a `[[binary]]`). Remove it, or start its name with `_` to keep it on purpose.

```narra error E0921
export function run(args: Array<string>): int {
  const unused = 5
  return 0
}
```

Code after a `return`, `break`, `continue` or `panic` is dead code, also an error.

```narra error E1500
export function run(args: Array<string>): int {
  return 0
  print("done")
}
```

Use `const` unless the name is assigned again. A `let` that is never assigned is an error. Passing `mutable name` counts as an assignment.

```narra error E1510
export function run(args: Array<string>): int {
  let count = 3
  return count
}
```

An exported function lists its effects after the result type: `uses alloc` when it builds strings, arrays or maps, `uses io` when it reads or prints. Private functions need no list. A package's binary entry (`entry` in `narra.toml`), usually `run`, needs none.

```narra error E0800
import * as Int from "std/int"

export function label(count: int): string {
  return Int.toString(count)
}
```

```narra
import * as Int from "std/int"

export function label(count: int): string uses alloc {
  return Int.toString(count)
}
```

A function that changes its argument takes `mutable` in the parameter and in every call.

```narra error E0315
import * as Arrays from "std/array"

export function addOne(items: Array<int>) uses alloc {
  Arrays.push(mutable items, 1)
}
```

```narra
import * as Arrays from "std/array"

export function addOne(mutable items: Array<int>) uses alloc {
  Arrays.push(mutable items, 1)
}
```

There are no exceptions. A function that can fail returns `Result<T, E>`; the caller checks it with `is Err` or `match`. Use `panic("message")` only for bugs.

```narra error E0927
export function check(value: int): int {
  if (value < 0) {
    throw "negative"
  }
  return value
}
```

```narra
export function check(value: int): Result<int, string> {
  if (value < 0) {
    return Err("negative")
  }
  return Ok(value)
}
```

Only `print`, `assert`, `panic` and `copy` are global functions. Everything else is imported from `std/...`, as `import * as Arrays from "std/array"`. Statements end at a line break, with no semicolons. There are no classes, `new` or `this`; use `struct`, `enum` and functions of the module.

## Tests

Tests sit in the same file, in `test "name" { ... }` blocks, with checks from `std/test`. `Test` functions work only inside a `test` block.

```narra
import * as Test from "std/test"

export function double(value: int): int {
  return value * 2
}

test "double doubles" {
  Test.eq(double(4), 8)
}
```

`narrac test PATH` prints one line per test and exits 1 when one fails. `-filter TEXT` runs the tests whose name contains the text.

## More

The guide is in `docs/guide/` of the Narra repository. Read the chapter for the topic before writing code you are unsure about: values and types (2), functions (3), structs and enums (4), `match` (5), `Option` and `Result` (6), traits (7), modules (8), `mutable` (9), effects and tests (10), the standard library (11), and the tools (12).
