# Lang specification (draft)

This document defines the Lang language. Each section states the rules first, then shows an example. `DECISIONS.md` explains why each rule was chosen over the alternatives.

File extension: `.lang`. Compiler command: `lang`.

## 1. Overview

- Lang code is a set of modules. One file is one module.
- Values are immutable by default. Safe code has no pointers or references, only values.
- The compiler frees memory with reference counting that it inserts itself. There is no garbage collector.
- When a value has exactly one owner, an update writes to it in place instead of copying it.
- The compiler emits C. A C compiler builds the final binary.

```reason
import { type shape } from "app/geometry";
import * as Float from "std/float";

export let area: (shape) => f64 = (s) =>
  switch (s) {
  | Circle(r) => Float.pi *. r *. r
  | Rect({width, height}) => width *. height
  };

test("rect area") {
  assert(area(Rect({width: 2.0, height: 3.0})) == 6.0);
};
```

## 2. Syntax

The syntax follows ReScript, the current form of ReasonML syntax.

- **Bindings.** `let x = expr;` binds a name. Bindings are immutable.
- **Mutable locals.** `let mut x = expr;` declares a local that can be reassigned with `x = expr;`. A mutable local cannot be captured by a closure.
- **Expressions.** `if`, `switch` and `{ ... }` blocks are expressions. A block's value is its last expression.
- **Functions.** `(a, b) => expr`. Functions are values and can be stored, passed and returned.
- **Labeled arguments.** `(~path, ~mode) => ...` declares them. Callers write `open(~path="a.txt", ~mode=Read)`. Labeled arguments can appear in any order.
- **Loops.** `for i in 0 to n - 1 { ... }`, `for item in list { ... }` and `while cond { ... }`. Loops have type `unit`.
- **Comments.** `//` to the end of the line. `/* ... */` for blocks. `///` before a declaration is a doc comment.
- **Formatting.** `lang fmt` defines the one valid layout. The compiler rejects nothing for layout, but CI is expected to run `lang fmt --check`.

There are no macros and no function overloading.

### Calls

- **A call passes every argument.** A function declared with two parameters is called with exactly two arguments. Passing fewer is a compile error.
- **No currying.** `(a, b) => expr` is one function of two parameters, not a function that returns a function. The type `(int, int) => int` is different from `(int) => (int) => int`.
- **No partial application.** To fix some arguments, write a lambda: `(x) => add(1, x)`.
- **No pipe operator.** Write nested calls or name the intermediate values with `let`.

```reason
// Not valid: add(1), x->add(1), x |> add(1)
let addOne = (x) => add(1, x);
let words = String.split(text, ~on=" ");
let counts = Array.map(words, String.length);
```

### Operators

- **Operators are built in.** Programs cannot define new operators or give existing operators new meanings.
- **Each operator works on one type family.**

| Operators | Operand types |
| --- | --- |
| `+` `-` `*` `/` `%` | Integers. Both operands have the same integer type. |
| `+%` `-%` `*%` | Integers, with wrapping on overflow. |
| `+.` `-.` `*.` `/.` | Floats. Both operands have the same float type. |
| `++` | `string` concatenation. |
| `<` `<=` `>` `>=` | Integers, floats, `char` and `string`. Both operands have the same type. |
| `==` `!=` | Any type without functions, compared structurally (section 4). |
| `&&` `\|\|` `!` | `bool`. |

- An integer operation whose operand types are still unknown after inference uses `int`.

## 3. Modules and imports

- **One file is one module.** Its path from the package root is its name: `src/app/text.lang` is `"app/text"`.
- **Names are private by default.** `export` makes a top-level declaration visible to other modules.
- **Imports are explicit.** A module sees only its own declarations, the names it imports, and the prelude.
- **Prelude.** The prelude holds `option`, `result`, `Some`, `None`, `Ok`, `Error`, the primitive types, `assert` and `panic`. Nothing else is in scope without an import.
- **Paths are package rooted.** `"app/text"` for this package, `"std/string"` for the standard library, `"pkgname/module"` for a dependency. Relative paths such as `"../text"` are not allowed.

Import forms:

```reason
import { classify, type token } from "app/text";   // named imports
import * as Str from "std/string";                 // namespace import
export { classify } from "app/text";               // named re-export
```

Rules:

- **No cycles.** The import graph must be acyclic. The compiler reports the cycle as a list of modules.
- **No `export *`.** Re-exports name each item.
- **No default exports.**
- **No import side effects.** Importing a module never runs code. Top-level `let` values must be constant expressions or `lazy` values (see section 9).
- **Constructors come with their type.** Importing `type token` makes its constructors usable in patterns. In a `switch`, constructors are resolved from the type being matched. Elsewhere, write `Token.Word(x)` when the bare name is not in scope or is ambiguous.
- **Nested modules.** `export module Utf8 = { ... }` declares a module inside a file. It follows the same `export` rules.
- **`lang move`.** Moving a file with `lang move old new` rewrites every import of it.

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
| `unit` | The type with one value, `()` |

### Composite types

