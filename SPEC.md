# Lang specification (draft)

This document defines the Lang language. Each section states the rules first, then shows an example. `DECISIONS.md` explains why each rule was chosen over the alternatives.

File extension: `.lang`. Compiler command: `lang`.

The previous draft used ReasonML syntax. It is kept at the git tag `reasonml-draft`.

## 1. Overview

- Lang uses Lua syntax with static types. Metatables are removed.
- Lang compiles to native code, the same target as C, Rust and Go. The compiler emits C, and a C compiler builds the final binary.
- Lang code is a set of modules. One file is one module. There are no globals.
- Values are immutable. Safe code has no pointers or references, only values.
- A function changes a caller's variable only through a `mutable` parameter, marked at the call: `damage(mutable hero, 10)`.
- The compiler frees memory with reference counting that it inserts itself. There is no garbage collector.
- When a value has exactly one owner, an update writes to it in place instead of copying it.

```lua
const Float = import("std/float")

export enum Shape
  Circle(radius: f64)
  Rect(width: f64, height: f64)
end

export function area(s: Shape): f64
  match s
    when Circle(r) then return Float.PI * r * r
    when Rect(w, h) then return w * h
  end
end

test "rect area" do
  assert(area(Rect(2.0, 3.0)) == 6.0)
end
```

## 2. Syntax

The syntax follows Lua 5.4. This section lists only the differences and additions.

### Bindings

- `const x = expr` binds a name that cannot be reassigned.
- `let x = expr` binds a name that can be reassigned with `x = expr`. A `let` binding cannot be captured by a closure.
- `local` is not a keyword. Every binding is local to its block or module.
- There are no globals. Assigning to an undeclared name is a compile error.
- Multiple assignment is allowed: `const a, b = f()` when `f` returns two values.
- **No shadowing inside a function.** A name cannot be declared again while an earlier binding with that name is in scope.
- **Unused names are errors.** An unused local, parameter, import or private function is a compile error. Name it `_` or start it with `_` to mark it as unused on purpose.

### Functions

- `function name(a: int, b: int): int ... end` declares a function. A function is private to its module unless declared with `export function` (section 3). A function declared inside another function is local to that function.
- Anonymous functions use the same form without a name: `function(x: int): int return x + 1 end`.
- **Recursion is declared.** A function that calls itself, directly or through other functions in the same module, is declared `recursive function`. A call that forms a cycle between functions that are not marked `recursive` is a compile error.
- **Tail calls in recursive functions are loops.** In a `recursive function`, a tail-position call to itself compiles to a jump. The stack does not grow. Other recursive calls use the stack. Stack overflow aborts the program.

```lua
recursive function gcd(a: int, b: int): int
  if b == 0 then return a end
  return gcd(b, a % b)          -- tail call, compiled to a loop
end

recursive function isEven(n: int): bool
  if n == 0 then return true end
  return isOdd(n - 1)
end

recursive function isOdd(n: int): bool
  if n == 0 then return false end
  return isEven(n - 1)
end
```

- Recursion through function values, such as a function passed as an argument, is not checked.
- **Parameters are constant.** A parameter cannot be reassigned or changed, unless it is declared `mutable` (section 6).
- **Multiple return values** are allowed and typed: `function divmod(a: int, b: int): (int, int)`.
- **A call passes every argument.** Passing fewer or more arguments than the function declares is a compile error.
- **No currying and no partial application.** To fix some arguments, write an anonymous function.
- **No method calls.** `obj:method(x)` is not valid. Write `Module.method(obj, x)`.
- **No varargs.** `...` is not valid. Pass an array instead.
- **No labeled arguments.** Pass a struct for functions with many options: `File.open(path, { mode = Read, create = true })`.

### Control flow

- `if cond then ... elseif cond then ... else ... end`. The condition must have type `bool`. There is no truthiness: `None`, `0` and `""` are not conditions.
- `while cond do ... end` and `repeat ... until cond`.
- `for i = first, last do ... end` counts from `first` to `last` inclusive, as in Lua. `for i = first, last, step do` takes a step.
- `for x in iter do ... end` loops over an iterator. `for i, x in Array.entries(xs) do` gives indexes from 0.
- **An iterator is a `Generator<T>`** (section 8). `Array.values`, `Array.entries`, `Map.keys`, `Map.entries` and `Str.chars` return generators. A function returns its own iterator by returning a generator.
- `for i, x in iter` works when the generator yields a `Tuple<A, B>`. The tuple is split into `i` and `x`.
- `break` and `goto` work as in Lua. `continue` is added.
- `match` is added (section 4).

