# 3. Functions and closures

The receipt code of chapter 2 lived in `run`. This chapter moves it into functions of its own. Then it passes functions around as values, to apply a pricing rule to every item and to build new rules out of old ones.

## Give a calculation a name

Every item line multiplies a price by a quantity and prints the result. Write that once as a function, and call it for each item:

```narra
import * as Int from "std/int"

function lineTotal(priceCents: int, quantity: int): int {
  return priceCents * quantity
}

function printLine(name: string, cents: int) {
  print(`${name}: ${Int.toString(cents)}`)
}

export function run(args: Array<string>): int {
  const teaCents = lineTotal(250, 2)
  const coffeeCents = lineTotal(320, 1)
  printLine("tea", teaCents)
  printLine("coffee", coffeeCents)
  printLine("total", teaCents + coffeeCents)
  return 0
}
```

```text
tea: 500
coffee: 320
total: 820
```

A function declaration names each parameter with its type, then gives the type of the result after the parentheses. `return` ends the function and hands back the value. `printLine` has no result type, so it returns nothing and is called as a statement.

A named function writes the type of every parameter, and its result type when it returns a value. The compiler does not guess them, because they are what a caller reads to learn how to use the function. Without `export`, a function is visible only in its own file.

A call passes exactly the arguments the function declares:

```narra error
// error: 7:19: this call passes 1 arguments, the function takes 2
function lineTotal(priceCents: int, quantity: int): int {
  return priceCents * quantity
}

export function run(args: Array<string>): int {
  return lineTotal(250)
}
```

## A function that calls itself

The shop gives change in coins of 50, 25, 10 and 5 cents. In how many different ways can it pay out 100 cents?

Counting them with loops is awkward, but the question splits into two smaller ones. Take the largest coin. Either the change uses at least one of it, or it uses none of it:

- With at least one 50, what is left to pay is 50 cents, still with all four coins.
- With no 50s, the full 100 cents must come from the other three coins.

The answer is the sum of the two, and each smaller question splits the same way. A function that answers it can call itself for the two parts:

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

function countWays(amount: int, coins: Array<int>, first: int): int {
  if (amount == 0) {
    return 1
  }
  if (amount < 0 || first == Arrays.length(coins)) {
    return 0
  }
  const usingIt = countWays(amount - coins[first], coins, first)
  const skippingIt = countWays(amount, coins, first + 1)
  return usingIt + skippingIt
}

export function run(args: Array<string>): int {
  const coins = [50, 25, 10, 5]
  print(Int.toString(countWays(100, coins, 0)))
  print(Int.toString(countWays(10, coins, 0)))
  return 0
}
```

```text
40
2
```

This is recursion: `countWays` calls `countWays`. `first` is the position of the largest coin still allowed, so "skip this coin" is `first + 1`. Each call works on a smaller problem: a smaller amount or fewer coins. The first two `if` statements stop at the cases with a known answer. Exactly 0 cents left is one way to pay. Less than 0, or no coins left, is no way. Every recursive function needs such stopping cases, or it calls itself forever.

`countWays(10, coins, 0)` is small enough to check by hand. The two ways are one 10-cent coin, or two 5-cent coins.

Each call that has not returned yet takes some memory, called stack space, until it returns. Here the calls go about 20 levels deep (100 cents in 5-cent steps), which is nothing. A recursion millions of levels deep runs out of stack, and the program stops.

There is one case where depth costs nothing: a call to itself that is the last thing a function does, as in `return addUpTo(n - 1, total + n)` below. The function has no work left after that call, so the compiler turns it into a jump back to the start of the function. The function then runs in the space of a single call, like a `while` loop:

```narra
import * as Int from "std/int"

function addUpTo(n: int, total: int): int {
  if (n == 0) {
    return total
  }
  return addUpTo(n - 1, total + n)
}

