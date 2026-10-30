# Narra specification (draft)

This document defines the Narra language. Each section states the rules first, then shows an example. `DECISIONS.md` explains why each rule was chosen over the alternatives.

File extension: `.nra`. Compiler command: `narra`.

## 1. Overview

- Narra uses TypeScript-style syntax with Rust-style semantics: static types, enums with fields, pattern matching, and no `null`.
- Narra compiles to native code, the same target as C, Rust and Go. The compiler emits C, and a C compiler builds the final binary.
- Narra code is a set of modules. One file is one module. There are no globals.
- Values are immutable. Safe code has no pointers or references, only values.
- A function changes a caller's variable only through a `mutable` parameter, marked at the call: `damage(mutable hero, 10)`.
- The compiler frees memory with reference counting that it inserts itself. There is no garbage collector.
- When a value has exactly one owner, an update writes to it in place instead of copying it.

```ts
import * as Float from "std/float"

export enum Shape {
  Circle(radius: f64)
  Rect(width: f64, height: f64)
}

export function area(s: Shape): f64 {
  return match (s) {
    Circle(r) => Float.PI * r * r
    Rect(w, h) => w * h
  }
}

test "rect area" {
  assert(area(Rect(2.0, 3.0)) == 6.0)
}
```

## 2. Syntax

The syntax follows TypeScript where the meaning is the same. Where Narra means something different, it uses different syntax or reports a compile error. The table under "Differences from TypeScript" lists those cases.

### Statements and layout

- Blocks use `{ }`.
- **No semicolons.** A line break ends a statement when the line could end there. A line that ends with an operator, `,`, `(`, `[` or `{` continues on the next line. A `;` is a compile error.
- `narra fmt` defines the one valid layout. CI is expected to run `narra fmt --check`.
- Comments: `//` to the end of the line, `/* ... */` for blocks, `///` before a declaration for documentation.

### Bindings

- `const x = expr` binds a name that cannot be reassigned.
- `let x = expr` binds a name that can be reassigned with `x = expr`. A `let` binding cannot be captured by a closure.
- There are no globals. Assigning to an undeclared name is a compile error.
- There are no compound assignments (`+=`, `++`). Write `x = x + 1`.
- **No shadowing inside a function.** A name cannot be declared again while an earlier binding with that name is in scope.
- **Unused names are errors.** An unused local, parameter, import or private function is a compile error. Name it `_` or start it with `_` to mark it as unused on purpose.
- **Destructuring:** `const { host, port: serverPort } = config` for structs (section 4), and `const #[quotient, remainder] = Int.divmod(a, b)` for tuples.

### Functions

- `function name(a: int, b: int): int { ... }` declares a function. A function is private to its module unless declared with `export function` (section 3). A function declared inside another function is local to that function.
- A function with no return type returns nothing.
- **Anonymous functions are arrow functions:** `(x: int): int => x + 1`, or `(x: int): int => { ... }` with a block. `function` is only used for named declarations.
- **`function*`** declares a generator and **`async function`** declares an async function (section 8). Arrow functions can be `async`: `async () => { ... }`.
- An arrow function passed as an argument takes its types from the parameter it is passed to. The types can be left out: `Array.map(prices, (p) => p * 2)`.
- **Function types** are written like arrow functions: `(item: T) => U`.
- **Parameters are constant.** A parameter cannot be reassigned or changed, unless it is declared `mutable` (section 6).
- **A call passes every argument.** Passing fewer or more arguments than the function declares is a compile error.
- **No currying and no partial application.** To fix some arguments, write an arrow function.
- **No method calls.** `xs.push(v)` is not valid. Write `Array.push(mutable xs, v)`. `value.field(args)` is only valid when `field` is a struct field that holds a function.
- **No varargs and no optional parameters.** Pass an array, or a struct with defaults: `File.open(path, OpenOptions { create: true })`.
- **Tail calls to the same function are loops.** A tail-position call from a function to itself compiles to a jump. The stack does not grow. This holds for every function. Other recursive calls use the stack, and stack overflow aborts the program.
- **Recursion needs no keyword.** The compiler finds recursive functions from the call graph of each module. `narra query recursion MODULE` lists them.

```ts
function gcd(a: int, b: int): int {
  if (b == 0) {
    return a
  }
  return gcd(b, a % b)          // tail call, compiled to a loop
}
```

### Closures

- An arrow function that captures nothing compiles to a plain C function pointer.
- An arrow function that captures values stores them in an environment. Captured heap values have their reference count incremented, and captured structs are copied.
- Closures capture only constant values (`const` bindings and parameters), so an environment never changes after it is created.
- **Function-type parameters cannot escape by default.** Inside the called function, such a parameter can only be called. It cannot be stored, returned, captured by another closure, or sent to a thread.
- **A non-escaping closure costs no allocation.** Its environment is placed on the caller's stack, and nothing is reference counted or freed.
- **`escaping`** marks a function-type parameter that may be stored or outlive the call: `function onClick(escaping handler: () => void)`. An escaping closure's environment is allocated on the heap.
- Struct fields and array elements of function type always hold escaping closures. `Thread.start`, `Async.spawn` and `Async.run` take escaping parameters.
- **`@inline`** on a function copies its body into each caller. On a higher-order function such as `Array.map`, this also places the closure's body at the call, which removes the indirect call.

```ts
@inline
export function map<T, U, effect E>(items: Array<T>, f: (item: T) => U uses E): Array<U> uses alloc, E {
  // f cannot escape, so the closure passed in needs no allocation
}

const scaled = Array.map(prices, (p) => p * rate)    // no closure allocation, no indirect call
```


### Control flow

- `if (cond) { ... } else if (cond) { ... } else { ... }`. The condition must have type `bool`. There is no truthiness: `None`, `0` and `""` are not conditions.
- `while (cond) { ... }`.
- `for (const item of iter) { ... }` loops over an iterator. `for (const #[i, item] of Array.entries(items))` gives indexes from 0.
- **An iterator is a `Generator<T>`** (section 8). `for ... of` iterates over its own copy of the generator, so a generator held in a variable is not advanced. `Array.values`, `Array.entries`, `Map.keys`, `Map.entries`, `Str.chars` and the `Range` functions return generators. A function returns its own iterator by returning a generator.
- **Counting loops use `Range`.** `Range.upTo(n)` counts from 0 to `n - 1`. `Range.between(start, end)` counts from `start` to `end - 1`. `Range.stepped(start, end, step)` takes a step.