### Literals

- **Strings** use double quotes only: `"text"`. Lua's long strings `[[ ... ]]` hold multi-line text.
- **Characters** use single quotes: `'a'`, `'\n'`, `'\u{1F600}'`. A character literal has type `char`.
- **Integers:** `42`, `0xFF`, `0o755`, `0b1010`. `_` can separate digits: `1_000_000`.
- **Floats:** `3.14`, `1.0e-9`, `1_000.5`.
- **A literal must fit its type.** `const x: u8 = 300` is a compile error.
- `true` and `false` have type `bool`.

### Comments and formatting

- `--` comments to the end of the line. `--[[ ... ]]` comments a block. `---` before a declaration is a doc comment.
- `lang fmt` defines the one valid layout. CI is expected to run `lang fmt --check`.

There are no macros, no metatables, no operator overloading and no function overloading.

### Operators

- **Operators are built in.** Programs cannot define operators or change their meaning.
- **Both operands have the same type.** There are no implicit conversions, so `1 + 2.0` is a compile error.

| Operators | Operand types |
| --- | --- |
| `+` `-` `*` `%` | Any integer type or any float type |
| `/` | Float types. Use `//` for integers |
| `//` | Floor division for any integer or float type |
| `^` | Float types |
| `+%` `-%` `*%` | Integer types, wrapping on overflow |
| `&` `\|` `~` `<<` `>>` | Integer types. `~` is xor when binary and bitwise not when unary |
| `..` | `string` |
| `<` `<=` `>` `>=` | Numbers, `char` and `string` |
| `==` `~=` | Any type without functions, compared structurally (section 4) |
| `and` `or` `not` | `bool` |
| `#` | Length of an array, map or string |

- A number literal with no other type information is `int` when written without a decimal point, and `f64` when written with one.

## 3. Modules and imports

- **One file is one module.** Its path from the package root is its name: `src/app/text.lang` is `"app/text"`.
- **Names are private by default.** `export` makes a top-level declaration visible to other modules: `export function`, `export enum`, `export struct`, `export type`, `export const`.
- **Imports are explicit.** A module sees only its own declarations, the names it imports, and the prelude.
- **Prelude.** The prelude holds `Option`, `Some`, `None`, `Result`, `Ok`, `Err`, `Array`, `Map`, `Tuple`, the primitive types, `assert` and `panic`. Nothing else is in scope without an import.
- **Paths are package rooted.** `"app/text"` for this package, `"std/string"` for the standard library, `"pkgname/module"` for a dependency. Relative paths are not allowed.

Import forms:

```lua
const Str = import("std/string")                  -- namespace import
const { classify, Token } = import("app/text")    -- named imports
```

There is no re-export statement. To expose an imported name, import it with `import` and export a new declaration:

```lua
const Text = import("app/text")

export const classify = Text.classify   -- a function
export type Token = Text.Token          -- a type, with its constructors
```

Rules:

- **`import` is a declaration, not a function call.** It must appear at the top level with a string literal path. It never runs code, and it never loads a module at runtime.
- **No cycles.** The import graph must be acyclic. The compiler reports the cycle as a list of modules.
- **No re-export statement and no default exports.** An exported name is always declared in the exporting file with `export function`, `export const`, `export type`, `export struct` or `export enum`.
- **`export type A = B`** declares an alias. When `B` is an enum, its constructors are usable through the alias.
- **No import side effects.** Top-level declarations must be functions, types, `const` constants or `lazy` values (section 10). There is no top-level `let`.
- **Constructors come with their enum.** Importing `Token` makes its constructors usable. In a `match`, constructors are resolved from the type being matched. Elsewhere, write `Token.Word(x)` when the bare name is ambiguous.
- **`lang move old new`** moves a file and rewrites every import of it.

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