export function run(args: Array<string>): int {
  print(Int.toString(addUpTo(10_000_000, 0)))
  return 0
}
```

```text
50000005000000
```

The running result travels in the parameter `total`, so nothing waits for the call to come back. `countWays` is different: it adds `usingIt` and `skippingIt` after its calls return, so each call keeps its stack space.

## Pass a function as a value

Back to pricing. You want to apply a rule to every price in a list: add tax, take off a discount. The loop is the same for every rule, and only the rule changes, so make the rule a parameter:

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

function applyToAll(prices: Array<int>, rule: (cents: int) => int): Array<int> {
  let result: Array<int> = []
  for (const price of prices) {
    Arrays.push(mutable result, rule(price))
  }
  return result
}

function withTax(cents: int): int {
  return cents + cents / 10
}

export function run(args: Array<string>): int {
  const prices = [250, 320, 180]
  const taxed = applyToAll(prices, withTax)
  print(Int.toString(taxed[0]))
  const discounted = applyToAll(prices, (cents) => cents - 50)
  print(Int.toString(discounted[0]))
  return 0
}
```

```text
275
200
```

`rule: (cents: int) => int` declares a parameter whose type is a function type: a function that takes an `int` and returns an `int`. Inside `applyToAll`, `rule(price)` calls it.

`withTax` is a named function, and passing its name passes the function itself. `(cents) => cents - 50` is an arrow function: a function without a name, written right where it is needed. Before `=>` come its parameters, after it the value it returns. Because it is passed to a parameter of a known function type, the compiler knows that `cents` is an `int`.

When an arrow function has more to do than one expression, give it a block body, which uses `return` like a named function:

```narra fragment
applyToAll(prices, (cents) => {
  if (cents > 300) {
    return cents - 50
  }
  return cents
})
```

An arrow function stored under a name has nothing to take its types from, so you write them. Leave them out, and the compiler asks for them:

```narra error
// error: 3:21: cannot tell the type of parameter "cents"; write it as `cents: Type`
export function run(args: Array<string>): int {
  const discount = (cents) => cents - 50
  return discount(100)
}
```

Write the type of each parameter in the parentheses: `const discount = (cents: int) => cents - 50` compiles. The compiler works out the result type from the body. You can still write it after the parentheses, as in `(cents: int): int => cents - 50`.

### A small helper inside a function

Sometimes a helper is useful only inside one function, and you want it there, not at the top of the file. A named function cannot go inside another one:

```narra error
// error: 3:3: a function inside a function is not supported; declare it at module level as a private function and pass the enclosing values as arguments
export function run(args: Array<string>): int {
  function lineTotal(priceCents: int, quantity: int): int {
    return priceCents * quantity
  }
  return lineTotal(250, 2)
}
```

"Module level" is the top level of the file, next to `run`, and "private" means without `export`. When the helper should stay inside the function, bind an arrow function to a `const` instead. That is the local form of a function:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const lineTotal = (priceCents: int, quantity: int) => priceCents * quantity
  print(Int.toString(lineTotal(250, 2) + lineTotal(320, 1)))
  return 0
}
```

```text
820
```

## One function for many types

`applyToAll` only works on lists of `int`. To turn the prices into labels you would need a second copy that returns `Array<string>`. A generic function avoids that. It names placeholder types in angle brackets, and every call fills them in:

```narra
import * as Int from "std/int"
import * as Str from "std/string"
import * as Arrays from "std/array"

function mapAll<T, U>(items: Array<T>, change: (item: T) => U): Array<U> {
  let result: Array<U> = []
  for (const item of items) {
    Arrays.push(mutable result, change(item))
  }
  return result
}

export function run(args: Array<string>): int {
  const prices = [250, 320]
  const labels = mapAll(prices, (cents) => `${Int.toString(cents)} cents`)
  for (const label of labels) {
    print(label)
  }
  const names = ["tea", "coffee"]
  const nameLengths = mapAll(names, (name) => Str.length(name))
  print(Int.toString(nameLengths[1]))
  return 0
}
```

```text
250 cents
320 cents
6
```

`T` is the type of the items going in and `U` the type coming out. In the first call the compiler sees an `Array<int>` and an arrow function that returns a `string`, so `T` is `int` and `U` is `string`. In the second call it is the other way round. You never write the types at the call. The compiler works them out from the arguments.

## Use values from around a function

A discount amount is often decided at run time, outside the arrow function that applies it. An arrow function can use the names around it:

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

function applyToAll(prices: Array<int>, rule: (cents: int) => int): Array<int> {
  let result: Array<int> = []
  for (const price of prices) {
    Arrays.push(mutable result, rule(price))
  }
  return result
}

export function run(args: Array<string>): int {
  const discountCents = 30
  const discounted = applyToAll([250, 320], (cents) => cents - discountCents)
  print(Int.toString(discounted[1]))
  return 0
}
```

