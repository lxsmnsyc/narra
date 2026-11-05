# 6. Option and Result

Some values are not always there, and some operations do not always work. A setting may be missing from a file, and the text `"80x"` is not a number. In Narra both cases show up in the type. A value that may be missing is an `Option`, and a result that may be an error is a `Result`. The compiler then makes sure you handle the missing value or the error before you use it.

This chapter builds a small reader for settings written as `key = value` lines, such as `port = 8080`.

## A value that may be missing

Start with a function that looks up a key. When it finds the key it has a string to give back. When it does not, it has nothing. Its result type, `Option<string>`, says that:

```narra
import * as Str from "std/string"

function findSetting(lines: Array<string>, key: string): Option<string> {
  for (const line of lines) {
    const equals = Str.indexOf(line, "=")
    if (equals is Some(position)) {
      if (Str.trim(Str.slice(line, 0, position)) == key) {
        return Some(Str.trim(Str.sliceFrom(line, position + 1)))
      }
    }
  }
  return None
}

export function run(args: Array<string>): int {
  const lines = ["host = example.org", "port = 8080"]
  const host = findSetting(lines, "host")
  if (host is Some(text)) {
    print(`host is ${text}`)
  }
  if (findSetting(lines, "user") is None) {
    print("no user set")
  }
  return 0
}
```

```text
host is example.org
no user set
```

An `Option<string>` is one of two things. `Some(text)` holds a string, and `None` holds nothing. The function returns `Some(...)` when it finds the key and `None` after the loop. `Option`, `Some` and `None` need no import.

`Str.indexOf` works the same way: it returns `Some(position)` when the text contains `"="` and `None` when it does not. That is why the loop opens it with `is`, as chapter 5 showed.

The caller cannot use an `Option<string>` as a string. If you forget to check, the program does not compile:

```narra error
// error: 17:20: expected string, found Option<string>
import * as Str from "std/string"

function findSetting(lines: Array<string>, key: string): Option<string> {
  for (const line of lines) {
    const equals = Str.indexOf(line, "=")
    if (equals is Some(position)) {
      if (Str.trim(Str.slice(line, 0, position)) == key) {
        return Some(Str.trim(Str.sliceFrom(line, position + 1)))
      }
    }
  }
  return None
}

export function run(args: Array<string>): int {
  print(findSetting(["host = example.org"], "host"))
  return 0
}
```

To get the string out, test the `Option` with `is` or `match`. Every place that reads a setting then decides what happens when it is missing.

## Falling back to a default

Often a missing setting has an obvious default. Writing an `if` for each one gets long, so `std/option` has functions for the common cases. Import it as `Options`:

```narra
import * as Options from "std/option"
import * as Str from "std/string"

function findSetting(lines: Array<string>, key: string): Option<string> {
  for (const line of lines) {
    const equals = Str.indexOf(line, "=")
    if (equals is Some(position)) {
      if (Str.trim(Str.slice(line, 0, position)) == key) {
        return Some(Str.trim(Str.sliceFrom(line, position + 1)))
      }
    }
  }
  return None
}

export function run(args: Array<string>): int {
  const lines = ["hostname = example.org"]

  const user = Options.withDefault(findSetting(lines, "user"), "guest")
  print(user)

  const host = Options.orElse(findSetting(lines, "host"), () => findSetting(lines, "hostname"))
  print(Options.withDefault(host, "localhost"))

  const address = Options.map(host, (name) => `https://${name}`)
  print(Options.withDefault(address, "no address"))
  return 0
}
```

```text
guest
example.org
https://example.org
```

What each call did:

- `Options.withDefault(option, fallback)` gives the content of a `Some`, or `fallback` for a `None`. There is no `user` key, so `user` is `"guest"`.
- `Options.orElse(option, f)` gives `option` when it is a `Some`, and otherwise calls `f` for another `Option`. Here it tries the key `host`, and then `hostname`. The second lookup runs only when the first one finds nothing.
- `Options.map(option, f)` applies `f` to the content of a `Some` and leaves a `None` as it is. The result is still an `Option`, because there may have been nothing to apply `f` to.

`withDefault` always computes its fallback, even when the option is a `Some`. When the fallback costs work, such as another lookup, use `orElse`, which calls its function only when it needs to.

The other functions in `std/option`:

| Function | Result |
| --- | --- |
| `Options.andThen(option, f)` | like `map`, for an `f` that itself returns an `Option` |
| `Options.isSome(option)`, `Options.isNone(option)` | a `bool`, without opening the option |
| `Options.okOr(option, error)` | a `Result`: `Ok` with the content, or `Err(error)` for a `None` |

## A result that may be an error

The port is a number, but in the file it is text. `Int.parse(text, 10)` turns text into an `int` in base 10, and it can fail. Its result type is `Result<int, string>`:

```narra
import * as Int from "std/int"

