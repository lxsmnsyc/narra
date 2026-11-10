# 6. Option and Result

You have met `Option` twice already. Reading a map in chapter 2 gives `Some(count)` or `None`, and the parser in chapter 5 returned an `Option<Command>`. Chapter 5 also met `Result`, the type `Int.parse` returns, and threw away its error message. This chapter writes functions that return both types, keeps the error messages, and shows the standard functions that make them shorter to use.

The program is a small reader for settings written as `key = value` lines, such as `port = 8080`.

## Return a value that may be missing

Start with a function that looks up a key in the lines. When it finds the key it has a string to give back, and when it does not it has nothing. Its result type, `Option<string>`, says that:

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
  if (findSetting(lines, "host") is Some(host)) {
    print(`host is ${host}`)
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

The function returns `Some(...)` with the value when the key matches, and `None` after the loop has looked at every line. `Str.indexOf` works the same way: it gives `Some(position)` when the line contains `"="`, and `None` when it does not.

The rest of the chapter keeps the settings in a `Map<string, string>` instead of lines. `settings["port"]` gives the same `Option<string>` as `findSetting(lines, "port")`, and the samples stay short.

## Fall back to a default

Often a missing setting has an obvious default. An `if` for each one gets long, so `std/option` has functions for the common cases. Import it as `Options`:

```narra
import * as Options from "std/option"

export function run(args: Array<string>): int {
  const settings: Map<string, string> = Map { "hostname": "example.org" }

  const user = Options.withDefault(settings["user"], "guest")
  print(user)

  const host = Options.orElse(settings["host"], () => settings["hostname"])
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
- `Options.orElse(option, f)` gives `option` when it is a `Some`, and otherwise calls `f` for another `Option`. Here it tries the key `host`, then `hostname`. The result is still an `Option`, since both keys may be missing, so the next line uses `withDefault` to get a string.
- `Options.map(option, f)` applies `f` to the content of a `Some` and leaves a `None` as it is. Its result is an `Option` too.

`withDefault` always computes its fallback, even when the option is a `Some`. When the fallback costs work, such as another lookup, use `orElse`, which calls its function only when it needs to.

The other functions in `std/option`:

| Function | Result |
| --- | --- |
| `Options.andThen(option, f)` | like `map`, for an `f` that itself returns an `Option` |
| `Options.isSome(option)`, `Options.isNone(option)` | a `bool`, without opening the option |
| `Options.okOr(option, error)` | a `Result`: `Ok` with the content, or `Err(error)` for a `None` |

## Keep the reason for a failure

The port is a number, but in the file it is text. In chapter 5, `parseNumber` turned the `Result` of `Int.parse` into an `Option` and dropped the reason. This time, keep it:

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

`Int.parse` returns a `Result<int, string>`. That is `Ok(value)` with an `int`, or `Err(error)` with a `string` that says what went wrong. The second type parameter is the type of the error, so a `Result` can carry a reason of any type, not only text. `Result`, `Ok` and `Err` need no import.

## Errors of your own

A string is fine for a message, but the code that calls your function may need to tell one failure from another. Declare the failures as an enum, and use it as the error type:

```narra
import * as Int from "std/int"

enum SettingError {
  Missing(key: string)
  NotANumber(key: string, text: string)
}

function readNumber(settings: Map<string, string>, key: string): Result<int, SettingError> {
  const found = settings[key]
  if (found is None) {
    return Err(SettingError.Missing(key))
  }
  const parsed = Int.parse(found.value, 10)
  if (parsed is Err(_)) {
    return Err(SettingError.NotANumber(key, found.value))
  }
  return Ok(parsed.value)
}

function report(settings: Map<string, string>, key: string): string {
  return match (readNumber(settings, key)) {
    Ok(number) => `${key} = ${Int.toString(number)}`
    Err(SettingError.Missing(name)) => `${name} is not set`
    Err(SettingError.NotANumber(name, text)) => `${name} should be a number, not "${text}"`
  }
}

export function run(args: Array<string>): int {
  const settings: Map<string, string> = Map { "port": "8080", "timeout": "soon" }
  print(report(settings, "port"))
  print(report(settings, "timeout"))
  print(report(settings, "retries"))
  return 0
}
```

```text
port = 8080
timeout should be a number, not "soon"
retries is not set
```

`readNumber` handles each failure at the line where it happens. When the key is missing it returns `Err(SettingError.Missing(key))` and stops. After that `if`, `found` is known to be a `Some`, so `found.value` reads the string, as chapter 5 showed. The same goes for `parsed`: after the `Err` test returns, `parsed.value` is the number.

This is how an error moves up through your functions: test, return the error, and read the value on the next line. `report` then gives each case its own message. Because `match` must cover every case, adding a third variant to `SettingError` makes `report` fail to compile until it handles the new case.

## A Result must be used

A function that returns a `Result` can fail, so calling it and ignoring the result is a compile error:

```narra error
// error: 7:3: this call returns a Result that is not checked; handle it, or discard it with `const _ = ...`
function save(_text: string): Result<int, string> {
  return Ok(1)
}

export function run(args: Array<string>): int {
  save("port = 8080")
  return 0
}
```

The message gives both ways out. Usually you handle the result with `is` or `match`. If the failure does not matter at this call, bind the result to `_`, and the code shows that you chose to drop it:

```narra
function save(_text: string): Result<int, string> {
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

enum SettingError {
  Missing(key: string)
  NotANumber(key: string, text: string)
}

function readNumber(settings: Map<string, string>, key: string): Result<int, SettingError> {
  const text = Options.okOr(settings[key], SettingError.Missing(key))
  return Results.andThen(text, (found) => {
    return Results.mapErr(Int.parse(found, 10), (_) => SettingError.NotANumber(key, found))
  })
}

export function run(args: Array<string>): int {
  const settings: Map<string, string> = Map { "port": "8080", "timeout": "soon" }
  print(Int.toString(Results.withDefault(readNumber(settings, "port"), 80)))
  print(Int.toString(Results.withDefault(readNumber(settings, "timeout"), 30)))
  const doubled = Results.map(readNumber(settings, "port"), (port) => port * 2)
  print(Int.toString(Results.withDefault(doubled, 0)))
  if (Results.err(readNumber(settings, "retries")) is Some(SettingError.Missing(name))) {
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

1. `Options.okOr` turns the `Option` from the map into a `Result`, with `SettingError.Missing(key)` as the error for `None`.
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

`Option` and `Result` are for failures that a caller can handle: a missing setting, a bad number, a file that does not exist. Other failures mean the program itself is wrong. Say the reader checks every port when it loads the settings, and a later function still finds a port below 1. That cannot happen unless the code has a bug, and there is nothing sensible to return. Stop the program with `panic("message")`, or state the condition with `assert(port >= 1)`. A panic prints the message and ends the program with exit code 101, and no code can catch it.

Some operations panic on their own. An index past the end of an array panics. Integer overflow does not panic; it wraps. Integer division does not panic either: the compiler accepts `total / count` only when `count` is a non-zero constant or a name you checked with `if (count != 0)` first. Chapter 10 shows how a test checks that a function panics.

Next: [7. Traits](07-traits.md).
