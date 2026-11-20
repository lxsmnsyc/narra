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

export function run(args: Array<string>): int {
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

Four more functions tidy text. `Str.trimStart` and `Str.trimEnd` remove white space from one end, `Str.padStart` and `Str.padEnd` grow a string to a width by repeating a `char`, and `Str.lastIndexOf` is `Str.indexOf` from the right. The width counts bytes, like `Str.length`, so `"café"` is five bytes wide and `Str.padStart("café", 7, '0')` gives `00café`. The fill is added as whole characters, so a fill of two bytes never splits, and a string already at the width comes back unchanged:

```narra
import * as Int from "std/int"
import * as Str from "std/string"

export function run(args: Array<string>): int {
  print(`[${Str.trimStart("  42  ")}]`)
  print(`[${Str.trimEnd("  42  ")}]`)
  print(Str.padStart("42", 5, '0'))
  print(Str.padStart("café", 7, '0'))
  print(`${Str.padEnd("name", 8, '.')}|`)
  if (Str.lastIndexOf("a/b/c", "/") is Some(index)) {
    print(Int.toString(index))
  }
  return 0
}
```

```text
[42  ]
[  42]
00042
00café
name....|
3
```

`std/char` also tells what kind of character you have: `Char.isDigit`, `Char.isLetter`, `Char.isAlphanumeric`, `Char.isUpper`, `Char.isLower` and `Char.isWhitespace`. They cover ASCII only: `Char.isLetter('é')` is `false`, and `'é'` is in none of the six. `std/unicode` has the Unicode tables for case, grapheme clusters and width, but no letter classes.

```narra
import * as Char from "std/char"
import * as Str from "std/string"

export function run(args: Array<string>): int {
  const characters = ['7', 'Q', 'q', ' ', 'é']
  for (const character of characters) {
    let kinds = ""
    if (Char.isDigit(character)) {
      kinds = `${kinds} digit`
    }
    if (Char.isLetter(character)) {
      kinds = `${kinds} letter`
    }
    if (Char.isUpper(character)) {
      kinds = `${kinds} upper`
    }
    if (Char.isWhitespace(character)) {
      kinds = `${kinds} whitespace`
    }
    print(`${Str.fromChar(character)}:${kinds}`)
  }
  return 0
}
```

```text
7: digit
Q: letter upper
q: letter
 : whitespace
é:
```

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

`(price) => price + price / 10` is an arrow function: a function without a name, written where it is used. [Chapter 3](03-functions-and-closures.md#pass-a-function-as-a-value) teaches it.

`Arrays.withCapacity(count)` makes an empty array with room for `count` elements, which saves growing it when you know the size in advance. `Arrays.values(items)` loops over the elements, as `for (const price of prices)` already does.

### Sort, search and shrink

These functions from `std/array` change the array they are given, so the call writes `mutable`. `Arrays.sort(mutable items)` puts numbers and strings in increasing order. `Arrays.sortBy(mutable items, compare)` takes a function that compares two elements and returns `Ordering.Less`, `Ordering.Equal` or `Ordering.Greater`; use it for another order or to sort by one field. Both keep elements that compare equal in their old order. `Arrays.reverse(mutable items)` turns the array around. `Arrays.pop(mutable items)` removes the last element and returns it as an `Option`, which is `None` for an empty array. `Arrays.truncate(mutable items, length)` keeps the first `length` elements and does nothing when the array is not longer. `Arrays.contains(items, wanted)` and `Arrays.indexOf(items, wanted)` look for an element, and `indexOf` returns its first position as an `Option`:

```narra
import * as Int from "std/int"
import * as Str from "std/string"
import * as Arrays from "std/array"

export function run(args: Array<string>): int uses io, alloc {
  let prices = [250, 320, 180, 400]
  Arrays.sort(mutable prices)
  print(Str.join(Arrays.map(prices, (price) => Int.toString(price)), " "))
  Arrays.sortBy(mutable prices, (first, second) => Int.compare(second, first))
  print(Str.join(Arrays.map(prices, (price) => Int.toString(price)), " "))
  if (Arrays.pop(mutable prices) is Some(cheapest)) {
    print(Int.toString(cheapest))
  }
  Arrays.truncate(mutable prices, 1)
  print(Int.toString(Arrays.length(prices)))
  print(Show.show(Arrays.contains(prices, 400)))
  return 0
}
```

```text
180 250 320 400
400 320 250 180
180
1
true
```

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

## Numbers of other sizes

An `int` is 64 bits wide, and it is right for most programs. Some data has a fixed width: a byte in a file, a colour channel, a 32-bit field in a network format. For those, Narra has sized integers, and a smaller float.

### Sized integers

The sized integer types are `i8`, `i16`, `i32`, `u8`, `u16`, `u32` and `u64`. The number is the width in bits, and `u` means unsigned, with no negative values. `i64` is another name for `int`. A `u8` holds 0 to 255, and an `Array<u8>` stores one byte per element.

A number written in the code takes the type that its place expects. The place can be an annotation, a parameter, a field, an array element or the other side of an operator. With nothing to go on, it is an `int`. A literal that does not fit the type is an error:

```narra error
// error: 3:20: integer literal 256 does not fit in u8
export function run(args: Array<string>): int {
  const byte: u8 = 256
  return 0
}
```

Arithmetic wraps at the width of the type, as an `int` wraps at 64 bits. A `u8` that adds 200 and 100 gets 44, because 300 is 256 and 44. Each sized type has its own module, named after it: `std/u8`, `std/i32` and so on. It has `toString`, `parse` and the checked arithmetic that `Int` has:

```narra
import * as Int from "std/int"
import * as U8 from "std/u8"

export function run(args: Array<string>): int {
  const level: u8 = 200
  const boost: u8 = 100
  print(U8.toString(level + boost))
  if (U8.addChecked(level, boost) is None) {
    print("does not fit in a u8")
  }
  if (U8.parse("255", 10) is Ok(value)) {
    print(Int.toString(Int.from(value) + 1))
  }
  if (U8.parse("256", 10) is Err(message)) {
    print(message)
  }
  return 0
}
```

```text
44
does not fit in a u8
256
number does not fit in 8 bits (0 to 255)
```

`parse(text, base)` reads a number from text in the given base, such as 10 or 16, and returns a `Result`: `Ok` with the number, or `Err` with a message. The error text names the range of the type. [Chapter 6](06-option-and-result.md) covers `Result`.

#### No mixing, and the conversions

Two sized types never mix, and neither does a sized type with an `int`. Adding an `i32` to an `int` is an error that names the way out:

```narra error
// error: 7:28: + needs two integers of the same type, found i32 and int; convert one side with I32.from, I32.tryFrom or I32.wrap
import * as I32 from "std/i32"

export function run(args: Array<string>): int {
  const total: i32 = 10
  const count = 3
  print(I32.toString(total + count))
  return 0
}
```

The conversion is a function in the module of the target type, and you choose how a value that does not fit is handled:

- `X.from(value)` is for values that always fit. `Int.from(small)` turns an `i32` into an `int`, and `I32.from(byte)` turns a `u8` into an `i32`. The compiler rejects a call where some value of the argument type would not fit.
- `X.tryFrom(value)` returns an `Option`, `None` when the value does not fit.
- `X.wrap(value)` keeps the low bits, so a value that does not fit wraps around.

```narra
import * as Int from "std/int"
import * as I32 from "std/i32"
import * as U8 from "std/u8"

export function run(args: Array<string>): int {
  const small: i32 = -5
  print(Int.toString(Int.from(small) * 1_000_000_000_000))
  const big = 3_000_000_000
  if (I32.tryFrom(big) is None) {
    print("3000000000 does not fit in an i32")
  }
  if (I32.tryFrom(7) is Some(seven)) {
    print(I32.toString(seven))
  }
  print(I32.toString(I32.wrap(big)))
  print(U8.toString(U8.wrap(300)))
  return 0
}
```

```text
-5000000000000
3000000000 does not fit in an i32
7
-1294967296
44
```

`from` takes a number that is already of a sized or `int` type. A literal such as `5` is an `int`, which does not always fit an `i32`, so `I32.from(5)` is an error. Write `const five: i32 = 5` to get a sized literal.

### The f32 float

`f32` is a 32-bit float beside the 64-bit `f64`. It takes half the memory and has fewer digits. Use it when a format or a library asks for it. A decimal literal is an `f32` when its place expects one, and an `f64` otherwise. The two do not mix. `std/f32` converts. `F32.from(value)` takes an `f64` (rounded to the nearest) or any integer. `Float.from(value)` takes an `f32` or any integer. `F32.truncate(value)` is the `f32` twin of `Float.truncate`.

```narra
import * as Float from "std/float"
import * as F32 from "std/f32"

export function run(args: Array<string>): int {
  const ratio: f32 = 0.1
  print(F32.toString(ratio * 3.0))
  print(Float.toString(Float.from(ratio)))
  print(Float.toString(0.1 + 0.2))
  return 0
}
```

```text
0.3
0.10000000149011612
0.30000000000000004
```

The `f32` value `0.1` is not exactly one tenth, and widening it to an `f64` shows the digits it holds. The `f64` sum `0.1 + 0.2` shows the same effect: decimal fractions are rounded in binary, in every language. Compare floats with a margin, not with `==`.

### How a float prints

`Float.toString`, `F32.toString` and a template hole `${...}` print the shortest digits that read back as the same number. A whole number keeps its `.0`, so a float never looks like an `int`. Values below `1e-7` or from `1e21` up use an exponent. NaN and the infinities have their own text:

```narra
import * as Float from "std/float"

export function run(args: Array<string>): int {
  print(Float.toString(20.0))
  print(Float.toString(1.5e-7))
  print(Float.toString(1e21))
  print(Float.toString(Float.sqrt(-1.0)))
  print(Float.toString(1.0 / 0.0))
  print(Float.toString(-0.0))
  return 0
}
```

```text
20.0
1.5e-7
1e+21
nan
inf
-0.0
```

Values from `1e-7` up to below `1e21` print in plain notation. Float division by zero needs no check, as [the division section](#dividing-by-a-name) says: it gives `inf`, `-inf` or `nan`.

### Fixed decimals and reading a float

`Float.toFixed(value, digits)` prints a float with exactly that many digits after the point, from 0 to 20. It rounds the way C's `%.*f` does, on the exact binary value, so `1.005` with two digits gives `1.00` (the stored number is just below `1.005`) and a tie such as `2.5` with no digits goes to the even neighbour, `2`. It returns `None` for a digit count outside 0 to 20. `Float.parse` is the other direction. Like `Int.parse` it returns a `Result`: digits with an optional sign, point and exponent, and nothing else (no spaces, `inf`, `nan` or hex), or an `Err` that says what is wrong. A number too small to represent reads as `0.0`, while a number too large is an `Err`.

```narra
import * as Float from "std/float"

export function run(args: Array<string>): int {
  const average = Float.fromInt(1000) / Float.fromInt(7)
  if (Float.toFixed(average, 2) is Some(text)) {
    print(text)
  }
  if (Float.toFixed(1.005, 2) is Some(text)) {
    print(text)
  }
  match (Float.parse("2.5e2")) {
    Ok(value) => print(Float.toString(value))
    Err(message) => print(message)
  }
  match (Float.parse("2.5x")) {
    Ok(value) => print(Float.toString(value))
    Err(message) => print(message)
  }
  return 0
}
```

```text
142.86
1.00
250.0
invalid character in the number
```

### Bits

The operators `&` (and), `|` (or), `^` (xor), `~` (not), `<<` and `>>` work on the bits of an integer of any width. Write the number in binary with `0b` or in hexadecimal with `0x` to see them:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const flags = 0b1010
  print(Int.toString(flags & 0b0110))
  print(Int.toString(flags | 0b0001))
  print(Int.toString(flags ^ 0b1111))
  print(Int.toString(~flags))
  print(Int.toString(1 << 4))
  print(Int.toString(-16 >> 2))
  return 0
}
```

```text
2
11
5
-11
16
-4
```

`<<` moves the bits left, which doubles the value for each step. `>>` moves them right. On a signed type it keeps the sign, so `-16 >> 2` is `-4`; on an unsigned type it fills with zeros. `~` flips every bit, so `~x` is `-x - 1` for a signed type. Both sides of an operator have the same type, as with `+`.

Like a divisor, a shift count must be proven. The count must be a constant from 0 to one less than the width (63 for an `int`, 7 for a `u8`). It can also be a name checked against both ends first. An unsigned count needs only the upper end:

```narra error
// error: 3:19: the shift count must be a constant from 0 to 63 or a name checked against both ends first; check the name first, as in `if (n >= 0 && n < 64) { ... }`
function shiftLeft(value: int, count: int): int {
  return value << count
}

export function run(args: Array<string>): int {
  return shiftLeft(1, 3)
}
```

With the check, the function compiles:

```narra
import * as Int from "std/int"

function shiftLeft(value: int, count: int): int {
  if (count >= 0 && count < 64) {
    return value << count
  }
  return 0
}

export function run(args: Array<string>): int {
  print(Int.toString(shiftLeft(1, 3)))
  return 0
}
```

```text
8
```

### The math module

`std/math`, imported as `Math`, has the usual number functions. The `int` ones are `Math.abs`, and `Math.min`, `Math.max` and `Math.clamp`, which take any two values of the same ordered type. `Math.clamp(value, low, high)` returns an `Option`, because `low` above `high` has no answer. The `f64` ones are `Math.absFloat`, `sqrt`, `floor`, `ceil`, `round`, `trunc`, `sin`, `cos`, `tan`, `atan2`, `exp` and `log`, plus the constants `Math.PI` and `Math.E`. `Math.pow` and `Math.powInt` are in [the powers section](#powers).

```narra
import * as Int from "std/int"
import * as Float from "std/float"
import * as Math from "std/math"

export function run(args: Array<string>): int {
  print(Int.toString(Math.abs(-4)))
  print(Int.toString(Math.min(3, 9)))
  if (Math.clamp(15, 0, 10) is Some(clamped)) {
    print(Int.toString(clamped))
  }
  print(Float.toString(Math.floor(2.7)))
  print(Float.toString(Math.round(2.5)))
  print(Float.toString(Math.sqrt(2.0)))
  print(Float.toString(Math.PI))
  return 0
}
```

```text
4
3
10
2.0
3.0
1.4142135623730951
3.141592653589793
```

### Random numbers

`std/random`, imported as `Random`, makes numbers from a seed. A random number generator changes as you draw from it, so pass it as `mutable` ([chapter 9](09-memory-and-mutable.md) explains the word). `Random.seeded(seed)` starts one, and the same seed gives the same numbers on every run and every platform, which makes a test repeatable. `Random.fromSystem()` seeds from the system and needs the `io` effect ([chapter 10](10-effects-and-testing.md)).

- `Random.nextInt(mutable random, low, high)` draws from `low` up to below `high`. It returns `None` when the range is empty.
- `Random.nextFloat(mutable random)` draws an `f64` from 0 up to below 1.
- `Random.nextBits(mutable random)` draws a `u64` of random bits.
- `Random.shuffle(mutable random, mutable items)` reorders an array.

```narra
import * as Random from "std/random"
import * as U64 from "std/u64"
import * as Int from "std/int"

export function run(args: Array<string>): int {
  let first = Random.seeded(7)
  let second = Random.seeded(7)
  print(U64.toString(Random.nextBits(mutable first)))
  print(U64.toString(Random.nextBits(mutable second)))
  for (const _round of [1, 2, 3]) {
    if (Random.nextInt(mutable first, 1, 7) is Some(roll)) {
      print(Int.toString(roll))
    }
  }
  if (Random.nextInt(mutable first, 5, 5) is None) {
    print("empty range")
  }
  let deck = [1, 2, 3, 4, 5]
  Random.shuffle(mutable first, mutable deck)
  for (const card of deck) {
    print(Int.toString(card))
  }
  return 0
}
```

```text
7191089600892374487
7191089600892374487
1
1
4
empty range
3
1
4
2
5
```

The generator is fast and small, and it is not for secrets: do not use it for keys or tokens.

### Hashing

`std/hash`, imported as `Hash`, turns a value into a number. `Hash.fnv1a(text)` is the FNV-1a hash of a string, as a `u64`. It is the same in every build and on every machine, so it is safe to store. `Hash.of(value)` hashes any value that can be a map key, and its result may change between builds, so use it only inside one run:

```narra
import * as U64 from "std/u64"
import * as Hash from "std/hash"

export function run(args: Array<string>): int {
  print(U64.toString(Hash.fnv1a("tea")))
  print(U64.toString(Hash.fnv1a("")))
  print(Hash.of("tea") == Hash.of("tea") ? "same" : "different")
  return 0
}
```

```text
6261350492309512777
14695981039346656037
same
```

Neither one is meant for security.

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