- **Records.** `type user = {name: string, age: int};` Records are structural values. Update with `{...u, age: 31}`.
- **Variants.** `type shape = Circle(f64) | Rect({width: f64, height: f64});`
- **Tuples.** `(int, string)`.
- **Arrays.** `array(t)` is a growable, contiguous sequence. Indexing is bounds checked.
- **Functions.** `(int, ~label: string) => bool`.
- **Option and result.** `option(t)` and `result(t, e)` are ordinary variants defined in the prelude.

There is no null. A missing value is `None`.

### Generics

- Type parameters are written `array('a)` and `type tree('a) = Leaf | Node(tree('a), 'a, tree('a))`, as in ReasonML.
- There are no typeclasses or traits. Behavior is passed explicitly, as a function or a module: `Array.sortBy(xs, Int.compare)`.
- **Structural equality.** `==` compares values structurally, and the compiler derives it for every type except functions. Using `==` on a type that contains a function is a compile error.
- **Hashing.** `Hash.of(x)` follows the same rule.

### Inference

- Types are inferred inside a module with Hindley-Milner inference.
- **Exported declarations must state their type**, including effects (section 7). Other modules type-check against these signatures only. They never read the body.
- Local bindings and private functions need no annotations.

### Numbers

- Integer overflow panics in debug builds and wraps in release builds. The operators `+%`, `-%` and `*%` always wrap.
- Float operators are separate: `+.`, `-.`, `*.`, `/.` (section 2).
- There are no implicit conversions. Use `I64.fromI32(x)`, `F64.fromInt(x)` and similar functions.

### Strings

- `string` holds valid UTF-8. Functions that build a string from bytes return `result`, and fail on invalid UTF-8.
- **Positions are byte offsets of type `byteIndex`.** `byteIndex` is a distinct type, not `int`. A character count cannot be passed where a position is expected.
- **Positions come from the string.** `String.indexOf`, `String.find` and iterators return `byteIndex` values. These always fall on a character boundary.
- **Slicing checks boundaries.** `String.slice(s, ~from, ~to)` panics when either position is not on a character boundary or is out of range.
- `String.length(s)` returns the byte length as `byteIndex`. `String.charCount(s)` counts characters and takes time proportional to the length.
- `ByteIndex.toInt` and `ByteIndex.unsafeFromInt` convert to and from `int`. The second one is checked at the next slice.
- Grapheme clusters, case mapping and normalization are in `std/unicode`.

```reason
let line = "key = value";
switch (String.indexOf(line, "=")) {
| Some(i) => {
    let key = String.trim(String.slice(line, ~from=ByteIndex.zero, ~to=i));
    let value = String.trim(String.sliceFrom(line, ByteIndex.next(line, i)));
    Some((key, value))
  }
| None => None
};
```

## 5. Errors

- **Recoverable errors use `result(t, e)`.**
- **`try expr`** unwraps an `Ok` value or returns the `Error` from the enclosing function. The error types must be equal, or the function must convert them with `Result.mapError`.
- **Panics abort the program.** `panic("message")`, failed `assert`, out-of-bounds indexing and debug-mode overflow all panic. A panic cannot be caught.
- There are no exceptions.

```reason
export let readConfig: (~path: string) => result(config, configError) uses io, alloc =
  (~path) => {
    let text = try Result.mapError(File.readText(~path), (e) => IoFailed(e));
    let config = try Config.parse(text);
    Ok(config)
  };
```

## 6. Memory model

### Values and ownership

- Every value has one or more owners: variables, fields, array slots, closure captures.
- Boxed values (records, variants with payloads, arrays, strings, closures) live on the heap with a reference count. Scalars and small unboxed records live inline.
- The compiler inserts reference count increments and decrements. It removes the ones it can prove unnecessary, as in Koka's Perceus.
- A value is freed when its count reaches zero. Freeing is deterministic, so file handles and sockets close at the point their last owner drops them.

### In-place updates

- An update such as `{...u, age: 31}`, `Array.set(xs, i, v)` or `Array.push(xs, v)` returns a new value.
- When the reference count of the input is 1, the compiler writes into the input instead of allocating. This is called reuse.
- Code that keeps one owner gets in-place mutation without writing any mutation.

```reason
let mut xs = Array.make(0, ~len=1000);
for i in 0 to 999 {
  xs = Array.set(xs, i, i * i);  // count is 1, so the write is in place
};
```

### Cycles

- Immutable values cannot form cycles, so reference counting frees everything without a cycle collector.
- There are no mutable reference cells in safe code. A mutable local (`let mut`) cannot be captured or stored, so it cannot create a cycle.
- **Graphs** use `Arena('a)` and `Handle('a)` from `std/arena`. A handle is an index plus a generation number. `Arena.get(arena, h)` returns `None` when the slot was freed or reused. Handles can point anywhere in the arena, including in cycles, and the arena frees all its slots at once.

### Threads

- Values sent to another thread with `Thread.spawn` or `Channel.send` are marked shared. Shared values use atomic reference counts. Unshared values use plain counts.
- An in-place update on a shared value copies first, so threads never see each other's writes.
- There is no shared mutable state in safe code. `std/sync` provides `Mutex('a)` and `Atomic('a)` for the cases that need it.

