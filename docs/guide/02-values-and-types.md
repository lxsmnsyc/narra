# 2. Values and types

This chapter builds a small shop receipt, one piece at a time. On the way you meet the two ways to name a value, and the values a program works with: numbers, text, true and false, lists, lookup tables and pairs.

## Name a value

A receipt starts with a price and a quantity. Give each a name with `const` and compute the total:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const priceCents = 250
  const quantity = 3
  const totalCents = priceCents * quantity
  print(`total: ${Int.toString(totalCents)} cents`)
  return 0
}
```

```text
total: 750 cents
```

`const priceCents = 250` names the value `250`. You did not write a type: the compiler sees that `250` is a whole number and gives `priceCents` the type `int`. You can write it out as `const priceCents: int = 250`, and the meaning is the same.

The text in backticks is a template string. Each `${...}` inside it is replaced by the value of the expression it holds. That expression must be a `string`, so the number goes through `Int.toString` first.

A `const` name keeps its value. If you try to give it a new one, the compiler stops you and says what to write instead:

```narra error
// error: 4:3: "totalCents" is constant; declare it with `let` to change it
export function run(args: Array<string>): int {
  const totalCents = 750
  totalCents = 700
  return 0
}
```

## Change a value over time

Some values do change. Say a customer saves 150 cents a week toward a 1000 cent purchase, and you want the number of weeks. Two values change as the weeks pass, so declare them with `let`:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const goalCents = 1000
  let savedCents = 0
  let weeks = 0
  while (savedCents < goalCents) {
    savedCents = savedCents + 150
    weeks = weeks + 1
  }
  print(`${Int.toString(weeks)} weeks, ${Int.toString(savedCents)} cents saved`)
  return 0
}
```

```text
7 weeks, 1050 cents saved
```

`while` runs its block again and again while the condition in parentheses is true. Each pass assigns new values to the two `let` names with `=`.

You add to a value with `+=`. `-=`, `*=`, `/=` and the other compound forms work the same way, and `weeks++` and `weeks--` add and subtract one. They are statements: they have no value, so `const copy = weeks++` is an error, and so is the prefix form `++weeks`.

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int uses io, alloc {
  let weeks = 0
  weeks += 2
  weeks++
  print(`${Int.toString(weeks)} weeks`)
  return 0
}
```

```text
3 weeks
```

Use `const` for every name that does not need a new value, and `let` only where it does. A reader then knows which names to watch. The compiler checks it: a `let` that is never assigned after its declaration is an error, and its fix writes `const` for you. As the sections on arrays and structs show, `const` fixes the name, not the contents: the elements of a `const` array can still change.

```narra error
// error: 3:3: "weeks" is never reassigned; declare it with `const`
export function run(args: Array<string>): int {
  let weeks = 7
  return weeks
}
```

An inner block, such as the body of an `if` or a loop, cannot declare a name that is still in scope from outside it. That way a name never silently changes meaning partway through a function:

```narra error
// error: 5:5: "weeks" is already declared in this function; shadowing is not allowed
export function run(args: Array<string>): int {
  const weeks = 7
  if (weeks > 4) {
    const weeks = 4
  }
  return 0
}
```

Pick another name, such as `cappedWeeks`. Two blocks side by side, such as two loops, can each declare a `const` of the same name. The first one is out of scope when the second begins.

## Whole numbers and decimals

Now split a bill. Seven items cost 1000 cents in all. What does one cost on average?

```narra
import * as Int from "std/int"
import * as Float from "std/float"

