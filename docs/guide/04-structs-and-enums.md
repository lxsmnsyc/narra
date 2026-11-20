# 4. Structs and enums

So far an item of the shop was a loose name and price, passed around side by side. This chapter gives the shop's data its own types. A struct groups named fields into one value. An enum says that a value is one of several kinds, each with its own data.

## Group fields into a struct

An item has a name, a price and a quantity. Declare a struct with those three fields, and write functions that take an item:

```narra
import * as Int from "std/int"

struct Item {
  name: string
  priceCents: int
  quantity: int = 1
}

function lineTotal(item: Item): int {
  return item.priceCents * item.quantity
}

function describe(item: Item): string {
  return `${Int.toString(item.quantity)} x ${item.name} = ${Int.toString(lineTotal(item))}`
}

export function run(args: Array<string>): int {
  const tea = Item { name: "tea", priceCents: 250, quantity: 2 }
  const coffee = Item { name: "coffee", priceCents: 320 }
  print(describe(tea))
  print(describe(coffee))
  return 0
}
```

```text
2 x tea = 500
1 x coffee = 320
```

`struct Item { ... }` declares a new type with one field per line. `Item { name: "tea", ... }` builds a value of it, and `item.priceCents` reads a field.

`quantity: int = 1` gives the field a default, so `coffee` leaves it out and gets 1. A field without a default must be given:

```narra error
// error: 9:16: missing field "priceCents" of Item
struct Item {
  name: string
  priceCents: int
  quantity: int = 1
}

export function run(args: Array<string>): int {
  const item = Item { name: "tea" }
  return item.quantity
}
```

When a variable has the name of the field, write the name alone: `Item { name, priceCents }` means `Item { name: name, priceCents: priceCents }`. Shorthand and `name: value` fields can be mixed in one literal.

```narra
import * as Int from "std/int"

struct Item {
  name: string
  priceCents: int
}

export function run(args: Array<string>): int {
  const name = "tea"
  const item = Item { name, priceCents: 250 }
  print(`${item.name} ${Int.toString(item.priceCents)}`)
  return 0
}
```

```text
tea 250
```

You can leave out the struct name when the compiler already knows which struct you mean. That is the case when the value goes to a parameter, a result or a name with a written type:

```narra
import * as Int from "std/int"

struct Item {
  name: string
  priceCents: int
  quantity: int = 1
}

function lineTotal(item: Item): int {
  return item.priceCents * item.quantity
}

export function run(args: Array<string>): int {
  const cocoa: Item = { name: "cocoa", priceCents: 280 }
  print(Int.toString(lineTotal(cocoa)))
  print(Int.toString(lineTotal({ name: "juice", priceCents: 190, quantity: 3 })))
  return 0
}
```

```text
280
570
```

## Change an item

A customer takes one more tea. Copy the item into a `let` name and change the field there:

```narra
import * as Int from "std/int"

struct Item {
  name: string
  priceCents: int
  quantity: int = 1
}

export function run(args: Array<string>): int {
  const tea = Item { name: "tea", priceCents: 250 }
  let moreTea = copy(tea)
  moreTea.quantity = moreTea.quantity + 1
  print(`${Int.toString(tea.quantity)} then ${Int.toString(moreTea.quantity)}`)

  const onSale = { ...tea, priceCents: 200 }
  print(`${Int.toString(tea.priceCents)} then ${Int.toString(onSale.priceCents)}`)

  if (tea != onSale && tea == Item { name: "tea", priceCents: 250 }) {
    print("compared field by field")
  }
  return 0
}
```

```text
1 then 2
250 then 200
compared field by field
```