function describe(text: string): string {
  return match (Int.parse(text, 10)) {
    Ok(port) => `port ${Int.toString(port)}`
    Err(reason) => `"${text}" is not a port: ${reason}`
  }
}

export function run(args: Array<string>): int {
  print(describe("8080"))
  print(describe("80x"))
  return 0
}
```

```text
port 8080
"80x" is not a port: invalid digit for the base
```

A `Result<int, string>` is `Ok(value)` with an `int`, or `Err(error)` with a `string` that says what went wrong. Like an `Option`, you open it with `match` or `is`. The difference is that a failure carries a reason, and the second type parameter says what type that reason has. `Result`, `Ok` and `Err` need no import either.

## Errors of your own

A string is fine for a message, but the code that calls your function may need to tell one failure from another. Declare the failures as an enum, and use it as the error type:

```narra
import * as Int from "std/int"
import * as Str from "std/string"

enum SettingError {
  Missing(key: string)
  NotANumber(key: string, text: string)
}

function findSetting(lines: Array<string>, key: string): Option<string> {
  for (const line of lines) {
    const equals = Str.indexOf(line, "=")
    if (equals is Some(position)) {
      if (Str.trim(Str.slice(line, 0, position)) == key) {
        return Some(Str.trim(Str.sliceFrom(line, position + 1)))
      }
    }
  }
  return None
}

function readNumber(lines: Array<string>, key: string): Result<int, SettingError> {
  const found = findSetting(lines, key)
  if (found is None) {
    return Err(SettingError.Missing(key))
  }
  const parsed = Int.parse(found.value, 10)
  if (parsed is Err(_)) {
    return Err(SettingError.NotANumber(key, found.value))
  }
  return Ok(parsed.value)
}

function report(lines: Array<string>, key: string): string {
  return match (readNumber(lines, key)) {
    Ok(number) => `${key} = ${Int.toString(number)}`
    Err(SettingError.Missing(name)) => `${name} is not set`
    Err(SettingError.NotANumber(name, text)) => `${name} should be a number, not "${text}"`
  }
}

export function run(args: Array<string>): int {
  const lines = ["port = 8080", "timeout = soon"]
  print(report(lines, "port"))
  print(report(lines, "timeout"))
  print(report(lines, "retries"))
  return 0
}
```

```text
port = 8080
timeout should be a number, not "soon"
retries is not set
```

`readNumber` handles each failure at the line where it happens. When the key is missing it returns `Err(SettingError.Missing(key))` and stops. After that `if`, the compiler knows that `found` is a `Some`, so `found.value` reads the string (chapter 5 calls this narrowing). The same goes for `parsed`: after the `Err` test returns, `parsed.value` is the number.

This is how an error moves up through your functions: test, return the error, and read the value on the next line. `report` then uses `match` to give each case its own message. Because `match` must cover every case, adding a third variant to `SettingError` makes `report` fail to compile until it handles the new case.

## A Result must be used

A function that returns a `Result` can fail, so calling it and ignoring the result is a compile error:

```narra error
// error: 7:3: this call returns a Result that is not checked; handle it, or discard it with `const _ = ...`
function save(text: string): Result<int, string> {
  return Ok(1)
}

