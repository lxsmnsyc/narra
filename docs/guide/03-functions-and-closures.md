# 3. Functions and closures

This chapter covers named functions, arrow functions, function types, the `escaping` rule for function parameters, generics and tail calls.

## Named functions

A function declaration writes every parameter type and the result type. It is private to its module unless it starts with `export`. A function with no result type returns nothing.

```narra
import * as Int from "std/int"

function gcd(a: int, b: int): int {
  if (b == 0) {
    return a
  }
  return gcd(b, a % b)
}

function announce(label: string, value: int) {
  print(`${label}: ${Int.toString(value)}`)
}

export function run(args: Array<string>): int {
  announce("gcd", gcd(1071, 462))
  return 0
}
```

```text
gcd: 21
```

The rules:

- A call passes every argument. There are no optional parameters, no varargs, no overloads and no currying. To fix some arguments, write an arrow function. To pass options, pass a struct with default field values (chapter 4).
- There are no method calls. `items.push(1)` is a compile error, and the message tells you to write `Arrays.push(mutable items, 1)`. `value.field(args) is valid only when `field` is a struct field that holds a function.
- You cannot declare a function inside a function. Declare it at module level and pass what it needs as arguments, or bind an arrow function to a `const`.
- Parameters are constant. A parameter changes only if it is declared `mutable` (chapter 9).
- Recursion needs no keyword.

```narra error
// error: 3:3: a function inside a function is not supported; declare it at module level as a private function and pass the enclosing values as arguments
export function run(args: Array<string>): int {
  function add(a: int): int {
    return a + 1
  }
  return add(1)
}
```

## Tail calls are loops

A call to the same function in tail position compiles to a jump, so it does not grow the stack. Other recursion uses the stack, and an overflow aborts the program.

```narra
import * as Int from "std/int"

function sumTo(n: int, total: int): int {
  if (n == 0) {
    return total
  }
  return sumTo(n - 1, total + n)
}

export function run(args: Array<string>): int {
  print(Int.toString(sumTo(10_000_000, 0)))
  return 0
}
```

```text
50000005000000
```

The call must be the whole `return` expression. `return n + sumTo(n - 1)` has work left after the call, so it uses the stack.

## Arrow functions and function types

An arrow function is an anonymous function: `(x: int): int => x + 1`, or with a block body `(x: int): int => { ... }`. A function type is written the same way: `(x: int) => int`. The `function` keyword is only for named declarations.

An arrow function passed as an argument takes its parameter types from the parameter it is passed to, so you can leave them out. A `const` arrow function needs its types written.

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

function map<T, U>(items: Array<T>, f: (item: T) => U): Array<U> {
  let out: Array<U> = []
  for (const item of Arrays.values(items)) {
    Arrays.push(mutable out, f(item))
  }
  return out
}

function double(x: int): int {
  return x * 2
}

export function run(args: Array<string>): int {
  const rate = 10
  const scaled = map([1, 2, 3], (price) => price * rate)
  print(Int.toString(scaled[2]))
  const labels = map(scaled, (price) => `#${Int.toString(price)}`)
  print(labels[0])
  const doubled = map([1, 2, 3], double)
  print(Int.toString(doubled[1]))
  const triple = (x: int): int => x * 3
  print(Int.toString(triple(4)))
  return 0
}
```

```text
30
#10
4
12
```

`map<T, U>` is a generic function. Its type arguments are inferred from the call, and you cannot write them explicitly yet. A named function is a value too, as `double` is above. The standard library does not provide `Arrays.map` in this version, so the sample defines `map` itself.

## Closures capture constants

A closure can read the `const` bindings and parameters around it. It cannot capture a `let` binding, because the environment of a closure never changes after it is made, so a closure can never see a later assignment.

```narra error
// error: 8:22: a closure cannot capture "count" because it is declared with let; closures capture only constants
function apply(f: () => int): int {
  return f()
}

export function run(args: Array<string>): int {
  let count = 1
  return apply(() => count)
}
```

Copy the value into a `const` first, and capture that.

## Function parameters do not escape

A parameter of function type can only be called inside the function. It cannot be stored, returned, or captured by another closure. In return, the compiler puts the closure's environment on the caller's stack, so passing a closure costs no allocation.

To keep a function beyond the call, mark the parameter `escaping`. Its environment is then allocated on the heap and reference counted.

```narra error
// error: 3:10: parameter "f" cannot escape: it can only be called or passed to a function parameter that is not escaping; declare it as `escaping f` to store, return or capture it
function same(f: (x: int) => int): (x: int) => int {
  return f
}

export function run(args: Array<string>): int {
  return same((x) => x)(1)
}
```

```narra
import * as Int from "std/int"

function adder(n: int): (x: int) => int {
  return (x: int): int => x + n
}

function compose(escaping f: (x: int) => int, escaping g: (x: int) => int): (x: int) => int {
  return (x) => g(f(x))
}

struct Button {
  label: string
  onClick: (times: int) => string
}

export function run(args: Array<string>): int {
  const addFive = adder(5)
  const both = compose(addFive, (x) => x * 2)
  print(Int.toString(both(1)))
  const button = Button { label: "ok", onClick: (times) => `clicked ${Int.toString(times)}` }
  print(button.onClick(3))
  return 0
}
```

```text
12
clicked 3
```

The function in a struct field or array element always counts as escaping, so it needs no keyword there. A parameter whose type is a type parameter (`T`) is escaping as well.

Next: [4. Structs and enums](04-structs-and-enums.md).
