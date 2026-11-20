# 12. Generators

A function returns once, with all of its result. A generator gives its values one at a time, and runs only as far as the next value each time it is asked. This chapter shows how to write one with `function*` and `yield` and how a loop reads it. It then covers the effects of a generator and how to resume one by hand. The last section covers coroutines, generators that also receive a value each time they go on.

## Write a generator

Declare a generator with `function*`. Its result type is `Generator<T>`, where `T` is the type of the values it gives. Inside the body, `yield value` gives one value:

```narra
import * as Int from "std/int"

function* countdown(start: int): Generator<int> {
  print("countdown starts")
  let n = start
  while (n > 0) {
    print(`about to give ${Int.toString(n)}`)
    yield n
    n--
  }
  print("countdown ends")
}

export function run(args: Array<string>): int {
  const numbers = countdown(3)
  print("made the generator")
  for (const n of numbers) {
    print(`got ${Int.toString(n)}`)
  }
  return 0
}
```

```text
made the generator
countdown starts
about to give 3
got 3
about to give 2
got 2
about to give 1
got 1
countdown ends
```

The order of the lines shows how a generator runs:

- Calling `countdown(3)` runs none of its body. It gives back a `Generator<int>` that is ready to start.
- The loop asks the generator for a value. The body runs up to the first `yield`, and the loop gets `3`.
- At `yield` the body stops where it is, with `n` kept as it was. The next time the loop asks, the body goes on from the line after the `yield`.
- When the body reaches its end, the generator is done, and the loop ends.

A generator is a top-level function, declared like any other. It can be exported with `export function*`, and it can have type parameters. There is no arrow form of a generator.

## Stop early

A generator can give values forever. The loop that reads it decides when to stop, and the body never runs past the last value the loop asked for:

```narra
import * as Int from "std/int"

function* ticketNumbers(first: int): Generator<int> {
  let number = first
  while (true) {
    yield number
    number++
  }
}

function* pricesUnder(prices: Array<int>, limit: int): Generator<int> {
  for (const price of prices) {
    if (price >= limit) {
      return
    }
    yield price
  }
}

export function run(args: Array<string>): int {
  for (const ticket of ticketNumbers(100)) {
    if (ticket == 103) {
      break
    }
    print(`ticket ${Int.toString(ticket)}`)
  }
  for (const price of pricesUnder([120, 250, 900, 80], 500)) {
    print(`price ${Int.toString(price)}`)
  }
  return 0
}
```

```text
ticket 100
ticket 101
ticket 102
price 120
price 250
```

`ticketNumbers` has no end, and `break` leaves the loop after three tickets. The generator is then dropped where it stopped.

`return` inside a generator ends it, as reaching the end of the body does. It takes no value, because the values of a generator are the ones it yields:

```narra error
// error: 4:3: return in a generator takes no value; yield the value, then return
function* numbers(): Generator<int> {
  yield 1
  return 2
}

export function run(args: Array<string>): int {
  return 0
}
```

## What a loop can read

`for ... of` reads several kinds of values, and only a generator among them runs code to make each value:

| Source | What the loop gets |
| --- | --- |
| An `Array<T>` | Each element, in order |
| `Arrays.entries(items)` | Each position and element, as a tuple |
| `Maps.keys(map)`, `Maps.entries(map)` | An array of the keys, or of the key and value tuples |
| `Str.chars(text)` | An array of the characters |
| A `Range` from `std/range` | Each number in the range |
| A `Generator<T>` | Each value the generator yields |

Arrays, ranges and maps are not generators. `Maps.keys` and `Str.chars` build an array, and a loop over an array walks the array itself. `Arrays.entries` and `Arrays.values` are allowed only at the head of a loop.

So a function with a `Generator<int>` parameter does not take an array or a range. A short generic generator turns any array into a generator:

```narra
import * as Int from "std/int"
import * as Ranges from "std/range"

function* each<T>(items: Array<T>): Generator<T> {
  for (const item of items) {
    yield item
  }
}

function* upTo(limit: int): Generator<int> {
  for (const i of Ranges.upTo(limit)) {
    yield i
  }
}

function sum(numbers: Generator<int>): int {
  let total = 0
  for (const n of numbers) {
    total = total + n
  }
  return total
}

export function run(args: Array<string>): int {
  const prices = [250, 320, 180]
  print(Int.toString(sum(each(prices))))
  print(Int.toString(sum(upTo(4))))
  return 0
}
```

```text
750
6
```

Write a parameter as `Array<T>` when the function needs every value at once, or its length. Write it as `Generator<T>` when the function reads the values once, in order. A caller can then pass values that are made one at a time and never stored together.

## Pass on another generator's values: `yield*`

`yield* other` gives every value of another generator, in turn, before the body goes on. A generator can also call itself to make a new generator. Together they walk a nested structure, such as a tree:

