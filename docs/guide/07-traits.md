# 7. Traits

A shop sells books and songs. Both have a title and a price, and both need a line in the catalog, but they are different structs. A trait lets one function work with both. The trait names an operation, and each type supplies its own version. Generic code can then ask for "any type that has this operation".

This chapter builds the catalog, then uses the traits that every type can have: printing, ordering, equality and hashing. It also shows traits on generic and distinct types, and an array that holds values of different types. It ends with `Shape`, which saves values as JSON and loads them back.

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

A function in an `implement` block has the signature the trait gives it. If your version does not need one of the parameters, write it with a leading `_`, as in `_book`, so the compiler knows you left it unused on purpose.

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

## Traits for a generic type

The shop sells two items together at one price, such as two books. `Bundle<T>` is a generic struct, as in [chapter 4](04-structs-and-enums.md#structs-and-enums-for-any-type), and its label is built from the labels of its two items. One `implement` block covers every `Bundle`, whatever `T` is:

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

implement Label for Book {
  function label(item: Book): string {
    return `${item.title} by ${item.author}`
  }
}

struct Bundle<T: Label> implements Label {
  first: T
  second: T
  cents: int
}

implement Label for Bundle<T> {
  function label(bundle: Bundle<T>): string {
    return `${Label.label(bundle.first)} + ${Label.label(bundle.second)}`
  }
}

function printCatalog<T: Label>(heading: string, items: Array<T>) {
  print(heading)
  for (const item of items) {
    print(`  ${Label.label(item)}`)
  }
}

export function run(args: Array<string>): int {
  const dune = Book { title: "Dune", author: "Herbert", cents: 990 }
  const emma = Book { title: "Emma", author: "Austen", cents: 450 }
  const pair = Bundle { first: dune, second: emma, cents: 1200 }
  printCatalog("Bundles", [pair])
  print(Int.toString(pair.cents))
  return 0
}
```

```text
Bundles
  Dune by Herbert + Emma by Austen
1200
```

`struct Bundle<T: Label>` puts a constraint on the type parameter of a struct, as `<T: Label>` does on a function. `implement Label for Bundle<T>` writes the type with the same parameter names as its declaration. The functions in the block see the constraint, so `Label.label(bundle.first)` is allowed. A `Bundle` has `Label` itself, so `printCatalog` takes an array of bundles.

The block has no constraints of its own. Leave `: Label` out of the struct, and the call in the block is an error:

```narra error
// error: 13:23: Label.label needs T to implement Label; declare it as <T: Label>
trait Label {
  function label(item: Self): string
}

struct Bundle<T> implements Label {
  first: T
  second: T
}

implement Label for Bundle<T> {
  function label(bundle: Bundle<T>): string {
    return Label.label(bundle.first)
  }
}

export function run(args: Array<string>): int {
  return 0
}
```

A constraint on a struct applies to every use of the struct. Building a bundle of numbers is an error, because `int` has no `Label`:

```narra error
// error: 12:16: int does not implement Label, which Bundle needs for type parameter T
trait Label {
  function label(item: Self): string
}

struct Bundle<T: Label> {
  first: T
  second: T
}

export function run(args: Array<string>): int {
  const pair = Bundle { first: 1, second: 2 }
  return pair.first
}
```

The same goes for a generic function that takes a `Bundle<T>`: it declares `<T: Label>` too, and the compiler says so when it is missing.

There is one `implement` block for all instances of a generic type. `implement Label for Bundle<int>` is an error that asks you to name the type as `Bundle<T>`. A generic enum takes its traits the same way, as in `enum Choice<T: Label> implements Label` with `implement Label for Choice<T>`.

## Finding the cheapest item with `Ordered`

To find the cheapest item, a function must compare two items. The standard trait for that is `Ordered`. Its one function, `Ordered.compare(a, b)`, returns an `Ordering`: `Ordering.Less`, `Ordering.Equal` or `Ordering.Greater`. Numbers and strings already implement it.

You could write `derives Ordered`, but a derived order compares the fields in the order they are declared. For `Book` that means by title first, and the shop wants price. So write the implementation yourself, and compare the cents. `cheapest` returns one of the items it was given, so its signature says `from items` ([chapter 9](09-memory-and-mutable.md#return-part-of-a-parameter)):

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

function cheapest<T: Ordered>(items: Array<T>): Option<T> from items {
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

Chapter 4 compared two structs with `==`, field by field, without any code for it. That works for every struct, enum, array, `Option` and `Result` whose contents can be compared. Two values are equal when their contents are equal, and the same goes for map keys:

```narra
import * as Int from "std/int"

struct Book {
  title: string
  tags: Array<string>
}

function countOf<T: Equatable>(items: Array<T>, wanted: T): int {
  let count = 0
  for (const item of items) {
    if (item == wanted) {
      count = count + 1
    }
  }
  return count
}

export function run(args: Array<string>): int {
  const first = Book { title: "Dune", tags: ["sf"] }
  const second = Book { title: "Dune", tags: ["sf"] }
  let stock: Map<Book, int> = Map {}
  stock[first] = 3
  if (stock[second] is Some(count)) {
    print(`the second copy finds ${Int.toString(count)} in stock`)
  }
  print(Int.toString(countOf([first, second], second)))
  print(Int.toString(countOf(["tea", "cake", "tea"], "tea")))
  return 0
}
```

```text
the second copy finds 3 in stock
2
2
```

`second` finds the entry stored under `first`, because the two books have the same contents. There is no separate test for "the same object": a Narra value has no identity apart from its contents (chapter 9 explains why).

The trait behind `==` is `Equatable`. A generic function that compares two `T` values needs it as a constraint, as `countOf` does. Without `<T: Equatable>`, the compiler says `== on T needs Equatable; declare the type parameter as <T: Equatable>`.

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

## Map keys and `Hashable`

A map finds a key by its hash, a number worked out from the key's contents. The trait behind it is `Hashable`, and the compiler implements it for every type without a function or a float inside. A map key needs both `Equatable` and `Hashable`. A decimal number cannot be a key, because two results of a calculation that print the same can differ in the last bit:

```narra error
// error: 7:14: Weight cannot be a map key: float keys are not allowed
struct Weight {
  kilograms: f64
}

export function run(args: Array<string>): int {
  let stock: Map<Weight, int> = Map {}
  stock[Weight { kilograms: 0.5 }] = 3
  return 0
}
```

The stock count above found a book by all of its fields. Say the shop wants one count per title, whatever the price. Write both traits yourself, and use only the title in each:

```narra
import * as Int from "std/int"
import * as Hash from "std/hash"

struct Book implements Equatable, Hashable {
  title: string
  cents: int
}

implement Equatable for Book {
  function equals(a: Book, b: Book): bool {
    return a.title == b.title
  }
}

implement Hashable for Book {
  function hash(book: Book): int {
    return Hash.of(book.title)
  }
}

export function run(args: Array<string>): int {
  let stock: Map<Book, int> = Map {}
  stock[Book { title: "Dune", cents: 990 }] = 3
  if (stock[Book { title: "Dune", cents: 700 }] is Some(count)) {
    print(`Dune ${Int.toString(count)}`)
  }
  return 0
}
```

```text
Dune 3
```

`Hash.of(value)` from `std/hash` returns the hash of any `Hashable` value. Two values that are equal must have the same hash, so write the two traits from the same fields. A type that implements one of the two but not the other cannot be a map key: the error says `it implements one of Equatable and Hashable but not the other`. A generic function that calls `Hash.of` on a `T` declares `<T: Hashable>`.

## A trait function that makes a value

So far every trait function took a `Self` value. A trait function can also return `Self` without taking one. The shop shows a placeholder entry for a title it does not stock yet, and books and songs make different placeholders:

```narra
import * as Arrays from "std/array"

trait Placeholder {
  function placeholder(title: string): Self
}

struct Book implements Placeholder derives Show {
  title: string
  cents: int
}

struct Song implements Placeholder derives Show {
  title: string
  seconds: int
}

implement Placeholder for Book {
  function placeholder(title: string): Book {
    return Book { title, cents: 0 }
  }
}

implement Placeholder for Song {
  function placeholder(title: string): Song {
    return Song { title, seconds: 30 }
  }
}

function placeholders<T: Placeholder>(titles: Array<string>): Array<T> {
  let made: Array<T> = []
  for (const title of titles) {
    Arrays.push(mutable made, Placeholder.placeholder<T>(title))
  }
  return made
}

export function run(args: Array<string>): int {
  const book = Placeholder.placeholder<Book>("Dune")
  const song: Song = Placeholder.placeholder("Blue")
  print(Show.show(book))
  print(Show.show(song))
  print(Show.show(placeholders<Song>(["Red", "Green"])))
  return 0
}
```

```text
Book { title: "Dune", cents: 0 }
Song { title: "Blue", seconds: 30 }
[Song { title: "Red", seconds: 30 }, Song { title: "Green", seconds: 30 }]
```

No argument says which version to call, so the call names the type: `Placeholder.placeholder<Book>("Dune")`. When the expected type says it, as in `const song: Song = ...`, you can leave it out. In the generic `placeholders`, `<T>` passes on whichever type the caller chose. Without a type, the call is an error:

```narra error
// error: 17:39: Placeholder.placeholder has Self only in its result; name the type, as in Placeholder.placeholder<Type>(...)
trait Placeholder {
  function placeholder(title: string): Self
}

struct Book implements Placeholder {
  title: string
}

implement Placeholder for Book {
  function placeholder(title: string): Book {
    return Book { title }
  }
}

export function run(args: Array<string>): int {
  const book = Placeholder.placeholder("Dune")
  print(book.title)
  return 0
}
```

## Values of different types in one array

An array holds values of one type, so `printCatalog` above prints either books or songs. A shelf that mixes them needs a type for "any value with a `Label`". That type is `Dynamic<Label>`. `Dynamic.of(value)` wraps a value in it, and you need no import for either:

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
  const shelf: Array<Dynamic<Label>> = [
    Dynamic.of(Book { title: "Dune", author: "Herbert", cents: 990 }),
    Dynamic.of(Song { title: "Blue", seconds: 215, cents: 129 }),
    Dynamic.of(Book { title: "Emma", author: "Austen", cents: 450 }),
  ]
  for (const item of shelf) {
    print(Label.label(item))
  }
  return 0
}
```

```text
Dune by Herbert
Blue (215 s)
Emma by Austen
```

Each `Dynamic<Label>` holds a value and remembers its type. `Label.label(item)` calls the version of that type: `Book`'s for a book, `Song`'s for a song. The choice happens when the program runs, while `printCatalog` picks its version when it compiles.

`Dynamic.of` needs to know which trait to wrap the value for. So it goes where the type is written, such as a `const` with a type, a parameter or an array of `Dynamic<Label>`. A bare `const item = Dynamic.of(book)` is an error that asks for that type.

A `Dynamic` lets you use the trait and nothing else. You cannot read the fields of the value inside, since the compiler no longer knows its type:

```narra error
// error: 18:14: Dynamic<Label> has no fields
trait Label {
  function label(item: Self): string
}

struct Book implements Label {
  title: string
}

implement Label for Book {
  function label(item: Book): string {
    return item.title
  }
}

export function run(args: Array<string>): int {
  const item: Dynamic<Label> = Dynamic.of(Book { title: "Dune" })
  print(item.title)
  return 0
}
```

A `Dynamic` also has no `==` and cannot be a map key. It does not count as a type with the trait, so it does not meet a constraint such as `<T: Label>`. Loop over the values and call the trait function on each, as above. Copies of a `Dynamic` share one value, which cannot be changed through any of them.

Only some traits work with `Dynamic`. Each function of the trait must take `Self` as its first parameter and use it nowhere else. `Label` and `Show` qualify, so `Array<Dynamic<Show>>` holds anything printable. `Ordered` does not, because `compare` takes two `Self` values, and neither does `Placeholder`, which returns `Self`. `Dynamic<Ordered>` is an error that names the function at fault.

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

## Traits on a distinct type

A distinct type from [chapter 4](04-structs-and-enums.md#keep-similar-numbers-apart) is another way to wrap a type. It has `==` and works as a map key when its base type does. It has no order and no `Show` until you give it them. List the traits after the base type, as a struct lists them after its name:

```narra
import * as Int from "std/int"

trait Label {
  function label(item: Self): string
}

distinct type CustomerId = int implements Label derives Ordered, Show

implement Label for CustomerId {
  function label(id: CustomerId): string {
    return `customer #${Int.toString(CustomerId.toInt(id))}`
  }
}