```text
290
```

An arrow function that uses names from around it is called a closure, and it captures those names. A closure captures `const` names and parameters by value, since they cannot change. It can also use a `let` name, as long as it is passed straight to a parameter that is not `escaping`, such as `action` below. The closure then works on the variable itself, so a write inside it is visible after the call:

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

function each(items: Array<int>, action: (item: int) => void) {
  for (const item of Arrays.values(items)) {
    action(item)
  }
}

export function run(args: Array<string>): int {
  let total = 0
  each([1, 2, 3], (item) => {
    total = total + item
  })
  print(Int.toString(total))
  return 0
}
```

```text
6
```

A closure that is stored, returned or passed to an `escaping` parameter cannot capture a `let` name:

```narra error
// error: 8:22: a closure cannot capture "count" because it is declared with let, unless it is passed straight to a function parameter that is not escaping
function apply(escaping rule: () => int): int {
  return rule()
}

export function run(args: Array<string>): int {
  let count = 1
  return apply(() => count)
}
```

A stored or returned closure can be called long after the function that made it has returned, when `count` no longer exists. A closure passed to a parameter that is not `escaping` can only run during the call, so the variable is still there. For a closure that outlives the call, copy the current value into a `const` (`const current = count`) and capture that.

While the closure runs, the variable it assigns belongs to it. The same call cannot also pass that variable as another argument, because the function could be reading it while the closure replaces it:

```narra error
// error: 8:8: "items" is assigned by a closure passed to this call, so no other argument can use it
function each(items: Array<int>, action: () => void) {
  action()
}

export function run(args: Array<string>): int {
  let items = [1, 2]
  each(items, () => {
    items = []
  })
  return 0
}
```

Two closures in one call may both read the same variable. Only a closure that assigns it makes the call exclusive.

## Make new functions from old ones

A function can also return a function. `makeDiscount` returns a rule that takes a fixed amount off:

```narra
import * as Int from "std/int"

function makeDiscount(amountCents: int): (cents: int) => int {
  return (cents: int): int => cents - amountCents
}

export function run(args: Array<string>): int {
  const tenOff = makeDiscount(10)
  print(Int.toString(tenOff(250)))
  return 0
}
```

```text
240
```

The returned closure keeps `amountCents`, so it still works after `makeDiscount` has returned.

Now you want a rule that applies another rule twice, and you want to keep it to use later. The first try does not compile:

```narra error
// error: 3:21: a closure cannot capture parameter "rule" because it is not escaping; declare it as `escaping rule`
function twice(rule: (cents: int) => int): (cents: int) => int {
  return (cents) => rule(rule(cents))
}

export function run(args: Array<string>): int {
  return twice((cents) => cents - 10)(100)
}
```

The closure returned by `twice` uses `rule`, so `rule` has to outlive the call to `twice`. A function parameter can only be used while the call is running, unless you mark it `escaping`, as the message says:

```narra
import * as Int from "std/int"

function twice(escaping rule: (cents: int) => int): (cents: int) => int {
  return (cents) => rule(rule(cents))
}

export function run(args: Array<string>): int {
  const twentyOff = twice((cents) => cents - 10)
  print(Int.toString(twentyOff(100)))
  return 0
}
```

```text
80
```

An `escaping` parameter may be returned, stored or captured. A closure passed to it needs a memory allocation for the values it captures, which a plain function parameter avoids. Add `escaping` only when the function keeps the parameter.

The next chapter gives the shop's data a shape of its own: items with named fields, and values that can be one of several kinds.

Next: [4. Structs and enums](04-structs-and-enums.md).