export function run(args: Array<string>): int {
  save("port = 8080")
  return 0
}
```

The message gives both ways out. Usually you handle the result with `is` or `match`. If the failure does not matter at this call, bind the result to `_`, and the code shows that you chose to drop it:

```narra
function save(text: string): Result<int, string> {
  return Ok(1)
}

export function run(args: Array<string>): int {
  const _ = save("port = 8080")
  print("saved, or not")
  return 0
}
```

```text
saved, or not
```

## Shorter code with `std/result`

`std/result`, imported as `Results`, has functions like those of `std/option`. With them, `readNumber` fits in one expression:

```narra
import * as Int from "std/int"
import * as Options from "std/option"
import * as Results from "std/result"
import * as Str from "std/string"

enum SettingError {
  Missing(key: string)
  NotANumber(key: string, text: string)
}

function findSetting(lines: Array<string>, key: string): Option<string> {
  for (const line of lines) {
    const equals = Str.indexOf(line, "=")
    if (equals is Some(position)) {
      if (Str.trim(Str.slice(line, 0, position)) == key) {
        return Some(Str.trim(Str.sliceFrom(line, position + 1)))
      }
    }
  }
  return None
}

function readNumber(lines: Array<string>, key: string): Result<int, SettingError> {
  const text = Options.okOr(findSetting(lines, key), SettingError.Missing(key))
  return Results.andThen(text, (found) => {
    return Results.mapErr(Int.parse(found, 10), (_) => SettingError.NotANumber(key, found))
  })
}

export function run(args: Array<string>): int {
  const lines = ["port = 8080", "timeout = soon"]
  print(Int.toString(Results.withDefault(readNumber(lines, "port"), 80)))
  print(Int.toString(Results.withDefault(readNumber(lines, "timeout"), 30)))
  const doubled = Results.map(readNumber(lines, "port"), (port) => port * 2)
  print(Int.toString(Results.withDefault(doubled, 0)))
  if (Results.err(readNumber(lines, "retries")) is Some(SettingError.Missing(name))) {
    print(`${name} is missing`)
  }
  return 0
}
```

```text
8080
30
16160
retries is missing
```

Step by step:

1. `Options.okOr` turns the `Option` from `findSetting` into a `Result`, with `SettingError.Missing(key)` as the error for `None`.
2. `Results.andThen(result, f)` calls `f` with the `Ok` value, and `f` returns the next `Result`. An `Err` skips `f` and passes through.
3. `Results.mapErr(result, f)` changes the error and leaves an `Ok` alone. It turns the parser's string into a `SettingError`.

In `run`, `Results.withDefault` gives a fallback for any error, and `Results.map` changes an `Ok` value. `Results.err` gives the error as an `Option`, and `Results.ok` gives the value.

Both versions of `readNumber` do the same thing. The helpers suit short chains that transform a value. The `if` and `return` form suits a function with several steps. Each failure has its own line, and the code reads from top to bottom.

When an arrow function passed to these helpers returns a bare `None` or `Err(...)`, the compiler cannot work out the missing type. It asks for an annotation:

```narra error
// error: 6:50: cannot tell the type of None here; add a type annotation
import * as Options from "std/option"

export function run(args: Array<string>): int {
  const port: Option<int> = Some(8080)
  const next = Options.andThen(port, (number) => None)
  return 0
}
```

Annotating the arrow function's result, as in `(number): Option<int> => None`, fixes it.

## When failure is a bug

`Option` and `Result` are for failures that a caller can handle: a missing setting, a bad number, a file that does not exist. Some failures mean the program itself is wrong, such as an index past the end of an array. For those, call `panic("message")`, or `assert(condition)`. A panic stops the program with the message and exit code 101, and no code can catch it. Chapter 10 shows how a test checks that a function panics.

Next: [7. Traits](07-traits.md).