A struct behaves like the arrays of chapter 2: `copy(tea)` makes a value of its own, and changing `moreTea` leaves `tea` as it was. Without `copy`, `let moreTea = tea` would make two names for one struct, so the compiler rejects it and asks for `copy(tea)` ([chapter 9](09-memory-and-mutable.md#a-second-name-for-a-value)). You can assign a field of a `const` or a `let` name.

Often you want a new value that differs in one or two fields. `{ ...tea, priceCents: 200 }` copies every field of `tea`, then sets the fields listed after it. It is a new value, and `tea` does not change. The last `if` shows that `==` and `!=` compare two structs field by field.

## A value of one of several kinds

Customers pay in different ways. Cash needs only an amount. A card payment also has the last four digits of the card. A voucher has a code. The kinds carry different data, so a struct with every possible field would be wrong for all of them. An enum lists the kinds, called variants, each with its own fields:

```narra
import * as Int from "std/int"

enum Payment {
  Cash(cents: int)
  Card(cents: int, lastDigits: string)
  Voucher(code: string)
}

function describe(payment: Payment): string {
  return match (payment) {
    Payment.Cash(cents) => `cash ${Int.toString(cents)}`
    Payment.Card(cents, lastDigits) => `card ...${lastDigits} ${Int.toString(cents)}`
    Payment.Voucher(code) => `voucher ${code}`
  }
}

export function run(args: Array<string>): int {
  const payments = [Payment.Cash(500), Payment.Card(820, "4242"), Payment.Voucher("SPRING")]
  for (const payment of payments) {
    print(describe(payment))
  }
  return 0
}
```

```text
cash 500
card ...4242 820
voucher SPRING
```

`Payment.Card(820, "4242")` builds the `Card` variant, with the enum name in front. All three values have the type `Payment`, so they fit in one array.

`match` looks at which variant a payment is. Each line, called an arm, has a pattern on the left of `=>` and a result on the right. The pattern `Payment.Card(cents, lastDigits)` matches a card payment and binds its two fields to names for that arm. Write `_` in place of a field you do not need, as in `Payment.Cash(_)`. Chapter 5 shows everything a pattern can do.

Say the shop starts to accept gift cards, and you add a variant. Every `match` on `Payment` that does not handle it now fails to compile:

```narra error
// error: 9:10: match does not cover GiftCard(_); add the cases or a `_` arm
enum Payment {
  Cash(cents: int)
  Voucher(code: string)
  GiftCard(cents: int)
}

function isVoucher(payment: Payment): bool {
  return match (payment) {
    Payment.Cash(_) => false
    Payment.Voucher(_) => true
  }
}

export function run(args: Array<string>): int {
  if (isVoucher(Payment.Cash(1))) {
    return 1
  }
  return 0
}
```

So a new variant cannot be forgotten. You either handle it, or add an arm with the pattern `_`, which matches anything that is left.

A variant can also have no fields at all, as in `enum Size { Small, Medium, Large }`. You write such a value without parentheses: `Size.Small`.

## Data that contains itself

The shop sells gift baskets. A basket holds products, and some of them are smaller baskets. The type of a product therefore refers to itself:

```narra
import * as Int from "std/int"

enum Product {
  Single(name: string, priceCents: int)
  Basket(name: string, contents: Array<Product>)
}

function price(product: Product): int {
  return match (product) {
    Product.Single(_, priceCents) => priceCents
    Product.Basket(_, contents) => {
      let total = 0
      for (const inner of contents) {
        total = total + price(inner)
      }
      return total
    }
  }
}

export function run(args: Array<string>): int {
  const tea = Product.Single("tea", 250)
  const mug = Product.Single("mug", 600)
  const small = Product.Basket("small", [tea, mug])
  const large = Product.Basket("large", [small, tea, Product.Single("cake", 900)])
  print(Int.toString(price(small)))
  print(Int.toString(price(large)))
  return 0
}
```

```text
850
2000
```

Recursion, from chapter 3, is the natural way to walk such data. `price` of a basket is the sum of the prices of its contents, and an inner basket is handled by the same function. A single product is the stopping case.

The `Basket` arm needs several statements to work out its result, so it gets a block. In a `match` used as a value, a `return` inside an arm's block gives the value of that arm. It does not leave `price`: the `match` takes the value, and `price` returns it. In a `match` used as a statement, a `return` in a block leaves the whole function, as anywhere else. Here are blocks of the same shape used both ways:

```narra
import * as Int from "std/int"

function shippingCents(weightGrams: int): int {
  const base = match (weightGrams) {
    0 => 0
    _ => {
      if (weightGrams > 1000) {
        return 900
      }
      return 400
    }
  }
  print(`base ${Int.toString(base)}`)
  return base + 50
}

function checkWeight(weightGrams: int): bool {
  match (weightGrams) {
    0 => {
      print("nothing to ship")
      return false
    }
    _ => print("weight ok")
  }
  return true
}

export function run(args: Array<string>): int {
  print(Int.toString(shippingCents(1500)))
  assert(checkWeight(200))
  assert(!checkWeight(0))
  return 0
}
```

```text
base 900
950
weight ok
nothing to ship
```

In `shippingCents`, `return 900` sets `base`, and the function goes on to print it. In `checkWeight`, `return false` ends the function at once. `assert` stops the program with a message when its condition is false. Here both conditions hold.

A struct cannot hold a field of its own type directly. Each order points to the one before it, and a first try is an error:

```narra error
// error: 2:1: type Order contains itself; put the recursive field in an enum, Box or array
struct Order {
  id: int
  previous: Order
}

export function run(args: Array<string>): int {
  return 0
}
```

A struct value holds its fields in place, so an `Order` that holds an `Order` would need infinite room. The message lists the ways out. An array or the fields of an enum variant you declare live in separate memory, as `Basket` above does.

The first order has no previous one, so the field should be optional: an `Option<Order>`, which is `Some(order)` or `None`, as in chapter 2. Unlike the enums you declare, `Option` holds its value in place (so does `Result`), so `previous: Option<Order>` gives the same error. For a single value, use a `Box`: it holds one value in separate memory, and the struct holds only a small reference to it. `Boxes.new(value)` makes a box, and `box.value` reads what is inside:

```narra
import * as Int from "std/int"
import * as Boxes from "std/box"

struct Order {
  id: int
  previous: Option<Box<Order>> = None
}

function countOrders(order: Order): int {
  return match (order.previous) {
    Some(box) => 1 + countOrders(box.value)
    None => 1
  }
}

export function run(args: Array<string>): int {
  const first = Order { id: 1 }
  const second = Order { id: 2, previous: Some(Boxes.new(first)) }
  const third = Order { id: 3, previous: Some(Boxes.new(second)) }
  print(Int.toString(countOrders(third)))
  return 0
}
```

```text
3
```

`previous` is either `Some(box)` or `None`, and it defaults to `None` for the first order. `Some` and `None`, and the `Ok` and `Err` of `Result`, are written without an enum name in front. Every other variant needs it.

The memory behind a box, an array or an enum variant is freed as soon as nothing uses it any more. You never free it yourself. Chapter 9 explains how that works.

## Structs and enums for any type

`Option<Order>` and `Box<Order>` take a type in angle brackets, and they work the same whatever that type is. Your own structs and enums can do this too. The shop counts its stock of items, but also of things that are not items, such as paper bags. A lookup can find a price or a name. Give the declaration a type parameter, as the generic functions of [chapter 3](03-functions-and-closures.md#one-function-for-many-types) have:

```narra
import * as Int from "std/int"

struct Item {
  name: string
  priceCents: int
}

struct Stock<T> {
  thing: T
  count: int = 0
}

enum Lookup<T> {
  Found(value: T)
  Missing(name: string)
}

function findPrice(items: Array<Item>, name: string): Lookup<int> {
  for (const item of items) {
    if (item.name == name) {
      return Lookup.Found(item.priceCents)
    }
  }
  return Lookup.Missing(name)
}

export function run(args: Array<string>): int {
  const tea = Stock { thing: Item { name: "tea", priceCents: 250 }, count: 12 }
  const bags = Stock { thing: "paper bag" }
  print(`${tea.thing.name}: ${Int.toString(tea.count)}`)
  print(`${bags.thing}: ${Int.toString(bags.count)}`)

  const items = [tea.thing]
  for (const name of ["tea", "cake"]) {
    match (findPrice(items, name)) {
      Lookup.Found(cents) => print(`${name} costs ${Int.toString(cents)}`)
      Lookup.Missing(missing) => print(`no ${missing}`)
    }
  }
  return 0
}
```

```text
tea: 12
paper bag: 0
tea costs 250
no cake
```

`struct Stock<T>` declares a struct with a type parameter `T`, and the field `thing` has that type. You do not write `T` when you build a value. The compiler works it out from the fields: `tea` is a `Stock<Item>` and `bags` is a `Stock<string>`. In a type, such as a parameter or a result, you write the type argument, as in `Lookup<int>`.

`Lookup.Missing(name)` has no field of type `T`, so the variant alone does not say what `T` is. In `findPrice` the result type says it. Where nothing does, the compiler asks for a type:

```narra error
// error: 8:25: cannot tell the type of Missing here; add a type annotation
enum Lookup<T> {
  Found(value: T)
  Missing(name: string)
}

export function run(args: Array<string>): int {
  const result = Lookup.Missing("cake")
  return 0
}
```

Write `const result: Lookup<int> = Lookup.Missing("cake")`, and the program compiles.

A type parameter can also require traits of the type, as in `struct Stock<T: Show>`. [Chapter 7](07-traits.md#traits-for-a-generic-type) explains traits and shows how.

## Give a type a second name

Some types are long to write, or a short name says better what they mean. `type Name = ...` gives a type a second name, called an alias:

```narra
import * as Int from "std/int"

enum Payment {
  Cash(cents: int)
  Voucher(code: string)
}

type Cart = Array<string>
type Pay = Payment
type PriceRule = (cents: int) => int
type Pair<T> = #[T, T]
type Shelf<V> = Map<string, V>

function describe(payment: Pay): string {
  return match (payment) {
    Pay.Cash(cents) => `cash ${Int.toString(cents)}`
    Pay.Voucher(code) => `voucher ${code}`
  }
}

function cheaper(prices: Pair<int>, rule: PriceRule): int {
  const #[first, second] = prices
  return rule(first < second ? first : second)
}

export function run(args: Array<string>): int {
  const cart: Cart = ["tea", "cake"]
  const payment: Payment = Pay.Cash(500)
  let shelf: Shelf<int> = Map {}
  shelf["tea"] = 12
  print(describe(payment))
  print(describe(Payment.Voucher("SPRING")))
  print(Int.toString(cheaper(#[250, 320], (cents) => cents - 50)))
  print(cart[1])
  if (shelf["tea"] is Some(count)) {
    print(Int.toString(count))
  }
  return 0
}
```

```text
cash 500
voucher SPRING
200
cake
12
```

An alias is the same type as the one it names. A `Cart` is an `Array<string>`, and a `Payment` can go where a `Pay` is expected and back. An alias of an enum also gives its variants: `Pay.Cash(500)` builds a `Payment`, and `Pay.Cash(cents)` works as a pattern.

An alias can take type parameters. `Pair<T>` is a tuple of two values of the same type, and `Shelf<V>` is a map from names to `V`. Every use writes the type arguments, so `Pair` alone is an error:

```narra error
// error: 5:17: type alias Pair takes 1 type arguments, found 0
type Pair<T> = #[T, T]

export function run(args: Array<string>): int {
  const prices: Pair = #[1, 2]
  return 0
}
```

An alias cannot refer to itself. `type Tree = Array<Tree>` is an error that says a recursive type needs an enum, as in [Data that contains itself](#data-that-contains-itself).

## Name the type of a value

Sometimes a type should follow a value you already have. `typeof name` in a type stands for the type of `name`, which can be a `const`, a `let`, a parameter or a function:

```narra
import * as Int from "std/int"

const SHOP_PRICES: Map<string, int> = Map { "tea": 250, "cake": 400 }

function priceOf(prices: typeof SHOP_PRICES, name: string): int {
  if (prices[name] is Some(cents)) {
    return cents
  }
  return 0
}

function halfPrice(cents: int): int {
  return cents / 2
}

function applyTwice(change: typeof halfPrice, cents: int): int {
  return change(change(cents))
}

export function run(args: Array<string>): int {
  let sale: typeof SHOP_PRICES = copy(SHOP_PRICES)
  sale["tea"] = applyTwice(halfPrice, priceOf(SHOP_PRICES, "tea"))
  print(Int.toString(priceOf(sale, "tea")))
  print(Int.toString(priceOf(sale, "cake")))
  return 0
}
```

```text
62
400
```

`typeof SHOP_PRICES` is `Map<string, int>`, and `typeof halfPrice` is the function type `(cents: int) => int`. If you later change the type of `SHOP_PRICES`, `priceOf` and `sale` follow it. Only a name can come after `typeof`, not a field such as `item.name` or any other expression.

A top-level constant written without a type has the type of its value. The compiler does not know that type yet when it reads the signatures of the module. So `typeof` of such a constant in a signature, a field or an alias is an error:

```narra error
// error: 4:31: the type of constant "STANDARD_DISCOUNT" is not known here; write it as `const STANDARD_DISCOUNT: Type = ...`
const STANDARD_DISCOUNT = 50

function applyDiscount(cents: typeof STANDARD_DISCOUNT): int {
  return cents
}

export function run(args: Array<string>): int {
  return applyDiscount(STANDARD_DISCOUNT)
}
```

Inside a function body, `typeof STANDARD_DISCOUNT` works.

## Keep similar numbers apart

Customer ids and order ids are both numbers, and nothing stops you from passing one where the other belongs. A distinct type is a new type with the same values as an existing one, which the compiler keeps apart from it:

```narra
import * as Int from "std/int"

distinct type CustomerId = int
distinct type OrderId = int

function describeCustomer(id: CustomerId): string {
  return `customer ${Int.toString(CustomerId.toInt(id))}`
}

export function run(args: Array<string>): int {
  const customer = CustomerId.from(7)
  print(describeCustomer(customer))
  if (customer == CustomerId.from(7)) {
    print("same customer")
  }
  return 0
}
```

```text
customer 7
same customer
```

`CustomerId.from(7)` makes a `CustomerId` from an `int`, and `CustomerId.toInt` turns it back. The conversions cost nothing when the program runs. A `CustomerId` can be compared with `==`, but it has no arithmetic, since adding two ids means nothing. Pass an `OrderId` by mistake, and the program does not compile:

```narra error
// error: 11:27: expected CustomerId, found OrderId
distinct type CustomerId = int
distinct type OrderId = int

function describeCustomer(id: CustomerId): int {
  return CustomerId.toInt(id)
}

export function run(args: Array<string>): int {
  const order = OrderId.from(42)
  return describeCustomer(order)
}
```

A distinct type is not an alias. `type Id = int` would make `Id` and `int` the same type, and the compiler would not keep them apart. A distinct type can also be ordered, printed and given operations of its own, as [chapter 7](07-traits.md#traits-on-a-distinct-type) shows.

## Hide the fields of a struct

Some structs have a rule their fields must follow. A time span, for example, must not end before it starts. If any code could build one with `Span { start: 5, end: 1 }`, nothing would enforce the rule. Mark the struct `opaque` when you export it, and code outside its module can no longer build it or read its fields. It must go through the functions the module exports. A program cannot show this in a single file, so here is the module, `src/span.nra`:

```narra file timeline src/span.nra
export opaque struct Span {
  start: int
  end: int
}

export function makeSpan(start: int, end: int): Result<Span, string> uses alloc {
  if (end < start) {
    return Err("the span ends before it starts")
  }
  return Ok(Span { start: start, end: end })
}

export function spanLength(span: Span): int {
  return span.end - span.start
}
```

`makeSpan` returns a `Result`, which is either `Ok` with the span or `Err` with a message. Chapter 6 covers it. An exported function that builds such a value writes `uses alloc` after its result type, because it allocates memory. Chapter 10 explains `uses`.

Another module, here `src/main.nra` of a package named `timeline`, imports the functions and goes through them:

```narra file timeline src/main.nra
import * as Int from "std/int"
import { makeSpan, spanLength } from "timeline/span"

export function run(args: Array<string>): int {
  match (makeSpan(2, 9)) {
    Ok(span) => print(Int.toString(spanLength(span)))
    Err(message) => print(message)
  }
  match (makeSpan(5, 1)) {
    Ok(span) => print(Int.toString(spanLength(span)))
    Err(message) => print(message)
  }
  return 0
}
```

```text
7
the span ends before it starts
```

It prints `7`, then `the span ends before it starts`. Writing `Span { start: 5, end: 1 }` or `span.start` in `main.nra` is an error. The message says that `Span` is an opaque struct of module `"timeline/span"`, and to use that module's functions. Every `Span` comes from `makeSpan`, so every `Span` follows the rule.

A distinct type can be `opaque` too. Other modules can then read its value with `toInt`, but only the declaring module can call `from`. Receipt numbers start at 1 and go up by one, and no other module should make up a number. Here is `src/numbers.nra` of a package named `receipts`:

```narra file receipts src/numbers.nra
export opaque distinct type ReceiptNumber = int

export function firstReceipt(): ReceiptNumber {
  return ReceiptNumber.from(1)
}

export function nextReceipt(number: ReceiptNumber): ReceiptNumber {
  return ReceiptNumber.from(ReceiptNumber.toInt(number) + 1)
}
```

`src/main.nra` gets its numbers from those two functions:

```narra file receipts src/main.nra
import * as Int from "std/int"
import { ReceiptNumber, firstReceipt, nextReceipt } from "receipts/numbers"

export function run(args: Array<string>): int {
  const first = firstReceipt()
  const second = nextReceipt(first)
  print(Int.toString(ReceiptNumber.toInt(second)))
  if (first != second) {
    print("two receipts")
  }
  return 0
}
```

```text
2
two receipts
```

`ReceiptNumber.from(5)` in `main.nra` is an error: `ReceiptNumber.from is not available outside module "receipts/numbers": ReceiptNumber is an opaque distinct type`.

## Export a name from another module

Say the shop's payment code lives in `src/payments.nra`. You want the rest of the program to import everything about the till from one module, `src/till.nra`. Narra has no statement that passes on another module's export as it is. Instead, the module imports the name and declares its own export with it: `export type` for a type and `export const` for a function. Here is `src/payments.nra` of a package named `shop`:

```narra file shop src/payments.nra
import * as Int from "std/int"

export enum Payment {
  Cash(cents: int)
  Voucher(code: string)
}

export function isVoucher(payment: Payment): bool {
  return match (payment) {
    Payment.Cash(_) => false
    Payment.Voucher(_) => true
  }
}

export function describe(payment: Payment): string uses alloc {
  return match (payment) {
    Payment.Cash(cents) => `cash ${Int.toString(cents)}`
    Payment.Voucher(code) => `voucher ${code}`
  }
}
```

`src/till.nra` exports all three names again:

```narra file shop src/till.nra
import * as Payments from "shop/payments"

export type Payment = Payments.Payment
export const isVoucher = Payments.isVoucher
export const describe = Payments.describe
```

`export type Payment = Payments.Payment` is an alias, so it is the same type, variants included. `export const isVoucher = Payments.isVoucher` is a constant that holds the function, and it keeps what the function declares: `describe` still `uses alloc`, so an exported function that calls it declares `uses alloc` too. `src/main.nra` imports all three from `"shop/till"`:

```narra file shop src/main.nra
import { Payment, isVoucher, describe } from "shop/till"

export function run(args: Array<string>): int {
  for (const payment of [Payment.Cash(500), Payment.Voucher("SPRING")]) {
    print(describe(payment))
    if (isVoucher(payment)) {
      print("no change to give")
    }
  }
  return 0
}
```

```text
cash 500
voucher SPRING
no change to give
```

[Chapter 8](08-modules-and-packages.md) covers modules and imports in full.

The next chapter looks closer at `match` and the patterns it uses, and at the other places patterns appear.

Next: [5. Pattern matching](05-pattern-matching.md).