A function with no return type returns nothing.

### Composite types

- **Structs.** Declared with `struct`. The full rules are in "Structs" below.

- **Enums.** Declared with `enum`. Each constructor can carry fields.

```lua
export enum Token
  Word(text: string)
  Number(value: int)
  Space
end
```

- **Arrays.** `Array<T>` is a growable, contiguous array. `{1, 2, 3}` builds one when the expected type is an array. Indexes start at 0. `xs[i]` is bounds checked and panics when out of range. `#xs` is the length.
- **Maps.** `Map<K, V>` is a hash map. `{ ["a"] = 1, ["b"] = 2 }` builds one when the expected type is a map. `m[k]` returns `Option<V>`.
- **Maps keep insertion order.** Iteration visits keys in the order they were first inserted. Removing a key and inserting it again moves it to the end.
- **Float keys are not allowed.** `Map<f64, V>` and keys that contain floats are compile errors.
- **Tuples.** `Tuple<int, string>` holds a fixed group of values, built with `Tuple(1, "a")`.
- **Multiple return values.** `: (int, int)` in a return type means the function returns two values. They are not a tuple and cannot be stored as one value.
- **Functions.** `function(int, string): bool`.

### Structs

A struct is a fixed set of named fields, stored inline like a C struct.

```lua
export struct ServerConfig
  host: string
  port: int = 8080
  tags: Array<string> = {}
end

struct Entry<K, V>
  key: K
  value: V
end
```

**Fields and defaults**

- Each field has a name and a type.
- A field can have a default value after `=`. The default must be a compile-time constant (section 10).
- A field without a default must be given every time the struct is built.

**Building a struct**

- `TypeName { field = value, ... }` builds a struct of that type. The fields can appear in any order.
- When the expected type is already known, the type name can be left out: `{ field = value }`. The type comes from an annotation, a parameter or a return type.
- Fields with defaults can be left out. All other fields must be given.
- Every field is written as `name = value`, even when the value is a variable with the same name.
- Type arguments of a generic struct are inferred from the fields: `Entry { key = "a", value = 1 }` is `Entry<string, int>`.

```lua
const config = ServerConfig { host = "localhost" }        -- port is 8080, tags is {}
const loopback: ServerConfig = { host = "127.0.0.1", port = 3000 }

function makeEntry(key: string, value: int): Entry<string, int>
  return { key = key, value = value }                      -- type taken from the return type
end
```

**Reading and updating**

- `s.field` reads a field.
- `s with { field = value }` returns a copy with some fields changed. The original is unchanged.
- `s.field = value` changes a field in place. `s` must be a `let` variable, a `mutable` parameter, or a field of one.
- To change a caller's struct inside a function, declare the parameter `mutable` (section 6).

```lua
const moved = config with { port = 9000 }

let current = config
current.port = 9001
current.tags[0] = "primary"
```

**Destructuring**

- `const { field1, field2 } = s` binds each field to a variable with the same name.
- `const { field = name } = s` binds the field to a different name.
- Fields that are not listed are ignored.
- The same pattern works in `match`. A literal after `=` tests the field's value instead of binding it.

```lua
const { host, port = serverPort } = config

match point
  when { x = 0.0, y = 0.0 } then print("origin")
  when { x = 0.0, y } then print("on the y axis at " .. F64.toString(y))
  when { x, y } then print(F64.toString(x) .. ", " .. F64.toString(y))
end
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

```lua
struct ListNode
  value: int
  next: Option<Box<ListNode>> = None
end

const list = ListNode { value = 1, next = Some(Box.new(ListNode { value = 2 })) }
```

### Optional values

- `Option<T>` is either `Some(value)` or `None`. A plain `T` always holds a value. There is no `nil`.
- `if x is Some(v) then` binds `v` inside the branch. After `if x is None then return ... end`, `x` is known to be `Some` for the rest of the block, and `x.value` reads the value.
- `Option.withDefault(x, fallback)` returns `fallback` when `x` is `None`.
- Type syntax uses names, not symbols: `Option<T>`, `Array<T>`, `Map<K, V>`, `Tuple<A, B>`, `Result<T, E>`.

### Pattern matching

`match` checks a value against patterns. It must cover every case, or have an `else` branch.

```lua
match token
  when Word(text) then print(text)
  when Number(n) if n < 0 then print("negative")
  when Number(n) then print(Int.toString(n))
  when Space then print(" ")
