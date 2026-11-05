# 6. Option and Result

Narra has no `null`, `undefined`, `throw` or `try`. A value that may be missing is an `Option<T>`. A call that may fail returns a `Result<T, E>`. Both are ordinary enums from the prelude, so you can use them without an import (`SPEC.md` section 4 "Optional values" and section 5; D9, D21).

If you come from TypeScript, think of `Option<T>` as `T | undefined` that the compiler makes you check, and of `Result<T, E>` as a return value in place of an exception. If you come from Rust, they are the same types, with one difference: there is no `?` operator and no `unwrap`. Every error path is a visible `if` or `match` (D9).

## Option

`Some(value)` holds a value and `None` holds nothing. A plain `T` always has a value.

```narra
import * as Int from "std/int"

function firstEven(numbers: Array<int>): Option<int> {
  for (const number of numbers) {
    if (number % 2 == 0) {
      return Some(number)
    }
  }
  return None
}

export function run(args: Array<string>): int {
  const found = firstEven([1, 3, 4, 5])
  if (found is Some(number)) {
    print(`found ${Int.toString(number)}`)
  }
  const missing = firstEven([1, 3])
  if (missing is None) {
    print("nothing found")
  }
  return 0
}
```

```text
found 4
nothing found
```

`is` tests a pattern and binds names inside the branch. Chapter 5 covers patterns; here they are the way to open an `Option`.

### Narrowing

After a test that leaves the function, the compiler knows which case remains. Then `.value` reads the content:

```narra
function describe(text: Option<string>): string {
  if (text is None) {
    return "empty"
  }
  return `text: ${text.value}`
}

export function run(args: Array<string>): int {
  print(describe(Some("hello")))
  print(describe(None))
  return 0
}
```

```text
text: hello
empty
```

Reading `.value` before the check is a compile error. The message names the missing check:

```narra error
// error: check it with `is` or `match` before reading "value"
export function run(args: Array<string>): int {
  const found: Option<int> = Some(1)
  return found.value
}
```

There is no `unwrap` and no `expect`. To read an `Option` you either test it or give a fallback (D69).

## Helpers in `std/option`

Small transformations do not need a branch. Import the module with the plural alias `Options` (D67, D72):

| Function | Meaning |
| --- | --- |
| `Options.map(x, f)` | Apply `f` to the content of a `Some` |
| `Options.andThen(x, f)` | Like `map`, where `f` itself returns an `Option` |
| `Options.withDefault(x, fallback)` | The content, or `fallback` for `None` |
| `Options.orElse(x, f)` | `x` itself, or the `Option` that `f()` returns for `None` |
| `Options.isSome(x)`, `Options.isNone(x)` | Test without binding |
| `Options.okOr(x, error)` | Turn an `Option` into a `Result` |

```narra
import * as Options from "std/option"
import * as Int from "std/int"

function half(number: int): Option<int> {
  if (number % 2 == 0) {
    return Some(number / 2)
  }
  return None
}

export function run(args: Array<string>): int {
  const eight: Option<int> = Some(8)
  const seven: Option<int> = Some(7)

  const doubled = Options.map(eight, (number) => number * 2)
  print(Int.toString(Options.withDefault(doubled, 0)))

  const quarter = Options.andThen(Options.andThen(eight, half), half)
  print(Int.toString(Options.withDefault(quarter, -1)))

  const odd = Options.andThen(seven, half)
  print(Int.toString(Options.withDefault(odd, -1)))

  const lazy = Options.orElse(odd, () => Some(100))
  print(Int.toString(Options.withDefault(lazy, 0)))

  const missing = Options.okOr(odd, "seven is odd")
  if (missing is Err(message)) {
    print(message)
  }
  return 0
}
```

```text
16
2
-1
100
seven is odd
```

Three rules to know:

- `withDefault` always evaluates its fallback. Use `orElse` when computing the fallback costs something.
- The function arguments are non-escaping and carry an effect variable (chapter 10). A callback that prints makes the caller `uses io`.
- When a callback returns a bare `None` or `Err(...)`, the compiler cannot guess the other type. Annotate the result: `const next: Option<int> = Options.andThen(x, (number) => None)`.

The helpers never return early from your function. They build values. When a missing value must stop the function, use `is` and `return`, as above.

## Result

`Ok(value)` is success and `Err(error)` is failure. The error type is part of the signature, so callers know what can go wrong (D9, D57).

```narra
import * as Int from "std/int"

enum ParseError {
  Empty
  NotANumber(text: string)
}

function parseCount(text: string): Result<int, ParseError> {
  if (text == "") {
    return Err(ParseError.Empty)
  }
  const parsed = Int.parse(text, 10)
  if (parsed is Err(reason)) {
    return Err(ParseError.NotANumber(text))
  }
  return Ok(parsed.value)
}

function report(text: string): string {
  return match (parseCount(text)) {
    Ok(count) => `count ${Int.toString(count)}`
    Err(ParseError.Empty) => "no input"
    Err(ParseError.NotANumber(bad)) => `not a number: ${bad}`
  }
}

export function run(args: Array<string>): int {
  print(report("42"))
  print(report(""))
  print(report("4x"))
  return 0
}
```

```text
count 42
no input
not a number: 4x
```

After `if (parsed is Err(reason)) { return ... }` the compiler knows `parsed` is `Ok`, so `parsed.value` is allowed. This is the whole error-propagation story: a test, a `return` and a read.

### A Result must be looked at

Dropping a `Result` is a compile error, so a failure cannot go unnoticed:

```narra error
// error: returns a Result that is not checked
function save(): Result<int, string> {
  return Ok(1)
}

export function run(args: Array<string>): int {
  save()
  return 0
}
```

If you really do not care, say so with `const _ = save()`.

### Helpers in `std/result`

`import * as Results from "std/result"` gives `map`, `mapErr`, `andThen`, `withDefault`, and `ok` and `err`, which return the `Ok` or `Err` content as an `Option` (D69).

```narra
import * as Results from "std/result"
import * as Options from "std/option"
import * as Int from "std/int"

function parse(text: string): Result<int, string> {
  return Results.mapErr(Int.parse(text, 10), (reason) => `bad number: ${text}`)
}

function positive(number: int): Result<int, string> {
  if (number > 0) {
    return Ok(number)
  }
  return Err("not positive")
}

export function run(args: Array<string>): int {
  const good = parse("5")
  const bad = parse("x")
  print(Int.toString(Results.withDefault(Results.map(good, (number) => number * 10), 0)))
  print(Int.toString(Results.withDefault(Results.andThen(good, positive), -1)))
  print(Int.toString(Results.withDefault(Results.andThen(parse("-3"), positive), -1)))
  print(Options.withDefault(Results.err(bad), "no error"))
  print(Options.withDefault(Results.err(good), "no error"))
  return 0
}
```

```text
50
5
-1
bad number: x
no error
```

`Results.andThen` needs the callback to return a `Result` with the same error type. These helpers transform a result you already hold. They do not replace the `is` and `return` on the error path.

## Panics are not errors

`panic("message")`, a failed `assert`, an index out of bounds and overflow in a debug build abort the program. A panic cannot be caught (D36). Use `Result` for failures a caller can handle, and a panic for a bug.

Next: [Traits](07-traits.md).
