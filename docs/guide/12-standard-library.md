# 12. The standard library

The earlier chapters used a few standard modules along the way: `std/int`, `std/string`, `std/array`, `std/option`. This chapter goes through the modules that larger programs, such as command-line tools, are built on. Each section stands on its own.

## Iterators: `std/iterator`

A generator, written `function*`, gives values one at a time, and a `for ... of` loop asks it for the next one only when it needs it. `std/iterator` holds the usual operations over generators:

```narra
import * as Iterators from "std/iterator"
import * as Int from "std/int"

function* naturals(): Generator<int> {
  let n = 0
  while (true) {
    yield n
    n = n + 1
  }
}

export function run(args: Array<string>): int uses io, alloc {
  const odd = Iterators.filter(naturals(), (value: int): bool => value % 2 == 1)
  const squares = Iterators.map(odd, (value: int): int => value * value)
  const firstFour = Iterators.take(squares, 4)
  for (const square of firstFour) {
    print(Int.toString(square))
  }
  const total = Iterators.reduce(
    Iterators.take(naturals(), 5),
    (sum: int, value: int): int => sum + value,
    0,
  )
  print(`sum ${Int.toString(total)}`)
  return 0
}
```

```text
1
9
25
49
sum 10
```

`naturals` never ends, and the program still stops. `filter`, `map` and `take` are lazy: each returns a new generator and does no work until a loop asks it for a value. The loop asks `take` for a value, `take` asks `map`, `map` asks `filter`, and `filter` asks `naturals`, until four squares have come through. Each stage gets a name with `const`, so the program reads from top to bottom.

The lazy operators are `map`, `filter`, `flatMap`, `take`, `drop`, `takeWhile`, `dropWhile`, `enumerate` (pairs of an index and a value), `zip` (pairs from two generators, ending with the shorter one), `concat`, `chunk` and `window`.

The eager operators run the generator to its end, or to the first answer: `reduce`, `toArray`, `forEach`, `some`, `every`, `find` and `count`. `find` returns an `Option`, since no value may match.

`chunk` groups the values into arrays of a given size, and `window` gives every run of that many values in a row. A size below 1 makes no sense, so both return an `Option`, and `None` for such a size:

```narra
import * as Iterators from "std/iterator"
import * as Int from "std/int"
import * as Str from "std/string"
import * as Arrays from "std/array"

function* letters(): Generator<string> {
  yield "a"
  yield "b"
  yield "c"
  yield "d"
  yield "e"
}

export function run(args: Array<string>): int uses io, alloc {
  if (Iterators.chunk(letters(), 2) is Some(pairs)) {
    for (const pair of pairs) {
      print(Str.join(pair, ""))
    }
  }
  if (Iterators.chunk(letters(), 0) is None) {
    print("no chunks of size 0")
  }
  const found = Iterators.find(letters(), (letter: string): bool => letter != "a" && letter != "b")
  print(match (found) { Some(letter) => letter, None => "none" })
  print(Int.toString(Arrays.length(Iterators.toArray(letters()))))
  return 0
}
```

```text
ab
cd
e
no chunks of size 0
c
5
```

### Callbacks that are kept

A lazy operator keeps its callback inside the generator it returns, and calls it later. So its callback is `escaping` (chapter 3), and it cannot capture a `let`:

```narra error
// error: 12:74: a closure cannot capture "limit" because it is declared with let, unless it is passed straight to a function parameter that is not escaping; to share state that changes with an escaping closure, keep it in a Cell from std/cell
import * as Iterators from "std/iterator"

function* digits(): Generator<int> {
  for (const digit of [1, 2, 3, 4, 5]) {
    yield digit
  }
}

export function run(args: Array<string>): int uses io, alloc {
  let limit = 3
  const small = Iterators.filter(digits(), (value: int): bool => value < limit)
  return Iterators.count(small)
}
```

Use a `const`, or a `Cell` for a value that changes. An eager operator calls its callback before it returns, so its callback is not escaping and may capture a `let`.

### The source does not move

An operator takes its generator as a value, like any other argument. A loop inside the operator runs on its own copy, so the generator you passed in is still at its start:

```narra
import * as Iterators from "std/iterator"
import * as Co from "std/coroutine"
import { Step } from "std/coroutine"

function* words(): Generator<string> {
  yield "one"
  yield "two"
  yield "three"
}

export function run(args: Array<string>): int uses io, alloc {
  let source = words()
  for (const word of Iterators.take(source, 2)) {
    print(`taken ${word}`)
  }
  for (const word of source) {
    print(`again ${word}`)
  }
  if (Co.resume(mutable source) is Step.Yielded(word)) {
    print(`by hand ${word}`)
  }
  if (Co.resume(mutable source) is Step.Yielded(word)) {
    print(`by hand ${word}`)
  }
  return 0
}
```

```text
taken one
taken two
again one
again two
again three
by hand one
by hand two
```

The second loop starts from `"one"` again, and runs the generator's work a second time. To take a few values and go on from there, resume the generator yourself with `Co.resume(mutable source)`: it moves `source` itself, and each resume goes on where the last one stopped.

### Effects

The result of a lazy operator has the effects of its source and its callback. Mapping over a generator that prints gives a generator that prints, and its type says `uses io`. A pure source with a pure callback gives a pure generator.