export function run(args: Array<string>): int {
  const totalCents = 1000
  const items = 7
  print(Int.toString(totalCents / items))
  print(Int.toString(totalCents % items))
  const average = Float.fromInt(totalCents) / Float.fromInt(items)
  print(Float.toString(average))
  if (Float.truncate(average * 10.0) is Some(tenths)) {
    print(Int.toString(tenths))
  }
  return 0
}
```

```text
142
6
142.85714285714286
1428
```

Narra has two kinds of numbers. An `int` is a whole number, and `/` on two `int` values drops the fraction, so `1000 / 7` is `142`. `%` gives the remainder, `6`. An `f64` is a decimal number, written with a decimal point, such as `2.5` or `10.0`. `Float.fromInt` turns an `int` into an `f64`, and `Float.truncate` turns an `f64` back into an `int` by dropping the fraction. It returns an `Option<int>`, because a NaN or a number too large for an `int` has no `int` to give: `None`.

The two kinds never mix on their own. Multiplying an `f64` by `3` is an error, because `3` is an `int`:

```narra error
// error: 4:23: * needs two numbers of the same type, found f64 and int
export function run(args: Array<string>): int {
  const price = 2.5
  const total = price * 3
  return 0
}
```

Write `3.0` instead, or convert with `Float.fromInt`. Every conversion is visible in the code, so you always know where a value may lose its fraction.

You can write long numbers with `_` between the digits, as in `1_000_000`. Hexadecimal and binary numbers start with `0x` and `0b`, as in `0xFF` and `0b101`.

An `int` holds values up to about nine quintillion (9223372036854775807). Going past that wraps around, in every build: the largest value plus 1 is the smallest. Hash functions rely on this. When an overflow would be a bug, `Int.addChecked`, `Int.subChecked` and `Int.mulChecked` return `None` instead of a wrapped number.

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const largest = 9223372036854775807
  print(Int.toString(largest + 1))
  return 0
}
```

```text
-9223372036854775808
```

### Dividing by a name

Division by zero has no answer, and Narra does not let a program reach it. The compiler accepts `/` and `%` on integers only when it can see that the divisor is not zero. A non-zero number written in the code is fine, which is why `totalCents / items` above compiled: `items` is a `const` set from `7`. A name that comes from somewhere else, such as a parameter, is not fine until you check it:

```narra error
// error: 3:23: the divisor must be a non-zero constant or a name checked against zero first; check the name first, as in `if (d != 0) { ... }`
function average(totalCents: int, items: int): int {
  return totalCents / items
}

export function run(args: Array<string>): int {
  return average(10, 2)
}
```

Check the name with `if`, and the compiler knows the answer inside the block. Any comparison that rules out zero works: `items != 0`, `items > 0`, `items >= 1` and the like. A check that ends the function early counts for the lines after it, and in `a && b` the check on the left counts for `b`:

```narra
import * as Int from "std/int"

function share(totalCents: int, people: int): int {
  if (people <= 0) {
    return 0
  }
  return totalCents / people
}

function perItem(totalCents: int, items: int): int {
  if (items != 0 && totalCents / items > 100) {
    return totalCents / items
  }
  return 0
}

export function run(args: Array<string>): int {
  print(Int.toString(share(900, 4)))
  print(Int.toString(perItem(900, 3)))
  return 0
}
```

```text
225
300
```

A check proves a fact about a local name, and only until the name changes. Assigning to it ends the fact, with `=`, `+=` or `++`. So does passing it as `mutable`, and so does the start of a loop that does either. A field such as `order.items` or an array element is never narrowed, and neither is an expression such as `count - 1`. Copy the value into a local with `const`, then check the local:

```narra error
// error: 9:24: the divisor must be a non-zero constant or a name checked against zero first; copy it into a local first, then check the local
struct Order {
  items: int
}

export function run(args: Array<string>): int {
  const order = Order { items: 3 }
  if (order.items != 0) {
    return 100 / order.items
  }
  return 0
}
```

The same rule covers `%`, `/=` and `%=`. Float division needs no check. [Chapter 3](03-functions-and-closures.md) shows how a function can require a non-zero argument, so that its callers prove it.

### Division when the divisor may be zero

