# 5. Pattern matching

In this chapter the shop gets a till that reads commands from lines of text such as `add tea 2` or `pay 500`. Turning a line of text into a command, and acting on it, means asking "what shape is this value?" over and over. Patterns are how Narra asks that question. You met them in `match` in chapter 4. Here you see what else they can test, and the other places they appear.

## Decide by kind, and by condition

The till knows three commands:

```narra
import * as Int from "std/int"

enum Command {
  Add(name: string, quantity: int)
  Remove(name: string)
  Pay(cents: int)
}

function describe(command: Command): string {
  return match (command) {
    Command.Add(name, quantity) if (quantity >= 10) => `bulk order of ${name}`
    Command.Add(name, quantity) => `add ${Int.toString(quantity)} ${name}`
    Command.Remove(name) => `remove ${name}`
    Command.Pay(cents) => `pay ${Int.toString(cents)}`
  }
}

export function run(args: Array<string>): int {
  const commands = [Command.Add("tea", 2), Command.Add("cups", 12), Command.Remove("tea"), Command.Pay(500)]
  for (const command of commands) {
    print(describe(command))
  }
  return 0
}
```

```text
add 2 tea
bulk order of cups
remove tea
pay 500
```

An arm can add a condition after its pattern, `if (quantity >= 10)`, called a guard. The arm is taken only when the pattern matches and the guard is true. `match` tries the arms from top to bottom and takes the first that fits. `Add("cups", 12)` goes to the bulk arm, and `Add("tea", 2)` falls past it to the next one.

Several patterns can share one arm. Join them with `|`, and the arm is taken when any of them matches:

```narra fragment
function isChange(command: Command): bool {
  return match (command) {
    Command.Add(_, _) | Command.Remove(_) => true
    Command.Pay(_) => false
  }
}
```

Every pattern in the list must bind the same names with the same types, because the arm body uses them whichever pattern matched. A guard applies to the whole list. The list can also sit inside a pattern, as in `Some(1 | 2)`. An `is` test takes one pattern only; use `match` when you need alternatives.

Here `match` is a value: the function returns whatever the chosen arm gives, so every arm gives a `string`. A `match` can also stand alone as a statement, as `checkWeight` did in chapter 4. Then each arm can run a block of statements, and a `return` in it returns from the function:

```narra
enum Command {
  Add(name: string, quantity: int)
  Remove(name: string)
  Pay(cents: int)
}

function isAccepted(command: Command): bool {
  match (command) {
    Command.Pay(cents) => {
      if (cents <= 0) {
        print("a payment must be positive")
        return false
      }
    }
    _ => print("checked")
  }
  return true
}

export function run(args: Array<string>): int {
  assert(isAccepted(Command.Remove("tea")))
  assert(!isAccepted(Command.Pay(0)))
  return 0
}
```

```text
checked
a payment must be positive
```

`_` matches anything, so the last arm covers `Add` and `Remove` together.

## Match on numbers, text and tuples

`match` works on any value, not only enums. A literal in a pattern, such as `0` or `"tea"`, matches only that value. A tuple pattern matches several values at once:

```narra
import * as Int from "std/int"

function quantityWord(quantity: int): string {
  return match (quantity) {
    0 => "no"
    1 => "one"
    2 => "two"
    _ => "many"
  }
}

function discountPercent(isMember: bool, totalCents: int): int {
  return match (#[isMember, totalCents >= 1000]) {
    #[true, true] => 15
    #[true, false] => 5
    #[false, true] => 10
    #[false, false] => 0
  }
}

export function run(args: Array<string>): int {
  print(`${quantityWord(0)} cups, ${quantityWord(2)} cups, ${quantityWord(40)} cups`)
  print(Int.toString(discountPercent(true, 1200)))
  print(Int.toString(discountPercent(false, 400)))
  return 0
}
```

```text
no cups, two cups, many cups
15
0
```