export function run(args: Array<string>): int {
  const ann = CustomerId.from(12)
  const bo = CustomerId.from(7)
  print(Label.label(ann))
  print(Show.show(bo))
  if (Ordered.compare(bo, ann) is Ordering.Less) {
    print("bo signed up first")
  }
  return 0
}
```

```text
customer #12
7
bo signed up first
```

A derived trait of a distinct type is the trait of its base type. `Show.show(bo)` prints `7`, as `Show.show(7)` does, and `Ordered.compare` compares the two numbers. The base type must have the trait:

```narra error
// error: 6:40: cannot derive Show for VisitLog: its base type Visit does not implement Show
struct Visit {
  minutes: int
}

distinct type VisitLog = Visit derives Show

export function run(args: Array<string>): int {
  return 0
}
```

`==` and the hash of a distinct type always compare the base value, so you cannot write your own `implement Equatable` or `implement Hashable` for one.

## Saving and loading values

The shop must keep its books between runs, so the program writes them to text and reads them back. The trait `Shape` does the first half of the job. It converts a value to a `Value`, a tree of nulls, booleans, numbers, text, lists and records. The tree does not depend on a file format. `Shape` also converts a `Value` back to your type. A format module such as `std/json` turns the tree into text and text into a tree.

Add `derives Shape` to a struct and the compiler writes both conversions. `Json.encode` writes any value that has `Shape`, and `Json.decode<Book>` reads text as a `Book`:

```narra
import * as Json from "std/json"