```ts
for (const i of Range.upTo(Array.length(items))) {
  print(items[i])
}
```

- `break` and `continue` work as in TypeScript, including labels: `outer: for (...) { ... break outer }`.
- There is no C-style `for (init; cond; step)`, no `do ... while`, no `switch` and no `goto`.
- `match` replaces `switch` (section 4).

### Literals

- **Strings** use double quotes: `"text"`. They have no interpolation.
- **Template strings** use backticks: `` `Hello ${name}` ``. Each `${...}` must hold an expression of type `string`. Template strings can span several lines.
- **Characters** use single quotes: `'a'`, `'\n'`, `'\u{1F600}'`. A character literal has type `char`. `'abc'` is a compile error.
- **Integers:** `42`, `0xFF`, `0o755`, `0b1010`. `_` can separate digits: `1_000_000`.
- **Floats:** `3.14`, `1.0e-9`, `1_000.5`.
- **A literal must fit its type.** `const x: u8 = 300` is a compile error.
- `true` and `false` have type `bool`.
- A number literal with no other type information is `int` when written without a decimal point, and `f64` when written with one.
- **Arrays:** `[1, 2, 3]`. `[]` needs its type from context.
- **Maps:** `Map { "a": 1, "b": 2 }`. A key that is not a literal is written `[key]: value`.
- **Tuples:** `#[1, "a"]` builds a tuple of type `#[int, string]`. It always makes a tuple, needs no expected type and has at least two elements. `[1, 2]` is always an array.

There are no macros, no operator overloading and no function overloading.

### Operators

- **Operators are built in.** Programs cannot define operators or change their meaning.
- **Both operands have the same type.** There are no implicit conversions, so `1 + 2.0` is a compile error.

| Operators | Operand types |
| --- | --- |
| `+` `-` `*` `/` `%` | Any integer type or any float type. `/` on integers truncates toward zero |
| `+%` `-%` `*%` | Integer types, wrapping on overflow |
| `&` `\|` `^` `<<` `>>` | Integer types. `^` is xor |
| `~` | Integer types, bitwise not |
| `<` `<=` `>` `>=` | Numbers, `char` and `string` |
| `==` `!=` | Types that implement `Equatable`, compared structurally (section 4) |
| `&&` `\|\|` `!` | `bool` |

- There is no string concatenation operator. Use a template string or `Str.concat(a, b)`.
- There is no length operator. Use `Array.length(items)`, `Str.length(s)` and `Map.size(m)`.
- There is no power operator. Use `Float.pow(x, y)` or `Int.pow(x, y)`.
- `Int.floorDiv(a, b)` rounds toward negative infinity.

### Differences from TypeScript

Code that is valid TypeScript but means something different in Narra is a compile error with a fix, never a silently different meaning.

| TypeScript | Narra |
| --- | --- |
| `const` fixes the binding, but the object can still change | `const` values never change |
| `==` converts types, `===` compares strictly | `==` compares structurally with no conversion. `===` is a compile error |
| `let` can be captured by closures | `let` cannot be captured |
| `null`, `undefined` | Not present. Use `Option<T>` with `None` |
| `T[]` is an array type | Not present. Write `Array<T>` (D23) |
| `number` | Not present. Use `int`, `i32`, `f64` and the other sized types |
| `enum` holds named numbers or strings | `enum` holds variants with fields, checked by `match` |
| `'text'` is a string | `'a'` is a `char`. Strings use `"text"` |
| `for (const k in obj)` loops over keys | Not present. Use `for (const k of Map.keys(m))` |
| `import X from "m"` imports a default export | Not present. Use `import * as X from "m"` |
| `xs.push(v)`, `s.length` | Not present. Use `Array.push(mutable xs, v)`, `Str.length(s)` |
| `a + b` joins strings | Not present. Use a template string |
| `class`, `this`, `new`, `interface` | Not present. Use `struct`, `enum` and `trait` |
| `try`, `catch`, `throw` | Not present. Use `Result` (section 5) |
| Calling an async function without `await` starts it in the background | A compile error. Use `await f()`, or start a task with `Async.spawn` (section 8) |
| `Promise<T>` | Not present. `Async.spawn` returns a `Task<T>` |
| `async function*` | Not present in the first version |

## 3. Modules and imports

- **One file is one module.** Its path from the package root is its name: `src/app/text.nra` is `"app/text"`.
- **Names are private by default.** `export` makes a top-level declaration visible to other modules: `export function`, `export const`, `export type`, `export struct`, `export enum`, `export trait`.
- **Imports are explicit.** A module sees only its own declarations, the names it imports, and the prelude.
- **Prelude.** The prelude holds `Option`, `Some`, `None`, `Result`, `Ok`, `Err`, `Array`, `Map`, `Range`, the primitive types, `print`, `assert` and `panic`. Nothing else is in scope without an import.
- **Paths are package rooted.** `"app/text"` for this package, `"std/string"` for the standard library, `"pkgname/module"` for a dependency. Relative paths are not allowed.

```ts
import * as Str from "std/string"            // namespace import
import { classify, Token } from "app/text"   // named imports
```

There is no re-export statement. To expose an imported name, import it and export a new declaration:

```ts
import * as Text from "app/text"

export const classify = Text.classify        // a function
export type Token = Text.Token               // a type, with its constructors
```

Rules:

- **Imports are static.** They appear at the top of the file with a string literal path. They never run code and never load a module at runtime. `import(...)` as an expression is a compile error.
- **No cycles.** The import graph must be acyclic. The compiler reports the cycle as a list of modules.
- **No re-export statement and no default exports.** `export { x } from "m"`, `export *` and `export default` are compile errors. An exported name is always declared in the exporting file.
- **`export type A = B`** declares an alias. When `B` is an enum, its constructors are usable through the alias.
- **`export opaque struct`** and **`export distinct type`** export a type while hiding its contents (section 4).
- **No import side effects.** Top-level declarations must be functions, types, traits, `const` constants or `lazy` values (section 10). There is no top-level `let`.
- **Constructors come with their enum.** Importing `Token` makes its constructors usable. In a `match`, constructors are resolved from the type being matched. Elsewhere, write `Token.Word(x)` when the bare name is ambiguous.
- **`narra move old new`** moves a file and rewrites every import of it.

## 4. Types

### Primitive types

| Type | Meaning |
| --- | --- |
| `i8` `i16` `i32` `i64` | Signed integers |
| `u8` `u16` `u32` `u64` | Unsigned integers |
| `int` | Alias for `i64` |
| `f32` `f64` | Floats |
| `bool` | `true` or `false` |
| `char` | A Unicode scalar value |
| `string` | Immutable UTF-8 text |
| `bytes` | Immutable byte sequence |

