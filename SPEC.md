# Lang specification (draft)

This document defines the Lang language. Each section states the rules first, then shows an example. `DECISIONS.md` explains why each rule was chosen over the alternatives.

File extension: `.lang`. Compiler command: `lang`.

The previous draft used ReasonML syntax. It is kept at the git tag `reasonml-draft`.

## 1. Overview

- Lang uses Lua syntax with static types. Metatables are removed.
- Lang compiles to native code, the same target as C, Rust and Go. The compiler emits C, and a C compiler builds the final binary.
- Lang code is a set of modules. One file is one module. There are no globals.
- Values are immutable. Safe code has no pointers or references, only values.
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

### Functions

- `function name(a: int, b: int): int ... end` declares a function. A function is private to its module unless declared with `export function` (section 3). A function declared inside another function is local to that function.
- Anonymous functions use the same form without a name: `function(x: int): int return x + 1 end`.
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
- `break` and `goto` work as in Lua. `continue` is added.
- `match` is added (section 4).

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

- **Structs.** Declared with `struct`. Built with a table constructor whose type is known from context. Updated with `with`, which returns a new struct.
- **Structs are stored inline,** like C structs. A struct field, array element or variable holds the struct's fields directly, with no heap allocation and no reference count of its own.
- **A struct cannot contain itself.** Use an enum, an array or `Box<T>` for recursive data.

```lua
export struct User
  name: string
  age: int
end

const u: User = { name = "Ana", age = 30 }
const older = u with { age = 31 }
```

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
- **Tuples.** `Tuple<int, string>` holds a fixed group of values, built with `Tuple(1, "a")`.
- **Multiple return values.** `: (int, int)` in a return type means the function returns two values. They are not a tuple and cannot be stored as one value.
- **Functions.** `function(int, string): bool`.

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
- **Hashing.** `Hash.of(x)` follows the same rule.

### Inference

- Types of variables are inferred. `const n = 1` is `int`.
- **Function parameters and return types are written for every function.** Anonymous functions passed as arguments are the exception. Their types come from the parameter they are passed to.
- Other modules type-check against exported signatures only. They never read function bodies.

### Numbers

- Integer overflow panics in debug builds and wraps in release builds. `+%`, `-%` and `*%` always wrap.
- Conversions are explicit: `I32.from(x)`, `F64.fromInt(x)`, `Int.truncate(f)`.

### Strings

- `string` holds valid UTF-8. Functions that build a string from bytes return `Result` and fail on invalid UTF-8.
- **Positions are byte offsets of type `ByteIndex`**, starting at 0. `ByteIndex` is a distinct type, not `int`.
- **Positions come from the string.** `Str.indexOf`, `Str.find` and iterators return `ByteIndex` values that always fall on a character boundary.
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

### In-place updates

- `xs[i] = v`, `m[k] = v`, `Array.push(xs, v)`, `u.age = v` and `u with { ... }` produce a new value and assign it.
- When the reference count of the old value is 1, the compiler writes into it instead of copying. This is called reuse.
- Code that keeps one owner gets in-place mutation with no copies.

```lua
let xs: Array<int> = Array.filled(0, 1000)
for i = 0, 999 do
  xs[i] = i * i  -- the count is 1, so the write is in place
end
```

- `xs[i] = v` and `u.age = v` require the target to be a `let` variable or a field of one.

### Cycles

- Immutable values cannot form cycles, so reference counting frees everything without a cycle collector.
- A `let` variable cannot be captured or stored, so it cannot create a cycle.
- **Graphs** use `Arena<T>` and `Handle<T>` from `std/arena`. A handle is an index plus a generation number. `Arena.get(arena, h)` returns `None` when the slot was freed or reused. Handles can point anywhere in the arena, including in cycles, and the arena frees all its slots at once.

### Threads

- Values sent to another thread with `Thread.spawn` or `Channel.send` are marked shared. Shared values use atomic reference counts. Unshared values use plain counts.
- An in-place update on a shared value copies first, so threads never see each other's writes.
- There is no shared mutable state in safe code. `std/sync` provides `Mutex<T>` and `Atomic<T>` for the cases that need it.

## 7. Effects

An effect marks something a function can do beyond computing a value. The set is fixed:

| Effect | Meaning |
| --- | --- |
| `alloc` | May allocate heap memory |
| `io` | May perform input or output, read the clock, or read the environment |
| `yield` | May suspend the current coroutine (section 8) |
| `unsafe` | Calls code that the compiler cannot check (section 9) |