struct Book derives Shape, Show {
  title: string
  author: string
  cents: int
  note: Option<string>
  tags: Array<string>
}

export function run(args: Array<string>): int {
  const book = Book {
    title: "Dune",
    author: "Herbert",
    cents: 990,
    note: None,
    tags: ["novel", "space"],
  }
  const text = Json.encode(book)
  print(text)
  const copy = Json.decode<Book>(text)
  if (copy is Ok(value)) {
    print(Show.show(value))
  }
  const sparse = Json.decode<Book>(`{"title":"Emma","author":"Austen","cents":500,"tags":[],"extra":1}`)
  if (sparse is Ok(value)) {
    print(Show.show(value.note))
  }
  return 0
}
```

```text
{"title":"Dune","author":"Herbert","cents":990,"tags":["novel","space"]}
Book { title: "Dune", author: "Herbert", cents: 990, note: None, tags: ["novel", "space"] }
None
```

A struct becomes a JSON object with one key per field, in the order the struct declares them. The rules for the other direction are these:

- An `Option` field that holds `None` is left out of the text. When the text has no such key, or has `null`, the field is `None`.
- A key that the struct does not have is ignored, as `"extra"` is above.
- A field with a default value, such as `year: int = 2000`, takes the default when the key is missing. Any other missing field is an error.

`Json.decode` returns a `Result`. When the text is not valid JSON, or does not fit the type, the `Err` holds a message. For a mismatch the message starts with the path to the bad part, from the root of the document:

```narra
import * as Json from "std/json"

