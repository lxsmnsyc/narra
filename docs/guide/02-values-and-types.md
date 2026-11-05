# 2. Values and types

This chapter covers `const` and `let`, numbers, strings and templates, `bool`, arrays, maps and tuples.

## Bindings: `const` and `let`

`const` binds a name that cannot be reassigned, and the value never changes either. `let` binds a name you may reassign.

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const limit = 4
  let total = 0
  let step = 0
  while (step < limit) {
    total = total + step
    step = step + 1
  }
  print(`total ${Int.toString(total)}`)
  return 0
}
```

```text
total 6
```

Where this differs from TypeScript ([SPEC.md section 2, Bindings](../../SPEC.md), D24 and D38 in [DECISIONS.md](../../DECISIONS.md)):

- `const` is deep. A `const` array cannot change, while a TypeScript one can.
- The compiler has no compound assignment yet. Write `total = total + step`, not `total += step` or `step++`. D76 accepts both forms, and the guide will use them when the compiler does.
- Shadowing is an error. You cannot declare a name while an earlier binding with that name is in scope.
- An unused local, parameter or import is an error. Start the name with `_` to say it is unused on purpose.
- A closure cannot capture a `let` binding (chapter 3).

```narra error
// error: 5:5: "value" is already declared in this function; shadowing is not allowed
export function run(args: Array<string>): int {
  const value = 1
  if (value > 0) {
    const value = 2
  }
  return 0
}
```

Types are inferred for locals: `const n = 1` is an `int`. Function parameters and results are always written.

## Numbers

| Type | Meaning |
| --- | --- |
| `int` | 64-bit signed integer (alias of `i64`) |
| `i8` `i16` `i32` `i64`, `u8` `u16` `u32` `u64` | sized integers |
| `f64`, `f32` | floats |

TypeScript has one `number`. Narra has these types and no implicit conversions between them (D30), so `1 + 2.0` is a compile error. Convert with `Float.fromInt` and `Float.truncate`.

```narra
import * as Int from "std/int"
import * as Float from "std/float"

export function run(args: Array<string>): int {
  print(Int.toString(7 / 2))
  print(Int.toString(-7 / 2))
  print(Int.toString(7 % 3))
  print(Int.toString(0xFF + 0b101 + 1_000))
  print(Float.toString(1.5 * 4.0))
  print(Float.toString(Float.sqrt(2.0)))
  const half = Float.fromInt(7) / 2.0
  print(Float.toString(half))
  print(Int.toString(Float.truncate(half)))
  return 0
}
```

```text
3
-3
1
1260
6.0
1.4142135623730951
3.5
3
```

- `/` on integers truncates toward zero.
- A number written without a decimal point is an `int`, and with one an `f64`.
- Integer overflow panics in a debug build and wraps in a release build. The operators `+%`, `-%` and `*%` always wrap.

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const big = 9223372036854775807
  print(Int.toString(big +% 1))
  return 0
}
```

```text
-9223372036854775808
```

## Strings, characters and templates

A `string` is immutable UTF-8 text in double quotes. A `char` is one Unicode scalar value in single quotes. In TypeScript, both quote styles make strings. Here `'a'` is a `char`.

A template string uses backticks. Each `${...}` holds an expression of type `string`, so a number goes through `Int.toString` first. There is no implicit conversion, and there is no `+` for strings (D12). Join with a template or with `Str.concat`.

```narra
import * as Int from "std/int"
import * as Str from "std/string"

export function run(args: Array<string>): int {
  const name = "Narra"
  const greeting = `Hello, ${name}!`
  print(greeting)
  print(Int.toString(Str.length(greeting)))
  print(Str.concat("ab", "cd"))
  print(Str.repeat("-", 5))
  print(Str.fromChar('λ'))
  if ("apple" < "banana") {
    print("strings compare in order")
  }
  return 0
}
```

```text
Hello, Narra!
13
abcd
-----
λ
strings compare in order
```

`Str.length` is a length in bytes. Positions in a string are byte offsets, and slicing checks that they fall on a character boundary (D15, D60).

```narra
import * as Int from "std/int"
import * as Str from "std/string"

export function run(args: Array<string>): int {
  const line = "key = value"
  const found = Str.indexOf(line, "=")
  if (found is Some(position)) {
    const key = Str.trim(Str.slice(line, 0, position))
    const value = Str.trim(Str.sliceFrom(line, position + 1))
    print(`${key}|${value}`)
  }
  print(Int.toString(Str.charCount("héllo")))
  print(Int.toString(Str.length("héllo")))
  return 0
}
```

