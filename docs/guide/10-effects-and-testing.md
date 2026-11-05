# 10. Effects and testing

An effect marks something a function can do beyond computing a value. The signature says it, and the compiler checks it, so you can tell from a declaration whether a call can allocate, read a file or print. The second half of this chapter covers `test` blocks and `std/test`.

## The three effects

| Effect | Meaning |
| --- | --- |
| `alloc` | May allocate heap memory |
| `io` | May perform input or output, read the clock, or read the environment |
| `unsafe` | Calls code the compiler cannot check |

The set is fixed. `unsafe` belongs to C interop, which this guide does not cover yet. Panics are not an effect: any function can panic.

## Declaring and inferring

An exported function declares its effects after the return type with `uses`. A private function or an arrow function does not: the compiler infers them. A function's effects include those of everything it calls.

```narra
import * as Int from "std/int"

// Private: the effects are inferred from the body.
function announce(number: int): string {
  print("announcing")
  return Int.toString(number)
}

// Exported: the effects are written in the signature.
export function report(number: int): string uses io, alloc {
  return announce(number)
}

export function twice(number: int): int {
  return number * 2
}

export function run(args: Array<string>): int {
  print(report(twice(21)))
  return 0
}
```

```text
announcing
42
```

`twice` is pure and has no `uses`. `report` calls `announce`, which prints and builds a string, so `report` must say `io, alloc`. The entry function `run` has every effect.

Calling an effectful function from one that does not declare the effect is a compile error. The message names the call that brings it:

```narra error
// error: Function `log` uses `io` (through `print`) but its signature does not declare it
export function log(message: string) {
  print(message)
}

export function run(args: Array<string>): int {
  log("x")
  return 0
}
```

Why make exported functions say it? Private functions can be changed freely, so inference keeps them light. An exported function is a promise to other modules, and its effects are part of that promise. Adding a `print` to an exported function is a visible change to its signature.

## Effect variables

A function that calls a function it receives has the effects of that function, whichever it is. An effect variable states this. It is declared in the type parameters with `effect`, and used only after `uses`:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

export function applyAll<T, U, effect E>(
  items: Array<T>,
  transform: (item: T) => U uses E
): Array<U> uses alloc, E {
  let results: Array<U> = []
  for (const item of items) {
    Arrays.push(mutable results, transform(item))
  }
  return results
}

export function run(args: Array<string>): int {
  // A pure callback: applyAll only allocates.
  const doubled = applyAll([1, 2, 3], (number) => number * 2)
  // A callback that prints: this call has io, and so does run.
  const noisy = applyAll([1, 2], (number) => {
    print("visiting")
    return number
  })
  print(Int.toString(doubled[2] + noisy[1]))
  return 0
}
```

```text
visiting
visiting
8
```

At each call, `E` takes the effects of the function passed in. This is how `Arrays.map`, `Options.map` and the other std helpers can take any callback: a pure callback keeps the call pure, an `io` callback makes the caller need `io`. The variable must appear in one of the parameters' function types, so the compiler knows where it comes from.

A callback that does `io` passed where the parameter's function type does not allow it is an error, because a plain function type is pure:

```narra error
// error: but its function type does not declare it
export function apply(f: (x: int) => int, x: int): int {
  return f(x)
}

export function run(args: Array<string>): int {
  const noisy = (x: int): int => {
    print("hi")
    return x
  }
  return apply(noisy, 1)
}
```

## Tests

A `test "name" { ... }` block sits next to the code it checks. `narra test` builds and runs the blocks. In a normal build, test blocks are removed, so they cost nothing in the program you ship.

Assertions come from `std/test`:

| Call | Passes when |
| --- | --- |
| `Test.eq(actual, expected)` | the values are equal. Needs `Equatable` and `Show`. |
| `Test.ne(actual, other)` | the values differ |
| `Test.panics(f)` | `f()` panics |
| `Test.panicsWith(f, message)` | `f()` panics with a message that contains `message` |
| `Test.golden(name, actual)` | `actual` matches the file `testdata/<name>.golden` |

A failed assertion is a panic that fails that test only. Both values are written with `Show`, and the failure names the `file:line:col` of the call.

`TestGuide` runs every sample with `run`, so the test file below is shown as a fragment. Save it as `age.nra` and run `narra test age.nra`.

```narra fragment
import * as Test from "std/test"
import * as Int from "std/int"

struct Point derives Show {
  x: int
  y: int
}

function divide(a: int, b: int): int {
  return a / b
}

function parseAge(text: string): Result<int, string> {
  const parsed = Int.parse(text, 10)
  if (parsed is Err(reason)) {
    return Err(`not a number: ${text}`)
  }
  return Ok(parsed.value)
}

test "parse accepts digits" {
  Test.eq(parseAge("42"), Ok(42))
}

test "parse rejects text" {
  Test.eq(parseAge("x"), Err("not a number: x"))
}

test "points compare by value" {
  Test.eq(Point { x: 1, y: 2 }, Point { x: 1, y: 2 })
  Test.ne(Point { x: 1, y: 2 }, Point { x: 2, y: 1 })
}

test "division by zero panics" {
  Test.panicsWith(() => divide(1, 0), "division by zero")
}

test "this one fails" {
  Test.eq(divide(7, 2), 4)
}
```

`narra test` runs each test in its own process, in source order, and goes on after a failure. It prints one line per test, then a summary:

```narra fragment
$ narra test age.nra
ok - parse accepts digits
ok - parse rejects text
ok - points compare by value
ok - division by zero panics
FAIL - this one fails: expected 4, got 3 (age.nra:39:10)
4 passed, 1 failed
```

The exit code is 1 when any test failed, otherwise 0. Output that a test prints appears before its line. A separate process per test means a crash, a leak or a timeout (10 seconds, or `NARRA_TEST_TIMEOUT`) fails that one test and nothing else.

Options:

- `narra test -filter parse age.nra` runs only the tests whose name contains `parse`. A filter that matches nothing prints `no tests match "parse"` and exits 1.
- `narra test -time age.nra` adds the milliseconds of each test.

### Testing with the pieces from this guide

Because errors are values, tests of failure paths are plain `Test.eq` calls on an `Err`. Use `Test.panics` for the cases that must abort, and keep in mind that its closure runs in a child process: nothing the closure changes reaches the test, and it needs a POSIX system.

`Test.eq` needs `Show` and `Equatable`. Types without functions are `Equatable` automatically, and `derives Show` (chapter 7) gives the text. Leaving it out is an error at the assertion.

### Rules for `std/test`

- Its functions can be called only inside a `test` block, arrow functions inside it included. A call in an ordinary function is a compile error, even if a test calls that function:

```narra error
// error: Test.eq can only be called inside a `test` block
import * as Test from "std/test"

function check(number: int) {
  Test.eq(number, 1)
}

test "a" {
  check(1)
}

export function run(args: Array<string>): int {
  return 0
}
```

- A test block may use any effect without declaring it. The `import` may stay in a file that is also built as a program, because the test blocks are not part of that build.

That is the end of the guide.
