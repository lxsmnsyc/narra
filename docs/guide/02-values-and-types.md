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

A `const` keeps its value. If you try to change one, the compiler stops you and says what to write instead:

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

You add to a value by writing the sum out in full. The short form is an error:

```narra error
// error: 4:9: compound assignment is not supported; write `x = x + y`
export function run(args: Array<string>): int {
  let weeks = 0
  weeks += 1
  return weeks
}
```

Use `const` for every name that does not need to change, and `let` only where it does. A reader then knows which values to watch.

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
  print(Int.toString(Float.truncate(average * 10.0)))
  return 0
}
```

```text
142
6
142.85714285714286
1428
```

Narra has two kinds of numbers. An `int` is a whole number, and `/` on two `int` values drops the fraction, so `1000 / 7` is `142`. `%` gives the remainder, `6`. An `f64` is a decimal number, written with a decimal point, such as `2.5` or `10.0`. `Float.fromInt` turns an `int` into an `f64`, and `Float.truncate` turns an `f64` back into an `int` by dropping the fraction.

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

An `int` holds values up to about nine quintillion (9223372036854775807). In a debug build, going past that stops the program with an overflow message. When you want the wrapped result on purpose, as in a hash function, use `+%`, `-%` or `*%`:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const largest = 9223372036854775807
  print(Int.toString(largest +% 1))
  return 0
}
```

```text
-9223372036854775808
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
  const dots = Str.repeat(".", width - Str.length(name) - Str.length(priceText))
  print(`${name}${dots}${priceText}`)
  print(Str.repeat("=", width))
  return 0
}
```

```text
tea..........250
================
```

A `string` is text in double quotes. The functions for strings come from `std/string`, imported here as `Str`: `Str.length` counts the bytes of a string, and `Str.repeat` repeats it.

To join strings, put them in a template string, as the line above does. `+` does not join strings, and the compiler points you to templates:

```narra error
// error: 3:15: strings are joined with a template string, not +
export function run(args: Array<string>): int {
  print("tea" + "...")
  return 0
}
```

Text often arrives as one line you need to take apart. `Str.split` cuts a string at each separator and gives the pieces as an array. Arrays come later in this chapter. For now, `parts[0]` is the first piece and `parts[1]` the second:

```narra
import * as Str from "std/string"

export function run(args: Array<string>): int {
  const line = "  coffee = 320 "
  const parts = Str.split(line, "=")
  const name = Str.trim(parts[0])
  const price = Str.trim(parts[1])
  print(`[${name}] costs [${price}]`)
  print(Str.slice(name, 0, 3))
  return 0
}
```

```text
[coffee] costs [320]
cof
```

`Str.trim` removes spaces at both ends, and `Str.slice(name, 0, 3)` takes the bytes from position 0 up to, but not including, position 3.

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

The functions for arrays come from `std/array`, imported as `Arrays`, and each takes the array as its first argument. `Arrays.push(mutable prices, 400)` adds an element at the end. The word `mutable` in the call marks the argument that the function changes. You can see at the call which of your values may change. Only a `let` name can be passed that way. Chapter 9 explains `mutable` in full.

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
  let squares = Arrays.filled(0, 4)
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

`Arrays.filled(0, 4)` makes an array of four zeros. The `#[position, square]` in the loop takes apart a pair of values. Pairs come at the end of this chapter.

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

### Arrays are values

What happens to a copy when you change the original? Keep a snapshot of the prices, then add one:

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

export function run(args: Array<string>): int {
  let prices: Array<int> = [250, 320]
  const snapshot = prices
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

The snapshot keeps two prices. Assigning an array to another name gives that name its own copy, so a change through one name never shows up through another. The copy is cheap: the two names share the elements until one of them changes, and only then are the elements copied.

A `const` array cannot change at all, not even its elements:

```narra error
// error: 4:3: "prices" is constant; declare it with `let` to change it
export function run(args: Array<string>): int {
  const prices = [250, 320]
  prices[0] = 200
  return 0
}
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

## Group a few values

Sometimes two values belong together, such as an item name and its price. A tuple groups a fixed number of values, which can have different types. It is written `#[...]`, both as a value and as a type:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const cheapest: #[string, int] = #["tea", 250]
  const #[name, priceCents] = cheapest
  print(`${name} at ${Int.toString(priceCents)}`)
  const items: Array<#[string, int]> = [#["tea", 250], #["coffee", 320]]
  for (const #[itemName, itemPrice] of items) {
    print(`${itemName}=${Int.toString(itemPrice)}`)
  }
  return 0
}
```

```text
tea at 250
tea=250
coffee=320
```

`const #[name, priceCents] = cheapest` takes the tuple apart into two names. That is the only way to read a tuple: it has no positions to index.

The next chapter moves the receipt code into functions, so that each piece has a name and can be used again.

Next: [3. Functions and closures](03-functions-and-closures.md).
