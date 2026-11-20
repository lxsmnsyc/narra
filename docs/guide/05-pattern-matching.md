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

An arm that can never be taken is an error as well. Here `_` already matches every value, so the arm after it is dead code:

```narra error
// error: 5:5: this arm is unreachable: the arms before it already match every value it matches
function sizeName(quantity: int): string {
  return match (quantity) {
    _ => "any"
    0 => "none"
  }
}

export function run(args: Array<string>): int {
  print(sizeName(0))
  return 0
}
```

The same holds for a repeated literal or a repeated variant. A guarded arm covers nothing for this check, because its guard can be false. A pattern after it is fine. Put the specific arms first and the catch-all last.

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

## Match an array

The till splits a line into words, so the words are an `Array<string>`. An array pattern is written like an array literal, and it tests the length as well as the elements. `[]` matches an empty array, `[first]` an array of exactly one element, and `["pay", amount]` an array of exactly two whose first element is `"pay"`. Elements are patterns, so literals, names and `_` all work. To accept any number of elements after the first few, end the pattern with `...`:

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

function describe(words: Array<string>): string {
  return match (words) {
    [] => "empty line"
    ["pay", amount] => `pay ${amount}`
    ["add", name, ..._] => `add ${name}`
    [first, ...rest] => `${first} with ${Int.toString(Arrays.length(rest))} more`
  }
}

export function run(args: Array<string>): int {
  print(describe([]))
  print(describe(["pay", "500"]))
  print(describe(["add", "tea", "2"]))
  print(describe(["dance", "a", "b"]))
  return 0
}
```

```text
empty line
pay 500
add tea
dance with 2 more
```

`..._` stands for "any further elements" and ignores them. `...rest` names them: `rest` is a new `Array` holding the remaining elements. The `...` part comes last, and a pattern can have only one. `["add", name, ..._]` matches three or more elements, `["add", name]` exactly two.

The compiler checks that the arms cover every length, as it does for enums. If only `[]` and `[_, _, _, ..._]` are listed, an array of one or two elements has no arm, and the error names the missing patterns:

```narra error
// error: 3:10: match does not cover [_], [_, _]; add the cases or a `_` arm
function describe(words: Array<string>): string {
  return match (words) {
    [] => "empty"
    [_, _, _, ..._] => "long"
  }
}

export function run(args: Array<string>): int {
  print(describe([]))
  return 0
}
```

An array pattern also works with `is`, as in `if (words is ["pay", amount])`, and inside other patterns, as in `Some([first, ..._])`.

A named rest copies the elements it keeps. That costs time in proportion to the number of elements, each time the pattern matches. A function that calls itself on `rest` for every element therefore does more and more copying as the array grows. For a long array, walk it with an index instead. `..._` copies nothing.

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

An `Array` has no fixed length, so a pattern that names elements cannot always match it. `const [a, b] = items` is an error, because `items` might hold one element or ten. The one array pattern that always matches is a lone rest, which makes a copy:

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

export function run(args: Array<string>): int {
  const items = [3, 4, 5]
  const [...copy] = items
  print(Int.toString(Arrays.length(copy)))
  if (items is [first, ..._]) {
    print(Int.toString(first))
  }
  const point: FixedArray<int, 3> = [1, 2, 3]
  const [x, y, z] = point
  print(Int.toString(x + y + z))
  return 0
}
```

```text
3
3
6
```

To test the length and bind in one step, use `is` or `match`, as `items is [first, ..._]` does. A `FixedArray` has a length the compiler knows. A pattern of exactly that many elements always matches, and it takes the array apart like a tuple. On a `FixedArray` the rest must be `..._`, since a named rest would need a type with a length of its own.

```narra error
// error: 4:9: this pattern does not match every Array<int>; use match or `is` to test a value
export function run(args: Array<string>): int {
  const items = [3, 4, 5]
  const [a, b] = items
  print("x")
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

## Calls: names, std functions, order

Three more things about calls turn up in the till's code.

A call can name its arguments, written like the fields of a struct literal. Chapter 3 used a name to skip a default. Names also make a call with several arguments of one type easy to read. They can come in any order, as long as the positional arguments come first:

```narra
import * as Int from "std/int"

function label(name: string, priceCents: int, quantity: int = 1): string {
  return `${Int.toString(quantity)} ${name} at ${Int.toString(priceCents)}`
}

export function run(args: Array<string>): int {
  print(label("tea", 250))
  print(label("tea", quantity: 3, priceCents: 250))
  print(label(name: "cake", priceCents: 900))
  return 0
}
```

```text
1 tea at 250
3 tea at 250
1 cake at 900
```

Every parameter is passed once, by position or by name. A `mutable` argument (chapter 9) keeps its word after the name:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

function fill(mutable buffer: Array<int>, value: int) {
  for (let index = 0; index < Arrays.length(buffer); index++) {
    buffer[index] = value
  }
}

export function run(args: Array<string>): int {
  let slots = [0, 0, 0]
  fill(buffer: mutable slots, value: 7)
  print(`${Int.toString(slots[0])} ${Int.toString(slots[1])} ${Int.toString(slots[2])}`)
  return 0
}
```

```text
7 7 7
```

A positional argument after a named one is an error:

```narra error
// error: 9:28: a positional argument cannot follow a named one; name it, or move it before the named arguments
import * as Int from "std/int"

function label(name: string, priceCents: int): string {
  return `${name} at ${Int.toString(priceCents)}`
}

export function run(args: Array<string>): int {
  print(label(name: "tea", 250))
  return 0
}
```

Names work when the callee is a declared function. A function held in a name, such as `const show = label`, takes its arguments by position only, and so do enum constructors and std functions. A parameter's name is part of what the function offers to its callers: renaming it breaks every call that uses the name.

A function from the standard library is a value too, when its types do not change from call to call. `Int.toString` can be stored, passed and returned like your own functions:

```narra
import * as Int from "std/int"
import * as Str from "std/string"
import * as Arrays from "std/array"

function formatAll(numbers: Array<int>, format: (int) => string): Array<string> {
  let out: Array<string> = []
  for (const number of numbers) {
    Arrays.push(mutable out, format(number))
  }
  return out
}

export function run(args: Array<string>): int {
  const format = Int.toString
  print(format(42))
  const words = formatAll([1, 2, 3], Int.toString)
  print(Str.join(words, "+"))
  return 0
}
```

```text
42
1+2+3
```

Some std functions cannot be values. `Arrays.length` works on an array of any element type, so it has no single function type. The same goes for a function that takes a `mutable` parameter, such as `Arrays.push`, and for one that takes a function, such as `Arrays.map`. The compiler asks for an arrow function that calls them:

```narra error
// error: 5:23: Arrays.length cannot be used as a value: its parameter and result types depend on the call; wrap it in an arrow function, such as `(items) => Arrays.length(items)`
import * as Arrays from "std/array"

export function run(args: Array<string>): int {
  const size = Arrays.length
  return 0
}
```

Last, the arguments of a call are evaluated from left to right, in the order they are written, also when they are named. This matters when an argument is a block that changes a `let` which an earlier argument reads. The earlier argument keeps the value it had when it was evaluated:

```narra
import * as Int from "std/int"

function show(first: int, second: int) {
  print(`${Int.toString(first)} then ${Int.toString(second)}`)
}

export function run(args: Array<string>): int {
  let counter = 0
  show(counter, {
    counter = counter + 10
    return counter
  })
  show(counter, 0)
  return 0
}
```

```text
0 then 10
10 then 0
```

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
