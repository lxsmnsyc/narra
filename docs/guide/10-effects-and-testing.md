# 10. Effects and testing

This chapter writes a small module that converts temperatures, then tests it. On the way it explains the `uses alloc` that chapter 8 needed. An exported function says in its signature whether it allocates memory or does input and output, and the compiler checks it.

## What a function does, in its signature

Start with a conversion and a function that formats a reading:

```narra error
// error: 9:10: Function `describe` uses `alloc` (through `Int.toString`) but its signature does not declare it; write `uses alloc`
import * as Int from "std/int"

export function toFahrenheit(celsius: int): int {
  return celsius * 9 / 5 + 32
}

export function describe(celsius: int): string {
  return `${Int.toString(celsius)} C is ${Int.toString(toFahrenheit(celsius))} F`
}

export function run(args: Array<string>): int {
  print(describe(21))
  return 0
}
```

`toFahrenheit` compiles: it only computes a number. `describe` builds new strings, and a new string needs memory from the heap. Allocating memory is an effect: something a function does besides computing its result. An exported function must list its effects after the result type, with `uses`. The message says which call brought the effect and what to write.

There are two effects you will meet:

| Effect | The function may |
| --- | --- |
| `alloc` | allocate heap memory: build strings, arrays, maps, closures, or enum values with fields |
| `io` | do input or output: print, read or write files, run a process, read the environment |

A function that prints needs `io` as well. With both declared, the program runs:

```narra
import * as Int from "std/int"

export function toFahrenheit(celsius: int): int {
  return celsius * 9 / 5 + 32
}

export function describe(celsius: int): string uses alloc {
  return `${Int.toString(celsius)} C is ${Int.toString(toFahrenheit(celsius))} F`
}

export function announce(celsius: int) uses io, alloc {
  print(describe(celsius))
}

export function run(args: Array<string>): int {
  announce(21)
  announce(-40)
  return 0
}
```

```text
21 C is 69 F
-40 C is -40 F
```

Now each signature tells a caller what to expect. `toFahrenheit` has no `uses`, so it does nothing but compute, and calling it twice with the same number gives the same answer. `describe` may allocate but never prints. `announce` may print. `run`, the function a program starts from, may do anything, so it never needs a `uses` list.

A panic is not an effect. Any function may panic, whatever its `uses` list says.

## Private functions work it out themselves

Only exported functions write `uses`. For a private function, the compiler finds the effects from the body:

```narra error
// error: 7:6: Function `report` uses `io` (through `log`) but its signature does not declare it; write `uses io`
function log(message: string) {
  print(message)
}

export function report(reading: int): int {
  log("reading taken")
  return reading
}

export function run(args: Array<string>): int {
  return report(1)
}
```

`log` has no `uses` list, and the compiler accepts that. It still knows that `log` prints, so the exported `report`, which calls `log`, must declare `io`. Effects pass up through every call. To find out what a private function does, follow its calls, or read the error when an exported caller misses something.

The split follows how code changes. A private function is used only inside its file, where you can change it and its callers together. An exported function is used by other files, and its `uses` list is part of what they rely on. If you later add a `print` to it, its signature changes, and the change is visible to anyone who reads it.

## Functions that take functions

An exported function that calls a function it receives has a problem: its effects depend on what it is given. Here is a first try:

```narra error
// error: 14:25: the argument for `visit` of `eachReading` uses `io`, but its function type does not declare it; write `uses io` in the type, or `uses E` with `<effect E>` on `eachReading`
export function eachReading(readings: Array<int>, visit: (reading: int) => void) {
  for (const reading of readings) {
    visit(reading)
  }
}

export function run(args: Array<string>): int {
  const readings = [21, 23]
  const noisy = (reading: int) => {
    print("visiting")
    return
  }
  eachReading(readings, noisy)
  return 0
}
```

In an exported function, the function type `(reading: int) => void` with no `uses` means a function with no effects. `eachReading` itself then has no effects either, and passing it a function that prints breaks that.

The message offers two fixes. Writing `uses io` in the parameter's type would allow a printing `visit`. But then `eachReading` would have `io` even when `visit` only computes. The better fix says "`eachReading` has the effects of whatever `visit` does", with an effect variable. It is declared with the type parameters as `effect E`, attached to the parameter's function type, and repeated in the function's own `uses` list:

```narra
import * as Int from "std/int"

export function eachReading<effect E>(readings: Array<int>, visit: (reading: int) => void uses E) uses E {
  for (const reading of readings) {
    visit(reading)
  }
}

export function run(args: Array<string>): int {
  eachReading([21, 23, 19], (reading) => {
    print(Int.toString(reading))
  })
  return 0
}
```

```text
21
23
19
```