### Composite types

- **Structs.** Declared with `struct`. The full rules are in "Structs" below.
- **Enums.** Declared with `enum`. Each constructor can carry fields.

```ts
export enum Token {
  Word(text: string)
  Number(value: int)
  Space
}
```

- **Arrays.** `Array<T>` is a growable, contiguous array on the heap. Elements are stored directly: an `Array<int>` holds raw 8-byte integers. Indexes start at 0. `items[i]` is bounds checked and panics when out of range.
- **Fixed-size arrays.** `FixedArray<T, N>` holds exactly `N` elements and is stored inline, like a struct. It needs no heap allocation. `[1.0, 2.0, 3.0, 4.0]` builds one when the expected type is a `FixedArray`. A generic function can take the length as a parameter: `function sum<const N: int>(values: FixedArray<f64, N>): f64`.
- **Vector types.** `std/simd` provides fixed-width vectors such as `F32x4`, `F64x2` and `I32x8`. Their operators work lane by lane and compile to the target's vector instructions.
- **Maps.** `Map<K, V>` is a hash map. `m[k]` returns `Option<V>`.
- **Maps keep insertion order.** Iteration visits keys in the order they were first inserted. Removing a key and inserting it again moves it to the end.
- **Float keys are not allowed.** `Map<f64, V>` and keys that contain floats are compile errors.
- **Tuples.** `#[int, string]` is a tuple type. It holds a fixed group of at least two values. `#[1, "a"]` builds one, `const #[n, s] = t` (or `let`) destructures one, and `#[0, name]` matches one in a pattern. The type is written like the literal, in every type position: `Array<#[int, string]>`, `#[#[int, int], string]`. A function that returns several values returns a tuple or a struct.
- **Functions.** `(a: int, b: string) => bool`.

### Structs

A struct is a fixed set of named fields, stored inline like a C struct.

```ts
export struct ServerConfig {
  host: string
  port: int = 8080
  tags: Array<string> = []
}

struct Entry<K, V> {
  key: K
  value: V
}
```

**Fields and defaults**

- Each field has a name and a type, one per line.
- A field can have a default value after `=`. The default must be a compile-time constant (section 10).
- A field without a default must be given every time the struct is built.

**Building a struct**

- `TypeName { field: value, ... }` builds a struct of that type. The fields can appear in any order.
- When the expected type is already known, the type name can be left out: `{ field: value }`. The type comes from an annotation, a parameter or a return type.
- Fields with defaults can be left out. All other fields must be given.
- Every field is written as `name: value`, even when the value is a variable with the same name. `{ name }` is a compile error in a literal.
- Type arguments of a generic struct are inferred from the fields: `Entry { key: "a", value: 1 }` is `Entry<string, int>`.

```ts
const config = ServerConfig { host: "localhost" }            // port is 8080, tags is []
const loopback: ServerConfig = { host: "127.0.0.1", port: 3000 }

function makeEntry(key: string, value: int): Entry<string, int> {
  return { key: key, value: value }                          // type taken from the return type
}
```

**Reading and updating**

- `s.field` reads a field.
- `{ ...s, field: value }` returns a copy of `s` with some fields changed. The spread must come first, there is only one, and the result has the same type as `s`. The original is unchanged.
- `s.field = value` changes a field in place. `s` must be a `let` variable, a `mutable` parameter, or a field of one.
- To change a caller's struct inside a function, declare the parameter `mutable` (section 6).

```ts
const moved = { ...config, port: 9000 }

let current = config
current.port = 9001
current.tags[0] = "primary"
```

**Destructuring**

- `const { field1, field2 } = s` binds each field to a variable with the same name.
- `const { field: name } = s` binds the field to a different name.
- Fields that are not listed are ignored.
- The same patterns work in `match`. A literal after `:` tests the field's value instead of binding it.

```ts
const { host, port: serverPort } = config

match (point) {
  { x: 0.0, y: 0.0 } => print("origin")
  { x: 0.0, y } => print(`on the y axis at ${F64.toString(y)}`)
  { x, y } => print(`${F64.toString(x)}, ${F64.toString(y)}`)
}
```

**Storage**

- A struct field, array element or variable holds the struct's fields directly. There is no heap allocation and no reference count for the struct itself.
- Copying a struct copies its fields. Fields stored on the heap, such as strings and arrays, have their reference count incremented.
- `==` compares structs field by field. It is a compile error when a field holds a function.
- **A struct cannot contain itself.** Use an array, an enum or `Box<T>` for recursive data.

**Boxes**

- `Box<T>` holds one value on the heap with a reference count.
- `Box.new(value)` creates one. `b.value` reads the value. `b.value = v` replaces it when `b` is a `let` variable, in place when the count is 1.
- Use a box for recursive structs and for large structs that are copied often.

```ts
struct ListNode {
  value: int
  next: Option<Box<ListNode>> = None
}

const list = ListNode { value: 1, next: Some(Box.new(ListNode { value: 2 })) }
```

### Opaque types

`export opaque struct` exports a struct's name but not its fields.

```ts
// module "app/range"
export opaque struct Range {
  start: int
  finish: int
}

export function create(start: int, finish: int): Result<Range, string> {
  if (start > finish) {
    return Err("start is after finish")
  }
  return Ok(Range { start: start, finish: finish })
}

export function length(r: Range): int {
  return r.finish - r.start
}
```

- **Other modules** can store, pass, copy and compare a `Range` with `==`. They cannot build one with a literal, read or assign its fields, spread it, or destructure it.
- **The declaring module** uses the struct like any other struct.
- **Fields are all public or all private.** There is no per-field visibility.
- Use an opaque struct when a type has a rule that every value must follow, such as `start <= finish`.

### Distinct types

`distinct type` declares a new type with the same representation as an existing one.

```ts
export distinct type UserId = int
export distinct type OrderId = int

const user = UserId.from(42)
const raw: int = UserId.toInt(user)

loadUser(OrderId.from(7))     // compile error: OrderId is not UserId
```

- **No implicit conversion.** A `UserId` is not an `int`, and an `OrderId` is not a `UserId`.
- **`T.from` and `T.toInt`** are generated. For a distinct type over another type, the second function is named after it, such as `T.toString` for a distinct `string`.
- **No operators carry over.** `user + 1` is a compile error. The declaring module defines any operations the type needs as functions.
- **No runtime cost.** A distinct type compiles to the same C type as its base.
- **`export opaque distinct type`** hides `from`, so only the declaring module can create values. Other modules can still call the unwrap function.

