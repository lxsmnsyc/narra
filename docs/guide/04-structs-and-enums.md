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
  let moreTea = tea
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

A struct behaves like the arrays of chapter 2: `let moreTea = tea` makes a copy, and changing `moreTea` leaves `tea` as it was. You can assign a field only through a `let` name.

Often you want a new value that differs in one or two fields. `{ ...tea, priceCents: 200 }` copies every field of `tea`, then sets the fields listed after it. It is a new value, and `tea` does not change.

`==` and `!=` compare two structs field by field.

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
  return isVoucher(Payment.Cash(1)) ? 1 : 0
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

Recursion, from chapter 3, is the natural way to walk such data. `price` of a basket is the sum of the prices of its contents, and an inner basket is handled by the same function. A single product is the stopping case. An arm whose result needs several statements gets a block, which ends with `return`.

A struct cannot hold a field of its own type directly. An order that points to the previous order is an error:

```narra error
// error: 2:1: type Order contains itself; put the recursive field in an enum, Box or array
struct Order {
  id: int
  previous: Option<Order>
}

export function run(args: Array<string>): int {
  return 0
}
```

A struct value holds its fields in place, so an `Order` that holds an `Order` would need infinite room. The message lists the ways out. An array or an enum variant with fields keeps its contents in separate memory, as `Basket` above does. For a single value, use a `Box`: it holds one value in separate memory, and the struct holds only a small reference to it. `Boxes.new(value)` makes a box, and `box.value` reads what is inside:

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

`Option<Box<Order>>` is either `Some(box)` or `None`, the same `Option` that reading a map gives. `Option` is an enum too. Its variants and the `Ok` and `Err` of `Result` are written without the enum name in front. Every other variant needs it. The first order has no previous one, so its field defaults to `None`.

The memory behind a box, an array or an enum variant is freed as soon as nothing uses it any more. You never free it yourself. Chapter 9 explains how that works.

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

## Hide the fields of a struct

Some structs have a rule their fields must follow. A time span, for example, must not end before it starts. If any code could build one with `Span { start: 5, end: 1 }`, nothing would enforce the rule. Mark the struct `opaque` when you export it, and code outside its module can no longer build it or read its fields. It must go through the functions the module exports. A program cannot show this in a single file, so here is the module, `src/span.nra`:

```narra fragment
export opaque struct Span {
  start: int
  end: int
}

export function makeSpan(start: int, end: int): Result<Span, string> {
  if (end < start) {
    return Err("the span ends before it starts")
  }
  return Ok(Span { start: start, end: end })
}

export function spanLength(span: Span): int {
  return span.end - span.start
}
```

Another module can import `Span`, `makeSpan` and `spanLength`, but `Span { start: 5, end: 1 }` and `span.start` are errors there. Every `Span` comes from `makeSpan`, so every `Span` follows the rule. `makeSpan` returns a `Result`, which is either `Ok` with the span or `Err` with a message. Chapter 6 covers it.

The next chapter looks closer at `match` and the patterns it uses, and at the other places patterns appear.

Next: [5. Pattern matching](05-pattern-matching.md).