When you want "no result" as an answer, `Int.divmod` and `Int.floorDiv` take any divisor and return an `Option`: `None` for zero. `Int.divmod(a, b)` gives the quotient and remainder as a pair, truncating like `/` and `%`. `Int.floorDiv(a, b)` rounds toward negative infinity, so it differs from `/` on negative numbers:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  print(Int.toString(-7 / 2))
  if (Int.floorDiv(-7, 2) is Some(rounded)) {
    print(Int.toString(rounded))
  }
  if (Int.divmod(7, 2) is Some(#[quotient, remainder])) {
    print(`${Int.toString(quotient)} remainder ${Int.toString(remainder)}`)
  }
  if (Int.divmod(7, 0) is None) {
    print("no answer for 0")
  }
  return 0
}
```

```text
-3
-4
3 remainder 1
no answer for 0
```

`Int.compare(a, b)` returns an `Ordering`: `Ordering.Less`, `Ordering.Equal` or `Ordering.Greater`.

### Powers

There is no power operator. `Math.powInt(base, exponent)` from `std/math` raises an `int` to a power, and `Math.pow(base, exponent)` does the same for `f64` values. `Math.powInt` returns an `Option<int>`, and a negative exponent gives `None`:

```narra
import * as Int from "std/int"
import * as Float from "std/float"
import * as Math from "std/math"

export function run(args: Array<string>): int {
  if (Math.powInt(2, 10) is Some(kilo)) {
    print(Int.toString(kilo))
  }
  print(Float.toString(Math.pow(2.0, 0.5)))
  return 0
}
```

```text
1024
1.4142135623730951
```

## Text

The receipt needs lines of text. Build one item line, padded with dots:

```narra
import * as Int from "std/int"
import * as Str from "std/string"

export function run(args: Array<string>): int {
  const name = "tea"
  const priceText = Int.toString(250)
  const width = 16
  if (Str.repeat(".", width - Str.length(name) - Str.length(priceText)) is Some(dots)) {
    print(`${name}${dots}${priceText}`)
  }
  if (Str.repeat("=", width) is Some(rule)) {
    print(rule)
  }
  return 0
}
```

```text
tea..........250
================
```

A `string` is text in double quotes. The functions for strings come from `std/string`, imported here as `Str`: `Str.length` counts the bytes of a string, and `Str.repeat` repeats it. A negative count cannot be repeated, so `Str.repeat` does not return a string. It returns an `Option<string>`: `Some(text)` when it worked, `None` when it did not. `if (... is Some(dots))` runs its block with the text bound to `dots`. The section on maps below explains `Option` fully.

To join strings, put them in a template string, as the line above does. `+` does not join strings, and the compiler points you to templates:

```narra error
// error: 3:15: strings are joined with a template string, not +
export function run(args: Array<string>): int {
  print("tea" + "...")
  return 0
}
```

To add to a string you already have, use `+=`. It is the one place `+` joins strings, and it is a statement like the other compound forms. `Str.concat(a, b)` joins two strings into a new one:

```narra
import * as Int from "std/int"
import * as Str from "std/string"

export function run(args: Array<string>): int {
  let line = "tea"
  line += " x "
  line += Int.toString(3)
  print(line)
  print(Str.concat("total: ", "750"))
  return 0
}
```

```text
tea x 3
total: 750
```

To add many pieces in a loop, use a string builder instead. `Str.builder()` starts an empty one, `Str.append(mutable out, text)` and `Str.appendChar(mutable out, char)` add to it, and `Str.build(out)` gives the finished string:

```narra
import * as Int from "std/int"
import * as Str from "std/string"

export function run(args: Array<string>): int {
  const prices = [250, 320, 180]
  let out = Str.builder()
  for (const price of prices) {
    Str.append(mutable out, Int.toString(price))
    Str.appendChar(mutable out, ',')
  }
  print(Str.build(out))
  return 0
}
```

```text
250,320,180,
```

Text often arrives as one line you need to take apart. `Str.split` cuts a string at each separator and gives the pieces as an array. Arrays come later in this chapter. For now, `parts[0]` is the first piece and `parts[1]` the second:

```narra
import * as Str from "std/string"

