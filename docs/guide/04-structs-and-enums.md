# 4. Structs and enums

Narra has no `class`, `this`, `new` or `interface`. Data is a `struct` (a fixed set of named fields) or an `enum` (a choice between variants that carry their own fields). Behavior is a function that takes the data, or a trait (chapter 7).

## Structs

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

struct ServerConfig {
  host: string
  port: int = 8080
  tags: Array<string> = []
}

function describe(config: ServerConfig): string {
  return `${config.host}:${Int.toString(config.port)} (${Int.toString(Arrays.length(config.tags))} tags)`
}

export function run(args: Array<string>): int {
  const config = ServerConfig { host: "localhost" }
  print(describe(config))

  const loopback: ServerConfig = { host: "127.0.0.1", port: 3000 }
  print(describe(loopback))

  const moved = { ...config, port: 9000 }
  print(describe(moved))
  print(describe(config))

  let current = config
  current.port = 9001
  Arrays.push(mutable current.tags, "primary")
  print(describe(current))
  print(describe(config))

  const { host, port: serverPort } = current
  print(`${host} ${Int.toString(serverPort)}`)
  return 0
}
```

```text
localhost:8080 (0 tags)
127.0.0.1:3000 (0 tags)
localhost:9000 (0 tags)
localhost:8080 (0 tags)
localhost:9001 (1 tags)
localhost:8080 (0 tags)
localhost 9001
```

What to notice, from [SPEC.md section 4, Structs](../../SPEC.md) and D20 and D26:

- A field may have a constant default. A field without one must be given when you build the struct.
- Write `TypeName { field: value }`. When the expected type is known, from an annotation, a parameter or a result type, you can leave the name out.
- Every field is written `name: value`. The shorthand `{ name }` is not an object literal here: in value position a `{ ... }` is a block expression (chapter 5).
- `{ ...s, field: v }` copies `s` and changes some fields. The spread comes first, and the original does not change.
- A struct is a value. `let current = config` makes a copy, and `current.port = 9001` changes only `current`. You can assign to a field only through a `let` variable or a `mutable` parameter.
- `==` compares structs field by field.
- A struct cannot contain itself. Use an `Array`, an `enum` or a `Box<T>` for recursive data.

## Enums

An enum is a set of variants, and each variant can carry named fields. This replaces the TypeScript `enum` (which only names numbers or strings), tagged unions with a `kind` field, and class hierarchies. Constructors are always written with the enum name: `Shape.Circle(1.0)`, `Shape.Dot`. The only bare constructors are `Some`, `None`, `Ok` and `Err`.

```narra
import * as Float from "std/float"

enum Shape {
  Circle(radius: f64)
  Rect(width: f64, height: f64)
  Dot
}

function area(shape: Shape): f64 {
  return match (shape) {
    Shape.Circle(radius) => 3.0 * radius * radius
    Shape.Rect(width, height) => width * height
    Shape.Dot => 0.0
  }
}

export function run(args: Array<string>): int {
  const shapes = [Shape.Circle(2.0), Shape.Rect(2.0, 3.5), Shape.Dot]
  for (const shape of shapes) {
    print(Float.toString(area(shape)))
  }
  assert(Shape.Dot == Shape.Dot)
  assert(Shape.Rect(1.0, 2.0) == Shape.Rect(1.0, 2.0))
  assert(Shape.Rect(1.0, 2.0) != Shape.Dot)
  return 0
}
```

```text
12.0
7.0
0.0
```

`match` must cover every variant, so adding a variant to the enum makes every `match` on it fail to compile until you handle it (chapter 5). Leaving one out is an error.

```narra error
// error: 8:3: match does not cover Blue; add the cases or a `_` arm
enum Color {
  Red
  Blue
}
export function run(args: Array<string>): int {
  const color = Color.Red
  match (color) { Color.Red => print("red") }
  return 0
}
```

A fielded variant lives on the heap, and the compiler counts its references. An enum value is a small handle that you copy freely (D55).

## Recursive types and `Box`

A struct cannot hold itself, and an enum cannot hold a value of its own type inline. A fielded enum variant can: the payload is on the heap, so a tree or a list is a recursive enum.

```narra
import * as Int from "std/int"