Rules:

- Exported functions declare their effects after the return type: `function f(s: string): int uses alloc`.
- The compiler infers effects for private functions and anonymous functions.
- A function's effects include the effects of everything it calls. Calling an `io` function from a function that does not declare `io` is a compile error.
- Panics are not an effect. Any function can panic.
- `main` has every effect except `yield`.

### Effect variables

A function that calls a function it receives has that function's effects. An effect variable states this in a signature.

```lua
export function map<T, U, effect E>(items: Array<T>, f: function(T): U uses E): Array<U> uses alloc, E
```

- Effect variables are declared in the type parameter list with the `effect` keyword, and appear only after `uses`.
- **They are bound by a parameter.** An effect variable in the function's own `uses` list must also appear in a parameter's function type.
- At each call, the variable takes the effects of the function passed in. `map(xs, Str.len)` has `alloc`. `map(paths, File.readText)` has `alloc, io`.

## 8. Coroutines

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

const c = counter(3)
for n in Co.values(c) do
  print(Int.toString(n))  -- 0, 1, 2
end
```

- **`Coroutine<In, Out>`** takes `In` on each resume and produces `Out` on each yield.
- **`Generator<Out>`** is a coroutine that takes no input on resume. `Co.generator(f)` makes one.
- **`Co.create(f)`** makes a coroutine from a function with the `yield` effect.
- **`Co.resume(c, input)`** returns `Yielded(value)`, `Done` or `Panicked(message)`. A panic inside a coroutine still aborts the program in release builds. `Panicked` is only reported in test builds.
- **`Co.yield(value)`** suspends the coroutine. It has the `yield` effect.
- **`Co.values(g)`** iterates over a `Generator<Out>`.
- **Yield crosses calls only through `yield` functions.** A function can call `Co.yield` only if it has the `yield` effect. Higher-order functions pass `yield` through effect variables like any other effect.
- **Generic code that yields is specialized.** A generic function called with a `yield` function is compiled as a separate state-machine copy for that effect. Code that never yields pays nothing.

## 9. Unsafe code and C interop

- **`unsafe do ... end`** blocks allow raw pointers (`Ptr<T>`), pointer arithmetic and unchecked indexing. A function containing an `unsafe` block has the `unsafe` effect unless it is marked `@trusted`, which states that the author checked it by hand.
- **`extern`** declares a C function: `extern function strlen(s: Ptr<u8>): u64 uses unsafe = "strlen"`.
- **`@repr(C)`** on a struct fixes its field order and padding to match C. Without it, the compiler may reorder fields to reduce padding.
- `lang audit` lists every `unsafe` block and `@trusted` function in a package.

## 10. Top-level declarations

- **`const`** at the top level declares a compile-time constant. Its value must be one of these: a literal, a constructor, a struct or array of constants, a named function, another constant, or a call to an `@const` function. Inside a function, `const` accepts any expression.
- **`lazy`** defers a computation until first use: `const table = lazy buildTable()`. The first `Lazy.force(table)` computes the value and later calls return it. Forcing is thread safe.
- Modules have no initialization code. Program startup runs nothing before `main`.

## 11. Code generation

- **Output is C.** One `.c` file per module, plus a small runtime: allocator, reference counting, panic handler, arena.
- **Debug builds** use a fast C compiler such as TCC with checks on.
- **Release builds** use Clang or GCC with `-O2`.
- **Coroutines** compile to a struct holding their live variables and a state number, plus a resume function with a `switch` on that number.
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
```

- **Version selection is minimal version selection,** as in Go. Each dependency resolves to the highest version that some package in the graph requires as its minimum. The result is the same on every machine and needs no solver.
- **One version per package** in a build. A major version change is a new package name, such as `json2`.
- **`lang.lock`** records the commit hash and content hash of every dependency. `lang build` fails when a fetched dependency does not match its hash.

## 16. Deferred

These are left out of the first version on purpose.

- **Async I/O.** Coroutines (section 8) are the base for it. An event loop in `std/async` that resumes coroutines when I/O is ready is the planned direction.
- **Functors and module parameters.** See `DECISIONS.md`, D6.
- **A native debug backend.** See `DECISIONS.md`, D10.
- **A bytecode VM.** Possible later for scripting and hot reload. The first target is native code only.