export function run(args: Array<string>): int {
  const line = "  coffee = 320 "
  if (Str.split(line, "=") is Some(parts)) {
    const name = Str.trim(parts[0])
    const price = Str.trim(parts[1])
    print(`[${name}] costs [${price}]`)
    if (Str.slice(name, 0, 3) is Some(start)) {
      print(start)
    }
  }
  return 0
}
```

```text
[coffee] costs [320]
cof
```

`Str.trim` removes spaces at both ends, and `Str.slice(name, 0, 3)` takes the bytes from position 0 up to, but not including, position 3. Both `Str.split` and `Str.slice` return an `Option`: `Str.split` gives `None` when the separator is empty, and `Str.slice` gives `None` when a position is outside the string or inside a character.

A string holds UTF-8 text, where a letter such as `é` takes two bytes. So `Str.length` counts bytes, and `Str.charCount` counts characters:

```narra
import * as Int from "std/int"
import * as Str from "std/string"

export function run(args: Array<string>): int {
  print(Int.toString(Str.length("café")))
  print(Int.toString(Str.charCount("café")))
  print(Str.fromChar('é'))
  return 0
}
```

```text
5
4
é
```

A single character in single quotes, such as `'é'`, is a `char`, not a `string`. `Str.fromChar` turns a `char` into a one-character string.

## Make a decision

Shops give a discount on large orders. `if` runs a block only when its condition is true, and `else` gives the other case:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const totalCents = 1250
  const isLarge = totalCents >= 1000
  if (isLarge) {
    print(`discount applied: ${Int.toString(totalCents - 100)}`)
  } else {
    print(`no discount: ${Int.toString(totalCents)}`)
  }
  if (isLarge && totalCents % 2 == 0) {
    print("even total")
  }
  return 0
}
```

```text
discount applied: 1150
even total
```

With more than two cases, chain them with `else if`. The conditions are tested from the top, and only the first true one runs its block:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const totalCents = 2400
  let discountCents = 0
  if (totalCents >= 5000) {
    discountCents = 500
  } else if (totalCents >= 2000) {
    discountCents = 200
  } else if (totalCents >= 1000) {
    discountCents = 100
  }
  print(`discount: ${Int.toString(discountCents)}`)
  return 0
}
```

```text
discount: 200
```

`totalCents >= 1000` is a `bool`, a value that is either `true` or `false`. The comparisons are `==`, `!=`, `<`, `<=`, `>` and `>=`, and you combine conditions with `&&` (and), `||` (or) and `!` (not).

A condition must be a `bool`. A number is not one, so `if (count)` does not compile. Write the comparison you mean, such as `count > 0`:

```narra error
// error: 4:7: expected bool, found int
export function run(args: Array<string>): int {
  const count = 3
  if (count) {
    print("not empty")
  }
  return 0
}
```

When both cases are values, `condition ? a : b` picks one. Only the chosen side runs:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const totalCents = 1250
  const label = totalCents >= 1000 ? "large" : "small"
  print(`${label} ${Int.toString(totalCents)}`)
  return 0
}
```

```text
large 1250
```

### Long lines

A line break ends a statement when the line could end there. A line that ends with an operator, a `,` or an opening `(`, `[` or `{` goes on to the next line. You can break after `+` or inside a call:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const totalCents = 1250
  const withTax = totalCents +
    totalCents / 10
  print(Int.toString(withTax))
  return 0
}
```

```text
1375
```

A line that starts with `?` or `:` does not continue the line above, so a `? :` over several lines needs parentheses around it:

```narra
export function run(args: Array<string>): int {
  const totalCents = 1250
  const label = (totalCents >= 1000
    ? "large"
    : "small")
  print(label)
  return 0
}
```

```text
large
```

Without the parentheses the compiler stops and tells you to add them:

```narra error
// error: 5:5: a conditional expression that spans lines must be wrapped in parentheses, as in `(cond ? a : b)`
export function run(args: Array<string>): int {
  const totalCents = 1250
  const label = totalCents >= 1000
    ? "large"
    : "small"
  print(label)
  return 0
}
```

## Lists of values

A receipt has many prices. An array holds a list of values of one type, written in square brackets. The type of a list of `int` values is `Array<int>`.

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

export function run(args: Array<string>): int {
  let prices: Array<int> = [250, 320, 180]
  Arrays.push(mutable prices, 400)
  prices[0] = 200
  let totalCents = 0
  for (const price of prices) {
    totalCents = totalCents + price
  }
  print(`${Int.toString(Arrays.length(prices))} items, ${Int.toString(totalCents)} cents`)
  print(`first: ${Int.toString(prices[0])}`)
  return 0
}
```

