# Narra specification (draft)

This document defines the Narra language. Each section states the rules first, then shows an example. `DECISIONS.md` explains why each rule was chosen over the alternatives.

File extension: `.nra`. Compiler command: `narra`.

Code samples leave out the imports of `Array`, `Map`, `Range` and `Box` functions (section 3): a sample that calls `Arrays.push` assumes `import * as Arrays from "std/array"`.

## 1. Overview

- Narra uses TypeScript-style syntax with Rust-style semantics: static types, enums with fields, pattern matching, and no `null`.
- Narra compiles to native code, the same target as C, Rust and Go. The compiler emits C, and a C compiler builds the final binary.
- Narra code is a set of modules. One file is one module. There are no globals.
- Every name holds its own value: a change through one name never shows through another. Safe code has no pointers or references, only values.
- A function changes a caller's data only through a `mutable` parameter, marked at the call: `damage(mutable hero, 10)`. It changes the contents (fields and elements), never the caller's variable itself.
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
    Shape.Circle(r) => Float.PI * r * r
    Shape.Rect(w, h) => w * h
  }
}

test "rect area" {
  assert(area(Shape.Rect(2.0, 3.0)) == 6.0)
}
```

## 2. Syntax

The syntax follows TypeScript where the meaning is the same. Where Narra means something different, it uses different syntax or reports a compile error. The table under "Differences from TypeScript" lists those cases. One difference has no diagnostic: assignment copies a value, so after `const b = a`, a change to `a` does not show in `b` (D81).

### Statements and layout

- Blocks use `{ }`.
- **No semicolons.** A line break ends a statement when the line could end there. A line that ends with an operator, `,`, `(`, `[` or `{` continues on the next line. A `;` is a compile error, except between the parentheses of a `for (init; condition; step)` loop (section 2, Control flow).
- `narra fmt` defines the one valid layout. CI is expected to run `narra fmt --check`.
- `narra fmt` reads a `[fmt]` table of `narra.toml` (the nearest one above the file). Keys, with their defaults:
  - `printWidth = 100`: the line width.
  - `tabWidth = 2`: spaces per indent level.
  - `useTabs = false`: indent with tabs.
  - `trailingComma = "all"`: a trailing comma in broken lists, or `"none"`.

  There are no other options.
- Comments: `//` to the end of the line, `/* ... */` for blocks, `///` before a declaration for documentation.

### Bindings

- `const x = expr` binds a name that cannot be reassigned. **`const` fixes the name, not the contents** (D81): the fields and elements of a local `const` struct, array, map or other contents type (section 6, Mutable parameters) can be assigned, and the `const` can be passed as `mutable`.
- `let x = expr` binds a name that can also be reassigned with `x = expr`. A closure passed straight to a non-escaping function parameter can capture it, by reference (section 2, Closures).
- **Every binding holds its own value.** `const other = hero` copies `hero`, so a later change to either one does not show in the other. The one exception is a `for (mutable item of items)` binding, which reaches an element of `items` (section 2, Control flow).
- **A plain parameter is read-only.** Its fields and elements cannot be assigned, and it cannot be passed as `mutable` (section 6).
- **Bindings that copy part of another value are read-only.** A `match` or `is` pattern binding (`x is Some(v)`), a `for (const p of players)` binding and a destructured binding (`const { a, b } = s`) hold a copy, so their fields and elements cannot be assigned and they cannot be passed as `mutable`. `for (const p of players) { p.hp = 0 }` is a compile error: write `for (mutable p of players)` (section 2, Control flow), write through the array, `players[index].hp = 0`, or make a copy first with `copy(p)` (D81, D94). A `mutable` marker exists only in that loop (D85); `x is Some(mutable y)`, `const { mutable a } = s` and a local `mutable name = place` do not.
- **Top-level constants never change** (D81). A top-level `const` cannot be changed at any depth (section 10).
- There are no globals. Assigning to an undeclared name is a compile error.
- **Compound assignment (D76).** `x += e` and the forms `-=` `*=` `/=` `%=` `&=` `|=` `^=` `<<=` `>>=` mean `x = x op e` with the target place evaluated once: `items[next()] += 1` calls `next` once. They are statements with no value, valid on every target `=` accepts (a `let`, a field or element of a changeable value, the contents of a `mutable` parameter). `x++` and `x--` are statements for an integer of any width; the prefix `++x` is a compile error (E0964). The value must have the type of the target (E0962), and `++` on a non-integer is E0963. Wrapping, division by zero and shift checks are those of the binary operator. The wrapping operators `+%=`, `-%=` and `*%=` do not exist. On a `string` target only `+=` exists and it joins strings (`+` alone is still an error). A Map index target (`counts[key] += 1`) panics when the key is missing, since the old value is needed (reading `counts[key]` gives an `Option`); write `Maps.set` first for a new key.
- **No shadowing inside a function.** A name cannot be declared again while an earlier binding with that name is in scope.
- **Unused names are errors.** An unused local, parameter, import or private function is a compile error. Name it `_` or start it with `_` to mark it as unused on purpose.
- **Destructuring:** `const { host, port: serverPort } = config` for structs (section 4), and `const #[quotient, remainder] = Int.divmod(a, b)` for tuples.

### Functions

- `function name(a: int, b: int): int { ... }` declares a function. A function is private to its module unless declared with `export function` (section 3). A function cannot be declared inside another function (E0930): declare it at module level as a private function and pass the enclosing values as arguments. A nested declaration in TypeScript captures the enclosing locals, and a module-level copy could read a different name where a local shadows it (D74). To keep a helper that captures, bind an arrow function to a `const`.
- A function with no return type returns nothing.
- **A function with a result must return on every path.** A path ends at `return`, a call to `panic`, an `if`/`else` whose branches all end, a `match` whose arms all end, or `while (true)` / `for (;;)` with no `break` out of it. Any other path that reaches the closing brace is a compile error that names the path (E0969, E0970 for an arrow function with a block body). Generators have no result value and are exempt (D92).
- **Anonymous functions are arrow functions:** `(x: int): int => x + 1`, or `(x: int): int => { ... }` with a block. `function` is only used for named declarations.
- **`function*`** declares a generator and **`async function`** declares an async function (section 8). Arrow functions can be `async`: `async () => { ... }`.
- An arrow function passed as an argument takes its types from the parameter it is passed to. The types can be left out: `Arrays.map(prices, (p) => p * 2)`.
- **Function types** are written like arrow functions: `(item: T) => U`.
- **Parameters are constant.** A parameter cannot be reassigned, and its contents cannot be changed unless it is declared `mutable` (section 6). A `mutable` parameter still cannot be reassigned.
- **A call passes every argument.** Passing fewer or more arguments than the function declares is a compile error.
- **Arguments are evaluated left to right**, in the order they are written. An argument that is a block expression, a `match` or a closure call can assign a `let` that an earlier argument reads; the earlier argument keeps the value it had when it was evaluated. After `show(items, { items = ["z"]; return 0 })`, `show` received the original `items`.
- **Named arguments.** A call can name its arguments, written like struct literal fields: `move(x: 1, y: 2)`. Positional arguments come first, then named ones in any order: `move(1, z: 3, y: 2)`. A positional argument after a named one is an error. Each parameter is passed exactly once, by position or by name. `mutable` follows the colon: `fill(buffer: mutable items, value: 0)`. Names work when the callee is a declared function or a trait function. A function value, an enum constructor, and a std function built into the compiler such as `Str.slice` take positions only (D87). The arguments are evaluated in the order they are written and passed in the order of the parameters. A parameter's name is part of the function's interface: renaming it breaks the calls that use the name.
- **No currying and no partial application.** To fix some arguments, write an arrow function.
- **No method calls.** `xs.push(v)` is not valid. Write `Arrays.push(mutable xs, v)`. `value.field(args)` is only valid when `field` is a struct field that holds a function.
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
- A closure captures constants (`const` bindings and non-mutable parameters) by value, except a local `const` whose contents it changes (next item).
- **A non-escaping closure captures `let` variables and `mutable` parameters by reference** (D77). An arrow function written directly as the argument of a non-escaping function-type parameter may read and assign them, as in TypeScript:

<!-- context
function each(items: Array<int>, f: (item: int) => void) {
  f(items[0])
}
---
const numbers = [1, 2]
-->
```ts
let total = 0
each(numbers, (item) => { total = total + item })
```

  Every `let` and `mutable` it uses is captured by reference, also when it only reads it. A local `const` whose fields or elements it assigns, or which it passes as `mutable`, directly or through a nested closure, is captured by reference too, by every closure between the change and the declaration, and is borrowed like a `let` it assigns (D81). A `const` that no closure in the chain changes stays a copy. A closure that is not of this kind cannot change a captured value. Every arrow function between the use and the declaration must be of this kind. A closure that is stored, returned, passed to an `escaping` parameter or through a function value, or that sits inside one of those, cannot capture a `let` or a `mutable` parameter (E0612, E0613).
- **An escaping closure shares changing state through a `Cell`** (D90). `const count = Cells.create(0)` is captured by value like any constant, and every copy of the handle reads and writes the same slot: `Cells.update(count, (value) => value + 1)`, `Cells.get(count)`. Section 8 describes `Cell<T>`.
- **A variable that such a closure assigns is borrowed for the call.** In that call it cannot be passed as `mutable`, read or assigned in another argument (including `items.length` or `xs[0]`), or captured by another closure argument (E0614). Two closures that only read the same variable are fine. A narrowing by `is` is not seen inside a closure that captures the variable, and a narrowing of a variable that a closure assigns ends at the statement of the call.
- **Function-type parameters cannot escape by default.** Inside the called function, such a parameter can only be called. It cannot be stored, returned, captured by another closure, or sent to a thread.
- **A non-escaping closure costs no allocation.** Its environment is placed on the caller's stack, and nothing is reference counted or freed.
- **`escaping`** marks a function-type parameter that may be stored or outlive the call: `function onClick(escaping handler: () => void)`. An escaping closure's environment is allocated on the heap.
- Struct fields and array elements of function type always hold escaping closures. `Thread.start`, `Async.spawn` and `Async.run` take escaping parameters.
- **`@inline`** on a function copies its body into each caller. On a higher-order function such as `Arrays.map`, this also places the closure's body at the call, which removes the indirect call.

```ts
@inline
export function map<T, U, effect E>(items: Array<T>, f: (item: T) => U uses E): Array<U> uses alloc, E {
  // f cannot escape, so the closure passed in needs no allocation
}

const scaled = Arrays.map(prices, (p) => p * rate)    // no closure allocation, no indirect call
```


### Control flow

- `if (cond) { ... } else if (cond) { ... } else { ... }`. The condition must have type `bool`. There is no truthiness: `None`, `0` and `""` are not conditions.
- `while (cond) { ... }`.
- `for (const item of iter) { ... }` loops over an iterator. `for (const #[i, item] of Arrays.entries(items))` gives indexes from 0.
- **`for (mutable item of items)` changes the elements in place** (D85). `item` reaches each element of the array, as a `mutable` parameter would, so `item.hp = 0` changes `items`. `for (const #[index, mutable item] of Arrays.entries(items))` also gives the index, and `Arrays.values(items)` may be written for `items`.
  - The source is a local place of type `Array<T>`: a local `const` or `let`, a `mutable` parameter, or a field or element path of one. A call result, a plain parameter, a top-level constant, a map, a range and a generator are compile errors.
  - `T` has contents (section 6, Mutable parameters): `for (mutable n of numbers)` over `Array<int>` is a compile error; write `numbers[index] = ...` in an `Arrays.entries` loop.
  - The binding follows the rules of a `mutable` parameter: its fields and elements can be assigned and it can be passed as `mutable`, but it cannot be assigned whole or escape.
  - The source is borrowed for the whole body: the body cannot read, assign, pass or capture the source or any path that overlaps it (`players`, `players[0]`, and `game` for a source `game.players`; `game.round` is fine).
  - The array is made unique before the first iteration (a shared array is copied once), and the loop takes no reference to it.
  - Not allowed in a `function*` or `async` body yet.

<!-- context
struct Player {
  hp: int
}
---
let players: Array<Player> = []
-->
```ts
for (mutable player of players) {
  player.hp = 0                // changes players
}
```

- **An iterator is a `Generator<T>`** (section 8). `for ... of` iterates over its own copy of the generator, so a generator held in a variable is not advanced. `Arrays.values`, `Arrays.entries`, `Maps.keys`, `Maps.entries`, `Str.chars` and the `Range` functions return generators. A function returns its own iterator by returning a generator.
- **Counting loops use `Range`.** `Ranges.upTo(n)` counts from 0 to `n - 1`. `Ranges.between(start, end)` counts from `start` to `end - 1`. `Ranges.stepped(start, end, step)` takes a step: a positive step counts up while the value is below `end`, a negative step counts down while it is above `end`, and a step of 0 panics. The loop stops before the value would pass the largest or smallest `int`.

<!-- context
import * as Arrays from "std/array"
import * as Ranges from "std/range"
---
const items = ["a", "b"]
-->
```ts
for (const i of Ranges.upTo(Arrays.length(items))) {
  print(items[i])
}
```

- `break` and `continue` work as in TypeScript, including labels: `outer: for (...) { ... break outer }`.
- **C-style `for`.** `for (init; condition; step) { ... }` has TypeScript's meaning, and `for (;;)` loops forever (D76).

<!-- context
const limit = 3
let total = 0
-->
```ts
for (let i = 0; i < limit; i = i + 1) {
  total = total + i
}
```

  - `init` is a `let` declaration or empty. Its name is scoped to the loop. `const` is error E0960, because the step could not change it.
  - `condition` has type `bool`, with no truthiness (error E0959), or is empty.
  - `step` is an assignment, a call or empty (error E0961 for anything else).
  - `continue` runs the step, then the condition. Labels work as for the other loops.
  - The semicolons belong to the parentheses. A `;` anywhere else is still an error.
  - The loop is a statement, not a value.
- There is no `do ... while`, no `switch` and no `goto`.
- `match` replaces `switch` (section 4).

#### Block expressions and `if` as a value

A block has two forms, told apart by position (D75). A block in statement position is a block statement. A block in value position is a block expression: after `=`, as a call argument, as an operand, as a `return` value, inside `${...}`, or as an element of an array, tuple or struct literal.

<!-- context
function lookup(code: int): string {
  return "name"
}
---
const code = 1
-->
```ts
const label = {
  const name = lookup(code)
  if (name == "") {
    return "unknown"
  }
  return name
}
```

- **The value comes from `return`.** Narra never returns implicitly, so the last expression of a block is not its value.
- **`return` stops at the nearest block expression**, as at a function body. A `return` anywhere inside it, also in nested block statements, loops and `match` arms, ends the block expression and not the function. A block expression cannot return from the enclosing function.
- **`break` and `continue` cannot leave a block expression.** A `break` or `continue` to a loop outside it is a compile error (E0932). Loops inside the block work as usual.
- **Every path must end in `return value`** (E0931), and every returned value has the same type. A bare `return` is an error. A block expression with no `return` at all, such as `const o = { a }`, is a compile error (E0933) that points to struct literals.
- A `{` that starts `{}`, `{ ...s, f: v }` or `{ name: value }` in value position is still a struct literal whose type comes from context. The exception is `{ name: for` and `{ name: while`, where `name:` labels a loop and the `{` starts a block expression.
- `if (c) { ... } else { ... }` in value position is an expression whose branches are block expressions, each ending in `return`. The `else` branch is required, and `else if` chains are allowed. At the start of a statement `if` is still a statement.
- `cond ? a : b` is a conditional expression, as in TypeScript. `cond` must be `bool` (no truthiness), `a` and `b` have the same type, and only the chosen branch runs. It binds looser than `||` and is right-associative: `a ? b : c ? d : e` is `a ? b : (c ? d : e)`. A missing `:` is E0934. The line rule above applies: a line that starts with `?` or `:` is a new statement, so a conditional expression that spans lines is wrapped in parentheses, where line breaks are ignored, and the Prettier layout works inside them: `const v = (cond` / `? a` / `: b)` on three lines. Without the parentheses the compiler reports E0934 and says to wrap it. A conditional expression is not a constant (a top-level `const` cannot use one), its value must be used (as a statement it is E0934; write an `if`), and both branches need a value. It is a block expression holding an `if` whose branches each `return` one value, so a self call in a returned branch (`return n < 0 ? f(n + 1) : n`) is still a tail call.

<!-- context
const n = 1
-->
```ts
const sign = if (n < 0) { return "negative" } else if (n == 0) { return "zero" } else { return "positive" }
```

- A self call in a `return` inside a block expression is not a tail call of the function.

### Literals

- **Strings** use double quotes: `"text"`. They have no interpolation.
- **Template strings** use backticks: `` `Hello ${name}` ``. Each `${...}` must hold an expression of type `string`. Template strings can span several lines. They take the same escapes as strings: `\n`, `\t`, `\r`, `\0`, `\\`, `\"`, `\'`, `` \` ``, `\$` and `\u{<hex>}`. Any other `\X` is a compile error, reported at the backslash. `\${` is a literal `${`, not an interpolation.
- **Characters** use single quotes: `'a'`, `'\n'`, `'\u{1F600}'`. A character literal has type `char`. `'abc'` is a compile error.
- **Integers:** `42`, `0xFF`, `0o755`, `0b1010`. `_` can separate digits: `1_000_000`.
- **Floats:** `3.14`, `1.0e-9`, `1_000.5`.
- **A literal must fit its type.** `const x: u8 = 300` is a compile error.
- `true` and `false` have type `bool`.
- A number literal with no other type information is `int` when written without a decimal point, and `f64` when written with one.
- **Arrays:** `[1, 2, 3]`. `[]` needs its type from context, as in `let xs: Array<int> = []`. Arrays have no literal form with a type argument.
- **Maps:** `Map { "a": 1, "b": 2 }`. A key that is not a literal is written `[key]: value`. The key and value types come from the entries or from context. An empty `Map {}` needs its type from context.
- **Type arguments on a literal.** `Map<string, int> {}`, `Map<string, int> { "a": 1 }` and `Pair<int, string> { a: 1, b: "x" }` write the type arguments after the name. The literal then has that type without help from context, so `count(Map<string, int> {})` needs no annotation. The compiler checks the arity, that each argument is a type, the map key rules (no float keys, `Equatable` and `Hashable` together) and every key, value and field against the written types. If an expected type also exists, the two must be the same type. A `<` after a capitalized name starts type arguments only when it is followed by a balanced `<...>` of type names, dotted names and tuple types and then `{`; in every other case it is the less-than operator, so `a < b` is unchanged. Function types are not accepted inside the arguments; use an annotation for those. There is no `{K: V}` type and no `#{...}` literal.
- **Tuples:** `#[1, "a"]` builds a tuple of type `#[int, string]`. It always makes a tuple, needs no expected type and has at least two elements. `[1, 2]` is always an array.

There are no macros, no operator overloading and no function overloading.

### Operators

- **Operators are built in.** Programs cannot define operators or change their meaning.
- **Both operands have the same type.** There are no implicit conversions, so `1 + 2.0` is a compile error.

| Operators | Operand types |
| --- | --- |
| `+` `-` `*` `/` `%` | Any integer type or any float type. `/` on integers truncates toward zero. `+` `-` `*` on integers wrap on overflow |
| `&` `\|` `^` `<<` `>>` | Integer types. `^` is xor |
| `~` | Integer types, bitwise not |
| `<` `<=` `>` `>=` | Numbers, `char` and `string` |
| `==` `!=` | Types that implement `Equatable`, compared structurally (section 4) |
| `&&` `\|\|` `!` | `bool` |

- There is no string concatenation operator. Use a template string or `Str.concat(a, b)`. The statement `s += t` joins strings (section 2, Bindings).
- Compound assignment (`+=` and the others) and `x++` / `x--` are statements, not operators, and have no value (section 2, Bindings).
- There is no length operator. Use `Arrays.length(items)`, `Str.length(s)` and `Maps.size(m)`.
- There is no power operator. Use `Float.pow(x, y)` or `Int.pow(x, y)`.
- `Int.floorDiv(a, b)` rounds toward negative infinity.

### Differences from TypeScript

Code that is valid TypeScript but means something different in Narra is a compile error with a fix, never a silently different meaning. The exception is value copies (the `const b = a` row): a copy followed by a change to one of the names compiles, because finding the change would need a whole-function check of every copy (D81).

| TypeScript | Narra |
| --- | --- |
| `const b = a` gives two names for one object, and a change through one shows through the other | `const` fixes the name, and its contents can change, as in TypeScript (D81). `const b = a` makes a copy: a change to `a` does not show in `b` |
| `==` converts types, `===` compares strictly | `==` compares structurally with no conversion. `===` is a compile error |
| `let` can be captured by any closure | `let` can be captured only by a closure passed to a non-escaping parameter; an escaping closure shares changing state through a `Cell` |
| `null`, `undefined` | Not present. Use `Option<T>` with `None` |
| `T[]` is an array type | Not present. Write `Array<T>` (D23) |
| `number` | Not present. Use `int`, `i32`, `f64` and the other sized types |
| `enum` holds named numbers or strings | `enum` holds variants with fields, checked by `match` |
| `'text'` is a string | `'a'` is a `char`. Strings use `"text"` |
| `for (const k in obj)` loops over keys | Not present. Use `for (const k of Maps.keys(m))` |
| `for (init; cond; step)` accepts `var`, `const`, several declarations and any condition | The same form, with `init` a single `let` or empty (`const` is E0960), `condition` a `bool` (E0959) and `step` an assignment or a call (E0961) |
| `import X from "m"` imports a default export | Not present. Use `import * as X from "m"` |
| `xs.push(v)`, `s.length` | Not present. Use `Arrays.push(mutable xs, v)`, `Str.length(s)` |
| `a + b` joins strings | Not present. Use a template string (`s += t` is allowed as a statement) |
| `y = x++`, `++x`, `a[i] = b += 1` | Compound assignment and `x++` are statements without a value; `++x` is an error (E0964) |
| `class`, `this`, `new`, `interface` | Not present. Use `struct`, `enum` and `trait` |
| `try`, `catch`, `throw` | Not present. Use `Result` (section 5) |
| Calling an async function without `await` starts it in the background | A compile error. Use `await f()`, or start a task with `Async.spawn` (section 8) |
| `const o = { a }` builds an object | Not present. `{ a }` in value position is a block expression, which needs `return` (E0933). Write a struct literal: `Point { x: 1 }` |
| `Promise<T>` | Not present. `Async.spawn` returns a `Task<T>` |
| `async function*` | Not present in the first version |
| A `function` declared inside a function | A compile error (E0930). Declare it at module level as a private function and pass the enclosing values as arguments |

## 3. Modules and imports

- **One file is one module.** Its path from the package root is its name: `src/app/text.nra` is `"app/text"`.
- **Names are private by default.** `export` makes a top-level declaration visible to other modules: `export function`, `export const`, `export type`, `export struct`, `export enum`, `export trait`.
- **Imports are explicit.** A module sees only its own declarations, the names it imports, and the prelude.
- **Prelude.** The prelude holds the types and constructors `Option`, `Some`, `None`, `Result`, `Ok`, `Err`, `Array`, `Map`, `Range`, `Box` and the primitive types, and the functions `print`, `assert`, `panic` and `copy` (section 6, Copies on shared values). Nothing else is in scope without an import. These four are the only global functions.
- **A type in the prelude does not bring its functions.** The functions of `Array`, `Map`, `Range` and `Box` live in `std/array`, `std/map`, `std/range` and `std/box` (D67). `Arrays.push(mutable xs, v)` needs `import * as Arrays from "std/array"`. Without the import the call is a compile error (E0908) whose fix adds the import line. A type the program declares named `Box` or `Map` hides the prelude one.
- **Std aliases are plural by convention.** The std modules of the capitalized prelude types are imported as `Arrays`, `Maps`, `Boxes`, `Ranges`, `Options` and `Results` (D72). Samples and fixes use these names. Any other alias that follows the rules below works too.
- **Standard modules.** `std/array`, `std/box`, `std/bytes`, `std/cell`, `std/char`, `std/env`, `std/f32`, `std/float`, `std/fs`, `std/hash`, `std/int`, `std/i8`, `std/i16`, `std/i32`, `std/u8`, `std/u16`, `std/u32`, `std/u64`, `std/io`, `std/json`, `std/map`, `std/math`, `std/option`, `std/path`, `std/process`, `std/random`, `std/range`, `std/result`, `std/shape`, `std/string`, `std/test` and `std/time`.
- **`std/shape` and `std/json` (D83).** `std/shape` declares `export enum Value { Null, Bool(value: bool), Int(value: int), Float(value: f64), Text(value: string), List(items: Array<Value>), Record(fields: Map<string, Value>) }`, a dynamic tree for structured data. `std/json` has `parse(text: string): Result<Value, string> uses alloc` and `stringify(value: Value): string uses alloc`. Write `import * as Json from "std/json"` and `import { Value } from "std/shape"`. `parse` follows RFC 8259: a number without fraction or exponent that fits `int` is `Value.Int`, every other number is `Value.Float`; a duplicate key keeps its first place and takes the last value; escapes and surrogate pairs decode, a lone surrogate is an error; nesting deeper than 512 is an error. Errors read `<line>:<column>: <reason>`. `stringify` is compact, writes non-ASCII as UTF-8 and a non-finite `Float` as `null`, and a whole `Float` keeps `.0`.
- **`std/math`, `std/time`, `std/random` and `std/hash`** are Narra source over the primitives of `std/internal/math` (the functions of C's `math.h`), `std/internal/clock` and `std/internal/entropy`.
  - `std/math` has the constants `PI` and `E`; `abs` and `absFloat`; `min`, `max` and `clamp` for any `Ordered` type, so for `int` and `f64`; `pow` and `powInt` (a negative exponent panics, overflow wraps like `*`; use `Int.mulChecked` to detect it); and `sqrt`, `floor`, `ceil`, `round` (half away from zero), `trunc`, `sin`, `cos`, `tan`, `atan2`, `exp` and `log`, all on `f64`. `clamp` panics when `low > high`.
  - `std/time` has `Instant` and `Duration` (both a count of nanoseconds). `now()` reads the monotonic clock, `elapsed(since)` reads it again, `between(earlier, later)` needs no clock, `unixMillis()` is the wall clock, `sleep(duration)` sleeps. `milliseconds`, `seconds` and `nanoseconds` build a `Duration`, `asMilliseconds` and `asNanoseconds` read it. Reading a clock and sleeping have the `io` effect (section 7).
  - `std/random` is SplitMix64: a `Random` holds one `u64` state, the same seed gives the same numbers on every run and platform, and a value is copied like any struct. `seeded(seed)` and `fromSystem() uses io` (seeded from `/dev/urandom`) make one; `nextInt(mutable random, low, high)` returns an `int` in `[low, high)` without modulo bias and panics when the range is empty; `nextFloat` returns an `f64` in `[0, 1)`; `nextBits` returns 64 random bits; `shuffle(mutable random, mutable items)` is a Fisher-Yates shuffle. It is not for cryptography.
  - `std/hash` has `Hash.of(x)`, the `Hashable` hash of any value (it may change between builds), and `fnv1a(text)`, FNV-1a 64 over the UTF-8 bytes of a string as a `u64`, which never changes. `of` is a keyword, and is accepted as a function name and after `.`, like `from`.
- **`std/internal/` is for the standard library.** A std module written in Narra reaches the C primitives it needs through modules under `std/internal/<name>`. Only a module whose path starts with `std/` may import them. Any other import is an error (E0426): `std/internal/<name> is internal to the standard library`. The error list of unknown modules does not name them.
- **Paths are package rooted.** `"app/text"` for this package, `"std/string"` for the standard library, `"pkgname/module"` for a dependency. Relative paths are not allowed.

```ts
import * as Str from "std/string"            // namespace import
import * as Arrays from "std/array"          // plural alias for a prelude type's module
import { classify, Token } from "app/text"   // named imports
import { map as mapArray } from "std/array"  // a named import with a new name
```

There is no re-export statement. To expose an imported name, import it and export a new declaration:

```ts
import * as Text from "app/text"

export const classify = Text.classify        // a function
export type Token = Text.Token               // a type, with its constructors
```

Rules:

- **Imports are static.** They appear at the top of the file with a string literal path. They never run code and never load a module at runtime. `import(...)` as an expression is a compile error.
- **An import alias may not equal a type name.** `import * as Shape` in a module that declares or imports a type `Shape`, or `import * as Array` (a prelude type), is a compile error, as in TypeScript. `Shape.Circle` then always names a constructor of `Shape`. The fix suggests the plural (`Arrays`). Two imports with the same alias are an error too.
- **Named imports can take a new name.** `import { map as mapArray } from "std/array"` makes the function `mapArray` in this module. The new name follows the rules of every other name in the module: it may not clash with a declaration, another import or a type name. `import { Teacher }` of an enum constructor is a compile error whose fix imports the enum instead.
- **No cycles.** The import graph must be acyclic. The compiler reports the cycle as a list of modules.
- **No re-export statement and no default exports.** `export { x } from "m"`, `export *` and `export default` are compile errors. An exported name is always declared in the exporting file.
- **`export type A = B`** declares an alias. When `B` is an enum, its constructors are usable through the alias.
- **`export opaque struct`** and **`export distinct type`** export a type while hiding its contents (section 4).
- **No import side effects.** Top-level declarations must be functions, types, traits, `const` constants or `lazy` values (section 10). There is no top-level `let`.
- **Constructors come with their enum and are always qualified.** Importing `Token` makes `Token.Word(x)` usable, in expressions and in patterns (D71). A bare `Word(x)` is a compile error whose fix writes `Token.Word(x)`. `Some`, `None`, `Ok` and `Err` are the exception: they are prelude names and stay bare. Constructors of different enums never clash, whatever their fields, and a user enum may have a constructor named `None` (`Maybe.None`); a bare `None` is always the prelude one. Two constructors with the same name in one enum are an error.
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
- **Enums.** Declared with `enum`. Each constructor can carry named fields, and is written with its enum's name: `Token.Word("hi")`, `Token.Space` (section 3).

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

<!-- context
struct ServerConfig {
  host: string
  port: int = 8080
  tags: Array<string> = []
}
---
-->
```ts
const config = ServerConfig { host: "localhost" }            // port is 8080, tags is []
const loopback: ServerConfig = { host: "127.0.0.1", port: 3000 }
```

<!-- context
struct Entry<K, V> {
  key: K
  value: V
}
-->
```ts
function makeEntry(key: string, value: int): Entry<string, int> {
  return { key: key, value: value }                          // type taken from the return type
}
```

**Reading and updating**

- `s.field` reads a field.
- `{ ...s, field: value }` returns a copy of `s` with some fields changed. The spread must come first, there is only one, and the result has the same type as `s`. The original is unchanged.
- `s.field = value` changes a field in place. `s` must be a local `const` or `let`, a `mutable` parameter, or a field or element of one (section 2, Bindings).
- To change a caller's struct inside a function, declare the parameter `mutable` (section 6).

<!-- context
struct ServerConfig {
  host: string
  port: int = 8080
  tags: Array<string> = []
}
---
const config = ServerConfig { host: "localhost" }
-->
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

<!-- context
import * as Float from "std/float"
struct ServerConfig {
  host: string
  port: int = 8080
}
struct Point {
  x: f64
  y: f64
}
---
const config = ServerConfig { host: "localhost" }
const point = Point { x: 1.0, y: 2.0 }
-->
```ts
const { host, port: serverPort } = config

match (point) {
  { x: 0.0, y: 0.0 } => print("origin")
  { x: 0.0, y } => print(`on the y axis at ${Float.toString(y)}`)
  { x, y } => print(`${Float.toString(x)}, ${Float.toString(y)}`)
}
```

**Storage**

- A struct field, array element or variable holds the struct's fields directly. There is no heap allocation and no reference count for the struct itself.
- Copying a struct copies its fields. Fields stored on the heap, such as strings and arrays, have their reference count incremented.
- `==` compares structs field by field. It is a compile error when a field holds a function.
- **A struct cannot contain itself.** Use an array, an enum or `Box<T>` for recursive data.

**Boxes**

- `Box<T>` holds one value on the heap with a reference count.
- `Boxes.new(value)` creates one. `b.value` reads the value. `b.value = v` replaces it when `b` is a local `const` or `let` or a `mutable` parameter, in place when the count is 1.
- Use a box for recursive structs and for large structs that are copied often.

<!-- context
import * as Boxes from "std/box"
-->
```ts
struct ListNode {
  value: int
  next: Option<Box<ListNode>> = None
}

const list = ListNode { value: 1, next: Some(Boxes.new(ListNode { value: 2 })) }
```

### Opaque types

`export opaque struct` exports a struct's name but not its fields.

```ts
// module "app/range"
export opaque struct Range {
  start: int
  finish: int
}

export function create(start: int, finish: int): Result<Range, string> uses alloc {
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

### Type aliases

`type Name = T` gives a type a second name. `export type` exports it.

```
type Tok = Token
type Pair<T> = #[T, T]
export type Token = Text.Token
```

- **The alias is the same type.** `Tok` and `Token` can be used in place of each other, and when the target is an enum, `Tok.Word(x)` builds a value and `Tok.Word(text)` works in a pattern.
- **Generic aliases** take type parameters, and every use gives the arguments. A wrong number of arguments is a compile error. Constructors through a generic alias (`Opt.Some(3)` for `type Opt<T> = Option<T>`) are not available yet.
- **No recursion.** An alias may not refer to itself, directly or through other aliases; recursive types need an enum.
- **Diagnostics** show a struct, enum or distinct type reached through an alias as `Tok (Token)`.
- **Nothing reaches the emitted C.**

### Distinct types

`distinct type` declares a new type with the same representation as an existing one.

```ts
export distinct type UserId = int
export distinct type OrderId = int
```

<!-- context
export distinct type UserId = int
export distinct type OrderId = int
function loadUser(id: UserId) {}
---
-->
<!-- error: E0119 -->
```ts
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
- **`std/option` helpers.** `import * as Options from "std/option"` adds `Options.map(x, f)`, `Options.andThen(x, f)` (`f` returns an `Option`), `Options.withDefault(x, fallback)`, `Options.orElse(x, f)` (`f` takes no argument and returns an `Option`), `Options.isSome(x)`, `Options.isNone(x)` and `Options.okOr(x, error)` (a `Result`). `withDefault` returns `fallback` when `x` is `None`. The fallback is always evaluated; use `orElse` for a lazy one. When a callback returns `None` or `Err(...)` and the type of the call is not known from context, annotate the result (`const y: Option<int> = Options.andThen(x, (v) => None)`). There is no `unwrap` or `expect`: read the value after an `is` test, or give a fallback.
- **Callbacks.** The function arguments of these helpers are non-escaping and take an effect variable, like `Arrays.map` in the effects section: a call uses the effects of its callback, so an `io` callback needs `uses io` on the caller. A capturing arrow function passed to a helper is allocated on the stack.
- **Helpers never return early.** They are expressions. The main style for a path that stops stays `is` and `match`.
- Type syntax uses names, not symbols: `Option<T>`, `Array<T>`, `Map<K, V>`, `Result<T, E>`. There is no `T | null` or `T?`. Tuples are the one exception: `#[A, B]` (see Tuples above).

### Pattern matching

`match` checks a value against patterns. It must cover every case, or end with a `_` arm.

<!-- context
import * as Int from "std/int"
enum Token {
  Word(text: string)
  Number(value: int)
  Space
}
enum Status {
  Active
  Suspended(reason: string)
}
---
const token = Token.Space
const status = Status.Active
-->
```ts
match (token) {
  Token.Word(text) => print(text)
  Token.Number(n) if (n < 0) => print("negative")
  Token.Number(n) => print(Int.toString(n))
  Token.Space => print(" ")
}

const label = match (status) {
  Status.Active => "active"
  Status.Suspended(reason) => `suspended: ${reason}`
  _ => "other"
}
```

- Each arm is `pattern => expression` or `pattern => { statements }`.
- Patterns can bind fields, test literals, and add an `if (...)` guard.
- **Alternatives (D98).** An arm may list patterns joined with `|`: `Token.Plus | Token.Minus => ...`. Alternatives nest inside a pattern: `Some(1 | 2)`. Every alternative binds the same names with the same types (E0967 for different names, E0968 for different types), so `Shape.Circle(r) | Shape.Sphere(r) => r` works. A guard covers the whole list: when it fails for one matching alternative, the next alternative that matches is tried with its own bindings, and the arm is taken if any alternative matches with the guard true (as in Rust). A binding keeps the full type of the value: there is no narrowing to the listed constructors and no union type (D73). Exhaustiveness counts each alternative, and an alternative that the arms and alternatives before it already match is an error (E0974) when it is built from constructors and bools; repeated integer or string literals are not reported. Alternatives are only for `match`; an `is` test and a `const`, `let` or `for ... of` destructuring take one pattern (E0974). `narrac fmt` writes a list on one line when it fits, otherwise one alternative per line, each after the first starting with `|`.
- **`match` as a value.** In value position (after `=`, as an argument, an operand or a `return` value), `match` produces a value. An arm written `pattern => expression` gives the expression. An arm written `pattern => { ... }` is a block expression (see "Block expressions" in section 2): its value comes from `return v`, a `return` inside it ends the arm and not the function, `break` and `continue` cannot leave it, and every path must end in `return`. All arms have the same type. A `match` in statement position keeps block-statement arms, where `return` leaves the function.
- `match` on an `Option<T>` uses `Some(v) =>` and `None =>`.

### Generics

- Type parameters use angle brackets: `function first<T>(items: Array<T>): Option<T>` and `struct Entry<K, V> { ... }`.
- A type parameter can require traits: `<T: Ordered>`, or `<T: Show & Ordered>` for several (see "Traits" below).
- Arithmetic operators do not work on a type parameter. Pass a function or require a trait instead.
- **Structural equality.** `==` compares values structurally. The compiler implements `Equatable` for every type that has no function inside it. Using `==` on a type that contains a function is a compile error.
- **Hashing.** `Hash.of(x)` needs `Hashable`. The compiler implements it for every type that has no function or float inside it.
- Generic code that uses `==` declares `<T: Equatable>`. Code that hashes declares `<T: Hashable>`.
- **Type arguments in a call.** `id<int>(1)`, `Mod.pair<int, string>(a, b)` and `first<Array<int>>(xs)` give the type arguments of a generic function. They fix the type parameters before the arguments are checked, so an argument that does not fit is an ordinary type mismatch. Write one type argument per type parameter or none; a different count is an error, and so are type arguments on a function that is not generic. A `<` after a name starts type arguments only when a balanced `<...>` of type names follows, closed by `>` directly followed by `(`. Every name in it that is not followed by `.` must be a type parameter of the enclosing declaration (so a lowercase `t` works), a primitive type (`int`, `f64`, `bool`, `char`, `string`, `void`), or start with a capital letter. A name of two or more characters with no lowercase letter (`MAX`, `LIMIT_2`) is a constant, not a type, so `a < b > (c)` and `both(x < MAX, MIN > (y))` are still comparisons. The remaining ambiguity is a value that is capitalized but not SCREAMING_CASE, as in `f(x < Limit, Cap > (y))`, which reads as a call with type arguments; parenthesize the second comparison, `(Cap > y)`. Single capital letters (`T`, `K`, `V`) are types. Constructors (`Some<int>(1)`) and builtin functions take no type arguments; annotate the variable instead, as in `const x: Option<int> = None()`. Function types are not accepted inside the arguments.
- **Floats follow IEEE 754 in `==`.** `NaN == NaN` is `false`. A struct that contains `NaN` is not equal to itself.

<!-- context
import * as Arrays from "std/array"
-->
```ts
export function contains<T: Equatable>(items: Array<T>, wanted: T): bool {
  for (const item of Arrays.values(items)) {
    if (item == wanted) {
      return true
    }
  }
  return false
}
```

### Traits

A trait is a named set of functions that a type implements. A type's traits are all listed where the type is declared.

<!-- context
import * as Int from "std/int"
-->
```ts
export trait Label {
  function label(value: Self): string
}

export struct Money implements Label, Ordered {
  cents: int
}

implement Label for Money {
  function label(value: Money): string {
    return `${Int.toString(value.cents / 100)}.${Int.toString(value.cents % 100)}`
  }
}

implement Ordered for Money {
  function compare(a: Money, b: Money): Ordering {
    return Int.compare(a.cents, b.cents)
  }
}

```

<!-- context
import * as Int from "std/int"
export trait Label {
  function label(value: Self): string
}
export struct Money implements Label {
  cents: int
}
implement Label for Money {
  function label(value: Money): string {
    return Int.toString(value.cents)
  }
}
---
const price = Money { cents: 1250 }
-->
```ts
const text = Label.label(price)
```

**Declaring and implementing**

- `trait Name { ... }` declares a trait. Its body lists function signatures. `Self` is the implementing type.
- **A trait has functions only.** No fields, no default implementations, no associated types, and no traits that require other traits.
- **Implementations live with the type.** A struct, enum or distinct type lists its traits with `implements`. Each listed trait needs an `implement Trait for Type { ... }` block in the same module, or a `derives` entry.
- **No implementations anywhere else.** A module cannot implement a trait for a type declared in another module. To add a trait to another module's type, wrap it in a `distinct type` or a struct.
- `derives` asks the compiler to write the implementation: `struct User derives Show, Ordered, Shape { ... }`.

**Calling**

- Trait functions are called through the trait name: `Label.label(price)`. There is no `price.label()` form.
- In generic code, `<T: Show>` allows `Show.show(value)` for a `value: T`.
- A trait function may mention `Self` only in its result, for example `function fromValue(value: Value): Result<Self, string>`. A call names the type that `Self` stands for: `Shape.fromValue<User>(value)`. Exactly one type argument is allowed. Without it, the call is an error unless the expected type fixes `Self`. In a generic body `<T: Shape>`, `Shape.fromValue<T>(value)` is resolved when the body is instantiated.

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
| `Shape` | Conversion to and from the `Value` tree of `std/shape`, for serialization | By the compiler for prelude types, by `derives` or `implement` for a struct or enum |

- A type can replace the automatic `Equatable` or `Hashable` with its own `implement` block.
- Primitive types, `string`, `Array<T>`, `Map<K, V>`, `Option<T>` and `Result<T, E>` implement the standard traits when their contents do.

**Serialization with `Shape`**

- `Shape` (D84) has two functions in `std/shape`: `toValue(value: Self): Value` and `fromValue(value: Value): Result<Self, string>`. Formats read and write `Value`, so `Shape` does not know JSON.
- `Json.encode(value)` is `Json.stringify(Shape.toValue(value))`. `Json.decode<User>(text)` is `Json.parse(text)` followed by `Shape.fromValue<User>`; a parse error is passed through.
- The compiler provides `Shape` for `bool`, `int`, `f64`, `char` (as a one-character `Text`), `string`, `Array<T>`, `Map<string, V>`, `Option<T>` (`None` is `Null`) and tuples of two to four items (a `List`), when their contents implement it. `Value` implements it as itself. A `Map` with other keys, a function type, a sized integer and a longer tuple do not: E0530 names the type that fails. A struct or enum derives it (`derives Shape`) or lists `implements Shape` and writes an `implement Shape for` block.
- Decoding is strict on types and ignores extra data: `int` accepts only `Value.Int`, `f64` accepts `Int` and `Float`. An error reads `<path>: <reason>` from the root, as in `.users[2].age: expected int, found Text`; at the root it is the reason alone.
- `derives Shape` writes the implementation. A struct is a `Record` with one entry per field, in declaration order; a field of type `Option<T>` holding `None` is left out, as `JSON.stringify` leaves out `undefined`. An `Option` field whose default is not `None` (`hint: Option<int> = Some(5)`) writes `None` as `Null` instead, so that it does not decode as its default. Decoding a struct gives a missing field its default (`age: int = 0`) and a missing `Option` field `None`, and decodes `Null` in an `Option` field as `None`; any other missing field is an error (`.age: missing field`). Unknown fields are ignored.
- An enum is a `Record` with a `"type"` entry naming the constructor, then one entry per field, as a TypeScript discriminated union: `{"type":"Circle","radius":1.0}`, `{"type":"Empty"}`. A missing, non-text or unknown `type` is an error naming the enum. A constructor field named `type` would clash, so deriving `Shape` on such an enum is E0532.
- Every field type must implement `Shape` (E0531 names the field and its type). A generic type derives `Shape` for the instances whose type arguments implement it, as `derives Show` does. Recursive enums (D55) derive it too.
- The derived `Shape` of an `export opaque struct` builds and reads its fields, so only the declaring module may use it; a `Shape` call or a `T: Shape` bound on it in another module is E0533.

### Inference

- Types of variables are inferred. `const n = 1` is `int`.
- **Function parameters and return types are written for every named function.** Arrow functions passed as arguments are the exception. Their types come from the parameter they are passed to.
- Other modules type-check against exported signatures only. They never read function bodies.

### Numbers

- **Integer overflow wraps in every build (D92).** `+`, `-`, `*` and unary `-` wrap at the width of the type. The operators `+%`, `-%`, `*%` and their compound forms do not exist.
- **Checked arithmetic.** `Int.addChecked(a, b)`, `subChecked` and `mulChecked` return `None` when the result does not fit, and the same functions exist in `std/i8` to `std/u64` (`U8.addChecked(a, b)`).
- Integer division or remainder by zero panics in every build. `MIN / -1` gives `MIN` and `MIN % -1` gives 0.
- **Sized integers (D82).** `i8`, `i16`, `i32`, `u8`, `u16`, `u32` and `u64` are separate types of that width; `i64` is the same type as `int`. All of them implement `Equatable`, `Hashable`, `Ordered` and `Show`, and work as map keys, fields, tuple elements, enum fields and generic arguments. `Array<u8>` stores one byte per element.
- **Literals take the type the context expects** (an annotation, a parameter, a field, an element, a return type, the other operand, a literal pattern), and `int` when nothing is expected. A literal must fit its type: `const x: u8 = 256` and `const y: u8 = -1` are E0294. `-128` is a valid `i8`.
- **There is no implicit conversion.** Arithmetic, comparison and bitwise operators need operands of the same type; `u8 + i32` is E0295, and the message names the conversion functions. Unary `-` on an unsigned type is E0296. Shift counts have the type of the value shifted, and a count at or above the width panics.
- **Overflow wraps at the type's own width:** `u8` `200 + 100` is 44 and `i8` `-128 - 1` is 127, in every build.
- Conversions are explicit functions named after the target type, in one module per type (`I32.from(x)`, `I32.tryFrom(x)`, `I32.wrap(x)`; modules `std/i8` to `std/u64`; `std/int` serves `i64`), and `Float.fromInt(x)`, `Float.truncate(f)`. `from` is for conversions where every value fits, `tryFrom` returns an `Option`, `wrap` keeps the low bits. Each of the seven modules also has `toString(x)` and `parse(text, base)`, and `std/int` has `from`, `tryFrom` and `wrap`. `X.from(v)` takes any integer type whose every value fits `X`; another integer type is E0940 and a value that is not an integer is E0941. `tryFrom` and `wrap` take any integer type. `parse` is `Int.parse` with the range of the type in its error text. A literal argument is an `int`, so `I32.from(5)` is E0940; write `const x: i32 = 5`.
- **`f32`** is a 32-bit float (C `float`) beside `f64`. A float literal takes `f32` when the context expects it and `f64` otherwise; an integer literal never becomes a float. A literal too large to be finite in `f32` is E0297. `f32` and `f64` do not mix in arithmetic or comparison (E0295). `f32` implements `Equatable`, `Ordered` and `Show` (the shortest text that reads back as the same `f32`) but not `Hashable`, so it is not a map key. `F32.from(v)` takes an `f64` (rounds to nearest) or any integer type, `F32.truncate(v)` gives an `int`, `F32.toString(v)` a string (module `std/f32`), and `Float.from(v)` takes an `f32` (exact) or any integer type. Another argument type is E0941.

### Strings

- `string` holds valid UTF-8. Functions that build a string from bytes return `Result` and fail on invalid UTF-8.
- **Positions are byte offsets of type `ByteIndex`**, starting at 0. `ByteIndex` is declared as `export opaque distinct type ByteIndex = u64`, so only `std/string` can create one.
- **Positions come from the string.** `Str.indexOf`, `Str.find` and iterators return `ByteIndex` values that always fall on a character boundary.
- **Slices share memory.** `Str.slice` and `Arrays.slice` return a view into the original buffer, with no copy. The view keeps the whole buffer alive. Use `Str.copy` or `Arrays.copy` to keep a small part of a large buffer without keeping the rest.
- **Slicing checks boundaries.** `Str.slice(s, from, to)` returns bytes `from` up to but not including `to`. It panics when a position is not on a character boundary or is out of range.
- `Str.length(s)` is the byte length as `ByteIndex`. `Str.charCount(s)` counts characters and takes time proportional to the length.
- Grapheme clusters, case mapping and normalization are in `std/unicode`.

### Bytes

- `bytes` is an immutable byte sequence with no UTF-8 guarantee. It has the representation of `string`, so `Bytes.fromString(text)` shares the buffer, and `Bytes.toString(data)` shares it too when the bytes are valid UTF-8. Otherwise it returns `Err("invalid UTF-8 at byte N")`, with the offset of the first bad byte.
- `std/bytes` (alias `Bytes`): `fromString(text): bytes`, `toString(data): Result<string, string>`, `length(data): int`, `at(data, index: int): u8`, `slice(data, start: int, end: int): bytes`, `concat(left, right): bytes`, `fromArray(items: Array<u8>): bytes` and `toArray(data): Array<u8>`. `at` and `slice` panic out of range. `slice` copies the bytes.
- `std/fs` has `readBytes(path): Result<bytes, string>` and `writeBytes(path, data: bytes): Result<int, string>`, which work as `readText` and `writeText` without the UTF-8 check.
- `bytes` implements `Equatable`, `Hashable`, `Ordered` (bytewise, a prefix first, and `<` works as on strings) and `Show`, which writes hex pairs: `bytes[68 69 0a]`. It works as a map key. It is not a contents type, so a `mutable` parameter cannot have it.

<!-- context
import * as Str from "std/string"
-->
```ts
export struct Pair {
  key: string
  value: string
}

export function parsePair(line: string): Option<Pair> uses alloc {
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
- **`std/result` helpers.** `import * as Results from "std/result"` adds `Results.map(r, f)`, `Results.mapErr(r, f)`, `Results.andThen(r, f)` (`f` returns a `Result` with the same error type), `Results.withDefault(r, fallback)`, and `Results.ok(r)` and `Results.err(r)`, which return the `Ok` or `Err` value as an `Option`. They work on a value you already hold. They never return from the caller, so they do not replace the visible `if` or `match` on the error path; they transform a result or give it a fallback.
- **Ignoring a `Result` is a compile error.** Discard one on purpose with `const _ = f()`.
- **Panics abort the program.** `panic("message")`, failed `assert`, and out-of-bounds indexing all panic. A panic cannot be caught.

<!-- context
import * as File from "std/fs"
struct Config {
  name: string
}
enum ConfigError {
  IoFailed(message: string)
}
-->
```ts
export function readConfig(path: string): Result<Config, ConfigError> uses io, alloc {
  const text = File.readText(path)
  if (text is Err(e)) {
    return Err(ConfigError.IoFailed(e))
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
- Boxed values (arrays, maps, strings, closures, `Box<T>` and recursive enums) live on the heap with a reference count. Numbers, `bool`, `char`, structs and enums that do not contain themselves and are at most 32 bytes live inline.
- A struct that holds boxed fields counts those fields, not itself. Copying the struct increments the count of each boxed field. An inline enum is the same: copying it increments the count of each boxed field of its variant.
- An enum is stored on the heap when a variant field contains the enum again through structs and enums held by value (an expression tree), or when its largest variant is over 32 bytes. Fields behind an `Array`, `Map`, `Box` or function do not count. `Some(1)`, `Ok(x)` and `Err(e)` of a small type allocate nothing.
- **`Box<T>`** puts a value on the heap with a reference count. Use it for recursive structs and for large structs that are passed around often.
- The compiler inserts reference count increments and decrements. It removes the ones it can prove unnecessary, as in Koka's Perceus.
- A value is freed when its count reaches zero. File handles, sockets and other resources close at the point their last owner drops them.
- **Large frees can be spread out.** When a large structure is freed, the runtime may release its memory in small batches during later allocations. Resources still close immediately. Only the return of memory to the allocator is delayed.
- **Running out of memory aborts the program,** the same as a panic.

### In-place updates

- `items[i] = v`, `m[k] = v`, `Arrays.push(mutable items, v)` and `u.age = v` change a local variable, `const` or `let`. `{ ...u, age: v }` returns a changed copy.
- When the reference count of the old value is 1, the compiler writes into it instead of copying. This is called reuse.
- Code that keeps one owner gets in-place mutation with no copies.

<!-- context
import * as Arrays from "std/array"
import * as Ranges from "std/range"
-->
```ts
let squares: Array<int> = Arrays.filled(0, 1000)
for (const i of Ranges.upTo(1000)) {
  squares[i] = i * i          // the count is 1, so the write is in place
}
```

- `items[i] = v` and `u.age = v` require the target to be a local `const` or `let`, a `mutable` parameter, or a field or element of one. A plain parameter and a top-level constant cannot be changed. Pattern, `for (const ...)` and destructured bindings are read-only; a `for (mutable item of items)` binding can be changed and changes `items` (section 2, Control flow).

### Copies on shared values

When an update targets a value whose reference count is greater than 1, the value is copied first. The copy is correct, but it costs time proportional to the value's size, and the source does not show it. `copy(x)` makes the copy visible (D94):

- **`copy(x)`** duplicates a value at any depth. At the call it is one reference count increment, so it costs nothing until a write. A later write to `x` or to the copy then copies only the objects on the written path, once each: after `const other = copy(team)`, `other.scores[1][0] = 30` copies the outer array of `other.scores` and one row, not the other rows.
- `copy` takes one argument of a type that a write changes in place: a struct (all-scalar ones too), `Array<T>`, `Map<K, V>`, `FixedArray<T, N>`, `Box<T>`, `StringBuilder`, `Generator<T>` or `Coroutine<In, Out>`, a tuple holding one of them, a distinct type over one, or a type parameter. Any other argument (numbers, `bool`, `char`, `string`, `bytes`, enums with `Option` and `Result`, function values) is a compile error whose fix removes the call, and so are a second argument and `copy(mutable x)` (E0975).
- `copy` is a prelude function, not a keyword. A local or parameter may be named `copy`, as one may be named `panic`; inside its scope a call `copy(x)` then calls the local. A top-level function named `copy` is never called: the prelude's `copy` is found first.
- **The debug copy check.** A debug build run with `NARRA_CHECK_COPIES` set counts every copy on write of an array, map, `Box`, `StringBuilder` or generator, and at exit reports the ones that no `copy(...)` explains, with exit code 103, like the leak check (D80): `copy check: copies on write that no copy(...) explains: array 2`. Each `copy(x)` gives the objects it shares one credit, a copy on write uses one credit of the object it copies and gives one to each object the two copies now share, and a credit ends when the object's count falls back to 1. So at most one hidden copy per `copy(...)` call goes unreported. All-scalar structs live inline and never copy on write; the compile-time shared-write rule covers them. Release builds have no credits and no hooks.

<!-- context
import * as Arrays from "std/array"
-->
```ts
const scores = [[1, 2], [3, 4]]
let other = copy(scores)
other[1][0] = 30            // copies the outer array and one row; scores is unchanged
Arrays.push(mutable other, [5, 6])
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

A function changes a caller's data only through a `mutable` parameter. The keyword appears in the declaration and again at the call. A `mutable` parameter reaches the contents of the caller's value, its fields and elements, and never the caller's variable (D81).

<!-- context
struct Vec2 {
  x: f64
  y: f64
}
struct Player {
  name: string
  hp: int
  position: Vec2
}
-->
<!-- error: E0949 -->
```ts
function damage(mutable player: Player, amount: int) {
  player.hp = player.hp - amount     // ok: changes a field
  player = { ...player, hp: 0 }      // compile error: a mutable parameter cannot be replaced
}
```

<!-- context
struct Vec2 {
  x: f64
  y: f64
}
struct Player {
  name: string
  hp: int
  position: Vec2
}
function damage(mutable player: Player, amount: int) {
  player.hp = player.hp - amount
}
---
-->
```ts
const hero = Player { name: "Ana", hp: 100, position: Vec2 { x: 0.0, y: 0.0 } }
damage(mutable hero, 10)      // hero.hp is now 90
```

<!-- error: E0948 -->
```ts
function bump(mutable n: int) {  // compile error: int has no contents to change
  n = n + 1
}
```

- **The parameter's type has contents.** A `mutable` parameter must have a contents type: a struct (an `opaque struct` too), `Array<T>`, `FixedArray<T, N>`, `Map<K, V>`, `Box<T>`, `StringBuilder`, `Generator<T>`, or `Coroutine<In, Out>`. Any other type is a compile error at the parameter: numbers, `bool`, `char`, `string`, `bytes`, distinct types, enums (`Option` and `Result` included), tuples, function types, the prelude `Range`, `Sender`, `Receiver`, `Shared` and `Atomic`. To give a new scalar or enum back to the caller, return it, or keep it in a field of a struct passed as `mutable`.
- **Generic parameters.** `mutable value: T` with a bare type parameter `T` is a compile error, whatever its constraint: no trait guarantees a contents type. A generic instance of a contents type is fine: `mutable items: Array<T>`. In a trait, `mutable value: Self` is allowed, every `implement` of the trait must be for a contents type, and a `Dynamic<Trait>` cannot be passed to such a function.
- **The callee changes contents, never the parameter.** Through a `mutable` parameter the callee may assign fields and elements at any depth, call std functions that change contents (`Arrays.push`, `Maps.set`, `Str.append`, `Co.resume`), and pass the value or a part of it on as `mutable`. Assigning the parameter itself (`player = other`, `count += 1`) is a compile error.
- **The argument is a local place.** It must be a local binding that holds a contents type, or a field or element path of one: a `const`, a `let`, or a `mutable` parameter. A literal, a call result, a plain parameter and a top-level constant are compile errors. Pattern, `for (const ...)` and destructured bindings are read-only copies and are compile errors too; a `for (mutable item of items)` binding can be passed (D85). Passing a `const` as `mutable` adds nothing that direct code lacks, since its fields can be assigned directly (section 2, Bindings).
- **Access is exclusive during the call.** A variable cannot be passed as `mutable` twice in one call. It cannot be passed as `mutable` and also read in another argument. `swap(mutable a, mutable a)` and `move(mutable p, p.x)` are compile errors.
- **A `mutable` variable is not used elsewhere in the same expression.** Within one statement's expression, a variable passed as `mutable` cannot also be read, assigned or passed again anywhere else, at any depth, including fields and elements of it. `p.hp + heal(mutable p)` and `pair(p, heal(mutable p))` are compile errors. Bind the call first: `const healed = heal(mutable p)`, then `p.hp + healed`. Two `mutable` calls on different variables are fine, and so are the two sides of `&&` and `||` and different arms of a `match`. These rules hold for every binding passed as `mutable`, `const` included.
- **A `mutable` parameter cannot escape.** It cannot be stored, returned, or captured by a closure that is not passed straight to a non-escaping parameter (D77). Copying its value with `copy(player)` is allowed.
- **No copies.** The compiler passes a pointer to the caller's value. Changes are written in place, with no copy and no reference count change for the value itself. A value shared with other owners is copied on its first change, as in any in-place update.
- **Other parameters are values.** A parameter without `mutable` is a read-only copy: its fields cannot be assigned, and it cannot be passed as `mutable`. To build new data instead of changing the caller's data, return a new value.
- **Enums and tuples are replaced, not changed** (D81). A struct inside an enum payload or a tuple is not a place: enum fields are fixed and tuples have no element assignment. A field that holds one is assigned whole: `state.current = Some(next)`. To change a struct held in an `Option`, read it with `is`, build the changed copy, and store it back. A `mutable` marker in a pattern (`x is Some(mutable y)`) does not exist (D85).

| Goal | Form |
| --- | --- |
| Change the caller's data | `damage(mutable hero, 10)` |
| Build new data | `const stronger = withBonus(hero, 5)` |
| Give back a new number | `count = bumped(count)` |

### Cycles

- Immutable values cannot form cycles, so reference counting frees everything without a cycle collector. Storing a value inside itself holds a second reference to it while the write runs, so the write copies first.
- A variable is never stored by reference: storing it stores a copy, and only a closure that cannot outlive the call captures a variable by reference, so no cycle forms.
- **Cycles through identity types leak** (D80). Identity types (`Shared<T>`, `Sender<T>`, `Receiver<T>`, `Task<T>` and `Cell<T>`; section 8) are handles to one object, so they can form a cycle. There is no cycle collector and no type rule against it. Memory safety holds, but the objects in the cycle are never freed and resources they hold never close. Every cycle passes through an identity object, so the debug leak check (`NARRA_CHECK_LEAKS`, section 13) reports it at exit, and a test that leaks fails. Two patterns leak:
  - A handle stored in its own payload: `Shared.update(state, ...)` that stores `Some(state)` in the value of `state`, or a `Sender` sent into its own channel.
  - A closure stored in a `Cell` that captures the same `Cell`.
- **Breaking a cycle.** Clear the slot before the last outside handle is dropped (`Cells.set(cell, None)`), or keep the graph in `std/arena` and refer to its nodes with handles.
- A later version may add a collector limited to identity objects. It would only free memory that this rule leaks, so no program changes meaning.
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
- **Function types carry effects** (D86): `(v: int) => int uses io`. A call through a function value has the effects of its type, and a function cannot be stored where the type lists fewer effects than it has. A function value's own type lists none, so a top-level constant that holds a function with effects writes its type: `const step: (v: int) => int uses io = loud`.
- **Generator types carry effects** (D78), as function types do: `Generator<string> uses io` is a generator whose body does I/O when it runs. `Generator<T>` alone is pure. The effects are charged where the body runs: a `for ... of` over the generator, a `yield*` of it, and `Co.resume`. Creating a generator charges only `alloc`, for its frame.
- A generator value may be used where a generator type with more effects is expected, not fewer: passing a `Generator<T> uses io` where `Generator<T>` is expected is a compile error.

### Effect variables

A function that calls a function it receives has that function's effects. An effect variable states this in a signature.

```ts
export function map<T, U, effect E>(items: Array<T>, f: (item: T) => U uses E): Array<U> uses alloc, E
```

- Effect variables are declared in the type parameter list with the `effect` keyword, and appear only after `uses`.
- **They are bound by a parameter.** An effect variable in the function's own `uses` list must also appear in a parameter's function type or generator type: `export function sum<effect E>(numbers: Generator<int> uses E): int uses E`.
- At each call, the variable takes the effects of the function or generator passed in. `map(names, Str.length)` has `alloc`. `map(paths, File.readText)` has `alloc, io`. A generator result type that names the variable takes them too: with `relay<effect E>(g: Generator<int> uses E): Generator<int> uses E, alloc`, `relay(lines(path))` is a `Generator<int> uses alloc, io`.
- **A function used as a value keeps its variables** (D91). After `const apply = twice`, `apply(loud, 1)` has the effects of `loud`, as `twice(loud, 1)` does, through a top-level or a local constant.

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

<!-- context
import * as Ranges from "std/range"
-->
```ts
function* counter(limit: int): Generator<int> {
  for (const i of Ranges.upTo(limit)) {
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
```

<!-- context
import * as Int from "std/int"
function* both(): Generator<int> {
  yield 1
}
---
-->
```ts
for (const n of both()) {
  print(Int.toString(n))
}
```

- **`function*`** declares a generator. Calling it returns a `Generator<Out>` without running any of its body. A generator is a top-level function (D74), and `export function*` exports it as `export function` does. There is no arrow form. The return type is required and must be `Generator<T>` or `Coroutine<In, Out>`. Generic generators are allowed.
- **`yield value`** suspends the generator and produces `value`, which must have type `T`. `yield` is a statement, not an expression. It is valid only directly inside a `function*` body or its nested statement blocks (`if`, `while`, `for`, labeled loops, `match` statement arms). It is not valid in an arrow function, a block expression, a conditional expression or a `match` used as a value.
- **`yield* other`** produces every value of another generator in turn. `other` must be a `Generator<T>` with the same `T`.
- **`return` in a generator takes no value.** It ends the generator, as falling off the end of the body does.
- **A generator outlives the call that creates it.** It cannot take a `mutable` parameter, and its function-type parameters must be `escaping`.
- **Effects** (section 7, D78). An exported generator writes the effects of its body on its return type: `export function* lines(path: string): Generator<string> uses io, alloc`. A private one gets them inferred into its type. A body that uses an effect its declared type omits is a compile error.
- **Calls of generator functions are not recursion.** A generator that calls itself, directly or through another generator, creates a new value, so D37 and tail loops do not apply.
- **`Coroutine<In, Out>`** (D79) is a generator that also receives a value on each resume. It is a separate type from `Generator<T>`: declare it with `function*` and the return type `Coroutine<In, Out>`, as in `function* total(): Coroutine<int, int>`. Inside it, `const input = yield output` yields `output` and, when the coroutine is resumed, binds the value passed to `Co.resume`. The `yield` with a value is valid only in a coroutine, and only as the initializer of a `const` or `let` that names one variable and has no type annotation; a plain `yield output` statement is also valid and drops the input. The input of the first resume is dropped, as in TypeScript, because no `yield` is waiting for it. `yield*` is not valid in a coroutine, and `for ... of` over a coroutine is a compile error (a loop has no input to send). A coroutine type carries effects like a generator type: `Coroutine<string, int> uses io`.
- **`Co.resume`** from `std/coroutine` (`import * as Co from "std/coroutine"`) runs a generator or a coroutine to its next `yield`, changing `g` in place: `Co.resume(mutable g)` for a `Generator<T>` and `Co.resume(mutable g, input)` for a `Coroutine<In, Out>`. It returns a `Step<Out>`, an enum of the same module (`import { Step } from "std/coroutine"`) with the constructors `Step.Yielded(value)`, `Step.Done` and `Step.Panicked(message)`, always qualified (D71). `Step` derives `Show` and `Equatable`. A finished generator or coroutine answers `Step.Done` to every later resume. The call charges the effects carried on the type of `g` (D78), and `alloc` for the `Step`. A shared generator is copied before it resumes (the count above 1), as in `for ... of`. `Panicked` is only reported in test builds (`narra test`): a panic inside the coroutine then ends the coroutine and `Co.resume` returns `Step.Panicked(message)` with the bare message; what the coroutine held at that moment is not freed. In other builds a panic aborts the program.
- **A generator is a value like any other.** Resuming one by hand requires a `mutable` argument. `for ... of` iterates over its own copy. `Generator<T>` has no `==`, no hashing and no `Show`.
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
- **Thread functions capture only constant values,** like every closure that can outlive its call. They cannot be `async`.
- **`Parallel.map(items, f)`** from `std/parallel` splits an array across threads and returns the results in order.

### Channels

```ts
const #[sender, receiver] = Channel.create<int>(16)     // holds up to 16 values

Thread.start(scope, () => {
  for (const i of Ranges.upTo(100)) {
    Channel.send(sender, i)                            // blocks while the channel is full
  }
  Channel.close(sender)
})

for (const value of Channel.values(receiver)) {       // ends when the channel is closed and empty
  print(Int.toString(value))
}
```

The builtin that builds channels must accept the type argument in `Channel.create<int>(16)`. Builtin functions reject type arguments today (E0326), so channels need an exception when they are built.

- **`Channel.create<T>(capacity)`** returns a `#[Sender<T>, Receiver<T>]`.
- **`Channel.send`** blocks while the channel is full. **`Channel.receive`** blocks until a value arrives.
- **`Channel.close(sender)`** closes the channel for every copy of the sender. Dropping the last copy also closes it. `Channel.receive` returns `None` once the channel is closed and empty.
- Blocking calls have the `io` effect.
- **A channel in a cycle stays open** (D80). A `Sender` sent through its own channel, or held in a `Shared` that the channel's values reach, keeps the channel alive: it is never freed, and dropping the other copies of the sender does not close it, so a receiver that reads until the close waits forever. When senders travel through channels, close the channel with `Channel.close(sender)` instead of relying on the last drop; the close ends the receiver's wait, and a receiver that reads the channel empty drops the sender it held and breaks the cycle. A `Sender<Reply>` sent inside a request (`Sender<#[Request, Sender<Reply>]>`) is fine: the reply channel does not hold the request channel.

### Cell

`Cell<T>` (D90) is an identity type for state that changes and is shared on one thread, such as a counter two escaping closures use. It is `Shared` without the lock. It is a Narra source module: `import * as Cells from "std/cell"` for the functions and `import { Cell } from "std/cell"` for the type.

<!-- context
import * as Cells from "std/cell"
-->
```ts
const count = Cells.create(0)
const next = (): int => {
  Cells.update(count, (value) => value + 1)
  return Cells.get(count)
}
const reset = (): int => {
  Cells.set(count, 0)
  return 0
}
```

- **`Cells.create(value)`** allocates the slot (`uses alloc`). **`Cells.get(cell)`** returns a copy of the value. **`Cells.set(cell, value)`** replaces it and drops the old value at the call. **`Cells.update(cell, change)`** stores `change(value)`; `change` is a function that returns the new value, because a scalar cannot be a `mutable` parameter (D81). None of the last three has an effect: a write to a `Cell` is not tracked.
- Copying a `Cell`, capturing it, or storing it in a struct copies the handle, so every copy sees every write. `Cell` is an `opaque struct`: its field is not reachable. `==` and `!=` on a `Cell` are a compile error (E0965), and a `Cell` is not `Hashable`, so it cannot be a `Map` key; compare the values with `Cells.get`.
- **A `Cell` cannot cross threads.** It uses plain reference counts. Use `Shared`. Threads are parked, so the compiler does not check this yet; the node that builds threads must reject a `Cell`, or a value that holds one, captured by a thread function or sent through a channel.
- **A closure stored in the `Cell` it captures leaks** (D80); clear it with `Cells.set(cell, None)` first.

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

- **Open (D81).** The sample above passes an `int` as a `mutable` arrow parameter, which D81 and E0028 do not allow. The node that builds `Shared` decides its form, for example a callback that returns the new value.
- **The lock is held only during the callback.** `Shared.update` and `Shared.read` take the lock, run the callback and release it. There is no way to take the lock and forget to release it.
- **The callback cannot be `async`,** so a task never suspends while holding a lock.
- **`Atomic<T>`** offers `Atomic.load`, `Atomic.store`, `Atomic.add` and `Atomic.compareExchange`.
- `==` on an identity type compares identity, not contents.
- **Identity types can form cycles** (D80, section 6 "Cycles"). A `Shared` whose value holds a handle to the same `Shared`, directly or through other identity objects, is never freed. `Task<T>` (Async functions below) and `Cell<T>` follow the same rule. The debug leak check reports the cycle; clear the slot before dropping the last outside handle, or use `std/arena`.
- **Deadlocks are possible** when callbacks nest locks in different orders. `narra audit` lists every nested `Shared.update`.

### Async functions

An event loop runs many tasks on one thread. A task suspends at each `await`, and the loop resumes it when the awaited I/O is ready. The syntax is TypeScript's.

```ts
import * as Async from "std/async"
import * as Net from "std/async/net"

async function handle(connection: Connection) uses io, alloc {
  const conn = connection
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
- **No async callbacks in normal higher-order functions.** `Arrays.map(items, async (x) => ...)` is a compile error. `std/async` provides `Async.map` and `Async.forEach` instead.
- **A `Loop` cannot escape,** with the same rules as a `Scope`.
- **Suspending I/O lives in `std/async/*`.** `std/async/net`, `std/async/file` and `std/async/time` mirror the blocking modules with the same function names. Their functions are `async`.
- **A task stays on the thread that runs its loop.** To use several cores, run one loop per thread inside `Thread.scope`. Pass work between the loops with channels.
- `await Async.sleep(milliseconds)` suspends the task for a time.

## 9. Unsafe code and C interop

- **`unsafe { ... }`** blocks allow raw pointers (`Ptr<T>`), pointer arithmetic and unchecked indexing. A function containing an `unsafe` block has the `unsafe` effect unless it is marked `@trusted`, which states that the author checked it by hand.
- **`extern`** declares a C function: `extern function strlen(s: Ptr<u8>): u64 uses unsafe = "strlen"`.
- **`@repr(C)`** on a struct fixes its field order and padding to match C. Without it, the compiler may reorder fields to reduce padding.
- **C-compatible enums.** `@repr(u8)`, `@repr(i32)` and the other integer types give an enum without fields explicit values. `Red = 1` in an enum without `@repr` is a compile error whose fix adds `@repr(i32)` (D73). `Enum.toInt` and `Enum.fromInt` convert, and `fromInt` returns `Option`.

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

- **`const`** at the top level declares a compile-time constant. It cannot be changed at any depth or passed as `mutable` (D81); a local `const` fixes only its name (section 2). Inside a function, `const` accepts any expression. A top-level value must be one of these:
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
  - Calling a `function*` allocates the struct with a reference count of 1 and state 0, and stores its parameters. It runs nothing of the body.
  - Every local, loop counter, temporary and expression result of the body is a field of the struct, so no C local is alive across a `yield`. Only scratch values used inside one expression stay C locals.
  - `yield v` stores `v` for the caller, records the next state and returns. Ending the body or `return` releases what is still held and sets the finished state; resuming a finished generator returns nothing again.
  - Dropping an unfinished generator releases what is live at its state and frees the struct. Resuming a shared generator first copies the struct and retains what is live in it, so `for ... of` and `yield*` work on their own copy.
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
  - `Test.golden(name: string, actual: string)` compares `actual` with the file `<name>.golden` in the package's `testdata/` directory (for a single file, in the file's own directory). It fails with `golden <name> differs at line N: expected "<line>", got "<line>"`, or when the file is missing. With the environment variable `NARRA_UPDATE_GOLDEN=1` it writes the file instead and passes. `name` may contain `/` but no path segment equal to `..`, and it is not empty. The missing-file message names `<name>.golden`, not the directory. When the test binary runs alone, the directory is `testdata/` relative to the working directory, unless `--golden-dir <dir>` is given. It declares `io`; test blocks may use any effect without declaring it.
  - A failed assertion is a panic that fails its test only.
- **`narra test`** runs each test in its own process, in source order, and goes on after a failure. It prints one line per test to standard output: `ok - <name>`, or `FAIL - <name>: <message> (<file>:<line>:<col>)`. Output a test prints appears before its line. The last line is `N passed, M failed`. The exit code is 1 when a test failed, otherwise 0. A test that crashes or leaks memory in a debug build fails with the signal or exit code.
  - `narra test -filter <text> file` runs only the tests whose name contains `<text>`.
  - `narra test -time file` appends ` (N ms)` to each test line. Without it no timing is printed.
  - A filter that matches no test prints `no tests match "<text>"` and exits 1. A file with no tests at all prints `0 passed, 0 failed` and exits 0.
  - Every test has a time limit: 10 seconds, or `NARRA_TEST_TIMEOUT` (a positive number of seconds). A test that runs past it fails with `FAIL - <name>: timed out after N s` and the run goes on. A `Test.panics` closure gets the same limit, or the time left on the test's limit minus one second (at least one second), whichever is smaller, so its own timeout message comes first.
  - With `NARRA_CHECK_LEAKS` set (debug builds of the project's own tests), a test that leaks fails with `leaked N allocations`.
  - The test binary reads `--filter <text>`, `--time` and `--golden-dir <dir>` (one or two dashes) from its own arguments, so it can be run directly. `narra test` passes the golden directory itself, so the generated C holds no path. Other arguments are an error (exit 2), and so is an option without its value. Flags may also follow the file name.
  - `f` runs in a child process, so nothing it changes or allocates reaches the test. Output it prints appears in order. `Test.panics` and `Test.panicsWith` therefore need a POSIX system (`fork`). A closure that runs past the timeout (10 seconds, or `NARRA_TEST_TIMEOUT`) fails.

<!-- context
import * as Int from "std/int"
import * as Str from "std/string"
function divide(a: int, b: int): int {
  return a / b
}
-->
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
- **TypeScript habits get a fix.** Each row of "Differences from TypeScript" (section 2) has a diagnostic, except the row on value copies (`const b = a`), which compiles. Its fix rewrites the TypeScript form into the Narra form. For example, `===` is rewritten to `==`, and `xs.push(v)` to `Arrays.push(mutable xs, v)`.

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

### Language server

`narra lsp` is a Language Server Protocol server on stdin and stdout. It works with any LSP client.

- Messages use the LSP framing: a `Content-Length` header, other headers ignored, then the JSON body.
- Lifecycle: `initialize`, `initialized`, `shutdown` and `exit`. `exit` ends the server with code 0 after `shutdown` and 1 before it. The end of the input without `exit` also ends it with code 1.
- `initialize` answers full text sync, `documentFormattingProvider` and a `positionEncoding`. The encoding is `"utf-8"` when the client offers it, else `"utf-16"`.
- A request before `initialize` gets error -32002, an unknown request -32601 and a message that cannot be read -32700. Unknown notifications are ignored.
- `textDocument/didOpen`, `didChange` (the full text) and `didClose` keep the editor's text, which replaces the file on disk. After each open or change the server checks the package that holds the file (the nearest `narra.toml`; a file with none is checked alone). It then sends `textDocument/publishDiagnostics` for every module of the package, with an empty list for a module that has no problem. A diagnostic's range starts and ends at its position. Its `code` is the diagnostic code.
- `didClose` checks the package again from the files on disk and empties the list of the closed document.
- `textDocument/formatting` formats with the `[fmt]` table of the nearest `narra.toml`. It returns one edit that replaces the whole document, an empty list when the document is already formatted, or `null` when it does not parse.

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
- **A cycle collector for identity objects.** Cycles through identity types leak and the leak check reports them (section 6, D80).
- **Functors and module parameters.** See `DECISIONS.md`, D6.
- **A native debug backend.** See `DECISIONS.md`, D10.
- **A bytecode VM.** Possible later for scripting and hot reload. The first target is native code only.