end
```

- Patterns can bind fields, test literals, and add an `if` guard.
- `match` on an `Option<T>` uses `when Some(v) then` and `when None then`.

### Generics

- Type parameters use angle brackets: `function first<T>(items: Array<T>): Option<T>` and `struct Entry<K, V> key: K value: V end`.
- There are no typeclasses or traits. Behavior is passed explicitly, as a function: `Array.sortBy(xs, Int.compare)`.
- Arithmetic operators do not work on a type parameter. Pass a function instead.
- **Structural equality.** `==` compares values structurally, and the compiler derives it for every type except functions. Using `==` on a type that contains a function is a compile error.
- **Hashing.** `Hash.of(x)` follows the same rule, and also rejects types that contain floats.
- **`Equatable` and `Hashable` constraints.** Generic code that uses `==` declares `<T: Equatable>`. Code that hashes declares `<T: Hashable>`. The compiler checks each type argument against the rules above. These two are the only constraints in the language, and users cannot define new ones.
- **Floats follow IEEE 754 in `==`.** `NaN == NaN` is `false`. A struct that contains `NaN` is not equal to itself.

```lua
export function contains<T: Equatable>(items: Array<T>, wanted: T): bool
  for item in Array.values(items) do
    if item == wanted then return true end
  end
  return false
end
```

### Inference

- Types of variables are inferred. `const n = 1` is `int`.
- **Function parameters and return types are written for every function.** Anonymous functions passed as arguments are the exception. Their types come from the parameter they are passed to.
- Other modules type-check against exported signatures only. They never read function bodies.

### Numbers

- Integer overflow panics in debug builds and wraps in release builds. `+%`, `-%` and `*%` always wrap.
- Integer division or remainder by zero panics in every build.
- Conversions are explicit: `I32.from(x)`, `F64.fromInt(x)`, `Int.truncate(f)`.

### Strings

- `string` holds valid UTF-8. Functions that build a string from bytes return `Result` and fail on invalid UTF-8.
- **Positions are byte offsets of type `ByteIndex`**, starting at 0. `ByteIndex` is a distinct type, not `int`.
- **Positions come from the string.** `Str.indexOf`, `Str.find` and iterators return `ByteIndex` values that always fall on a character boundary.
- **Slices share memory.** `Str.slice` and `Array.slice` return a view into the original buffer, with no copy. The view keeps the whole buffer alive. Use `Str.copy` or `Array.copy` to keep a small part of a large buffer without keeping the rest.
- **Slicing checks boundaries.** `Str.slice(s, from, to)` returns bytes `from` up to but not including `to`. It panics when a position is not on a character boundary or is out of range.
- `#s` is the byte length as `ByteIndex`. `Str.charCount(s)` counts characters and takes time proportional to the length.
- Grapheme clusters, case mapping and normalization are in `std/unicode`.

```lua
export struct Pair
  key: string
  value: string
end

export function parsePair(line: string): Option<Pair>
  const found = Str.indexOf(line, "=")
  if found is None then
    return None
  end
  const i = found.value
  const key = Str.trim(Str.slice(line, 0, i))
  const value = Str.trim(Str.sliceFrom(line, Str.nextChar(line, i)))
  return Some({ key = key, value = value })
end
```

## 5. Errors

- **Errors are return values.** A function that can fail returns `Result<T, E>`, with constructors `Ok(value)` and `Err(error)`.
- **The caller checks the result.** There is no `try`, no `catch` and no operator that returns early. Every error path is a visible `if` or `match`.
- **`is` tests a pattern and narrows the type.** After `if r is Err(e) then return ... end`, the compiler knows `r` is `Ok`. `r.value` then reads the value. The same narrowing works for `Option<T>` with `None`.
- **Ignoring a `Result` is a compile error.** Discard one on purpose with `const _ = f()`.
- **Panics abort the program.** `panic("message")`, failed `assert`, out-of-bounds indexing and debug-mode overflow all panic. A panic cannot be caught.
- There are no exceptions and no `pcall`.