`discountPercent` puts two conditions into a tuple and lists every combination as a table. The compiler checks that the arms cover every case, through tuples, `bool` values, enums and structs. Remove the `#[false, false]` arm and the program no longer compiles.

An `int` has too many values to list, so a `match` on one needs a `_` arm or a name that catches the rest:

```narra error
// error: 3:10: match does not cover every int value; add a `_` arm
function quantityWord(quantity: int): string {
  return match (quantity) {
    0 => "no"
    1 => "one"
  }
}

export function run(args: Array<string>): int {
  print(quantityWord(1))
  return 0
}
```

A struct pattern names the fields to test. A literal tests a field, and a name binds it:

```narra
import * as Int from "std/int"

struct Item {
  name: string
  priceCents: int
  quantity: int = 1
}

function describe(item: Item): string {
  return match (item) {
    { quantity: 0 } => "out of stock"
    { name, priceCents: 0 } => `${name} is free`
    { name, quantity } => `${Int.toString(quantity)} ${name}`
  }
}

export function run(args: Array<string>): int {
  print(describe(Item { name: "tea", priceCents: 250, quantity: 0 }))
  print(describe(Item { name: "napkin", priceCents: 0 }))
  print(describe(Item { name: "cake", priceCents: 900, quantity: 3 }))
  return 0
}
```

```text
out of stock
napkin is free
3 cake
```

Fields you do not name are not tested. `{ name }` is short for `{ name: name }`: it binds the field `name` to a name `name`. Patterns nest as deep as the value does, so `Some(Command.Add(name, 1))` matches an `Option` that holds an `Add` of exactly one item.

## Take a value apart

When a pattern always matches, you can use it in a `const` or `let` to take a struct or a tuple apart, without a `match`:

```narra
import * as Int from "std/int"

struct Item {
  name: string
  priceCents: int
  quantity: int = 1
}

function splitTotal(totalCents: int, people: int): #[int, int] {
  if (people == 0) {
    return #[0, 0]
  }
  return #[totalCents / people, totalCents % people]
}

export function run(args: Array<string>): int {
  const item = Item { name: "cake", priceCents: 900, quantity: 3 }
  const { name, priceCents: price } = item
  print(`${name} costs ${Int.toString(price)}`)

  const #[share, leftOver] = splitTotal(1000, 3)
  print(`${Int.toString(share)} each, ${Int.toString(leftOver)} left over`)
  return 0
}
```

```text
cake costs 900
333 each, 1 left over
```

`{ name, priceCents: price }` binds the field `name` to `name`, and the field `priceCents` to the name `price`. A function that has more than one result returns a tuple, and the caller takes it apart, as with `splitTotal`.

A pattern that could fail to match does not work here, because there would be nothing to bind when it fails:

```narra error
// error: 9:9: this pattern does not match every Item; use match or `is` to test a value
struct Item {
  name: string
  quantity: int
}

export function run(args: Array<string>): int {
  const item = Item { name: "tea", quantity: 1 }
  const { name, quantity: 1 } = item
  print(name)
  return 0
}
```

## Read a value that may be missing

Now the till turns a typed line into a command. Not every line is a valid command, so the function returns an `Option<Command>`: `Some(command)` when the line makes sense, `None` when it does not.