At each call, `E` becomes the effects of the function passed in. Here that function prints, so this call of `eachReading` has `io`. Passed a function that only computes, the same `eachReading` would have no effects. The `std/option` and `std/result` helpers from chapter 6 are written this way, which is why a callback that prints makes the call print.

A private function does not need this. A function-type parameter of a private function takes on the effects of whatever is passed in, without an effect variable.

## Writing tests

Tests go in the same file as the code, in `test` blocks. Here is the temperature module with a parser, an average, and five tests. Save it as `temperature.nra`:

```narra fragment
import * as Arrays from "std/array"
import * as Int from "std/int"
import * as Str from "std/string"
import * as Test from "std/test"

export function toFahrenheit(celsius: int): int {
  return celsius * 9 / 5 + 32
}

export function parseCelsius(text: string): Result<int, string> uses alloc {
  if (!Str.endsWith(text, "C")) {
    return Err(`missing unit: ${text}`)
  }
  return Int.parse(Str.slice(text, 0, Str.length(text) - 1), 10)
}

export function average(readings: Array<int>): int {
  let total = 0
  for (const reading of readings) {
    total = total + reading
  }
  return total / Arrays.length(readings)
}

test "water freezes at 32 F" {
  Test.eq(toFahrenheit(0), 32)
}

test "parse reads a reading" {
  Test.eq(parseCelsius("21C"), Ok(21))
}

test "parse rejects a missing unit" {
  Test.eq(parseCelsius("21"), Err("missing unit: 21"))
}

test "average of no readings panics" {
  Test.panicsWith(() => average([]), "division by zero")
}

test "water boils at 211 F" {
  Test.eq(toFahrenheit(100), 211)
}
```

A `test "name" { ... }` block holds statements, like a function body with no parameters. The checks come from the module `std/test`:

- `Test.eq(actual, expected)` passes when the two values are equal. `Test.ne(actual, other)` passes when they differ.
- `Test.panicsWith(f, message)` calls `f` and passes when it panics with a message that contains `message`. `Test.panics(f)` passes on any panic.

Because errors are values (chapter 6), a test of a failure compares with an `Err` like any other value, as `"parse rejects a missing unit"` does. `Test.panicsWith` is for failures that stop the program, such as the division by zero in `average([])`.

Run the tests with `narra test`:

```sh
narra test temperature.nra
```

```text
ok - water freezes at 32 F
ok - parse reads a reading
ok - parse rejects a missing unit
ok - average of no readings panics
FAIL - water boils at 211 F: expected 211, got 212 (temperature.nra:42:10)
4 passed, 1 failed
```

`narra test` runs the tests in the order they appear and prints a line for each. A failure shows both values and the line and column of the check that failed. Here the test is wrong, since water boils at 212 F. The runner goes on after a failure. It exits with code 1 when any test failed and 0 otherwise, so a script can stop on a failure.

Each test runs in a process of its own. A test that crashes, leaks memory or runs longer than 10 seconds fails on its own, and the tests after it still run. To allow a slower test, set the limit in seconds with the `NARRA_TEST_TIMEOUT` environment variable.

To run some of the tests, give part of their names with `-filter`. `-time` adds how long each test took:

```sh
narra test -filter parse temperature.nra
```

```text
ok - parse reads a reading
ok - parse rejects a missing unit
2 passed, 0 failed
```

A filter that matches no test prints `no tests match "..."` and exits with code 1, so a typo in the filter does not pass silently.

`narra build` and `narra run` leave the test blocks out, so tests add nothing to the program you ship. The file can import `std/test` and still build as a program.

### Values a test can compare

`Test.eq` prints both values when it fails, so their type needs `Show` as well as `==` (chapter 7). A struct without `Show` stops at the check:

```narra error
// error: Point does not implement Show
import * as Test from "std/test"

struct Point {
  x: int
}

test "points" {
  Test.eq(Point { x: 1 }, Point { x: 1 })
}

export function run(args: Array<string>): int {
  return 0
}
```

Adding `derives Show` to `Point` fixes it.

### Checks belong in test blocks

The `Test` functions only work inside a `test` block, including arrow functions written inside one. A helper function that calls `Test.eq` is an error, even when only tests call the helper:

```narra error
// error: 5:10: Test.eq can only be called inside a `test` block
import * as Test from "std/test"

function checkFreezing(celsius: int) {
  Test.eq(celsius * 9 / 5 + 32, 32)
}

test "freezing" {
  checkFreezing(0)
}

export function run(args: Array<string>): int {
  return 0
}
```

`checkFreezing` is an ordinary function, so it is part of every build of the file, and only `narra test` builds include the code behind `std/test`. Have the helper return the value, and check it in the test: `Test.eq(toFahrenheit(0), 32)`.

A test block may print or do any other effect without a `uses` list.