enum Expression {
  Number(value: int)
  Add(left: Expression, right: Expression)
  Multiply(left: Expression, right: Expression)
  Negate(inner: Expression)
}

function evaluate(expression: Expression): int {
  return match (expression) {
    Expression.Number(value) => value
    Expression.Add(left, right) => evaluate(left) + evaluate(right)
    Expression.Multiply(left, right) => evaluate(left) * evaluate(right)
    Expression.Negate(inner) => -evaluate(inner)
  }
}

function render(expression: Expression): string {
  return match (expression) {
    Expression.Number(value) => Int.toString(value)
    Expression.Add(left, right) => `(${render(left)} + ${render(right)})`
    Expression.Multiply(left, right) => `(${render(left)} * ${render(right)})`
    Expression.Negate(inner) => `-${render(inner)}`
  }
}

export function run(args: Array<string>): int {
  const expression = Expression.Add(
    Expression.Multiply(Expression.Number(2), Expression.Number(3)),
    Expression.Negate(Expression.Number(4)),
  )
  print(`${render(expression)} = ${Int.toString(evaluate(expression))}`)
  return 0
}
```

```text
((2 * 3) + -4) = 2
```

`Box<T>` is for a struct that must refer to itself. It holds one value on the heap with a reference count: `Boxes.new(value)` makes one and `box.value` reads it.

```narra
import * as Int from "std/int"
import * as Boxes from "std/box"

struct Node {
  value: int
  next: Option<Box<Node>> = None
}

function sum(head: Option<Box<Node>>): int {
  let total = 0
  let current = head
  while (true) {
    if (current is Some(box)) {
      total = total + box.value.value
      current = box.value.next
    } else {
      return total
    }
  }
  return total
}

export function run(args: Array<string>): int {
  const list = Some(Boxes.new(Node { value: 1, next: Some(Boxes.new(Node { value: 2 })) }))
  print(Int.toString(sum(list)))
  return 0
}
```

```text
3
```

## Opaque and distinct types

A `distinct type` makes a new type from an existing one, so the compiler keeps `UserId` and `OrderId` apart even though both are `int` (D40, D62). You convert by name: `UserId.from(42)` and `UserId.toInt(id)`. Arithmetic and `Show` do not carry over, but `==` does when the base type supports it.

```narra
import * as Int from "std/int"

distinct type UserId = int
distinct type OrderId = int

function lookup(id: UserId): int {
  return UserId.toInt(id) + 1
}

export function run(args: Array<string>): int {
  const user = UserId.from(42)
  const order = OrderId.from(42)
  print(Int.toString(lookup(user)))
  assert(user == UserId.from(42))
  print(Int.toString(OrderId.toInt(order)))
  return 0
}
```

```text
43
42
```

Passing an `OrderId` where a `UserId` is wanted is a compile error.

An `opaque struct` is exported without its fields, so only its own module can build it or read its fields (D39). That is how a module keeps an invariant. This needs two modules, so it is a fragment. In `src/range.nra`:

```narra fragment
export opaque struct Span {
  start: int
  end: int
}

export function makeSpan(start: int, end: int): Result<Span, string> {
  if (end < start) {
    return Err("end before start")
  }
  return Ok(Span { start: start, end: end })
}

export function spanLength(span: Span): int {
  return span.end - span.start
}
```

Another module imports `Span`, `makeSpan` and `spanLength`, and it cannot write `Span { ... }` or read `span.start`. A distinct type can be exported opaque too (`export opaque distinct type`). See [SPEC.md section 4](../../SPEC.md).

Next: [5. Pattern matching](05-pattern-matching.md).