```lua
export function readConfig(path: string): Result<Config, ConfigError> uses io, alloc
  const text = File.readText(path)
  if text is Err(e) then
    return Err(IoFailed(e))
  end

  const config = Config.parse(text.value)
  if config is Err(e) then
    return Err(e)
  end

  return Ok(config.value)
end
```

## 6. Memory model

### Values and ownership

- Every value has one or more owners: variables, fields, array slots, closure captures.
- Boxed values (enums with fields, arrays, maps, strings, closures and `Box<T>`) live on the heap with a reference count. Numbers, `bool`, `char` and structs live inline.
- A struct that holds boxed fields counts those fields, not itself. Copying the struct increments the count of each boxed field.
- **`Box<T>`** puts a value on the heap with a reference count. Use it for recursive structs and for large structs that are passed around often.
- The compiler inserts reference count increments and decrements. It removes the ones it can prove unnecessary, as in Koka's Perceus.
- A value is freed when its count reaches zero. Freeing is deterministic, so file handles and sockets close at the point their last owner drops them.
- **Running out of memory aborts the program,** the same as a panic.

### In-place updates

- `xs[i] = v`, `m[k] = v`, `Array.push(mutable xs, v)` and `u.age = v` change a `let` variable. `u with { ... }` returns a changed copy.
- When the reference count of the old value is 1, the compiler writes into it instead of copying. This is called reuse.
- Code that keeps one owner gets in-place mutation with no copies.

```lua
let xs: Array<int> = Array.filled(0, 1000)
for i = 0, 999 do
  xs[i] = i * i  -- the count is 1, so the write is in place
end
```

- `xs[i] = v` and `u.age = v` require the target to be a `let` variable, a `mutable` parameter, or a field of one.

### Mutable parameters

A function changes a caller's variable only through a `mutable` parameter. The keyword appears in the declaration and again at the call.

```lua
function damage(mutable player: Player, amount: int)
  player.hp = player.hp - amount
end

let hero = Player { name = "Ana", hp = 100, position = Vec2 { x = 0.0, y = 0.0 } }
damage(mutable hero, 10)      -- hero.hp is now 90
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
| `yield` | May suspend the current coroutine or async task (section 8) |
| `unsafe` | Calls code that the compiler cannot check (section 9) |

Rules:

- Exported functions declare their effects after the return type: `function f(s: string): int uses alloc`.
- The compiler infers effects for private functions and anonymous functions.
- A function's effects include the effects of everything it calls. Calling an `io` function from a function that does not declare `io` is a compile error.
- Panics are not an effect. Any function can panic.
- The entry function (section 15) has every effect except `yield`. `Async.run` provides a context where `yield` is allowed.

### Effect variables

A function that calls a function it receives has that function's effects. An effect variable states this in a signature.

```lua
export function map<T, U, effect E>(items: Array<T>, f: function(T): U uses E): Array<U> uses alloc, E
```

- Effect variables are declared in the type parameter list with the `effect` keyword, and appear only after `uses`.
- **They are bound by a parameter.** An effect variable in the function's own `uses` list must also appear in a parameter's function type.
- At each call, the variable takes the effects of the function passed in. `map(xs, Str.len)` has `alloc`. `map(paths, File.readText)` has `alloc, io`.

## 8. Coroutines and concurrency

Lang has three tools. Each covers one type of concurrent work:

| Work | Tool |
| --- | --- |
| Suspend and resume a computation on one thread | Coroutines |
| Run computations in parallel on several cores | Scoped threads and channels |
| Wait on many I/O operations at once | Async tasks on an event loop |

A panic on any thread or task aborts the whole program.

### Coroutines

Coroutines keep Lua's model: create, resume and yield. They are typed, and the compiler turns each coroutine into a state machine, so no separate stack is allocated.

```lua
const Co = import("std/coroutine")

function counter(limit: int): Generator<int>
  return Co.generator(function() uses yield
    for i = 0, limit - 1 do
      Co.yield(i)
    end
  end)
end

let c = counter(3)
for n in Co.values(mutable c) do
  print(Int.toString(n))  -- 0, 1, 2