```text
4 items, 1100 cents
first: 200
```

`prices[0]` is the first element, because positions start at 0. Reading a position past the end stops the program with a message. `for (const price of prices)` runs its block once per element, with `price` bound to each element in turn.

The functions for arrays come from `std/array`, imported as `Arrays`, and each takes the array as its first argument. `Arrays.push(mutable prices, 400)` adds an element at the end. The word `mutable` in the call marks the argument that the function changes. You can see at the call which of your values may change. A `const` or a `let` name can be passed that way. Chapter 9 explains `mutable` in full.

Writing the call as a method on the array, `prices.push(400)`, is an error that points you back to the module function:

```narra error
// error: 6:10: there are no method calls; call a module function such as Module.push(value, ...)
import * as Arrays from "std/array"

export function run(args: Array<string>): int {
  let prices: Array<int> = [250]
  prices.push(400)
  return Arrays.length(prices)
}
```

To count through positions, loop over a range. `Ranges.upTo(4)` counts 0, 1, 2, 3, and `Arrays.entries` gives each position together with its element:

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"
import * as Ranges from "std/range"

export function run(args: Array<string>): int {
  let squares: Array<int> = []
  if (Arrays.filled(0, 4) is Some(zeros)) {
    squares = copy(zeros)
  }
  for (const i of Ranges.upTo(4)) {
    squares[i] = i * i
  }
  for (const #[position, square] of Arrays.entries(squares)) {
    print(`${Int.toString(position)}: ${Int.toString(square)}`)
  }
  return 0
}
```

```text
0: 0
1: 1
2: 4
3: 9
```

`Arrays.filled(0, 4)` makes an array of four zeros, as `Some(zeros)`; a negative count gives `None`. The `#[position, square]` in the loop takes apart a pair of values. Pairs come at the end of this chapter.

To count down or in strides, give the range a step. `Ranges.stepped(start, end, step)` stops before `end`, as the plain range does. A negative step counts down while the value is above `end`:

```narra
import * as Int from "std/int"
import * as Ranges from "std/range"

export function run(args: Array<string>): int {
  for (const i of Ranges.stepped(10, 0, -4)) {
    print(Int.toString(i))
  }
  for (const i of Ranges.stepped(0, 10, 5)) {
    print(Int.toString(i))
  }
  return 0
}
```

```text
10
6
2
0
5
```

A step of 0 written in the code is a compile error. A step of 0 that is only known when the program runs gives an empty range. The range stops before the value would pass the largest or smallest `int`, so a range near the end of `int` does not wrap around.

### Stop early or skip an element

You want the first price above 300, and you want to skip prices of 0, which are free samples. `break` leaves a loop at once, and `continue` skips to the next pass:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const prices = [250, 0, 320, 180, 400]
  let paidItems = 0
  for (const price of prices) {
    if (price == 0) {
      continue
    }
    paidItems = paidItems + 1
    if (price > 300) {
      print(`first price above 300: ${Int.toString(price)}`)
      break
    }
  }
  print(`paid items looked at: ${Int.toString(paidItems)}`)
  return 0
}
```

```text
first price above 300: 320
paid items looked at: 2
```

Both work in `while` loops too.

### Count with a `for` header

When a loop counts with a name of its own, use the C-style `for`. Its header holds the start, the condition and the step, separated by `;`. The name is a `let`, and the step is an assignment or a call, such as `day++`:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  for (let day = 1; day <= 3; day++) {
    print(Int.toString(day))
  }
  return 0
}
```

