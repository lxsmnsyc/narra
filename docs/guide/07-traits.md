# 7. Traits

A shop sells books and songs. Both have a title and a price, and both need a line in the catalog, but they are different structs. A trait lets one function work with both. The trait names an operation, and each type supplies its own version. Generic code can then ask for "any type that has this operation".

This chapter builds the catalog, then uses the traits that every type can have: printing, ordering and equality.

## Giving two types the same operation

Here are the two structs, and a trait `Label` with one function. `Self` in the trait stands for whichever type implements it:

```narra
import * as Int from "std/int"

trait Label {
  function label(item: Self): string
}

struct Book implements Label {
  title: string
  author: string
  cents: int
}

struct Song implements Label {
  title: string
  seconds: int
  cents: int
}

implement Label for Book {
  function label(item: Book): string {
    return `${item.title} by ${item.author}`
  }
}

implement Label for Song {
  function label(item: Song): string {
    return `${item.title} (${Int.toString(item.seconds)} s)`
  }
}

export function run(args: Array<string>): int {
  const book = Book { title: "Dune", author: "Herbert", cents: 990 }
  const song = Song { title: "Blue", seconds: 215, cents: 129 }
  print(Label.label(book))
  print(Label.label(song))
  return 0
}
```

```text
Dune by Herbert
Blue (215 s)
```

Three pieces work together:

1. `trait Label { ... }` lists the functions a type must provide, with their signatures and no bodies.
2. `struct Book implements Label` says, where `Book` is declared, that it has this trait. Someone reading the struct sees all of its traits in one place.
3. `implement Label for Book { ... }` writes the functions for `Book`, with `Self` replaced by `Book`.

You call the function through the trait, as `Label.label(book)`. The compiler looks at the type of the argument and picks `Book`'s version.

If a struct lists a trait but you forget the `implement` block, the compiler tells you which one is missing:

```narra error
// error: 6:24: Book implements Label but has no `implement Label for Book` block
trait Label {
  function label(item: Self): string
}

struct Book implements Label {
  title: string
}

export function run(args: Array<string>): int {
  return 0
}
```

## One function for every labeled type

So far `Label.label` saves nothing over two plain functions. The gain comes in a function that works for any type with a label. A type parameter with a constraint, `<T: Label>`, says "any type `T` that implements `Label`":

```narra
import * as Int from "std/int"

trait Label {
  function label(item: Self): string
}

struct Book implements Label {
  title: string
  author: string
  cents: int
}

struct Song implements Label {
  title: string
  seconds: int
  cents: int
}

implement Label for Book {
  function label(item: Book): string {
    return `${item.title} by ${item.author}`
  }
}

implement Label for Song {
  function label(item: Song): string {
    return `${item.title} (${Int.toString(item.seconds)} s)`
  }
}

function printCatalog<T: Label>(heading: string, items: Array<T>) {
  print(heading)
  for (const item of items) {
    print(`  ${Label.label(item)}`)
  }
}

export function run(args: Array<string>): int {
  printCatalog("Books", [Book { title: "Dune", author: "Herbert", cents: 990 }])
  printCatalog("Songs", [
    Song { title: "Blue", seconds: 215, cents: 129 },
    Song { title: "Red", seconds: 180, cents: 99 },
  ])
  return 0
}
```

```text
Books
  Dune by Herbert
Songs
  Blue (215 s)
  Red (180 s)
```

`printCatalog` is written once. Each call fills in `T` from the array it gets, `Book` or `Song`, and `Label.label(item)` uses that type's version.

The constraint is what allows `Label.label(item)` inside the function. Without it, `T` could be any type at all, so the compiler refuses the call and says what to add:

```narra error
// error: 8:22: Label.label needs T to implement Label; declare it as <T: Label>
trait Label {
  function label(item: Self): string
}

function printCatalog<T>(heading: string, items: Array<T>) {
  for (const item of items) {
    print(Label.label(item))
  }
}

export function run(args: Array<string>): int {
  return 0
}
```

The constraint also protects the callers. Passing an array of a type without `Label` is an error at the call:

```narra error
// error: Ticket does not implement Label, which printCatalog needs for type parameter T
trait Label {
  function label(item: Self): string
}

struct Ticket {
  seat: int
}

function printCatalog<T: Label>(heading: string, items: Array<T>) {
  for (const item of items) {
    print(Label.label(item))
  }
}

export function run(args: Array<string>): int {
  printCatalog("Tickets", [Ticket { seat: 4 }])
  return 0
}
```

## Printing any value with `Show`

Writing a label for every type is work. Often the text only needs to show what is inside a value, such as in a log line or a test failure. For that, Narra has a standard trait, `Show`, and the compiler can write it for you. Add `derives Show` to the declaration:

```narra
struct Book derives Show {
  title: string
  cents: int
}

enum Format derives Show {
  Paper
  Ebook(megabytes: int)
}

export function run(args: Array<string>): int {
  const book = Book { title: "Dune", cents: 990 }
  print(Show.show(book))
  print(Show.show([Format.Paper, Format.Ebook(3)]))
  print(Show.show(Some(42)))
  return 0
}
```