end
```

- **`Coroutine<In, Out>`** takes `In` on each resume and produces `Out` on each yield.
- **`Generator<Out>`** is a coroutine that takes no input on resume. `Co.generator(f)` makes one.
- **`Co.create(f)`** makes a coroutine from a function with the `yield` effect.
- **`Co.resume(mutable c, input)`** runs the coroutine until its next yield, changing `c` in place. It returns `Yielded(value)`, `Done` or `Panicked(message)`. A panic inside a coroutine still aborts the program in release builds. `Panicked` is only reported in test builds.
- **`Co.yield(value)`** suspends the coroutine. It has the `yield` effect.
- **`Co.values(mutable g)`** iterates over a `Generator<Out>`, advancing `g` in place.
- **A coroutine is a value like any other.** Resuming one requires a `mutable` argument, so the change is visible at the call. Copying a coroutine with `const` and then resuming the copy leaves the original unchanged.
- **Yield crosses calls only through `yield` functions.** A function can call `Co.yield` only if it has the `yield` effect. Higher-order functions pass `yield` through effect variables like any other effect.
- **Generic code that yields is specialized.** A generic function called with a `yield` function is compiled as a separate state-machine copy for that effect. Code that never yields pays nothing.

### Threads

Threads run in a scope. Every thread started in a scope finishes before the scope returns.

```lua
const Thread = import("std/thread")

const total = Thread.scope(function(scope: Scope): int uses alloc
  const left = Thread.start(scope, function(): int return sum(firstHalf) end)
  const right = Thread.start(scope, function(): int return sum(secondHalf) end)
  return Thread.join(left) + Thread.join(right)
end)
```

- **`Thread.scope(f)`** calls `f` with a `Scope`. It waits for every thread started in that scope, then returns the result of `f`.
- **`Thread.start(scope, f)`** runs `f` on a new thread and returns a `Thread<T>`. `Thread.join(t)` waits for it and returns its result.
- **A `Scope` cannot escape.** It cannot be stored, returned, or captured by a thread function. Only the function passed to `Thread.scope` can start threads in it.
- **Thread functions capture only constant values,** like every closure. They cannot have the `yield` effect.
- **`Parallel.map(items, f)`** from `std/parallel` splits an array across threads and returns the results in order.

### Channels

```lua
const sender, receiver = Channel.create(16)      -- holds up to 16 values

Thread.start(scope, function() uses io
  for i = 0, 99 do
    Channel.send(sender, i)                       -- blocks while the channel is full
  end
  Channel.close(sender)
end)

for value in Channel.values(receiver) do         -- ends when the channel is closed and empty
  print(Int.toString(value))
end
```

- **`Channel.create<T>(capacity)`** returns a `Sender<T>` and a `Receiver<T>`.
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

```lua
const hits = Shared.create(0)

Shared.update(hits, function(mutable count: int)
  count = count + 1
end)

const now = Shared.read(hits, function(count: int): int return count end)
```

- **The lock is held only during the callback.** `Shared.update` and `Shared.read` take the lock, run the callback and release it. There is no way to take the lock and forget to release it.
- **The callback cannot have the `yield` effect,** so a task never suspends while holding a lock.
- **`Atomic<T>`** offers `Atomic.load`, `Atomic.store`, `Atomic.add` and `Atomic.compareExchange`.
- `==` on an identity type compares identity, not contents.
- **Deadlocks are possible** when callbacks nest locks in different orders. `lang audit` lists every nested `Shared.update`.

### Async tasks

An event loop runs many tasks on one thread. A task suspends at each I/O call that would block, and the loop resumes it when the I/O is ready.

```lua
const Async = import("std/async")
const Net = import("std/async/net")

function handle(connection: Connection) uses io, yield, alloc
  let conn = connection
  while true do
    const data = Net.read(mutable conn, 4096)     -- suspends until data arrives
    if data is Err(_) then return end
    if #data.value == 0 then return end
    const written = Net.write(mutable conn, data.value)
    if written is Err(_) then return end
  end
end