struct Book derives Shape, Show {
  title: string
  cents: int
  note: Option<string>
  year: int = 2000
}

function report(text: string) {
  const result = Json.decode<Book>(text)
  if (result is Err(message)) {
    print(`error: ${message}`)
  } else {
    print(`ok: ${Json.encode(result.value)}`)
  }
}

export function run(args: Array<string>): int {
  report(`{"title":"Dune","cents":"990"}`)
  report(`{"title":"Dune"}`)
  report(`{"title":"Dune","cents":990`)
  report(`{"title":"Dune","cents":990.5}`)
  report(`[1]`)
  report(`{"title":"Dune","cents":990}`)
  return 0
}
```

```text
error: .cents: expected int, found Text
error: .cents: missing field
error: 1:28: unexpected end of input
error: .cents: expected int, found Float
error: expected Record, found List
ok: {"title":"Dune","cents":990,"year":2000}
```

The third message is a syntax error and reads `<line>:<column>: <reason>`. The others name the path to the bad part. `.cents` is the key `cents` of the top object, and `.books[1]` would be the second item of the array under `books`. The kinds in the messages (`Text`, `Float`, `Record`, `List`) are the cases of `Value`. Decoding checks types strictly: an `int` field accepts a whole number and nothing else, so `990.5` and `"990"` are errors. An `f64` field accepts both `990` and `990.5`.

An enum becomes an object with a `"type"` key that names the constructor, followed by the constructor's fields. It fails to decode when `"type"` is missing or names no constructor. Any type that has `Shape` can be a field, so structs, enums, arrays and maps with `string` keys nest as deep as you need:

```narra
import * as Json from "std/json"