```text
Book { title: "Dune", cents: 990 }
[Paper, Ebook(3)]
Some(42)
```

`derives Show` writes the `implement Show for Book` block, field by field. Numbers, strings, arrays, `Option` and `Result` already have `Show`, so a struct made of them can derive it. A field whose type has no `Show` stops the derive, and the error names the field.

When you want different text, leave out `derives` and write `implements Show` with your own `implement Show for Book` block, as with `Label` above. A `Show` you write yourself is used everywhere a derived one would be, including inside `Show.show` of an array of books.

## Finding the cheapest item with `Ordered`

To find the cheapest item, a function must compare two items. The standard trait for that is `Ordered`. Its one function, `Ordered.compare(a, b)`, returns an `Ordering`: `Ordering.Less`, `Ordering.Equal` or `Ordering.Greater`. Numbers and strings already implement it.

You could write `derives Ordered`, but a derived order compares the fields in the order they are declared. For `Book` that means by title first, and the shop wants price. So write the implementation yourself, and compare the cents:

```narra
struct Book implements Ordered derives Show {
  title: string
  cents: int
}

implement Ordered for Book {
  function compare(a: Book, b: Book): Ordering {
    return Ordered.compare(a.cents, b.cents)
  }
}

function cheapest<T: Ordered>(items: Array<T>): Option<T> {
  let best: Option<T> = None
  for (const item of items) {
    if (best is Some(current)) {
      if (Ordered.compare(item, current) is Ordering.Less) {
        best = Some(item)
      }
    } else {
      best = Some(item)
    }
  }
  return best
}

function report<T: Ordered & Show>(items: Array<T>): string {
  const found = cheapest(items)
  if (found is Some(item)) {
    return Show.show(item)
  }
  return "nothing for sale"
}

export function run(args: Array<string>): int {
  print(report([Book { title: "Dune", cents: 990 }, Book { title: "Emma", cents: 450 }]))
  print(report([7, 3, 9]))
  const none: Array<Book> = []
  print(report(none))
  return 0
}
```

```text
Book { title: "Emma", cents: 450 }
3
nothing for sale
```

`cheapest` returns an `Option<T>`, because an empty array has no cheapest item (chapter 6). `report` needs two traits, so its constraint joins them with `&`: `T: Ordered & Show`. The same `report` works on books and on plain numbers, since `int` has both traits.

A struct can list traits with `implements` and with `derives` at once, as `Book` does here.

## Equality comes free

`==` and `!=` compare two values of the same type, field by field. You do not write anything for this: every struct, enum, array, `Option` and `Result` has it as long as everything inside can be compared.

```narra
struct Book {
  title: string
  tags: Array<string>
}

export function run(args: Array<string>): int {
  const first = Book { title: "Dune", tags: ["sf"] }
  const second = Book { title: "Dune", tags: ["sf"] }
  if (first == second) {
    print("same book")
  }
  if (Some(first) != None) {
    print("an option of a book compares too")
  }
  let stock: Map<Book, int> = Map {}
  stock[first] = 3
  if (stock[second] is Some(count)) {
    print("the second copy finds the same entry")
  }
  return 0
}
```

```text
same book
an option of a book compares too
the second copy finds the same entry
```

Two values are equal when their contents are equal. There is no separate test for "the same object", because a Narra value has no identity apart from its contents (chapter 9 explains why). The same holds for map keys: `second` finds the entry stored under `first`. Behind this are two traits, `Equatable` for `==` and `Hashable` for map keys, which the compiler provides for you.

A function cannot be compared, so a type that holds one has no `==`:

```narra error
// error: 8:10: cannot use == on Button: it contains a function
struct Button {
  onClick: () => int
}

export function run(args: Array<string>): int {
  const ok = Button { onClick: () => 1 }
  if (ok == ok) {
    return 1
  }
  return 0
}
```

Compare the fields you care about instead, such as `first.title == second.title`.

## Traits for a type you did not declare

Only the module that declares a type can implement a trait for it. You cannot add `Show` to `int`, for instance:

```narra error
// error: 2:1: cannot implement Show for int: the type is declared in another module; wrap it in a struct
implement Show for int {
  function show(value: int): string {
    return "int"
  }
}

export function run(args: Array<string>): int {
  return 0
}
```

This keeps each type's traits in one place, next to its declaration, so no other file can change how a type prints or compares. When you want your own behavior for someone else's type, wrap it in a struct of your own and implement the trait there:

```narra
import * as Int from "std/int"

struct Price implements Show {
  cents: int
}

implement Show for Price {
  function show(price: Price): string {
    const euros = price.cents / 100
    const rest = price.cents % 100
    const padding = rest < 10 ? "0" : ""
    return `€${Int.toString(euros)}.${padding}${Int.toString(rest)}`
  }
}

export function run(args: Array<string>): int {
  print(Show.show(Price { cents: 990 }))
  print(Show.show([Price { cents: 1205 }, Price { cents: 5 }]))
  return 0
}
```

```text
€9.90
[€12.05, €0.05]
```

Next: [8. Modules and packages](08-modules-and-packages.md).