### Optional values

- `Option<T>` is either `Some(value)` or `None`. A plain `T` always holds a value. There is no `null` or `undefined`.
- `if (x is Some(v)) { ... }` binds `v` inside the branch. After `if (x is None) { return ... }`, `x` is known to be `Some` for the rest of the block, and `x.value` reads the value.
- `Option.withDefault(x, fallback)` returns `fallback` when `x` is `None`.
- Type syntax uses names, not symbols: `Option<T>`, `Array<T>`, `Map<K, V>`, `Result<T, E>`. There is no `T | null` or `T?`. Tuples are the one exception: `#[A, B]` (see Tuples above).

### Pattern matching

`match` checks a value against patterns. It must cover every case, or end with a `_` arm.

```ts
match (token) {
  Word(text) => print(text)
  Number(n) if (n < 0) => print("negative")
  Number(n) => print(Int.toString(n))
  Space => print(" ")
}

const label = match (status) {
  Active => "active"
  Suspended(reason) => `suspended: ${reason}`
  _ => "other"
}
```

- Each arm is `pattern => expression` or `pattern => { statements }`.
- Patterns can bind fields, test literals, and add an `if (...)` guard.
- **`match` as a value.** When every arm is a single expression, `match` produces a value.
- `match` on an `Option<T>` uses `Some(v) =>` and `None =>`.

### Generics

- Type parameters use angle brackets: `function first<T>(items: Array<T>): Option<T>` and `struct Entry<K, V> { ... }`.
- A type parameter can require traits: `<T: Ordered>`, or `<T: Show & Ordered>` for several (see "Traits" below).
- Arithmetic operators do not work on a type parameter. Pass a function or require a trait instead.
- **Structural equality.** `==` compares values structurally. The compiler implements `Equatable` for every type that has no function inside it. Using `==` on a type that contains a function is a compile error.
- **Hashing.** `Hash.of(x)` needs `Hashable`. The compiler implements it for every type that has no function or float inside it.
- Generic code that uses `==` declares `<T: Equatable>`. Code that hashes declares `<T: Hashable>`.
- **Floats follow IEEE 754 in `==`.** `NaN == NaN` is `false`. A struct that contains `NaN` is not equal to itself.

```ts
export function contains<T: Equatable>(items: Array<T>, wanted: T): bool {
  for (const item of Array.values(items)) {
    if (item == wanted) {
      return true
    }
  }
  return false
}
```

### Traits

A trait is a named set of functions that a type implements. A type's traits are all listed where the type is declared.

```ts
export trait Show {
  function show(value: Self): string
}

export struct Money implements Show, Ordered {
  cents: int
}

implement Show for Money {
  function show(value: Money): string {
    return `${Int.toString(value.cents / 100)}.${Int.toString(value.cents % 100)}`
  }
}

implement Ordered for Money {
  function compare(a: Money, b: Money): Ordering {
    return Int.compare(a.cents, b.cents)
  }
}

const label = Show.show(price)
```

**Declaring and implementing**

- `trait Name { ... }` declares a trait. Its body lists function signatures. `Self` is the implementing type.
- **A trait has functions only.** No fields, no default implementations, no associated types, and no traits that require other traits.
- **Implementations live with the type.** A struct, enum or distinct type lists its traits with `implements`. Each listed trait needs an `implement Trait for Type { ... }` block in the same module, or a `derives` entry.
- **No implementations anywhere else.** A module cannot implement a trait for a type declared in another module. To add a trait to another module's type, wrap it in a `distinct type` or a struct.
- `derives` asks the compiler to write the implementation: `struct User derives Show, Ordered, Shape { ... }`.

**Calling**

- Trait functions are called through the trait name: `Show.show(price)`. There is no `price.show()` form.
- In generic code, `<T: Show>` allows `Show.show(value)` for a `value: T`.

**Trait values**

- `Dynamic<Trait>` holds a value of any type that implements the trait. `Dynamic.of(value)` creates one when the expected type is known.
- `Array<Dynamic<Drawable>>` holds values of different types. `Drawable.draw(item, mutable canvas)` calls the implementation of each item's own type.
- A `Dynamic<Trait>` stores the value in a box, plus a pointer to the implementation.

**Standard traits**

| Trait | Purpose | Implemented |
| --- | --- | --- |
| `Equatable` | `==` and `!=` | Automatically for types without functions |
| `Hashable` | `Hash.of`, map keys | Automatically for types without functions or floats |
| `Ordered` | `compare`, sorting | By `derives` or `implement` |
| `Show` | Text for logs, tests and errors | By `derives` or `implement` |
| `Shape` | A description of fields and constructors, for serialization | By `derives` only |

- A type can replace the automatic `Equatable` or `Hashable` with its own `implement` block.
- Primitive types, `string`, `Array<T>`, `Map<K, V>`, `Option<T>` and `Result<T, E>` implement the standard traits when their contents do.

**Serialization with `Shape`**

- `derives Shape` makes the compiler emit a constant description of the type: field names, field types, defaults and enum constructors.
- Libraries such as `std/json` read the description: `Json.encode(user)` and `Json.decode(text, User)` work for any type that derives `Shape`.
- An opaque struct can derive `Shape` only for use inside its own module.

### Inference

- Types of variables are inferred. `const n = 1` is `int`.
- **Function parameters and return types are written for every named function.** Arrow functions passed as arguments are the exception. Their types come from the parameter they are passed to.
- Other modules type-check against exported signatures only. They never read function bodies.

### Numbers

- Integer overflow panics in debug builds and wraps in release builds. `+%`, `-%` and `*%` always wrap.
- Integer division or remainder by zero panics in every build.
- Conversions are explicit: `I32.from(x)`, `F64.fromInt(x)`, `Int.truncate(f)`.

### Strings

- `string` holds valid UTF-8. Functions that build a string from bytes return `Result` and fail on invalid UTF-8.
- **Positions are byte offsets of type `ByteIndex`**, starting at 0. `ByteIndex` is declared as `export opaque distinct type ByteIndex = u64`, so only `std/string` can create one.
- **Positions come from the string.** `Str.indexOf`, `Str.find` and iterators return `ByteIndex` values that always fall on a character boundary.
- **Slices share memory.** `Str.slice` and `Array.slice` return a view into the original buffer, with no copy. The view keeps the whole buffer alive. Use `Str.copy` or `Array.copy` to keep a small part of a large buffer without keeping the rest.
- **Slicing checks boundaries.** `Str.slice(s, from, to)` returns bytes `from` up to but not including `to`. It panics when a position is not on a character boundary or is out of range.
- `Str.length(s)` is the byte length as `ByteIndex`. `Str.charCount(s)` counts characters and takes time proportional to the length.
- Grapheme clusters, case mapping and normalization are in `std/unicode`.