```text
1
2
3
```

`continue` runs the step before it tests the condition again. `for (;;)` loops until a `break`. The `;` belongs to this header and is an error anywhere else.

### Leave an outer loop

`break` and `continue` act on the innermost loop. To act on an outer one, put a label before it, as in `outer: for ...`, and name it: `break outer` or `continue outer`. Here the search through rows skips the rest of the first row and stops at the first 4:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const rows = [[1, 2], [3, 4], [5, 6]]
  outer: for (const row of rows) {
    for (const cell of row) {
      if (cell == 1) {
        continue outer
      }
      if (cell == 4) {
        break outer
      }
      print(Int.toString(cell))
    }
  }
  return 0
}
```

```text
3
```

### Arrays are values

What happens to a copy when you change the original? Keep a snapshot of the prices, then add one:

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

export function run(args: Array<string>): int {
  const prices: Array<int> = [250, 320]
  const snapshot = copy(prices)
  Arrays.push(mutable prices, 400)
  print(Int.toString(Arrays.length(prices)))
  print(Int.toString(Arrays.length(snapshot)))
  return 0
}
```

```text
3
2
```

The snapshot keeps two prices. `copy(prices)` gives `snapshot` a value of its own, so a change through one name never shows up through the other. The copy is cheap: the two names share the elements until one of them changes, and only then are the elements copied. `const snapshot = prices` without `copy` would make two names for one array, so the compiler rejects it and asks for `copy(prices)` ([chapter 9](09-memory-and-mutable.md#a-second-name-for-a-value)).

`prices` is a `const`, and it still grows. A `const` fixes the name: `prices` names the same array for the whole function, and its elements can change, through `prices[0] = 200` or `Arrays.push(mutable prices, 400)`. A `let` adds one thing, giving the name a new value with `prices = [...]`:

```narra error
// error: 4:3: "prices" is constant; declare it with `let` to change it
export function run(args: Array<string>): int {
  const prices = [250, 320]
  prices = [200]
  return 0
}
```

### More array functions

Two more functions from `std/array` return a new array and leave the old one alone. `Arrays.map(items, change)` applies a function to each element, and `Arrays.slice(items, start, end)` takes the elements from `start` up to, but not including, `end`. Like `Str.slice`, it returns `None` when a position is out of range:

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

export function run(args: Array<string>): int {
  const prices = [250, 320, 180, 400]
  const withTax = Arrays.map(prices, (price) => price + price / 10)
  for (const price of withTax) {
    print(Int.toString(price))
  }
  if (Arrays.slice(prices, 1, 3) is Some(middle)) {
    print(Int.toString(Arrays.length(middle)))
  }
  return 0
}
```

```text
275
352
198
440
2
```

`Arrays.withCapacity(count)` makes an empty array with room for `count` elements, which saves growing it when you know the size in advance. `Arrays.values(items)` loops over the elements, as `for (const price of prices)` already does.

## Look up a value by name

The shop keeps a count of each item in stock. A `Map` looks up a value by a key. `Map<string, int>` maps item names to counts:

```narra
import * as Int from "std/int"
import * as Maps from "std/map"

export function run(args: Array<string>): int {
  let stock: Map<string, int> = Map { "tea": 12, "coffee": 5 }
  stock["coffee"] = 4
  const newItem = "cocoa"
  stock[newItem] = 9
  for (const name of Maps.keys(stock)) {
    if (stock[name] is Some(count)) {
      print(`${name}: ${Int.toString(count)}`)
    }
  }
  print(`${Int.toString(Maps.size(stock))} items`)
  return 0
}
```

```text
tea: 12
coffee: 4
cocoa: 9
3 items
```

`Map { "tea": 12, "coffee": 5 }` writes a map with two entries. Assigning `stock["coffee"]` changes an entry, or adds it when the key is new. `Maps.keys` lists the keys in the order they were added.

Reading `stock[name]` is different from reading an array. The key may not be in the map, so the result is not an `int` but an `Option<int>`: either `Some(count)`, holding the count, or `None`. Use it as an `int` and the compiler says so:

```narra error
// error: 4:27: expected int, found Option<int>
export function run(args: Array<string>): int {
  const stock: Map<string, int> = Map { "tea": 12 }
  const count: int = stock["tea"]
  return count
}
```

`if (stock[name] is Some(count))` checks which case you have. When it is `Some`, the block runs with `count` bound to the value inside. Narra has no `null`: a value that may be missing has a type that says so, and you check it before you use it. Chapter 5 shows every way to do that.

Removing an entry and starting from an empty map look like this:

```narra
import * as Int from "std/int"
import * as Maps from "std/map"

export function run(args: Array<string>): int {
  let stock = Map<string, int> {}
  stock["tea"] = 12
  const removed = Maps.remove(mutable stock, "tea")
  if (removed && stock["tea"] is None) {
    print("tea is gone")
  }
  print(Int.toString(Maps.size(stock)))
  return 0
}
```

```text
tea is gone
0
```

`Maps.remove` returns `true` when the key was there and `false` when it was not. An empty map has no entries to tell the compiler its types, so you write them: `Map<string, int> {}`.

### Count with a map

A common job is counting: how often does each item appear in a list of sales? You cannot write `counts[item] += 1`, because a key that is not there yet has no old value to add to:

```narra error
// error: 5:3: only variables, fields and array elements take a compound assignment
export function run(args: Array<string>): int {
  let counts = Map<string, int> {}
  const item = "tea"
  counts[item] += 1
  return 0
}
```

`Maps.update(mutable counts, key, initial, change)` sets the entry to `change(old)`, where `old` is the current value, or to `change(initial)` when the key is missing. With `initial` 0 and a `change` that adds 1, it counts:

```narra
import * as Int from "std/int"
import * as Maps from "std/map"

export function run(args: Array<string>): int {
  const sales = ["tea", "coffee", "tea", "tea"]
  let counts = Map<string, int> {}
  for (const item of sales) {
    Maps.update(mutable counts, item, 0, (count) => count + 1)
  }
  for (const #[item, count] of Maps.entries(counts)) {
    print(`${item}: ${Int.toString(count)}`)
  }
  return 0
}
```

```text
tea: 3
coffee: 1
```

`Maps.entries` gives each key with its value as a pair, in the order the keys were added.

## Group a few values

Sometimes two values belong together, such as an item name and its price. A tuple groups a fixed number of values, which can have different types. It is written `#[...]`, both as a value and as a type:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const cheapest: #[string, int] = #["tea", 250]
  const #[name, priceCents] = cheapest
  print(`${name} at ${Int.toString(priceCents)}`)
  print(`${cheapest[0]} at ${Int.toString(cheapest[1])}`)
  const items: Array<#[string, int]> = [#["tea", 250], #["coffee", 320]]
  for (const #[itemName, itemPrice] of items) {
    print(`${itemName}=${Int.toString(itemPrice)}`)
  }
  return 0
}
```

```text
tea at 250
tea at 250
tea=250
coffee=320
```

`const #[name, priceCents] = cheapest` takes the tuple apart into two names. `cheapest[0]` reads one position, here `"tea"`, and `cheapest[1]` reads `250`. The index must be a number written in the code, so the compiler knows the type of the result and the read cannot fail. A tuple has no way to change one position, so `cheapest[0] = "milk"` is an error:

```narra error
// error: 4:3: only variables, fields and array elements can be assigned
export function run(args: Array<string>): int {
  let pair = #["tea", 250]
  pair[0] = "milk"
  return 0
}
```

The next chapter moves the receipt code into functions, so that each piece has a name and can be used again.

Next: [3. Functions and closures](03-functions-and-closures.md).