```narra
import * as Int from "std/int"
import * as Str from "std/string"
import * as Arrays from "std/array"

enum Command {
  Add(name: string, quantity: int)
  Remove(name: string)
  Pay(cents: int)
}

function parseNumber(text: string): Option<int> {
  return match (Int.parse(text, 10)) {
    Ok(value) => Some(value)
    Err(_) => None
  }
}

function parseCommand(line: string): Option<Command> {
  const pieces = Str.split(line, " ")
  if (pieces is None) {
    return None
  }
  const words = pieces.value
  match (#[words[0], Arrays.length(words)]) {
    #["add", 3] => {
      if (parseNumber(words[2]) is Some(quantity)) {
        return Some(Command.Add(words[1], quantity))
      }
    }
    #["remove", 2] => {
      return Some(Command.Remove(words[1]))
    }
    #["pay", 2] => {
      if (parseNumber(words[1]) is Some(cents)) {
        return Some(Command.Pay(cents))
      }
    }
    _ => print(`unknown command: ${line}`)
  }
  return None
}

export function run(args: Array<string>): int {
  const lines = ["add tea 2", "add tea lots", "dance", "pay 500"]
  for (const line of lines) {
    if (parseCommand(line) is Some(_)) {
      print(`ok: ${line}`)
    } else {
      print(`skipped: ${line}`)
    }
  }
  return 0
}
```

```text
ok: add tea 2
skipped: add tea lots
unknown command: dance
skipped: dance
ok: pay 500
```

`Int.parse` reads a number from text. It can fail, so it returns a `Result`: `Ok(value)` with the number, or `Err(message)` with the reason. `parseNumber` turns that into an `Option`, keeping only whether a number was there.

`value is Pattern` is a `bool` condition that is true when the value matches. When the pattern binds names, as `Some(quantity)` does, those names exist inside the `if` block, where the match is known to hold. That is how you get at the value inside an `Option`: you cannot use it without first checking that it is there.

The other way round also works. Check for `None` and leave early, and the value is known to be there for the rest of the function. Read it with `.value`:

```narra
import * as Int from "std/int"
import * as Maps from "std/map"

function priceWithTax(prices: Map<string, int>, name: string): int {
  const found = prices[name]
  if (found is None) {
    return 0
  }
  return found.value + found.value / 10
}

export function run(args: Array<string>): int {
  const prices: Map<string, int> = Map { "tea": 250 }
  print(Int.toString(priceWithTax(prices, "tea")))
  print(Int.toString(priceWithTax(prices, "cake")))
  print(Int.toString(Maps.size(prices)))
  return 0
}
```

```text
275
0
1
```

After the early `return`, the compiler knows `found` is a `Some`, so `found.value` is allowed. This works on `const` names and parameters, and on a `let` name as long as nothing assigns to it in between. Without the check, `.value` is an error:

```narra error
// error: 4:16: this Option<int> may hold any of its variants; check it with `is` or `match` before reading "value"
export function run(args: Array<string>): int {
  const found: Option<int> = Some(1)
  return found.value
}
```

An `is` that binds names must be the whole condition of its `if`. On the other side of an `&&`, the names would not exist. The message also mentions the conditional expression, `? :`, which comes at the end of this chapter:

```narra error
// error: 4:13: an `is` pattern can only bind names when it is the whole condition of an if or a conditional expression
export function run(args: Array<string>): int {
  const found: Option<int> = Some(12)
  if (found is Some(quantity) && quantity >= 10) {
    print("bulk")
  }
  return 0
}
```

Nest the second test inside instead, or put it in a pattern guard with `match`:

```narra
export function run(args: Array<string>): int {
  const found: Option<int> = Some(12)
  if (found is Some(quantity)) {
    if (quantity >= 10) {
      print("bulk")
    }
  }
  match (found) {
    Some(quantity) if (quantity >= 10) => print("bulk again")
    _ => print("small")
  }
  return 0
}
```

```text
bulk
bulk again
```

## Compute a value in several steps

Sometimes a value needs a few statements to work out: a lookup, a check, a fallback. A block in a place where a value is expected computes one. Its `return` gives the value of the block, not of the function:

```narra
import * as Int from "std/int"

function receiptLine(name: string, priceCents: int, quantity: int): string {
  const amount = {
    const total = priceCents * quantity
    if (total >= 1000) {
      return `${Int.toString(total - 100)} (100 off)`
    }
    return Int.toString(total)
  }
  return `${name}: ${amount}`
}

export function run(args: Array<string>): int {
  print(receiptLine("tea", 250, 2))
  print(receiptLine("cake", 900, 3))
  return 0
}
```