```ts
export struct Pair {
  key: string
  value: string
}

export function parsePair(line: string): Option<Pair> {
  const found = Str.indexOf(line, "=")
  if (found is None) {
    return None
  }
  const i = found.value
  return Some(Pair {
    key: Str.trim(Str.slice(line, 0, i)),
    value: Str.trim(Str.sliceFrom(line, Str.nextChar(line, i))),
  })
}
```

## 5. Errors

- **Errors are return values.** A function that can fail returns `Result<T, E>`, with constructors `Ok(value)` and `Err(error)`.
- **The caller checks the result.** There is no `try`, no `catch`, no `throw` and no operator that returns early. Every error path is a visible `if` or `match`.
- **`is` tests a pattern and narrows the type.** After `if (r is Err(e)) { return ... }`, the compiler knows `r` is `Ok`. `r.value` then reads the value. The same narrowing works for `Option<T>` with `None`.
- **Ignoring a `Result` is a compile error.** Discard one on purpose with `const _ = f()`.
- **Panics abort the program.** `panic("message")`, failed `assert`, out-of-bounds indexing and debug-mode overflow all panic. A panic cannot be caught.

```ts
export function readConfig(path: string): Result<Config, ConfigError> uses io, alloc {
  const text = File.readText(path)
  if (text is Err(e)) {
    return Err(IoFailed(e))
  }

  const config = Config.parse(text.value)
  if (config is Err(e)) {
    return Err(e)
  }

  return Ok(config.value)
}
```

## 6. Memory model

### Values and ownership

- Every value has one or more owners: variables, fields, array slots, closure captures.
- Boxed values (enums with fields, arrays, maps, strings, closures and `Box<T>`) live on the heap with a reference count. Numbers, `bool`, `char` and structs live inline.
- A struct that holds boxed fields counts those fields, not itself. Copying the struct increments the count of each boxed field.
- **`Box<T>`** puts a value on the heap with a reference count. Use it for recursive structs and for large structs that are passed around often.
- The compiler inserts reference count increments and decrements. It removes the ones it can prove unnecessary, as in Koka's Perceus.
- A value is freed when its count reaches zero. File handles, sockets and other resources close at the point their last owner drops them.
- **Large frees can be spread out.** When a large structure is freed, the runtime may release its memory in small batches during later allocations. Resources still close immediately. Only the return of memory to the allocator is delayed.
- **Running out of memory aborts the program,** the same as a panic.

### In-place updates

- `items[i] = v`, `m[k] = v`, `Array.push(mutable items, v)` and `u.age = v` change a `let` variable. `{ ...u, age: v }` returns a changed copy.
- When the reference count of the old value is 1, the compiler writes into it instead of copying. This is called reuse.
- Code that keeps one owner gets in-place mutation with no copies.

```ts
let squares: Array<int> = Array.filled(0, 1000)
for (const i of Range.upTo(1000)) {
  squares[i] = i * i          // the count is 1, so the write is in place
}
```

- `items[i] = v` and `u.age = v` require the target to be a `let` variable, a `mutable` parameter, or a field of one.

### Copies on shared values

When an update targets a value whose reference count is greater than 1, the value is copied first. The copy is correct, but it costs time proportional to the value's size, and the source does not show it. Three tools make these copies visible:

- **`narra check --report-copies`** lists every update where the compiler cannot prove the count is 1.
- **`NARRA_TRACE_COPIES=1`** makes a program log every copy at runtime, with its source location and size.
- **`@unique`** on a `let` declares that the variable is never shared. In debug builds, an update that would copy it panics. In release builds the check is removed.

```ts
@unique let buffer: Array<u8> = Array.withCapacity(65536)
Array.push(mutable buffer, 0)        // debug builds panic here if buffer is shared
const view = buffer                  // shares buffer: the next update would panic in debug
```

### Regions

A region allocates many values from one block and frees them all at once.

```ts
import * as Region from "std/region"

const summary = Region.run((region) => {
  const tree = Parser.parse(source)        // every allocation here comes from the region
  return Summary.of(tree)                  // the result is copied out
})
// the region and everything in it has been freed
```

- **`Region.run(f)`** runs `f` with a region. Every heap allocation made while `f` runs comes from the region. Values in a region have no reference count updates.
- **Values leave a region only by copying.** The result of `f` is copied to the normal heap. A value sent to a channel, stored in a `Shared`, or returned through the result is copied when it leaves.
- **The region is freed when `f` returns,** in one step.
- **A region cannot escape,** with the same rules as a `Scope`.
- Regions suit work with a clear end, such as handling one request or compiling one file.

### Mutable parameters

A function changes a caller's variable only through a `mutable` parameter. The keyword appears in the declaration and again at the call.

```ts
function damage(mutable player: Player, amount: int) {
  player.hp = player.hp - amount
}

let hero = Player { name: "Ana", hp: 100, position: Vec2 { x: 0.0, y: 0.0 } }
damage(mutable hero, 10)      // hero.hp is now 90
```

- **The argument must be changeable.** It must be a `let` variable, a `mutable` parameter, or a field or element of one. A `const`, a literal or a function result is a compile error.
- **Access is exclusive during the call.** A variable cannot be passed as `mutable` twice in one call. It cannot be passed as `mutable` and also read in another argument. `swap(mutable a, mutable a)` and `move(mutable p, p.x)` are compile errors.
- **A `mutable` parameter cannot escape.** It cannot be stored, captured by a closure, or returned. Copying its value with `const copy = player` is allowed.
- **No copies.** The compiler passes a pointer to the caller's variable. Changes are written in place, with no copy and no reference count change for the variable itself.
- **Other parameters are values.** A parameter without `mutable` is a constant copy. To build new data instead of changing the caller's data, return a new value.

| Goal | Form |
| --- | --- |
| Change the caller's data | `damage(mutable hero, 10)` |
| Build new data | `const stronger = withBonus(hero, 5)` |

### Cycles