export function run(args: Array<string>): int    -- entry function set in lang.toml
  Async.run(function(loop: Loop) uses io, yield, alloc
    let listener = Net.listen("127.0.0.1", 8080)
    while true do
      const accepted = Net.accept(mutable listener)
      if accepted is Ok(conn) then
        Async.spawn(loop, function() uses io, yield, alloc handle(conn) end)
      end
    end
  end)
  return 0
end
```

- **`Async.run(f)`** runs an event loop on the current thread until `f` and every task it started have finished. It removes the `yield` effect, so it can be called from the entry function or from a thread.
- **`Async.spawn(loop, f)`** starts a task on that loop and returns a `Task<T>`. **`Async.await(t)`** suspends until the task finishes and returns its result. **`Async.all(tasks)`** waits for several.
- **A `Loop` cannot escape,** with the same rules as a `Scope`.
- **Suspending I/O lives in `std/async/*`.** `std/async/net`, `std/async/file` and `std/async/time` mirror the blocking modules with the same function names. Their functions have `uses io, yield`.
- **A task stays on the thread that runs its loop.** To use several cores, run one loop per thread inside `Thread.scope`. Pass work between the loops with channels.
- `Async.sleep(milliseconds)` suspends the task for a time.

## 9. Unsafe code and C interop

- **`unsafe do ... end`** blocks allow raw pointers (`Ptr<T>`), pointer arithmetic and unchecked indexing. A function containing an `unsafe` block has the `unsafe` effect unless it is marked `@trusted`, which states that the author checked it by hand.
- **`extern`** declares a C function: `extern function strlen(s: Ptr<u8>): u64 uses unsafe = "strlen"`.
- **`@repr(C)`** on a struct fixes its field order and padding to match C. Without it, the compiler may reorder fields to reduce padding.
- **C-compatible enums.** `@repr(u8)`, `@repr(i32)` and the other integer types give an enum without fields explicit values. `Enum.toInt` and `Enum.fromInt` convert, and `fromInt` returns `Option`.

```lua
@repr(u8)
export enum Color
  Red = 1
  Green = 2
  Blue = 4
end
```

- **Callbacks from C.** A function marked `@export(C, "name")` gets a C-compatible symbol. Its parameters and result must be `@repr(C)` types, numbers or `Ptr<T>`.
- **C strings.** `std/ffi` provides `CString`. `CString.from(s)` copies a string with a terminating zero byte, and `CString.toString(p)` copies back and returns `Result` on invalid UTF-8.
- **Linking.** A `[link]` table in `lang.toml` lists C libraries and search paths (section 15).
- `lang audit` lists every `unsafe` block and `@trusted` function in a package.

## 10. Top-level declarations

- **`const`** at the top level declares a compile-time constant. Inside a function, `const` accepts any expression. A top-level value must be one of these:
  - a literal or a constructor
  - a struct or array of constants
  - a named function or another constant
  - a call to an `@const` function
- **`lazy`** defers a computation until first use: `const table = lazy buildTable()`. The first `Lazy.force(table)` computes the value and later calls return it. Forcing is thread safe.
- Modules have no initialization code. Program startup runs nothing before the entry function.

## 11. Code generation

- **Output is C.** One `.c` file per module, plus a small runtime: allocator, reference counting, panic handler, arena.
- **Debug builds** use a fast C compiler such as TCC with checks on.
- **Release builds** use Clang or GCC with `-O2`.
- **Coroutines** compile to a struct holding their live variables and a state number, plus a resume function with a `switch` on that number.
- **`mutable` parameters** compile to C pointer parameters.
- **Self-calls in tail position** inside a `recursive function` compile to loops (section 2). Other recursion uses the C stack. Stack overflow aborts the program.
- **Polymorphic functions are compiled once.** A value of a type parameter is passed as one machine word in a uniform form.
- **Uniform form.** The lowest bit tells generic code what the word holds. A clear bit means a heap pointer, which needs reference counting. A set bit means an inline value.

| Type | Uniform form |
| --- | --- |
| Boxed values (enums with fields, arrays, maps, strings, closures, `Box<T>`) | Pointer, unchanged |
| `i8` to `i32`, `u8` to `u32`, `bool`, `char`, `f32`, enum constructors without fields | Inline, shifted left one bit with the tag bit set |
| `i64`, `u64`, `int`, `f64` | Copied into a heap box with a reference count |
| Structs | Copied into a heap box with a reference count |

- **Boxing happens only at the boundary.** Non-generic code uses raw C types. The compiler converts a value to the uniform form when it passes the value into generic code, and converts it back when it receives one.
- **`@specialize`** on a function asks the compiler to emit a separate copy for each concrete type it is called with. Use it in hot paths only.

## 12. Build model

- The compiler reads a module's `import` declarations before anything else and builds the module graph from them.
- Each module produces an interface file (`.langi`) with its exported signatures. A module type-checks against its dependencies' interface files only.
- When a module changes but its interface file stays the same, modules that import it are not re-checked or recompiled.
- Modules with no dependency between them compile in parallel.
- Builds are deterministic. The same inputs produce byte-identical C and binaries.
- Compile-time evaluation is limited to top-level `const` declarations and `@const` functions.

## 13. Tests and contracts

- **`test "name" do ... end`** blocks sit next to the code. `lang test` runs them. They are removed from non-test builds.
- **Contracts.** `requires` and `ensures` lines at the start of a function body are checked in debug builds and removed in release builds. `ensures` refers to the return value as `result`.

```lua
export function sqrt(x: f64): f64
  requires x >= 0.0
  ensures result >= 0.0
  return Float.sqrt(x)
end
```

## 14. Tooling for agents

### Diagnostics

`lang check --json` prints one JSON object per diagnostic:

```json
{
  "code": "E0412",
  "severity": "error",
  "message": "Function `load` uses `io` but its signature does not declare it",
  "file": "src/app/config.lang",
  "range": {"start": [12, 4], "end": [12, 22]},
  "fixes": [
    {"title": "Add `io` to the signature", "edits": [{"file": "src/app/config.lang", "range": {"start": [8, 40], "end": [8, 40]}, "text": ", io"}]}
  ]
}
```

- Error codes are stable. `lang explain E0412` prints the rule and an example fix.
- Fixes are exact edits that an agent can apply without rewriting them.

### Queries

| Command | Output |
| --- | --- |
| `lang query type-of FILE:LINE:COL` | The type and effects at a position |
| `lang query def NAME` | Where a name is defined |
| `lang query refs NAME` | Every use of an exported name |
| `lang query callers FUNCTION` | Every function that calls it |
| `lang query interface MODULE` | The module's exported signatures |

All query commands accept `--json`.

### Speed target

`lang check` on an unchanged package with one edited file returns in under 100 ms.

## 15. Packages (sketch)

- **A package is a directory** with a `lang.toml` manifest and a `src/` directory.
- **The package name is the import path prefix.** Package `json` exposes `src/parse.lang` as `"json/parse"`.
- **Dependencies are git repositories** pinned to version tags. There is no central registry.

```toml
[package]
name = "app"
version = "0.1.0"

[dependencies]
json = { git = "https://github.com/example/lang-json", version = "1.2.0" }

[[binary]]
name = "app"
entry = "app/cli.run"

[link]
libraries = ["sqlite3"]
search = ["/opt/homebrew/lib"]
```

- **Entry points are set in `lang.toml`.** Each `[[binary]]` names the binary and its entry function as `"module.function"`. A package with no `[[binary]]` is a library.
- **The entry function** has the signature `export function run(args: Array<string>): int`. The result is the process exit code. It has every effect except `yield`.
- **`[link]`** lists C libraries to link and directories to search.

- **Version selection is minimal version selection,** as in Go. Each dependency resolves to the highest version that some package in the graph requires as its minimum. The result is the same on every machine and needs no solver.
- **One version per package** in a build. A major version change is a new package name, such as `json2`.
- **`lang.lock`** records the commit hash and content hash of every dependency. `lang build` fails when a fetched dependency does not match its hash.

## 16. Deferred

These are left out of the first version on purpose.

- **Task cancellation and timeouts.** A task runs until it returns. A cancellation design for async tasks and threads is not decided.
- **Functors and module parameters.** See `DECISIONS.md`, D6.
- **A native debug backend.** See `DECISIONS.md`, D10.
- **A bytecode VM.** Possible later for scripting and hot reload. The first target is native code only.
