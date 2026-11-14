# 12. Narra for TypeScript programmers

This chapter is optional. It is for readers who already write TypeScript and want to know which habits carry over to Narra and which ones the compiler stops. Each topic shows the Narra form, with the error the TypeScript habit produces, and links to the chapter that teaches it.

## What carries over, and what does not

Much of a Narra file looks like TypeScript. These carry over: `function` declarations with typed parameters, `const` and `let`, arrow functions and template strings. So do `for ... of`, `import * as`, `import { ... }`, `export` and generics in angle brackets, and `if`, `while`, `? :`, `&&`, `||` and `!`.

```narra
import * as Int from "std/int"

function total(prices: Array<int>): int {
  let sum = 0
  for (const price of prices) {
    sum = sum + price
  }
  return sum
}

export function run(args: Array<string>): int {
  const prices = [250, 320, 180]
  const label = total(prices) > 500 ? "large" : "small"
  print(`${Int.toString(total(prices))} cents, ${label}`)
  return 0
}
```

```text
750 cents, large
```

Underneath, four things differ, and most of the rest of this chapter follows from them:

- Values are not shared. A change through one name never shows through another, and `copy(x)` makes the second value.
- There is no `null` and no `undefined`. A value that may be missing has the type `Option<T>`.
- There are no exceptions. A function that can fail returns a `Result<T, E>`.
- There is no garbage collector. Memory is freed by reference counting, at the moment the last copy of a value goes away.

