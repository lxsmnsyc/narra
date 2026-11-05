# 7. Traits

A trait is a named set of functions that a type implements. A trait has functions only, and a type lists all its traits where the type is declared.

## Declaring and implementing

A trait body lists signatures. `Self` is the implementing type. A struct, enum or distinct type names its traits after `implements`, and each one needs an `implement Trait for Type` block in the same module.

```narra
import * as Float from "std/float"

trait Area {
  function area(shape: Self): f64
  function name(shape: Self): string
}

struct Square implements Area {
  side: f64
}

struct Rect implements Area {
  width: f64
  height: f64
}

implement Area for Square {
  function area(shape: Square): f64 {
    return shape.side * shape.side
  }

  function name(shape: Square): string {
    return "square"
  }
}

implement Area for Rect {
  function area(shape: Rect): f64 {
    return shape.width * shape.height
  }

  function name(shape: Rect): string {
    return "rect"
  }
}

export function run(args: Array<string>): int {
  const square = Square { side: 3.0 }
  const rect = Rect { width: 2.0, height: 4.5 }
  print(`${Area.name(square)}: ${Float.toString(Area.area(square))}`)
  print(`${Area.name(rect)}: ${Float.toString(Area.area(rect))}`)
  return 0
}
```

```text
square: 9.0
rect: 9.0
```

Two rules to remember:

- **You call a trait function through the trait name.** `Area.area(square)`, not `square.area()`. There is no method call syntax.
- **A trait has no fields, no default bodies, no associated types and no supertraits.** It lists functions and nothing else.

Listing a trait without implementing it is an error:

```narra error
// error: Money implements Show but has no `implement Show for Money` block or `derives Show`
struct Money implements Show {
  cents: int
}

export function run(args: Array<string>): int {
  return 0
}
```

## Implementations live with the type

Only the module that declares a type can implement a trait for it. You cannot implement `Show` for `int`, or for a type from another module:

```narra error
// error: cannot implement Show for int: the type is declared in another module
implement Show for int {
  function show(value: int): string {
    return "int"
  }
}

export function run(args: Array<string>): int {
  return 0
}
```

To give a trait to a type you do not own, wrap it in a struct or a `distinct type` (chapter 4) and implement the trait on the wrapper. The payoff is that a type's traits are always found in one place, so there are no conflicting implementations and no import that changes behavior.

## Constraints on generics

A type parameter can require traits with `<T: Trait>`, and `&` combines them. Inside the function, trait calls on a `T` are allowed.

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

struct Money implements Ordered, Show {
  cents: int
}

implement Ordered for Money {
  function compare(a: Money, b: Money): Ordering {
    return Ordered.compare(a.cents, b.cents)
  }
}

implement Show for Money {
  function show(value: Money): string {
    return `${Int.toString(value.cents)}c`
  }
}

function largest<T: Ordered & Show>(items: Array<T>, first: T): string {
  let best = first
  for (const item of Arrays.values(items)) {
    if (Ordered.compare(item, best) is Ordering.Greater) {
      best = item
    }
  }
  return Show.show(best)
}

export function run(args: Array<string>): int {
  const prices = [Money { cents: 250 }, Money { cents: 990 }, Money { cents: 40 }]
  print(largest(prices, Money { cents: 0 }))
  print(largest([3, 9, 4], 0))
  return 0
}
```

```text
990c
9
```

`int` already implements `Ordered` and `Show`, so `largest([3, 9, 4], 0)` works without any code from you. Calling `largest` with a type that lacks a required trait is an error that names the missing trait:

```narra error
// error: Money does not implement Show, which label needs for type parameter T
struct Money {
  cents: int
}

function label<T: Show>(value: T): string {
  return Show.show(value)
}

export function run(args: Array<string>): int {
  print(label(Money { cents: 1 }))
  return 0
}
```

Generic code is checked once against the constraints, not once per call. Without the bound `T: Show`, calling `Show.show(value)` on a `T` is an error.

## Standard traits and `derives`

The compiler knows five traits (the fifth, `Shape`, is not implemented yet):

| Trait | Purpose | How a type gets it |
| --- | --- | --- |
| `Equatable` | `==` and `!=` | Automatic for types without functions |
| `Hashable` | map keys, `Hashable.hash` | Automatic for types without functions or floats |
| `Ordered` | `Ordered.compare`, sorting | `derives` or `implement` |
| `Show` | text for logs, tests and errors | `derives` or `implement` |
| `Shape` | a description of fields, for serialization | `derives` only |

`derives` asks the compiler to write the implementation, field by field, in declaration order. For an enum, constructors compare in declaration order first, then their fields.

```narra
struct Point derives Show, Ordered {
  x: int
  y: int
}

enum Figure derives Show, Ordered {
  Dot(at: Point)
  Label(text: string)
  Empty
}

function describeOrder(ordering: Ordering): string {
  return match (ordering) {
    Ordering.Less => "less"
    Ordering.Equal => "equal"
    Ordering.Greater => "greater"
  }
}

export function run(args: Array<string>): int {
  const first = Point { x: 1, y: 2 }
  const second = Point { x: 1, y: 3 }
  print(Show.show(first))
  print(describeOrder(Ordered.compare(first, second)))
  print(describeOrder(Ordered.compare(second, first)))
  print(Show.show([Figure.Dot(first), Figure.Label("hi"), Figure.Empty]))
  print(describeOrder(Ordered.compare(Figure.Dot(second), Figure.Label("a"))))
  return 0
}
```

```text
Point { x: 1, y: 2 }
less
greater
[Dot(Point { x: 1, y: 2 }), Label("hi"), Empty]
less
```

A field whose type lacks the trait stops the `derives`. A struct holding a function cannot derive `Show`, for instance.

### Equality and hashing are automatic

You do not write `derives Equatable`. Any struct, enum, array, map, `Option` or `Result` whose contents can be compared supports `==` and `!=`, and also works as a map key unless a float is involved.

```narra
struct Plain {
  id: int
  tags: Array<string>
}

enum Move {
  Walk(steps: int)
  Rest
}

export function run(args: Array<string>): int {
  const left = Plain { id: 1, tags: ["x", "y"] }
  const right = Plain { id: 1, tags: ["x", "y"] }
  if (left == right && Hashable.hash(left) == Hashable.hash(right)) {
    print("equal, same hash")
  }
  if (Move.Walk(2) != Move.Walk(3) && Move.Rest == Move.Rest) {
    print("enums compare by constructor and fields")
  }
  if ([[1, 2], [3]] == [[1, 2], [3]]) {
    print("arrays compare by content")
  }
  return 0
}
```

```text
equal, same hash
enums compare by constructor and fields
arrays compare by content
```

Equality is always by value. There is no identity comparison, and `==` on a type that holds a function is a compile error. A type can replace the automatic `Equatable` or `Hashable` with its own `implement` block.

## Not yet

`Dynamic<Trait>` (values of different types behind one trait) and `derives Shape` are planned but not implemented yet, so this guide does not teach them.

Next: [Modules and packages](08-modules-and-packages.md).