```text
key|value
5
6
```

## Booleans and conditions

`bool` is `true` or `false`, and a condition must be a `bool`. There is no truthiness: `0`, `""` and `None` are not conditions (D21). `==` compares by structure with no conversion, and `===` is a compile error.

```narra error
// error: 3:7: expected bool, found int
export function run(args: Array<string>): int {
  if (1) {
    return 1
  }
  return 0
}
```

## Arrays

`Array<T>` is a growable array. There is no `T[]` type (D23). Indexes start at 0, and an index out of range panics. The functions on arrays come from `std/array`, imported as `Arrays` by convention, and there are no methods (D12): `Arrays.push(mutable items, 4)`, not `items.push(4)`.

The word `mutable` at the call says that the function may change your variable. Chapter 9 explains the rules.

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"
import * as Ranges from "std/range"

export function run(args: Array<string>): int {
  let items: Array<int> = [1, 2, 3]
  Arrays.push(mutable items, 4)
  items[0] = 10
  let total = 0
  for (const item of Arrays.values(items)) {
    total = total + item
  }
  print(`${Int.toString(Arrays.length(items))} items, total ${Int.toString(total)}`)
  let squares = Arrays.filled(0, 3)
  for (const i of Ranges.upTo(3)) {
    squares[i] = i * i
  }
  for (const #[index, value] of Arrays.entries(squares)) {
    print(`${Int.toString(index)}: ${Int.toString(value)}`)
  }
  return 0
}
```

```text
4 items, total 19
0: 0
1: 1
2: 4
```

`Arrays.values` and `Arrays.entries` are the ways to loop over an array, and `Ranges.upTo(n)` and `Ranges.between(start, end)` are the way to count. There is no `for ... in`. A C-style `for (init; cond; step)` loop was accepted in D76 but the compiler does not have it yet, so use `while` when the step is not a range.

An array behaves as a value. Assigning it copies it, and the copy happens only when one of the two is changed (copy on write, [SPEC.md section 6](../../SPEC.md)).

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

export function run(args: Array<string>): int {
  let original: Array<int> = [1, 2, 3]
  const snapshot = original
  Arrays.push(mutable original, 4)
  print(Int.toString(Arrays.length(original)))
  print(Int.toString(Arrays.length(snapshot)))
  return 0
}
```

```text
4
3
```

In TypeScript both names would point to the same array. In Narra `snapshot` keeps its three items.

## Maps

`Map<K, V>` is a hash map that keeps insertion order. Write one with `Map { key: value }`, and an empty one with a type from context. A key that is not a literal goes in brackets. Reading `m[key]` gives an `Option<V>`, because the key may be missing. There is no `null` (chapter 5 shows how to read an `Option`). Float keys are an error.

```narra
import * as Int from "std/int"
import * as Maps from "std/map"

export function run(args: Array<string>): int {
  let ages: Map<string, int> = Map { "ada": 36, "grace": 45 }
  const newcomer = "linus"
  ages[newcomer] = 28
  ages["ada"] = 37
  for (const name of Maps.keys(ages)) {
    if (ages[name] is Some(age)) {
      print(`${name} ${Int.toString(age)}`)
    }
  }
  print(Int.toString(Maps.size(ages)))
  if (ages["nobody"] is None) {
    print("no such key")
  }
  const removed = Maps.remove(mutable ages, "grace")
  assert(removed)
  const empty = Map<string, int> {}
  print(Int.toString(Maps.size(empty)))
  return 0
}
```

```text
ada 37
grace 45
linus 28
3
no such key
0
```

Writing the type arguments, as in `Map<string, int> {}`, gives the literal its type without an annotation (D68).

## Tuples

A tuple groups two or more values of different types. The type and the literal both use `#[...]`, because `[a, b]` is always an array (D53). A function that returns several values returns a tuple.

```narra
import * as Int from "std/int"

function divmod(a: int, b: int): #[int, int] {
  return #[a / b, a % b]
}

export function run(args: Array<string>): int {
  const #[quotient, remainder] = divmod(17, 5)
  print(`${Int.toString(quotient)} remainder ${Int.toString(remainder)}`)
  const pair: #[string, bool] = #["ready", true]
  const #[label, flag] = pair
  print(`${label} ${flag ? "yes" : "no"}`)
  return 0
}
```

```text
3 remainder 2
ready yes
```

A tuple has no element access by index. Take it apart with a pattern, as above. The conditional expression `c ? a : b` works as in TypeScript, with a `bool` condition.

Next: [3. Functions and closures](03-functions-and-closures.md).