A program starts at an exported function `run` that returns an exit code ([chapter 1](01-getting-started.md#say-hello)), not at the top of a file.

## Statements and operators

A statement ends at the end of its line. A `;` is an error that says so:

```narra error
// error: 3:17: semicolons are not used; end the statement with a line break
export function run(args: Array<string>): int {
  print("hello");
  return 0
}
```

`==` compares by contents, the same way for numbers, strings, arrays and structs. There is no `===`, since Narra has only one equality:

```narra error
// error: 4:13: use `==`; there is no strict equality operator
export function run(args: Array<string>): int {
  const count = 1
  if (count === 1) {
    print("one")
  }
  return 0
}
```

`+=`, `-=` and the other compound forms, and `count++` and `count--`, work as in TypeScript, with one difference: they are statements with no value. `const copy = count++` and the prefix `++count` are errors ([chapter 2](02-values-and-types.md#change-a-value-over-time)):

```narra error
// error: 4:3: there is no prefix `++`; write `x++` as a statement
export function run(args: Array<string>): int {
  let count = 0
  ++count
  return count
}
```

Loops are `for ... of`, `while`, `break` and `continue`. There is no `for ... in`, and the error points to `of`:

```narra error
// error: 4:19: use `of`; `for ... in` is not part of the language
export function run(args: Array<string>): int {
  const names = ["tea", "cake"]
  for (const name in names) {
    print(name)
  }
  return 0
}
```

The C-style `for (let i = 0; i < n; i++)` loop works as in TypeScript. To walk a range, loop over `Ranges.upTo(n)`; for positions and elements, use `Arrays.entries` ([chapter 2](02-values-and-types.md#lists-of-values)).

## Numbers

TypeScript has one `number`. Narra has two number types: `int` for whole numbers and `f64` for decimals. The TypeScript name is an error:

```narra error
// error: 2:24: there is no `number`; use `int` or `f64`
function double(value: number): number {
  return value * 2
}

export function run(args: Array<string>): int {
  return 0
}
```

The two types never convert on their own, and `/` on two `int` values drops the fraction. `Float.fromInt` and `Float.truncate` convert between them ([chapter 2](02-values-and-types.md#whole-numbers-and-decimals)):

```narra
import * as Float from "std/float"
import * as Int from "std/int"

export function run(args: Array<string>): int {
  print(Int.toString(7 / 2))
  print(Float.toString(Float.fromInt(7) / 2.0))
  return 0
}
```

```text
3
3.5
```

## Strings and printing

In TypeScript, single and double quotes both make strings. In Narra, double quotes make a `string`, and single quotes make a `char`, which holds exactly one character:

```narra error
// error: 3:16: a character literal holds exactly one character; use double quotes for strings
export function run(args: Array<string>): int {
  const name = 'tea'
  print(name)
  return 0
}
```

`+` adds numbers and never joins strings. Use a template string ([chapter 2](02-values-and-types.md#text)):

```narra error
// error: 4:25: strings are joined with a template string, not +
export function run(args: Array<string>): int {
  const name = "tea"
  const line = "item: " + name
  print(line)
  return 0
}
```

Nothing turns into text on its own. `print` takes only a `string`, and so does every `${...}` part of a template. A number goes through `Int.toString` or `Float.toString` first:

```narra error
// error: 4:12: a template string part must be a string, found int; convert it, for example with Int.toString
export function run(args: Array<string>): int {
  const count = 3
  print(`${count} items`)
  return 0
}
```

There is no `console.log` that prints any value. For a struct or an enum, derive the `Show` trait and print `Show.show(value)`. It plays the part of `JSON.stringify` or `String(value)` ([chapter 7](07-traits.md#printing-any-value-with-show)):

```narra
struct Item derives Show {
  name: string
  cents: int
}

export function run(args: Array<string>): int {
  print(Show.show(Item { name: "tea", cents: 250 }))
  print(Show.show([1, 2, 3]))
  return 0
}
```

```text
Item { name: "tea", cents: 250 }
[1, 2, 3]
```

`.length` is not a property. Strings and arrays are measured with module functions, `Str.length` and `Arrays.length`:

```narra error
// error: 4:21: there is no .length; use Array.length(x) or Str.length(x)
export function run(args: Array<string>): int {
  const name = "tea"
  const size = name.length
  return 0
}
```

The fix imports the modules and calls their functions, with the alias names `Str` and `Arrays` that the rest of this guide uses:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"
import * as Str from "std/string"

export function run(args: Array<string>): int {
  print(Int.toString(Str.length("tea")))
  print(Int.toString(Arrays.length([1, 2])))
  return 0
}
```

```text
3
2
```

## Conditions

There is no truthiness. A condition must be a `bool`, so `if (count)` and `if (name)` do not compile. Write the comparison you mean, such as `count > 0` or `Str.length(name) > 0` ([chapter 2](02-values-and-types.md#make-a-decision)):

```narra error
// error: 4:7: expected bool, found string
export function run(args: Array<string>): int {
  const name = "tea"
  if (name) {
    print(name)
  }
  return 0
}
```

The ternary `cond ? a : b` has the same form as in TypeScript, and its condition must be a `bool` too ([chapter 5](05-pattern-matching.md#compute-a-value-in-several-steps)).

## Arrays and maps

`T[]` is not a type. Write `Array<T>`:

```narra error
// error: 4:34: there is no `T[]` array type; write `Array<T>`
import * as Arrays from "std/array"

export function total(prices: int[]): int {
  return Arrays.length(prices)
}

export function run(args: Array<string>): int {
  return total([1, 2])
}
```

There are no method calls, on arrays or on anything else. `items.push(x)` is `Arrays.push(mutable items, x)`, where `mutable` marks the argument the function changes ([chapter 2](02-values-and-types.md#lists-of-values)):

```narra error
// error: 4:9: there are no method calls; call a module function such as Module.push(value, ...) (value has type Array<int>)
export function run(args: Array<string>): int {
  let items: Array<int> = []
  items.push(1)
  return 0
}
```

`Arrays` is not a global like the `Array` of TypeScript. It is the name you give the module `std/array` when you import it. Without the import, the name does not exist ([chapter 8](08-modules-and-packages.md#what-needs-no-import)):

```narra error
// error: 4:3: unknown module "Arrays"; import it first
export function run(args: Array<string>): int {
  let items = [1]
  Arrays.push(mutable items, 2)
  return 0
}
```

A `const` means what it means in TypeScript: it fixes the name, and the value's elements can still change. Only giving the name a new value needs `let` ([chapter 2](02-values-and-types.md#arrays-are-values)):

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const prices = [250, 320]
  prices[0] = 200
  Arrays.push(mutable prices, 410)
  print(Int.toString(prices[0] + prices[2]))
  return 0
}
```

```text
610
```

A top-level constant is the exception: it is fixed at every depth, since a constant that could change would be a global variable.

Reading a map gives an `Option`, not a `V | undefined`. `stock["tea"]` has the type `Option<int>`, and you check it before you use it ([chapter 2](02-values-and-types.md#look-up-a-value-by-name)):

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const stock: Map<string, int> = Map { "tea": 12 }
  if (stock["tea"] is Some(count)) {
    print(`tea: ${Int.toString(count)}`)
  }
  if (stock["cake"] is None) {
    print("no cake")
  }
  return 0
}
```

```text
tea: 12
no cake
```

Map keys compare by contents too. In TypeScript, two objects with the same fields are two different `Map` keys. In Narra, they find the same entry ([chapter 7](07-traits.md#equality-comes-free)).

## Copies instead of shared references

In TypeScript, `const b = a` on an array or object gives two names for one object, and a change through one name shows through the other:

```ts
const scores = [10, 20]
const before = scores
scores.push(30)
console.log(before.length) // 3
```

In Narra, a value has one meaning wherever it is stored: a change through one name never shows up through another. To get that, Narra makes the copy visible. `copy(scores)` gives `before` a value of its own, so it keeps two elements. The copy is cheap: the two names share the elements until one of them changes ([chapter 2](02-values-and-types.md#arrays-are-values), [chapter 9](09-memory-and-mutable.md#why-copies-are-cheap)):

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

struct Player {
  name: string
  hp: int
}

export function run(args: Array<string>): int {
  const scores = [10, 20]
  const before = copy(scores)
  Arrays.push(mutable scores, 30)
  print(Int.toString(Arrays.length(before)))

  const hero = Player { name: "Ana", hp: 100 }
  const other = copy(hero)
  other.hp = 50
  print(Int.toString(hero.hp))
  return 0
}
```

```text
2
100
```

A value has no identity apart from its contents, so there is no "same object" test. The TypeScript habit that breaks is `const before = scores` followed by a change to `scores`. In TypeScript the change shows through `before`. Narra neither copies silently nor shares the change, so it rejects the write instead: a write to a value that another name still uses is a compile error that suggests `copy(...)` ([chapter 9](09-memory-and-mutable.md#writing-to-a-shared-value)):

```narra error
// error: 8:15: Arrays.push(mutable scores, ...) may change scores, which before still uses at line 9; take a copy where it is shared, with copy(...)
import * as Arrays from "std/array"
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const scores = [10, 20]
  const before = scores
  Arrays.push(mutable scores, 30)
  print(Int.toString(Arrays.length(before)))
  return 0
}
```

So TypeScript code that relied on sharing either gives the same result or stops with an error that names the fix.

A `for ... of` loop variable, a name bound by a pattern (`found is Some(player)`) and a destructured name hold copies too. Since TypeScript code that changes them expects the original to change, the compiler rejects a change to their fields or elements. For the loop over an array, write `for (mutable player of players)`: then `player.hp = 0` changes `players`, as in TypeScript, and the body cannot use `players` itself ([chapter 9](09-memory-and-mutable.md#heal-the-whole-party)). Elsewhere, write through the source, `players[index].hp = 0`, or copy the value first with `copy(player)`.

## Missing values

There is no `null` and no `undefined`. A value that may be missing has the type `Option<T>`, which is `Some(value)` or `None`. Write `None` where TypeScript returns `null` ([chapter 6](06-option-and-result.md#return-a-value-that-may-be-missing)):

```narra error
// error: 3:10: there is no null; use `None` with an `Option` type
function findPrice(name: string): Option<int> {
  return null
}

export function run(args: Array<string>): int {
  return 0
}
```

`undefined` gets the same answer:

```narra error
// error: 3:17: there is no undefined; use `None` with an `Option` type
export function run(args: Array<string>): int {
  const price = undefined
  return 0
}
```

`Option<T>` is like `T | undefined`, except that the compiler makes you check it. `is` takes the place of `x !== null` checks and type guards. After an early return on `None`, the compiler knows the value is a `Some`, and `.value` reads it ([chapter 5](05-pattern-matching.md#read-a-value-that-may-be-missing)):

```narra
import * as Int from "std/int"

function describe(stock: Map<string, int>, name: string): string {
  const found = stock[name]
  if (found is None) {
    return `${name}: none`
  }
  return `${name}: ${Int.toString(found.value)}`
}

export function run(args: Array<string>): int {
  const stock: Map<string, int> = Map { "tea": 12 }
  print(describe(stock, "tea"))
  print(describe(stock, "cake"))
  return 0
}
```

```text
tea: 12
cake: none
```

There is no `??` and no `?.`. The functions in `std/option` cover them ([chapter 6](06-option-and-result.md#fall-back-to-a-default)):

| TypeScript | Narra |
| --- | --- |
| `a ?? b` | `Options.withDefault(a, b)` |
| `a ?? f()`, where `f` runs only when needed | `Options.orElse(a, () => f())` |
| `a?.b` | `Options.map(a, (x) => x.b)` |

`withDefault` always computes its fallback, as any function argument is computed. TypeScript's `??` is lazy, so `orElse` is the form to use when the fallback costs work:

```narra
import * as Options from "std/option"

export function run(args: Array<string>): int {
  const settings: Map<string, string> = Map { "hostname": "example.org" }
  print(Options.withDefault(settings["user"], "guest"))
  const host = Options.orElse(settings["host"], () => settings["hostname"])
  const address = Options.map(host, (name) => `https://${name}`)
  print(Options.withDefault(address, "no address"))
  return 0
}
```

```text
guest
https://example.org
```

## Errors without exceptions

There is no `throw`. A function that can fail returns a `Result<T, E>`, which is `Ok(value)` or `Err(error)` ([chapter 6](06-option-and-result.md#errors-of-your-own)):

```narra error
// error: 3:3: there is no `throw`; return `Err(value)`, or call `panic` for a bug
function parsePort(text: string): Result<int, string> {
  throw "bad port"
}

export function run(args: Array<string>): int {
  return 0
}
```

There is no `try` or `catch` either. The caller checks the `Result` with `is` or `match`:

```narra error
// error: 3:3: there is no `try` or `catch`; return a `Result` and check it with `is Err` or `match`
export function run(args: Array<string>): int {
  try {
    return 1
  } catch (error) {
    return 2
  }
}
```

The same program written with `Result`:

```narra
import * as Int from "std/int"

function parsePort(text: string): Result<int, string> {
  const parsed = Int.parse(text, 10)
  if (parsed is Err(reason)) {
    return Err(`"${text}": ${reason}`)
  }
  if (parsed.value < 1) {
    return Err(`"${text}": below 1`)
  }
  return Ok(parsed.value)
}

export function run(args: Array<string>): int {
  for (const text of ["8080", "0", "80x"]) {
    match (parsePort(text)) {
      Ok(port) => print(`port ${Int.toString(port)}`)
      Err(message) => print(message)
    }
  }
  return 0
}
```

```text
port 8080
"0": below 1
"80x": invalid digit for the base
```

The error moves up by hand: test, return the `Err`, and read the value on the next line.

A TypeScript promise or return value can be dropped without a word. A `Result` that is dropped is a compile error. `const _ = save(text)` drops it on purpose, where TypeScript would write `void save(text)` ([chapter 6](06-option-and-result.md#a-result-must-be-used)):

```narra error
// error: 7:3: this call returns a Result that is not checked; handle it, or discard it with `const _ = ...`
function save(text: string): Result<int, string> {
  return Ok(1)
}

export function run(args: Array<string>): int {
  save("port = 8080")
  return 0
}
```

Chains of `.then(...)` on a promise become calls to `Results.andThen` and `Results.map` ([chapter 6](06-option-and-result.md#shorter-code-with-stdresult)). A bug that no caller can handle stops the program with `panic("message")`, and no code can catch a panic ([chapter 6](06-option-and-result.md#when-failure-is-a-bug)).

## Functions

A named function always writes the type of each parameter and of its result. Only arrow functions passed straight to another function leave their types out ([chapter 3](03-functions-and-closures.md#pass-a-function-as-a-value)). The result type is not inferred: a function without one returns nothing, so its `return value` is an error:

```narra error
// error: 3:3: this function returns nothing
function double(value: int) {
  return value * 2
}

export function run(args: Array<string>): int {
  return double(0)
}
```

A parameter can have a default, as in TypeScript, but the default must be a constant: `depth: int = 0` works, `depth: int = start()` does not ([chapter 3](03-functions-and-closures.md#arguments-a-caller-may-leave-out)). There is no `name?: T`; write `name: Option<T> = None`, and the caller passes `Some(3)`. There are no rest parameters or overloads. Each function has one list of parameters, and two functions in a module cannot share a name. Give each version its own name:

```narra error
// error: 5:1: function "describe" is declared twice
function describe(cents: int): string {
  return "a price"
}
function describe(name: string): string {
  return name
}

export function run(args: Array<string>): int {
  return 0
}
```

A function cannot be declared inside another function. Declare it at module level, or use an arrow function ([chapter 3](03-functions-and-closures.md#a-small-helper-inside-a-function)):

```narra error
// error: 3:3: a function inside a function is not supported; declare it at module level as a private function and pass the enclosing values as arguments
export function run(args: Array<string>): int {
  function double(value: int): int {
    return value * 2
  }
  return double(0)
}
```

`firstOr` returns an element of `items`, so it says `from items`, and it returns a copy of `fallback`, because a result can share only the one parameter `from` names ([chapter 9](09-memory-and-mutable.md#return-part-of-a-parameter)). A call may write type arguments, as in `firstOr<int>([], 0)`, the same as in TypeScript. You rarely need them, because the compiler works them out from the arguments. Write them when no argument says the type, as with `[]`. Write all of them or none. The other way is to annotate the variable that holds the value ([chapter 3](03-functions-and-closures.md#one-function-for-many-types)):

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

function firstOr<T>(items: Array<T>, fallback: T): T from items {
  if (Arrays.length(items) == 0) {
    return copy(fallback)
  }
  return items[0]
}

export function run(args: Array<string>): int {
  print(Int.toString(firstOr<int>([], 0)))
  const none: Array<int> = []
  print(Int.toString(firstOr(none, 1)))
  print(firstOr(["tea", "cake"], "nothing"))
  return 0
}
```

```text
0
1
tea
```

There is no automatic partial application. A function that takes one argument and returns another function is written out, as `makeDiscount` does in [chapter 3](03-functions-and-closures.md#make-new-functions-from-old-ones).

## Closures

A TypeScript closure sees the variable itself, so it sees every later assignment. A Narra closure passed straight to a parameter that is not `escaping` does the same: it sees and assigns the variable itself. Any other closure captures a copy, and it may capture only constants. Capturing a `let` there is an error ([chapter 3](03-functions-and-closures.md#use-values-from-around-a-function), [chapter 9](09-memory-and-mutable.md#closures-and-let)).

So the classic TypeScript counter does not compile:

```ts
function makeCounter() {
  let count = 0
  return () => ++count
}
```

```narra error
// error: 5:5: a closure cannot capture "count" because it is declared with let, unless it is passed straight to a function parameter that is not escaping
function makeCounter(): () => int {
  let count = 0
  return () => {
    count = count + 1
    return count
  }
}

export function run(args: Array<string>): int {
  const next = makeCounter()
  return next()
}
```

Keep the count in a variable of the caller. A function returns the new value, or changes a counter held in a struct through a `mutable` parameter:

```narra
import * as Int from "std/int"

struct Counter {
  count: int
}

function addOne(mutable counter: Counter) {
  counter.count = counter.count + 1
}

function nextCount(count: int): int {
  return count + 1
}

export function run(args: Array<string>): int {
  const counter = Counter { count: 0 }
  addOne(mutable counter)
  addOne(mutable counter)
  print(Int.toString(counter.count))
  let count = counter.count
  count = nextCount(count)
  print(Int.toString(count))
  return 0
}
```

```text
2
3
```

A `mutable` parameter reaches the contents of a value, so its type must have contents to change, and `mutable count: int` is an error. It cannot be returned, or captured by a closure that outlives the call, so `addOne` returns nothing. The caller reads its own `counter`.

A function parameter can be called while the function runs, and nothing more. To return it, store it or capture it in a closure that outlives the call, mark it `escaping`. TypeScript has no such marker, since every function value there may be kept ([chapter 3](03-functions-and-closures.md#make-new-functions-from-old-ones)):

```narra error
// error: 3:21: a closure cannot capture parameter "rule" because it is not escaping; declare it as `escaping rule`
function twice(rule: (cents: int) => int): (cents: int) => int {
  return (cents) => rule(rule(cents))
}

export function run(args: Array<string>): int {
  return twice((cents) => cents - 10)(100)
}
```

Written as `escaping rule`, the function compiles.

## Classes and objects

There are no classes. Data is a `struct`, and behavior is functions in a module that take the struct as a parameter ([chapter 4](04-structs-and-enums.md#group-fields-into-a-struct)):

```narra error
// error: 2:1: there is no `class`; use a `struct` and functions in a module
class Point {
  x: int
}

export function run(args: Array<string>): int {
  return 0
}
```

There is no `new`. A struct literal names the type, as in `Point { x: 1 }`:

```narra error
// error: 5:17: there is no `new`; write a struct literal such as `Name { field: value }`, or call a function
struct Point { x: int }

export function run(args: Array<string>): int {
  const point = new Point(1)
  return point.x
}
```

There is no `this`. A function gets the value it works on as a parameter:

```narra error
// error: 3:10: there is no `this`; pass the value as a parameter
export function run(args: Array<string>): int {
  return this.count
}
```

Declaring a variable named `this` gets the same error.

```narra error
// error: 3:9: there is no `this`; pass the value as a parameter
export function run(args: Array<string>): int {
  const this = 1
  return 0
}
```

The Narra form of a small class with a method:

```narra
import * as Int from "std/int"

struct Point {
  x: int
  y: int
}

function moved(point: Point, dx: int, dy: int): Point {
  return { ...point, x: point.x + dx, y: point.y + dy }
}

function describe(point: Point): string {
  return `(${Int.toString(point.x)}, ${Int.toString(point.y)})`
}

export function run(args: Array<string>): int {
  const start = Point { x: 1, y: 2 }
  print(describe(moved(start, 3, 4)))
  print(describe(start))
  return 0
}
```

```text
(4, 6)
(1, 2)
```

The spread update `{ ...point, x: 4 }` works as in TypeScript and builds a new value ([chapter 4](04-structs-and-enums.md#change-an-item)). Field shorthand works too: `Item { name }` means `Item { name: name }`. It needs the type name, though. A plain `{ name }` does not build an object: without a type name, `{ ... }` in value position is a block.

```narra
struct Item { name: string }

export function run(args: Array<string>): int {
  const name = "tea"
  const item = Item { name }
  print(item.name)
  return 0
}
```

```text
tea
```

A class with private fields becomes an `export opaque struct`. Code outside the module cannot build it or read its fields, and goes through the functions the module exports ([chapter 4](04-structs-and-enums.md#hide-the-fields-of-a-struct)).

## Interfaces and traits

There is no `interface`. A data shape is a `struct`, and a set of operations that several types share is a `trait`:

```narra error
// error: 2:8: there is no `interface`; use a `struct` for data or a `trait` for behavior
export interface Point {
  x: int
}

export function run(args: Array<string>): int {
  return 0
}
```

A trait is narrower than an interface. It holds only functions, and a type lists all of its traits where it is declared, with `implements` or `derives` ([chapter 7](07-traits.md#giving-two-types-the-same-operation)). Typing is not structural: a struct has a trait only when it lists it, even when a function of the right shape exists:

```narra error
// error: 13:13: Ticket does not implement Label, which show needs for type parameter T
trait Label {
  function label(item: Self): string
}
struct Ticket { seat: int }
function label(item: Ticket): string {
  return "ticket"
}
function show<T: Label>(item: T): string {
  return Label.label(item)
}
export function run(args: Array<string>): int {
  print(show(Ticket { seat: 1 }))
  return 0
}
```

A trait function is called through the trait, as `Label.label(book)`, where TypeScript writes `book.label()`. The method form is the same error as `items.push`:

```narra error
// error: 14:14: there are no method calls; call a module function such as Module.label(value, ...) (value has type Book)
trait Label {
  function label(item: Self): string
}
struct Book implements Label { title: string }
implement Label for Book {
  function label(item: Book): string {
    return item.title
  }
}

export function run(args: Array<string>): int {
  const book = Book { title: "Dune" }
  print(book.label())
  return 0
}
```

Only the module that declares a type can implement a trait for it. For `int` or a type from another module, wrap it in a struct of your own ([chapter 7](07-traits.md#traits-for-a-type-you-did-not-declare)).

## Unions and enums

A Narra `enum` takes the place of three TypeScript patterns: an `enum`, a tagged union with a `kind` field, and a small class hierarchy. Each variant can carry its own fields, and `match` takes the place of `switch` ([chapter 4](04-structs-and-enums.md#a-value-of-one-of-several-kinds)):

```ts
type Shape =
  | { kind: "circle"; radius: number }
  | { kind: "square"; side: number }
```

```narra
import * as Int from "std/int"

enum Shape {
  Circle(radius: int)
  Square(side: int)
}

function area(shape: Shape): int {
  return match (shape) {
    Shape.Circle(radius) => 3 * radius * radius
    Shape.Square(side) => side * side
  }
}

export function run(args: Array<string>): int {
  for (const shape of [Shape.Circle(2), Shape.Square(3)]) {
    print(Int.toString(area(shape)))
  }
  return 0
}
```

```text
12
9
```

An enum whose variants carry no fields, such as a `Size` with the variants `Small` and `Large`, plays the part of a plain TypeScript `enum`.

`match` has no fallthrough and no `break`. It must cover every variant. Add a variant, and each `match` that misses it fails to compile. A TypeScript `switch` needs a `never` check in its `default` case for that ([chapter 5](05-pattern-matching.md#decide-by-kind-and-by-condition)):

```narra error
// error: 8:10: match does not cover Square(_); add the cases or a `_` arm
enum Shape {
  Circle(radius: int)
  Square(side: int)
}

function area(shape: Shape): int {
  return match (shape) {
    Shape.Circle(radius) => 3 * radius * radius
  }
}

export function run(args: Array<string>): int {
  return 0
}
```

A branded type such as `number & { __brand: "UserId" }` is a `distinct type`. It has the same values as `int`, and the compiler keeps the two apart ([chapter 4](04-structs-and-enums.md#keep-similar-numbers-apart)):

```narra
import * as Int from "std/int"

distinct type UserId = int

function describeUser(id: UserId): string {
  return `user ${Int.toString(UserId.toInt(id))}`
}

export function run(args: Array<string>): int {
  print(describeUser(UserId.from(7)))
  return 0
}
```

```text
user 7
```

## Modules

The import forms are a subset of ES modules: `import * as Name from "path"` and `import { a, b as c } from "path"`. There are no default imports:

```narra error
// error: 2:8: use `import * as Name from "path"` or `import { a, b } from "path"`; there are no default imports
import Int from "std/int"

export function run(args: Array<string>): int {
  return Int.abs(0 - 1)
}
```

There is no `export default`, no `export *` and no `export { x } from`. Each name is exported where it is declared, with `export` in front of the declaration ([chapter 8](08-modules-and-packages.md#what-a-module-can-hold)):

```narra error
// error: 2:8: expected a declaration, found "default"
export default function double(value: int): int {
  return value * 2
}

export function run(args: Array<string>): int {
  return 0
}
```

A module cannot hold a top-level `let`, only functions, types and constants. State that changes lives in `run` and is passed to the functions that need it:

```narra error
// error: 2:1: expected a declaration, found "let"
let count = 0

export function run(args: Array<string>): int {
  return 0
}
```

An import path starts with the package name, as in `"words/count"`, where TypeScript writes a relative path such as `"./count"` ([chapter 8](08-modules-and-packages.md#splitting-a-program-into-files)). A relative path is an unknown module:

```narra error
// error: 2:1: unknown module "./count"
import * as Count from "./count"

export function run(args: Array<string>): int {
  return 0
}
```

The standard modules for the built-in types are imported under plural names, `Arrays`, `Maps`, `Options` and `Results`. An alias cannot share a name with a type, where TypeScript would allow `import * as Array` ([chapter 8](08-modules-and-packages.md#what-needs-no-import)):

```narra error
// error: 2:1: the import alias Array is also a type name; rename it, for example `import * as Arrays from "std/array"`
import * as Array from "std/array"

export function run(args: Array<string>): int {
  return Array.length([1, 2])
}
```

## Changing a value in a function

A TypeScript function that receives an object can change it, and the caller cannot tell from the call. In Narra, a parameter is a read-only copy unless it is declared `mutable`, and then the call writes `mutable` too ([chapter 9](09-memory-and-mutable.md#changing-a-player-in-a-function)):

```narra
import * as Int from "std/int"

struct Player {
  name: string
  hp: int
}

function damage(mutable player: Player, amount: int) {
  player.hp = player.hp - amount
}

export function run(args: Array<string>): int {
  const hero = Player { name: "Ana", hp: 100 }
  damage(mutable hero, 10)
  print(Int.toString(hero.hp))
  return 0
}
```

```text
90
```

Compared with a TypeScript object parameter, a `mutable` parameter:

- is changed only when the function declares it, and the change is visible at the call;
- cannot be returned, stored or captured by the function;
- gives the function the only access to the variable during the call, so no other argument of the same call can use it;
- needs a local `const` or `let`, a `mutable` parameter or a part of one as the argument;
- must have a type with contents, such as a struct, an `Array` or a `Map`, and cannot be replaced whole.

Assigning to a field of a plain parameter is an error:

```narra error
// error: 8:3: parameter "player" is read-only; declare it as `mutable player` to change the caller's value, or copy it first with `copy(player)`
struct Player {
  name: string
  hp: int
}

function damage(player: Player, amount: int) {
  player.hp = player.hp - amount
}

export function run(args: Array<string>): int {
  return 0
}
```

For a parameter, the fix is `mutable player`, or returning a new `Player` built with a spread update.

## Effects

A TypeScript function that prints and a pure function have the same type. In Narra, an exported function lists its effects after its result type with `uses`. It writes `alloc` when it allocates memory, as building a string does, and `io` when it does input or output, such as printing. The list is part of the function's API, the same way its result type is ([chapter 10](10-effects-and-testing.md#what-a-function-does-in-its-signature)):

```narra error
// error: 5:25: Function `describe` uses `alloc` (through `Int.toString`) but its signature does not declare it; write `uses alloc`
import * as Int from "std/int"

export function describe(cents: int): string {
  return `${Int.toString(cents)} cents`
}

export function run(args: Array<string>): int {
  print(describe(250))
  return 0
}
```

With `uses alloc` after `: string`, the program compiles. Private functions work out their effects themselves, so only exported functions write the list.

## Tests

Tests live in the same file as the code, in `test` blocks, where TypeScript projects use separate `*.test.ts` files and Jest or Vitest. `narrac test file.nra` runs them ([chapter 10](10-effects-and-testing.md#writing-tests)):

```narra fragment
import * as Test from "std/test"

function double(value: int): int {
  return value * 2
}

test "double doubles" {
  Test.eq(double(4), 8)
}
```

| Jest or Vitest | `std/test` |
| --- | --- |
| `expect(a).toEqual(b)` | `Test.eq(a, b)` |
| `expect(a).not.toEqual(b)` | `Test.ne(a, b)` |
| `expect(f).toThrow(message)` | `Test.panicsWith(f, message)` |

## If you also know Rust or Go

- `Option` and `Result` match Rust's types, but there is no `?` operator and no `unwrap`. Check with `is` or `match`, or use the `std/option` and `std/result` functions.
- A trait is like a Rust trait, with functions only. Only the module that declares a type implements a trait for it, which is stricter than Rust's orphan rule. `derives Show` is like `#[derive(Debug)]`.
- A file is a module, like a Rust `mod`, and `export` takes the place of `pub`.
- A `mutable` parameter is close to Rust's `&mut`: exclusive access for the length of a call. There is no borrow checker, no lifetimes and no `&`.
- Unlike Go, there is no garbage collector: reference counting frees a value when its last copy goes away.
- `uses` is closer to a list of capabilities than to Rust's `async` or `unsafe` function coloring.

This is the last chapter. [The index](README.md) lists every chapter and the features that are planned.