- Immutable values cannot form cycles, so reference counting frees everything without a cycle collector.
- A `let` variable or `mutable` parameter cannot be captured or stored, so it cannot create a cycle.
- **Graphs** use `Arena<T>` and `Handle<T>` from `std/arena`. A handle is an index plus a generation number. `Arena.get(arena, h)` returns `None` when the slot was freed or reused. Handles can point anywhere in the arena, including in cycles, and the arena frees all its slots at once.

### Values across threads

- A value captured by a thread function or sent through a channel is marked shared. Shared values use atomic reference counts. Unshared values use plain counts.
- An in-place update on a shared value copies first, so threads never see each other's writes.
- Every value can be sent to another thread. Safe code has no shared mutable memory, so no data race is possible. Section 8 covers threads, channels and the few types that are shared on purpose.

## 7. Effects

An effect marks something a function can do beyond computing a value. The set is fixed:

| Effect | Meaning |
| --- | --- |
| `alloc` | May allocate heap memory |
| `io` | May perform input or output, read the clock, or read the environment |
| `unsafe` | Calls code that the compiler cannot check (section 9) |

Rules:

- Exported functions declare their effects after the return type: `function f(s: string): int uses alloc { ... }`.
- The compiler infers effects for private functions and arrow functions.
- A function's effects include the effects of everything it calls. Calling an `io` function from a function that does not declare `io` is a compile error.
- Panics are not an effect. Any function can panic.
- The entry function (section 15) has every effect. It is not `async`. `Async.run` runs async code from it.
- `async` and `function*` are not effects. They mark the type of function a declaration is (section 8).

### Effect variables

A function that calls a function it receives has that function's effects. An effect variable states this in a signature.

```ts
export function map<T, U, effect E>(items: Array<T>, f: (item: T) => U uses E): Array<U> uses alloc, E
```

- Effect variables are declared in the type parameter list with the `effect` keyword, and appear only after `uses`.
- **They are bound by a parameter.** An effect variable in the function's own `uses` list must also appear in a parameter's function type.
- At each call, the variable takes the effects of the function passed in. `map(names, Str.length)` has `alloc`. `map(paths, File.readText)` has `alloc, io`.

## 8. Generators and concurrency

Narra has three tools. Each covers one type of concurrent work:

| Work | Tool |
| --- | --- |
| Produce values one at a time, on demand | Generators |
| Run computations in parallel on several cores | Scoped threads and channels |
| Wait on many I/O operations at once | Async functions on an event loop |

A panic on any thread or task aborts the whole program.

### Generators

A generator is a function that can suspend itself with `yield` and be resumed later. The syntax is TypeScript's. The compiler turns each generator into a state machine, so no separate stack is allocated.

```ts
function* counter(limit: int): Generator<int> {
  for (const i of Range.upTo(limit)) {
    yield i
  }
}

function* evens(limit: int): Generator<int> {
  for (const n of counter(limit)) {
    if (n % 2 == 0) {
      yield n
    }
  }
}

function* both(): Generator<int> {
  yield* counter(3)
  yield* evens(10)
}

for (const n of both()) {
  print(Int.toString(n))
}
```

- **`function*`** declares a generator. Calling it returns a `Generator<Out>` without running any of its body.
- **`yield value`** suspends the generator and produces `value`. It is valid only directly inside a `function*` body, not in a nested function.
- **`yield* other`** produces every value of another generator in turn.
- **`Coroutine<In, Out>`** is a generator that also receives a value on each resume. Inside it, `const input = yield output` gets that value.
- **`Co.resume(mutable g, input)`** from `std/coroutine` runs a generator or coroutine to its next `yield`, changing `g` in place. It returns `Yielded(value)`, `Done` or `Panicked(message)`. `Panicked` is only reported in test builds. In other builds a panic aborts the program.
- **A generator is a value like any other.** Resuming one by hand requires a `mutable` argument. `for ... of` iterates over its own copy.
- **Only generator bodies become state machines.** A `yield` cannot cross a normal function call, so other functions compile as plain C functions.

### Threads

Threads run in a scope. Every thread started in a scope finishes before the scope returns.

```ts
import * as Thread from "std/thread"

const total = Thread.scope((scope) => {
  const left = Thread.start(scope, () => sum(firstHalf))
  const right = Thread.start(scope, () => sum(secondHalf))
  return Thread.join(left) + Thread.join(right)
})
```

- **`Thread.scope(f)`** calls `f` with a `Scope`. It waits for every thread started in that scope, then returns the result of `f`.
- **`Thread.start(scope, f)`** takes `f` as an escaping parameter and runs it on a new thread and returns a `Thread<T>`. `Thread.join(t)` waits for it and returns its result.
- **A `Scope` cannot escape.** It cannot be stored, returned, or captured by a thread function. Only the function passed to `Thread.scope` can start threads in it.
- **Thread functions capture only constant values,** like every closure. They cannot be `async`.
- **`Parallel.map(items, f)`** from `std/parallel` splits an array across threads and returns the results in order.

### Channels

```ts
const #[sender, receiver] = Channel.create<int>(16)     // holds up to 16 values

Thread.start(scope, () => {
  for (const i of Range.upTo(100)) {
    Channel.send(sender, i)                            // blocks while the channel is full
  }
  Channel.close(sender)
})

for (const value of Channel.values(receiver)) {       // ends when the channel is closed and empty
  print(Int.toString(value))
}
```

- **`Channel.create<T>(capacity)`** returns a `#[Sender<T>, Receiver<T>]`.
- **`Channel.send`** blocks while the channel is full. **`Channel.receive`** blocks until a value arrives.
- **`Channel.close(sender)`** closes the channel for every copy of the sender. Dropping the last copy also closes it. `Channel.receive` returns `None` once the channel is closed and empty.
- Blocking calls have the `io` effect.

### Shared state

A few types have identity: copying one copies a handle to the same object. They are the only way two threads can see the same changing data.

| Type | Use |
| --- | --- |
| `Sender<T>`, `Receiver<T>` | Channel ends |
| `Shared<T>` | A value guarded by a lock |
| `Atomic<T>` | An integer or `bool` with atomic operations |

```ts
const hits = Shared.create(0)

Shared.update(hits, (mutable count) => {
  count = count + 1
})

const now = Shared.read(hits, (count) => count)
```

- **The lock is held only during the callback.** `Shared.update` and `Shared.read` take the lock, run the callback and release it. There is no way to take the lock and forget to release it.
- **The callback cannot be `async`,** so a task never suspends while holding a lock.
- **`Atomic<T>`** offers `Atomic.load`, `Atomic.store`, `Atomic.add` and `Atomic.compareExchange`.
- `==` on an identity type compares identity, not contents.
- **Deadlocks are possible** when callbacks nest locks in different orders. `narra audit` lists every nested `Shared.update`.