enum Item derives Shape, Show {
  Book(title: string, cents: int)
  Song(title: string, seconds: int)
  Gift
}

struct Order derives Shape {
  items: Array<Item>
  prices: Map<string, int>
}

export function run(args: Array<string>): int {
  const order = Order { items: [Item.Book("Dune", 990), Item.Gift], prices: Map { "Dune": 990 } }
  print(Json.encode(order))
  const bad = Json.decode<Item>(`{"type":"Film"}`)
  if (bad is Err(message)) {
    print(message)
  }
  return 0
}
```

```text
{"items":[{"type":"Book","title":"Dune","cents":990},{"type":"Gift"}],"prices":{"Dune":990}}
.type: Item has no constructor "Film"
```

The compiler already provides `Shape` for `bool`, `int`, `f64`, `string`, `Array`, `Map<string, V>`, `Option` and tuples of two to four items, when their contents have it. A sized integer such as `i8` has no `Value` form, so a struct with such a field cannot derive `Shape`:

```narra error
// error: 2:21: cannot derive Shape for Book: field "stars" has type i8, which does not implement Shape: i8 has no Value form
struct Book derives Shape {
  title: string
  stars: i8
}

export function run(args: Array<string>): int {
  return 0
}
```

You can also work with the tree yourself. `Json.parse` turns text into a `Value` without any type. This suits a document whose shape you do not know in advance. `Json.stringify` writes a `Value` back. To call `Shape` or name a case of `Value`, import them from `std/shape`. `derives Shape` itself needs no import:

```narra
import * as Json from "std/json"
import { Shape, Value } from "std/shape"

enum Item derives Shape, Show {
  Book(title: string, cents: int)
  Song(title: string, seconds: int)
  Gift
}

export function run(args: Array<string>): int {
  const tree = Json.parse(`{"a":[1,2.5,"x",null,true]}`)
  if (tree is Ok(value)) {
    print(Json.stringify(value))
    if (value is Value.Record(_fields)) {
      print("a record")
    }
  }
  const direct = Shape.toValue(Item.Song("Blue", 215))
  print(Json.stringify(direct))
  if (Shape.fromValue<Item>(direct) is Ok(item)) {
    print(Show.show(item))
  }
  return 0
}
```

```text
{"a":[1,2.5,"x",null,true]}
a record
{"type":"Song","title":"Blue","seconds":215}
Song("Blue", 215)
```

`Json.encode(value)` is `Json.stringify(Shape.toValue(value))`, and `Json.decode<T>(text)` is `Json.parse` followed by `Shape.fromValue<T>`. [Chapter 13](13-standard-library.md#configuration-files-stdtoml) uses the same `Shape` to read TOML files into structs.

Next: [8. Modules and packages](08-modules-and-packages.md).
