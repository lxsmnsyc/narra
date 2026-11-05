# 5. Pattern matching

A pattern describes the shape of a value, tests it, and binds names to the parts. Narra uses patterns in `match`, in `is`, and in `const` and `let` declarations. They replace `switch`, `null` checks and manual field access.

## `match`

`match` checks a value against patterns in order and must cover every case, or end with a `_` arm. Each arm is `pattern => expression`, or `pattern => { statements }`. An arm can add a guard with `if (...)`.

```narra
import * as Int from "std/int"

enum Token {
  Word(text: string)
  Number(value: int)
  Space
}

function describe(token: Token): string {
  return match (token) {
    Token.Word(text) => `word ${text}`
    Token.Number(value) if (value < 0) => "negative number"
    Token.Number(value) => `number ${Int.toString(value)}`
    Token.Space => "space"
  }
}

export function run(args: Array<string>): int {
  const tokens = [Token.Word("hi"), Token.Number(-3), Token.Number(42), Token.Space]
  for (const token of tokens) {
    print(describe(token))
  }
  match (tokens[0]) {
    Token.Word(text) => {
      print(`first is ${text}`)
    }
    _ => print("first is something else")
  }
  return 0
}
```

```text
word hi
negative number
number 42
space
first is hi
```

Used as a value (after `=`, as an argument or after `return`), `match` gives the value of the arm, and every arm has the same type. As a statement, an arm with a block can `return` from the function. A missing case is a compile error that names it (see the example in chapter 4). `match` works on any value: integers, strings, bools, structs and tuples, not just enums. TypeScript's `switch` falls through and checks nothing; this does neither.

## Literals, structs and tuples

Patterns nest, and a literal inside one tests the value instead of binding it.

```narra
import * as Int from "std/int"

struct Point {
  x: int
  y: int
}

function locate(point: Point): string {
  return match (point) {
    { x: 0, y: 0 } => "origin"
    { x: 0, y } => `on the y axis at ${Int.toString(y)}`
    { x, y: 0 } => `on the x axis at ${Int.toString(x)}`
    { x, y } => `at ${Int.toString(x)}, ${Int.toString(y)}`
  }
}

function classify(pair: #[int, bool]): string {
  return match (pair) {
    #[0, _] => "zero"
    #[n, true] if (n > 100) => "big and marked"
    #[_, true] => "marked"
    #[_, false] => "plain"
  }
}

export function run(args: Array<string>): int {
  print(locate(Point { x: 0, y: 0 }))
  print(locate(Point { x: 0, y: 5 }))
  print(locate(Point { x: 7, y: 0 }))
  print(locate(Point { x: 1, y: 2 }))
  print(classify(#[0, true]))
  print(classify(#[200, true]))
  print(classify(#[5, true]))
  print(classify(#[5, false]))
  return 0
}
```

```text
origin
on the y axis at 5
on the x axis at 7
at 1, 2
zero
big and marked
marked
plain
```

The compiler checks exhaustiveness through nested enums, bools, structs and tuples, so a missing combination is an error. Integers and strings have too many values to list, so a `match` on them needs a `_` arm.

## Destructuring

A `const` or `let` takes a pattern too, when the pattern always matches. Use it to take a struct or a tuple apart.

```narra
import * as Int from "std/int"

struct Config {
  host: string
  port: int
}

function divmod(a: int, b: int): #[int, int] {
  return #[a / b, a % b]
}

export function run(args: Array<string>): int {
  const config = Config { host: "localhost", port: 8080 }
  const { host, port: serverPort } = config
  print(`${host} ${Int.toString(serverPort)}`)
  const #[quotient, remainder] = divmod(17, 5)
  print(`${Int.toString(quotient)} ${Int.toString(remainder)}`)
  const pairs: Array<#[string, int]> = [#["a", 1], #["b", 2]]
  for (const #[name, count] of pairs) {
    print(`${name}=${Int.toString(count)}`)
  }
  return 0
}
```

```text
localhost 8080
3 2
a=1
b=2
```

A pattern that can fail does not work in a declaration. `const { x: 0, y } = point` is an error, because `x` might not be `0`. Use `match` or `is` for that.

## `is` and narrowing

`value is Pattern` is a `bool` expression. It binds the names of the pattern in the branch where it is true. This is how you read an `Option`, since there is no `null` and no way to get the value without checking it.

```narra
import * as Int from "std/int"
import * as Maps from "std/map"

function describeAge(ages: Map<string, int>, name: string): string {
  if (ages[name] is Some(age)) {
    return `${name} is ${Int.toString(age)}`
  }
  return `${name} is unknown`
}

function firstOrZero(items: Array<int>): int {
  const found: Option<int> = items[0] > 0 ? Some(items[0]) : None
  if (found is None) {
    return 0
  }
  return found.value
}

export function run(args: Array<string>): int {
  const ages: Map<string, int> = Map { "ada": 36 }
  print(describeAge(ages, "ada"))
  print(describeAge(ages, "alan"))
  print(Int.toString(Maps.size(ages)))
  print(Int.toString(firstOrZero([5])))
  print(Int.toString(firstOrZero([-5])))
  return 0
}
```

```text
ada is 36
alan is unknown
1
5
0
```

After `if (found is None) { return 0 }`, the compiler knows `found` is `Some` for the rest of the block, and `found.value` reads the value. Without such a test, `found.value` is an error: "check it with `is` or `match` before reading". Narrowing works on `const` bindings and parameters, and on `let` bindings that are not assigned in between. The same works for `Result` with `Err` (chapter 6). An `is` that binds names must be the whole condition of an `if` or a conditional expression, so `x is Some(v) && v > 0` is an error; nest two `if` statements instead.

```narra error
// error: 4:16: this Option<int> may hold any of its variants; check it with `is` or `match` before reading "value"
export function run(args: Array<string>): int {
  const found: Option<int> = Some(1)
  return found.value
}
```

## Block expressions and the ternary

In TypeScript, `{ ... }` in value position is an object literal. In Narra it is a block expression: a block whose value comes from `return`. A `return` inside it ends the block, not the function. Every path must end in `return value` ([SPEC.md section 2](../../SPEC.md), D75). To build a struct you write its name, `Point { x: 1 }`, so `const o = { a }` is an error rather than a silent object.

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

function lookup(code: int): string {
  if (code == 1) {
    return "one"
  }
  return ""
}

function label(code: int): string {
  const result = {
    const name = lookup(code)
    if (name == "") {
      return "unknown"
    }
    return `<${name}>`
  }
  return result
}

function firstOver(items: Array<int>, limit: int): int {
  const found = {
    for (const item of Arrays.values(items)) {
      if (item > limit) {
        return item
      }
    }
    return -1
  }
  return found
}

function sign(n: int): string {
  return n < 0 ? "negative" : n == 0 ? "zero" : "positive"
}

export function run(args: Array<string>): int {
  print(label(1))
  print(label(2))
  print(Int.toString(firstOver([3, 8, 12], 5)))
  print(sign(-4))
  print(sign(0))
  const category = if (firstOver([1], 5) < 0) { return "none" } else { return "some" }
  print(category)
  return 0
}
```

```text
<one>
unknown
8
negative
zero
none
```

`c ? a : b` works as in TypeScript, with a `bool` condition and the same type in both branches. It chains to the right. An `if` in value position needs an `else` and each branch returns its value. `break` and `continue` cannot leave a block expression.

Next: [6. Option and Result](06-option-and-result.md).