```narra
import * as Int from "std/int"

enum Tree {
  Leaf(value: int)
  Node(left: Tree, right: Tree)
}

function* leaves(tree: Tree): Generator<int> {
  match (tree) {
    Tree.Leaf(value) => {
      yield value
    }
    Tree.Node(left, right) => {
      yield* leaves(left)
      yield* leaves(right)
    }
  }
}

export function run(args: Array<string>): int {
  const tree = Tree.Node(Tree.Node(Tree.Leaf(1), Tree.Leaf(2)), Tree.Leaf(3))
  for (const value of leaves(tree)) {
    print(Int.toString(value))
  }
  return 0
}
```

```text
1
2
3
```

`yield* leaves(left)` is the same as a loop that yields each value of `leaves(left)`. The other generator must give values of the same type.

## A generator is a value

A generator can be stored in a `const`, passed to a function and returned, like any other value. A loop does not advance the generator it reads. It reads its own copy. So a second loop over the same `const` starts from the first value again, and runs the body's work a second time:

```narra
function* drinks(): Generator<string> {
  print("making the list")
  yield "tea"
  yield "coffee"
}

export function run(args: Array<string>): int {
  const menu = drinks()
  for (const drink of menu) {
    print(`first ${drink}`)
  }
  for (const drink of menu) {
    print(`again ${drink}`)
  }
  return 0
}
```

```text
making the list
first tea
first coffee
making the list
again tea
again coffee
```

To read a few values and go on from there later, resume the generator by hand (see "Resume a generator by hand" below).

Generators cannot be compared with `==`, used as map keys or printed with `Show`.

## Rules for the body

A generator keeps running after the call that made it has returned. Two rules for its parameters follow from that.

A generator cannot take a `mutable` parameter (chapter 9), since the generator would write to the caller's value long after the call:

```narra error
// error: 2:17: generator drain cannot take a mutable parameter "orders"; a generator is a value that is resumed later
function* drain(mutable orders: Array<int>): Generator<int> {
  for (const order of orders) {
    yield order
  }
}

export function run(args: Array<string>): int {
  return 0
}
```

A function-type parameter must be `escaping` (chapter 3), since the generator keeps it and calls it later:

```narra error
// error: 2:39: parameter "scale" of generator scaled has a function type, so it must be `escaping`: the generator outlives the call
function* scaled(numbers: Array<int>, scale: (n: int) => int): Generator<int> {
  for (const n of numbers) {
    yield scale(n)
  }
}

export function run(args: Array<string>): int {
  return 0
}
```

With `escaping` the generator compiles, and a closure passed to it captures values as an escaping closure does:

```narra
import * as Int from "std/int"

function* scaled(numbers: Array<int>, escaping scale: (n: int) => int): Generator<int> {
  for (const n of numbers) {
    yield scale(n)
  }
}

export function run(args: Array<string>): int {
  const factor = 3
  for (const n of scaled([1, 2, 5], (value) => value * factor)) {
    print(Int.toString(n))
  }
  return 0
}
```

```text
3
6
15
```

`yield` is a statement, and it belongs to the generator's own body. It can stand inside `if`, `while`, `for` and the arms of a `match` statement. It cannot stand inside an arrow function, even one written in the generator:

```narra error
// error: 4:5: yield cannot be used inside an arrow function; it is valid only directly in a function* body
function* labels(items: Array<int>): Generator<int> {
  const each = (n: int): void => {
    yield n
  }
}

export function run(args: Array<string>): int {
  return 0
}
```

A function called from a generator runs to its end before the generator goes on. Only the generator's own body can stop at a `yield`.

## Effects of a generator

A generator's body may print, read files or allocate, and those effects (chapter 10) happen when the body runs, not when the generator is made. So they are part of the generator's type: `Generator<int> uses io` is a generator whose body does input or output. A plain `Generator<int>` has no effects.

The effects count where the body runs: in a `for ... of` over the generator, in a `yield*` of it, and in a resume by hand. Making the generator costs only `alloc`, for the memory that keeps its state:

```narra
import * as Int from "std/int"

export function* loud(count: int): Generator<int> uses (io, alloc) {
  let n = 0
  while (n < count) {
    print(`making ${Int.toString(n)}`)
    yield n
    n++
  }
}

export function makeLoud(): Generator<int> uses (io, alloc) uses alloc {
  return loud(2)
}

function sum<effect E>(numbers: Generator<int> uses E): int uses E {
  let total = 0
  for (const n of numbers) {
    total = total + n
  }
  return total
}

export function run(args: Array<string>): int uses io, alloc {
  const numbers = makeLoud()
  print("made it")
  print(Int.toString(sum(numbers)))
  return 0
}
```

```text
made it
making 0
making 1
1
```