### Async functions

An event loop runs many tasks on one thread. A task suspends at each `await`, and the loop resumes it when the awaited I/O is ready. The syntax is TypeScript's.

```ts
import * as Async from "std/async"
import * as Net from "std/async/net"

async function handle(connection: Connection) uses io, alloc {
  let conn = connection
  while (true) {
    const data = await Net.read(mutable conn, 4096)
    if (data is Err(_)) {
      return
    }
    if (Bytes.length(data.value) == 0) {
      return
    }
    const written = await Net.write(mutable conn, data.value)
    if (written is Err(_)) {
      return
    }
  }
}

export function run(args: Array<string>): int {        // entry function set in narra.toml
  Async.run(async (loop) => {
    let listener = Net.listen("127.0.0.1", 8080)
    while (true) {
      const accepted = await Net.accept(mutable listener)
      if (accepted is Ok(conn)) {
        Async.spawn(loop, async () => await handle(conn))
      }
    }
  })
  return 0
}
```

- **`async function`** declares a function that can suspend. Its body compiles to a state machine.
- **`await` is required on every call to an async function.** A call without `await` is a compile error. Every point where a task can suspend is visible at the call.
- **Async functions run only inside an event loop.** They are called with `await` from other async functions, or started with `Async.run` and `Async.spawn`.
- **`Async.run(f)`** runs an event loop on the current thread until `f` and every task it started have finished. It is a normal function, so it can be called from the entry function or from a thread.
- **`Async.spawn(loop, f)`** starts a task on that loop and returns a `Task<T>`. **`await Async.join(t)`** waits for the task and returns its result.
- **`await Async.all(functions)`** runs several async functions at the same time and returns their results in order.
- **No async callbacks in normal higher-order functions.** `Array.map(items, async (x) => ...)` is a compile error. `std/async` provides `Async.map` and `Async.forEach` instead.
- **A `Loop` cannot escape,** with the same rules as a `Scope`.
- **Suspending I/O lives in `std/async/*`.** `std/async/net`, `std/async/file` and `std/async/time` mirror the blocking modules with the same function names. Their functions are `async`.
- **A task stays on the thread that runs its loop.** To use several cores, run one loop per thread inside `Thread.scope`. Pass work between the loops with channels.
- `await Async.sleep(milliseconds)` suspends the task for a time.

## 9. Unsafe code and C interop

- **`unsafe { ... }`** blocks allow raw pointers (`Ptr<T>`), pointer arithmetic and unchecked indexing. A function containing an `unsafe` block has the `unsafe` effect unless it is marked `@trusted`, which states that the author checked it by hand.
- **`extern`** declares a C function: `extern function strlen(s: Ptr<u8>): u64 uses unsafe = "strlen"`.
- **`@repr(C)`** on a struct fixes its field order and padding to match C. Without it, the compiler may reorder fields to reduce padding.
- **C-compatible enums.** `@repr(u8)`, `@repr(i32)` and the other integer types give an enum without fields explicit values. `Enum.toInt` and `Enum.fromInt` convert, and `fromInt` returns `Option`.

```ts
@repr(u8)
export enum Color {
  Red = 1
  Green = 2
  Blue = 4
}
```

- **Callbacks from C.** A function marked `@export(C, "name")` gets a C-compatible symbol. Its parameters and result must be `@repr(C)` types, numbers or `Ptr<T>`.
- **C strings.** `std/ffi` provides `CString`. `CString.from(s)` copies a string with a terminating zero byte, and `CString.toString(p)` copies back and returns `Result` on invalid UTF-8.
- **Linking.** A `[link]` table in `narra.toml` lists C libraries and search paths (section 15).
- `narra audit` lists every `unsafe` block and `@trusted` function in a package.

## 10. Top-level declarations

- **`const`** at the top level declares a compile-time constant. Inside a function, `const` accepts any expression. A top-level value must be one of these:
  - a literal or a constructor
  - a struct or array of constants
  - a named function or another constant
  - a call to an `@const` function
- **`lazy`** defers a computation until first use: `const table = lazy buildTable()`. The first `Lazy.force(table)` computes the value and later calls return it. Forcing is thread safe.
- Modules have no initialization code. Program startup runs nothing before the entry function.

## 11. Code generation

- **Output is C.** One `.c` file per module, plus a small runtime: allocator, reference counting, panic handler, arenas and regions.
- **Vector types** from `std/simd` compile to C compiler vector extensions.
- **Debug builds** use a fast C compiler such as TCC with checks on.
- **Release builds** use Clang or GCC with `-O2`.
- **Generators and async functions** compile to a struct and a resume function. The struct holds their live variables and a state number. The resume function switches on that number.
- **`mutable` parameters** compile to C pointer parameters.
- **Self-calls in tail position** compile to loops in every function (section 2). Other recursion uses the C stack. Stack overflow aborts the program.
- **Closures.** A non-escaping closure's environment is a C struct on the caller's stack. An escaping closure's environment is a reference-counted heap object. `@inline` functions are emitted as `static inline` C functions, and their bodies are copied into callers.
- **Trait constraints** compile to a hidden parameter that points to the implementation table.
- **Constant parameters pass by pointer.** A parameter larger than two machine words is passed as a pointer to the caller's value. Parameters are constant and access is exclusive (section 6), so this changes nothing a program can observe.
- **Reads do not copy.** `items[i].field` and `s.inner.field` read through a pointer. The element is not copied and its reference count is not touched.
- **Generic code uses layout tables.** A generic function that is not specialized receives one hidden table per type parameter. The table holds the type's size, its alignment, and functions to copy, drop and move a value. Values are stored and passed in their normal form, never boxed.
- **Containers store raw values.** `Array<T>` stores elements directly, using the size from the layout table. `Array<int>` holds raw integers in both generic and specialized code.
- **Specialization.**
  - Debug builds compile each generic function once, with layout tables, for fast builds.
  - Release builds compile a separate copy for each type a generic function is used with inside the same package.
  - Calls across packages use layout tables unless the function is marked `@specialize` or `@inline`.
  - `generics = "compact"` in the `[build]` table of `narra.toml` uses layout tables everywhere, for smaller binaries.
- **`@specialize`** on a function always emits a copy per concrete type, in every build.

## 12. Build model