```text
tea: 500
cake: 2600 (100 off)
```

`total` exists only inside the block, so it cannot be confused with anything later in the function. Every path through the block must end in `return` with a value.

For a choice between two values, there are two shorter forms. An `if` with an `else` can be a value, with a `return` in each branch. And `condition ? first : second` picks `first` when the condition is true and `second` otherwise:

```narra
import * as Int from "std/int"

function plural(word: string, count: int): string {
  return count == 1 ? word : `${word}s`
}

export function run(args: Array<string>): int {
  const totalCents = 1250
  const label = if (totalCents >= 1000) { return "large order" } else { return "order" }
  print(label)
  print(`${Int.toString(1)} ${plural("cup", 1)}, ${Int.toString(3)} ${plural("cup", 3)}`)
  return 0
}
```

```text
large order
1 cup, 3 cups
```

Both branches give a value of the same type. A value `if` needs its `else`, since without one there would be no value when the condition is false.

## The till

Here is the whole till, putting the chapter together. It reads the lines, keeps a map of what is in the basket, and handles payment:

```narra
import * as Int from "std/int"
import * as Str from "std/string"
import * as Arrays from "std/array"
import * as Maps from "std/map"

enum Command {
  Add(name: string, quantity: int)
  Remove(name: string)
  Pay(cents: int)
}

function parseNumber(text: string): Option<int> {
  return match (Int.parse(text, 10)) {
    Ok(value) => Some(value)
    Err(_) => None
  }
}

function parseCommand(line: string): Option<Command> {
  const pieces = Str.split(line, " ")
  if (pieces is None) {
    return None
  }
  const words = pieces.value
  match (#[words[0], Arrays.length(words)]) {
    #["add", 3] => {
      if (parseNumber(words[2]) is Some(quantity)) {
        return Some(Command.Add(words[1], quantity))
      }
    }
    #["remove", 2] => {
      return Some(Command.Remove(words[1]))
    }
    #["pay", 2] => {
      if (parseNumber(words[1]) is Some(cents)) {
        return Some(Command.Pay(cents))
      }
    }
    _ => {}
  }
  return None
}

function totalCents(prices: Map<string, int>, basket: Map<string, int>): int {
  let total = 0
  for (const name of Maps.keys(basket)) {
    if (basket[name] is Some(quantity)) {
      const price = prices[name] is Some(cents) ? cents : 0
      total = total + price * quantity
    }
  }
  return total
}

export function run(args: Array<string>): int {
  const prices: Map<string, int> = Map { "tea": 250, "cake": 900 }
  let basket = Map<string, int> {}
  const lines = ["add tea 2", "add cake 1", "add cake 1", "remove tea", "pay 1000", "pay 2000"]
  for (const line of lines) {
    if (parseCommand(line) is Some(command)) {
      match (command) {
        Command.Add(name, quantity) => {
          const already = basket[name] is Some(count) ? count : 0
          basket[name] = already + quantity
        }
        Command.Remove(name) => {
          Maps.remove(mutable basket, name)
        }
        Command.Pay(cents) => {
          const due = totalCents(prices, basket)
          if (cents < due) {
            print(`${Int.toString(cents)} is not enough, ${Int.toString(due)} due`)
          } else {
            print(`paid, change ${Int.toString(cents - due)}`)
          }
        }
      }
    }
  }
  return 0
}
```

```text
1000 is not enough, 1800 due
paid, change 200
```

Each piece is one you have seen. An enum holds the commands, and `match` with tuple and literal patterns reads them. `Option` and `is` handle lines that do not parse. `basket[name] is Some(count) ? count : 0` reads a map entry with a fallback, using `is` as the condition of `? :`.

Next: [6. Option and Result](06-option-and-result.md).