- `loud` is exported, so it writes the effects of its body on its result type. More than one effect goes in parentheses, as in a function type.
- `makeLoud` only makes the generator. Its own list, after the result type, is `uses alloc`, without `io`.
- `sum` loops over the generator, so it has the generator's effects. The effect variable `E` (chapter 10) says so, and the call `sum(numbers)` has `io` and `alloc`.

A private generator gets the effects of its body in its type without writing them. A generator with effects cannot go where a type with fewer effects is expected, because the function there would run them without declaring them:

```narra error
// error: 20:18: expected Generator<int>, found Generator<int> uses (alloc, io)
import * as Int from "std/int"

function* loud(items: Array<int>): Generator<int> {
  for (const n of items) {
    print(Int.toString(n))
    yield n
  }
}

function sum(numbers: Generator<int>): int {
  let total = 0
  for (const n of numbers) {
    total = total + n
  }
  return total
}

export function run(args: Array<string>): int uses io, alloc {
  return sum(loud([1, 2, 3]))
}
```

The lazy operators of `std/iterator` in [chapter 13](13-standard-library.md) are generators written this way. Their results carry the effects of the generator and the callbacks they are given.

## Resume a generator by hand

`Co.resume` from `std/coroutine` runs a generator to its next `yield` and gives back a `Step`:

- `Step.Yielded(value)` holds the value the generator gave.
- `Step.Done` means the body has ended. A generator that is done answers `Step.Done` to every later resume.
- `Step.Panicked(message)` appears only when the program runs as tests (chapter 10). A panic inside the generator then ends the generator, and the test sees the message. In a normal run, a panic stops the program.

`Co.resume` changes the generator it is given, so it takes a `mutable` argument, and the generator is in a `let`:

```narra
import * as Co from "std/coroutine"
import { Step } from "std/coroutine"

function* tickets(): Generator<string> {
  yield "A1"
  yield "A2"
}

export function run(args: Array<string>): int {
  let queue = tickets()
  print(Show.show(Co.resume(mutable queue)))
  if (Co.resume(mutable queue) is Step.Yielded(ticket)) {
    print(`next ticket ${ticket}`)
  }
  print(Show.show(Co.resume(mutable queue)))
  print(Show.show(Co.resume(mutable queue)))
  return 0
}
```

```text
Yielded("A1")
next ticket A2
Done
Done
```

Unlike a loop, each resume goes on where the last one stopped. Resume by hand to read the values of two generators in turn, or to read a few values now and the rest later. `Step` derives `Show` and `Equatable`, and its constructors are always written with `Step.`.

## Coroutines: generators that receive values

A coroutine is a generator that also receives a value each time it goes on. Its result type is `Coroutine<In, Out>`: it receives values of type `In` and yields values of type `Out`. Inside the body, `const input = yield output` yields `output`, stops, and binds `input` to the value the next resume sends in. `Co.send(mutable g, input)` is that resume:

```narra
import * as Co from "std/coroutine"
import { Step } from "std/coroutine"
import * as Str from "std/string"

function* longest(): Coroutine<string, string> {
  let best = ""
  while (true) {
    const word = yield best
    if (Str.length(word) > Str.length(best)) {
      best = word
    }
  }
}

export function run(args: Array<string>): int {
  let tracker = longest()
  const _start = Co.send(mutable tracker, "")
  for (const word of ["fig", "banana", "kiwi"]) {
    if (Co.send(mutable tracker, word) is Step.Yielded(best)) {
      print(`after ${word}: ${best}`)
    }
  }
  return 0
}
```

```text
after fig: fig
after banana: banana
after kiwi: banana
```

The first send starts the body, which runs to the first `yield`. No `yield` waits for that first input yet, so it is dropped. Each later send is the value of the `yield` that waits, and gets back the next value the body yields.

The rules that differ from a plain generator:

- `const input = yield output` (or `let`) works only in a coroutine, with one name and no type annotation. A plain `yield output` statement is allowed too, and drops the input.
- `Co.send` takes two arguments and works on a coroutine. `Co.resume` takes one and works on a generator.
- `yield*` is not allowed in a coroutine.
- `for ... of` cannot read a coroutine, because a loop has no value to send:

```narra error
// error: 10:26: for ... of cannot iterate a coroutine, which needs an input on each resume; call Co.send(mutable g, input)
function* echo(): Coroutine<string, string> {
  while (true) {
    const word = yield ""
    print(word)
  }
}

export function run(args: Array<string>): int {
  for (const word of echo()) {
    print(word)
  }
  return 0
}
```

A coroutine type carries effects as a generator type does: `Coroutine<string, int> uses io`.

A generator that waits for the network or a timer between its values is an async generator, declared with `async function*`. [Chapter 14](14-async.md) covers it.

Next: [13. The standard library](13-standard-library.md).