## 7. Effects

An effect marks something a function can do beyond computing a value. The set is fixed:

| Effect | Meaning |
| --- | --- |
| `alloc` | May allocate heap memory |
| `io` | May perform input or output, read the clock, or read the environment |
| `unsafe` | Calls code that the compiler cannot check (section 8) |

Rules:

- Exported functions declare their effects after the return type: `(string) => int uses alloc`.
- The compiler infers effects for private functions and lambdas.
- A function's effects include the effects of everything it calls. Calling an `io` function from a function that does not declare `io` is a compile error.
- Panics are not an effect. Any function can panic.
- `main` has every effect.

### Effect variables

A function that calls a function it receives has that function's effects. An effect variable states this in a signature.

```reason
export let map: (array('a), ('a) => 'b uses 'e) => array('b) uses alloc, 'e = ...;
```

- An effect variable is written like a type variable, but appears only after `uses`.
- **It is bound by a parameter.** An effect variable in the result `uses` list must also appear in a parameter's function type.
- At each call, the variable takes the effects of the function passed in. `map(xs, String.length)` has `alloc`. `map(paths, File.readText)` has `alloc, io`.
- The compiler infers effect variables for private functions and lambdas. Only exported higher-order functions write them.

## 8. Unsafe code and C interop

- **`unsafe { ... }`** blocks allow raw pointers (`ptr('a)`), pointer arithmetic and unchecked indexing. A function containing an `unsafe` block has the `unsafe` effect unless it is marked `@trusted`, which states that the author checked it by hand.
- **`external`** declares a C function: `external strlen: (ptr(u8)) => u64 uses unsafe = "strlen";`
- **Layout attributes.** `@repr(C)` gives a record C struct layout. `@unboxed` stores a record inline instead of on the heap. `@unboxed` records cannot be recursive.
- `lang audit` lists every `unsafe` block and `@trusted` function in a package.

## 9. Top-level values

- A top-level `let` must be a constant expression: literals, constructors, records and arrays of constants, and calls to functions marked `@const`.
- `lazy expr` defers a computation until first use. The first `Lazy.force(x)` computes the value and later calls return it. Forcing is thread safe.
- Modules have no initialization code. Program startup runs nothing before `main`.

## 10. Code generation

- **Polymorphic functions are compiled once.** A value of a type parameter is passed as one machine word in a uniform form.
- **Uniform form.** The lowest bit tells generic code what the word holds. A clear bit means a heap pointer, which needs reference counting. A set bit means an inline value.

| Type | Uniform form |
| --- | --- |
| Boxed values (records, variants with payloads, arrays, strings, closures) | Pointer, unchanged |
| `i8` to `i32`, `u8` to `u32`, `bool`, `char`, `f32`, `unit`, constructors without payloads | Inline, shifted left one bit with the tag bit set |
| `i64`, `u64`, `int`, `f64` | Copied into a heap box with a reference count |
| `@unboxed` records | Copied into a heap box with a reference count |

- **Boxing happens only at the boundary.** Non-generic code uses raw C types. The compiler converts a value to the uniform form when it passes the value into generic code, and converts it back when it receives one.
- **`@specialize`** on a function asks the compiler to emit a separate copy for each concrete type it is called with. Use it in hot paths only.
- **Output is C.** One `.c` file per module, plus a small runtime (allocator, reference counting, panic handler, arena).
- **Debug builds** use a fast C compiler such as TCC with checks on.
- **Release builds** use Clang or GCC with `-O2`.

## 11. Build model

- The compiler reads a module's imports before anything else and builds the module graph from them.
- Each module produces an interface file (`.langi`) with its exported signatures. A module type-checks against its dependencies' interface files only.
- When a module changes but its interface file stays the same, modules that import it are not re-checked or recompiled.
- Modules with no dependency between them compile in parallel.
- Builds are deterministic. The same inputs produce byte-identical C and binaries.
- Compile-time evaluation is limited to constant expressions and `@const` functions.

## 12. Tests and contracts

- **`test("name") { ... }`** blocks sit next to the code. `lang test` runs them. They are removed from non-test builds.
- **Contracts.** `requires` and `ensures` clauses on functions are checked in debug builds and removed in release builds.

```reason
export let sqrt: (f64) => f64 =
  (x) => requires x >=. 0.0 ensures result >=. 0.0 {
    Float.sqrt(x)
  };
```

## 13. Tooling for agents

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
    {"title": "Add `io` to the signature", "edits": [{"file": "src/app/config.lang", "range": {"start": [8, 52], "end": [8, 52]}, "text": ", io"}]}
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

## 14. Packages (sketch)

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

## 15. Deferred

These are left out of the first version on purpose.

- **Async.** Threads and channels are the only concurrency. When programs need many cheap tasks, an `async` effect (section 7) is the preferred direction, because it keeps one library for sync and async callers.
- **Functors.** See `DECISIONS.md`, D6.
- **A native debug backend.** See `DECISIONS.md`, D10.