- The compiler reads a module's `import` declarations before anything else and builds the module graph from them.
- Each module produces an interface file (`.nrai`) with its exported signatures. A module type-checks against its dependencies' interface files only.
- When a module changes but its interface file stays the same, modules that import it are not re-checked or recompiled.
- Modules with no dependency between them compile in parallel.
- Builds are deterministic. The same inputs produce byte-identical C and binaries.
- Compile-time evaluation is limited to top-level `const` declarations, `@const` functions and build conditions.

### Build conditions

Code that differs by platform branches on constants from `std/build`. The compiler keeps only the matching branch.

```ts
import * as Build from "std/build"

export function configDir(): string uses io {
  if (Build.os == Os.Windows) {
    return Env.get("APPDATA")
  } else if (Build.os == Os.MacOS) {
    return `${Env.home()}/Library/Application Support`
  } else {
    return `${Env.home()}/.config`
  }
}
```

- **`Build.os`, `Build.arch` and `Build.debug`** are compile-time constants. An `if` that tests only these constants is resolved at compile time. The output contains no runtime check.
- **Every branch is type-checked on every build.** A Windows branch is checked when building for Linux, so branches for other platforms cannot break unnoticed.
- **`@target(os = Os.Windows)`** on a declaration makes it exist only on that target. Calling it outside a branch that guarantees that target is a compile error.

```ts
@target(os = Os.Windows)
extern function GetTickCount64(): u64 uses unsafe = "GetTickCount64"
```

- A `@target` C function is linked only when building for its target.
- `narra check --all-targets` type-checks the package for every supported target.
- There are no per-platform file names and no preprocessor.

## 13. Tests and contracts

- **`test "name" { ... }`** blocks sit next to the code. `narra test` runs them. They are removed from non-test builds.
- **`std/test`** has assertions for test blocks. Its functions can be called only inside a `test` block, arrow functions inside it included; a call anywhere else is a compile error. The import itself may stay in a file that is also built as a program, because test blocks are not part of that build.
  - `Test.eq<T: Equatable & Show>(actual: T, expected: T)` panics with `expected <expected>, got <actual>` when the values differ. `Test.ne(actual, other)` panics with `expected a value other than <other>` when they are equal. Both values are written with `Show`, and the panic names the call's `file:line:col`.
  - `Test.panics(f: () => void)` calls `f` and panics unless `f` panics. `Test.panicsWith(f, message: string)` also needs the panic message to contain `message`.
  - A failed assertion is a panic: the test run stops and exits with code 101.
  - `f` runs in a child process, so nothing it changes or allocates reaches the test. Output it prints appears in order. `Test.panics` and `Test.panicsWith` therefore need a POSIX system (`fork`). A closure that runs past the timeout (10 seconds, or `NARRA_TEST_TIMEOUT`) fails.

```ts
import * as Test from "std/test"

test "parse" {
  Test.eq(Int.parse("ff", 16), Ok(255))
  Test.ne(Str.trim(" a "), " a ")
  Test.panicsWith(() => divide(1, 0), "division by zero")
}
```

- **Contracts.** `requires` and `ensures` lines between a function's signature and its body are checked in debug builds and removed in release builds. `ensures` refers to the return value as `result`.

```ts
export function sqrt(x: f64): f64
  requires x >= 0.0
  ensures result >= 0.0
{
  return Float.sqrt(x)
}
```

## 14. Tooling for agents

### Diagnostics

`narra check --json` prints one JSON object per diagnostic:

```json
{
  "code": "E0412",
  "severity": "error",
  "message": "Function `load` uses `io` but its signature does not declare it",
  "file": "src/app/config.nra",
  "range": {"start": [12, 4], "end": [12, 22]},
  "fixes": [
    {"title": "Add `io` to the signature", "edits": [{"file": "src/app/config.nra", "range": {"start": [8, 40], "end": [8, 40]}, "text": ", io"}]}
  ]
}
```

- Error codes are stable. `narra explain E0412` prints the rule and an example fix.
- Fixes are exact edits that an agent can apply without rewriting them.
- **TypeScript habits get a fix.** Each row of "Differences from TypeScript" (section 2) has a diagnostic. Its fix rewrites the TypeScript form into the Narra form. For example, `===` is rewritten to `==`, and `xs.push(v)` to `Array.push(mutable xs, v)`.

### Queries

| Command | Output |
| --- | --- |
| `narra query type-of FILE:LINE:COL` | The type and effects at a position |
| `narra query def NAME` | Where a name is defined |
| `narra query refs NAME` | Every use of an exported name |
| `narra query callers FUNCTION` | Every function that calls it |
| `narra query interface MODULE` | The module's exported signatures |
| `narra query recursion MODULE` | The module's recursive functions |

All query commands accept `--json`.

### Speed target

`narra check` on an unchanged package with one edited file returns in under 100 ms.

## 15. Packages (sketch)

- **A package is a directory** with a `narra.toml` manifest and a `src/` directory.
- **The package name is the import path prefix.** Package `json` exposes `src/parse.nra` as `"json/parse"`.
- **Dependencies are git repositories** pinned to version tags. There is no central registry.

```toml
[package]
name = "app"
version = "0.1.0"

[dependencies]
json = { git = "https://github.com/example/narra-json", version = "1.2.0" }

[[binary]]
name = "app"
entry = "app/cli.run"

[link]
libraries = ["sqlite3"]
search = ["/opt/homebrew/lib"]

[build]
generics = "compact"    # optional: layout tables everywhere, for smaller binaries
```

- **Entry points are set in `narra.toml`.** Each `[[binary]]` names the binary and its entry function as `"module.function"`. A package with no `[[binary]]` is a library.
- **The entry function** has the signature `export function run(args: Array<string>): int`. The result is the process exit code. It has every effect, and it is not `async`.
- **`[link]`** lists C libraries to link and directories to search.
- **Version selection is minimal version selection,** as in Go. Each dependency resolves to the highest version that some package in the graph requires as its minimum. The result is the same on every machine and needs no solver.
- **One version per package** in a build. A major version change is a new package name, such as `json2`.
- **`narra.lock`** records the commit hash and content hash of every dependency. `narra build` fails when a fetched dependency does not match its hash.

## 16. Deferred

These are left out of the first version on purpose.

- **Task cancellation and timeouts.** A task runs until it returns. A cancellation design for async tasks and threads is not decided.
- **Functors and module parameters.** See `DECISIONS.md`, D6.
- **A native debug backend.** See `DECISIONS.md`, D10.
- **A bytecode VM.** Possible later for scripting and hot reload. The first target is native code only.
