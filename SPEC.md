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
- A source line number or column at or past 1048576 (2^20) is a compile error (E1004).

### Bindings

- `const x = expr` binds a name that cannot be reassigned. **`const` fixes the name, not the contents** (D81): the fields and elements of a local `const` struct, array, map or other contents type (section 6, Mutable parameters) can be assigned, and the `const` can be passed as `mutable`.
- `let x = expr` binds a name that can also be reassigned with `x = expr`. A closure passed straight to a non-escaping function parameter can capture it, by reference (section 2, Closures).
- **Every binding holds its own value.** `const other = hero` copies `hero`, so a later change to either one does not show in the other. The one exception is a `for (mutable item of items)` binding, which reaches an element of `items` (section 2, Control flow).
- **A plain parameter is read-only.** Its fields and elements cannot be assigned, and it cannot be passed as `mutable` (section 6).
- **Bindings that copy part of another value are read-only.** A `match` or `is` pattern binding (`x is Some(v)`), a `for (const p of players)` binding and a destructured binding (`const { a, b } = s`) hold a copy, so their fields and elements cannot be assigned and they cannot be passed as `mutable`. `for (const p of players) { p.hp = 0 }` is a compile error: write `for (mutable p of players)` (section 2, Control flow), write through the array, `players[index].hp = 0`, or make a copy first with `copy(p)` (D81, D94). A `mutable` marker exists only in that loop (D85); `x is Some(mutable y)`, `const { mutable a } = s` and a local `mutable name = place` do not.
- **Top-level constants never change** (D81). A top-level `const` cannot be changed at any depth (section 10).
- There are no globals. Assigning to an undeclared name is a compile error.
- **Compound assignment (D76).** `x += e` and the forms `-=` `*=` `/=` `%=` `&=` `|=` `^=` `<<=` `>>=` mean `x = x op e` with the target place evaluated once: `items[next()] += 1` calls `next` once. They are statements with no value, valid on every target `=` accepts (a `let`, a field or element of a changeable value, the contents of a `mutable` parameter). `x++` and `x--` are statements for an integer of any width; the prefix `++x` is a compile error (E0964). The value must have the type of the target (E0962), and `++` on a non-integer is E0963. Wrapping, divisor proofs and shift count proofs are those of the binary operator: `x /= d` needs a proven `d`. The wrapping operators `+%=`, `-%=` and `*%=` do not exist. On a `string` target only `+=` exists and it joins strings (`+` alone is still an error). A Map index target (`counts[key] += 1`) is a compile error, since a missing key has no old value (reading `counts[key]` gives an `Option`); write `Maps.update(mutable counts, key, 0, (count) => count + 1)`, which starts from the initial value when the key is missing (D92).
- **No shadowing inside a function.** A name cannot be declared again while an earlier binding with that name is in scope.
- **Unused names are errors.** An unused local, parameter, import or private function is a compile error. Name it `_` or start it with `_` to mark it as unused on purpose.
- **Destructuring:** `const { host, port: serverPort } = config` for structs (section 4), and `const #[quotient, remainder] = Int.divmod(a, b)` for tuples.
- **Array destructuring (D103).** A destructuring must match every value, so on an `Array<T>` only `const [...rest] = items` works: it binds a new Array with every element, a copy. `const [a, b] = items` and `for (const [a, b] of rows)` over an Array are compile errors (E0715), since the length is not known; test it with `match` or `is` (section 4, Pattern matching). On a `FixedArray<T, N>` with a known `N`, a pattern of exactly `N` elements, or `..._` after at most `N` elements, works in `const`, `let`, `for ... of` and nested in tuple and struct patterns: `const [x, y, z] = point`. Another length is E0723. In a generic body where `N` is a constant parameter, only `[..._]` matches every value (E0723).

### Functions

- `function name(a: int, b: int): int { ... }` declares a function. A function is private to its module unless declared with `export function` (section 3). A function cannot be declared inside another function (E0930): declare it at module level as a private function and pass the enclosing values as arguments. A nested declaration in TypeScript captures the enclosing locals, and a module-level copy could read a different name where a local shadows it (D74). To keep a helper that captures, bind an arrow function to a `const`.
- A function with no return type returns nothing.
- **A function with a result must return on every path.** A path ends at `return`, a call to `panic` or `Process.exit`, an `if`/`else` whose branches all end, a `match` whose arms all end, or `while (true)` / `for (;;)` with no `break` out of it. Any other path that reaches the closing brace is a compile error that names the path (E0969, E0970 for an arrow function with a block body). Generators have no result value and are exempt (D92).
- **Anonymous functions are arrow functions:** `(x: int): int => x + 1`, or `(x: int): int => { ... }` with a block. `function` is only used for named declarations.
- **`function*`** declares a generator and **`async function`** declares an async function (section 8). Arrow functions can be `async`: `async () => { ... }`, and `async { ... }` is an async block.
- An arrow function passed as an argument takes its types from the parameter it is passed to. The types can be left out: `Arrays.map(prices, (p) => p * 2)`.
- **Function types** are written like arrow functions: `(item: T) => U`.
- **Standard functions are values** when their signature is fixed: `Int.toString`, `Str.length` and `Fs.readText` can be passed, stored and returned, and `import { toString } from "std/int"` names one directly. Its effects follow it: `Fs.readText` has `alloc, io`, so a function type that takes it lists both. A standard function whose types depend on the call (`Arrays.length`, `Maps.size`), one with a `mutable` parameter and one that takes a function (`Arrays.map`) are not values (E1013); write an arrow function, `(items) => Arrays.length(items)`.
- **Parameters are constant.** A parameter cannot be reassigned, and its contents cannot be changed unless it is declared `mutable` (section 6). A `mutable` parameter still cannot be reassigned.
- **A call passes every argument** that has no default. Passing fewer arguments, or more than the function declares, is a compile error.
- **Arguments are evaluated left to right**, in the order they are written. An argument that is a block expression, a `match` or a closure call can assign a `let` that an earlier argument reads; the earlier argument keeps the value it had when it was evaluated. After `show(items, { items = ["z"]; return 0 })`, `show` received the original `items`.
- **Named arguments.** A call can name its arguments, written like struct literal fields: `move(x: 1, y: 2)`. Positional arguments come first, then named ones in any order: `move(1, z: 3, y: 2)`. A positional argument after a named one is an error. Each parameter is passed exactly once, by position or by name. `mutable` follows the colon: `fill(buffer: mutable items, value: 0)`. Names work when the callee is a declared function or a trait function. A function value, an enum constructor, and a std function built into the compiler such as `Str.slice` take positions only (D87). The arguments are evaluated in the order they are written and passed in the order of the parameters. A parameter's name is part of the function's interface: renaming it breaks the calls that use the name.
- **No currying and no partial application.** To fix some arguments, write an arrow function.
- **No method calls.** `xs.push(v)` is not valid. Write `Arrays.push(mutable xs, v)`. `value.field(args)` is only valid when `field` is a struct field that holds a function.
- **Parameter defaults (D115, not built yet: node parameter-defaults).** A parameter of a declared function can have a default after `=`: `function render(node: Node, depth: int = 0, indent: string = "  ")`. A call may leave it out: `render(tree)`, `render(tree, 2)`, or `render(tree, indent: "\t")`, which names a later parameter to skip `depth`.
  - Defaults are trailing: once a parameter has one, every later parameter has one (E1210).
  - A default is a compile-time constant, by the rule of a struct field default (section 4): a literal, a unary operator on a constant, a constructor with constant arguments, or an array of constants (E1212). It cannot read another parameter, a local, a top-level constant or a function, and cannot call an `@const` function.
  - A `mutable` parameter has no default (E1211). Arrow function parameters and function types have none (E1213), nor have trait functions and their `implement` blocks (E1214) or the functions of an `extern` block (E1215). `name?: T` is a compile error (E1216); write `name: Option<T> = None`.
  - The compiler fills each left-out default into the call, in parameter order, after named arguments are matched, so the function is still one C function with a fixed parameter list. The default is built at each call that leaves it out. A call that leaves out an array default has the `alloc` effect; no other default has an effect. A `requires` pattern on the parameter is checked against the filled value.
  - A default whose type uses a type parameter must type-check for every type argument, as for a generic struct field (`fallback: Option<T> = None`). A left-out argument gives no evidence for inference.
  - A function value takes every argument. The type of `render` as a value is `(node: Node, depth: int, indent: string) => string`, with no defaults, and a call through it passes all three.
  - A default is part of the function's interface, like a parameter name: changing it changes every call that leaves it out.
- **No varargs.** Pass an array.
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
- Struct fields and array elements of function type always hold escaping closures. `Thread.start` and `Async.then` take escaping parameters. An `async` arrow function or block captures like an escaping closure wherever it is written (section 8).
- **`@inline`** on a function copies its body into each call. An arrow function passed for a function-type parameter that the body only calls is copied too, to each place the body calls the parameter, so the call has no C call, no closure environment and no indirect call. The arrow reads the caller's variables in place.

```ts
@inline
function applyTwice(value: int, transform: (item: int) => int): int {
  return transform(transform(value))
}

const rate = 3
const result = applyTwice(2, (item) => item * rate)    // no call, no closure, no indirect call
```

  - `@inline` is a guarantee, not a hint. Every call is copied, also across modules and packages, and a generic function is copied with the call's type arguments.
  - Copying never changes behavior. Arguments are evaluated once, left to right; a `mutable` parameter still writes the caller's place; copies, releases, panics and effects are those of the ordinary call. A `return` in the body ends the copy, and a `return` in a block arrow ends that arrow.
  - A function-type parameter that the body passes on, stores in a closure or reads other than in a call gets a real closure value, built at the call. A named function or a variable holding a function is passed as a value.
  - An `@inline` function cannot be recursive: a call to itself, directly or through other `@inline` functions, is E1030, and the message names the chain (`isEven -> isOdd -> isEven`). A call to a function without `@inline` ends the chain.
  - `@inline` needs a body that runs to its end inside the call: on a generator (`function*`) or a function of an `extern { } from` block it is E1031. It is not written in an `implement` block (E0990), since those calls go through the trait.
  - An `@inline` function used as a value (`const f = applyTwice`, or passed as an argument) has an ordinary C function for that value. Calls are still copied. A program that only calls it has no C function for it.


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
- **`for await (const chunk of stream) { ... }`** reads an `AsyncGenerator<T>` in an async body, waiting for each value (section 8, Async generators). Not built yet.
- **Counting loops use `Range`.** `Ranges.upTo(n)` counts from 0 to `n - 1`. `Ranges.between(start, end)` counts from `start` to `end - 1`. `Ranges.stepped(start, end, step)` takes a step: a positive step counts up while the value is below `end`, a negative step counts down while it is above `end`, and `Ranges.stepped` returns `None` for a step of 0. The loop stops before the value would pass the largest or smallest `int`.

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
- **Tagged templates** (D114, not built yet): `` tag`a${x}b` `` is the call `tag(["a", "b"], [x])`. The parts array always has one more item than the values; an empty part is `""`. The tag is a name or a dotted name (`tag`, `Sql.query`) on the same line as the backtick; after `)`, `]` or a literal a template is E1200. A tag is a function or function value of two parameters, neither `mutable` (E1201): the parts, `Array<string>` with the escapes above or `Template.Raw` from `std/template` with the source text as written, and the values, `Array<T>` for one type `T`. Each hole is checked with `T` as its expected type, and a hole of another type is E1202; a generic tag takes `T` from the holes. The tag's effects count at the use site, plus `alloc` for the two arrays. Raw parts decode no escape and report none: `` \d `` is the two characters `\` and `d`, and `` \` `` and `\${` keep their backslash. A template at the start of a line, after a line that ends with a name, is E1203: a newline ends the statement, so it is not tagged as it would be in TypeScript. `Template.raw` is TypeScript's `String.raw`; it is `@const`, so `` const DIGITS = Template.raw`\d+` `` is computed while compiling (section 10).

  ```text
  import * as Arrays from "std/array"
  import * as Ranges from "std/range"
  import * as Str from "std/string"
  import * as Template from "std/template"

  function quoted(parts: Array<string>, values: Array<string>): string {
    let out = Str.builder()
    for (const index of Ranges.between(0, Arrays.length(values))) {
      Str.append(mutable out, `${parts[index]}"${values[index]}"`)
    }
    Str.append(mutable out, parts[Arrays.length(values)])
    return Str.build(out)
  }

  const greeting = quoted`hello ${name}!`      // hello "ada"! when name is "ada"
  const pattern = Template.raw`\d+\.${digits}` // the text \d+\. then digits
  ```
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
| `+` `-` `*` `/` `%` | Any integer type or any float type. `/` on integers truncates toward zero and needs a proven divisor (below). `+` `-` `*` on integers wrap on overflow |
| `&` `\|` `^` `<<` `>>` | Integer types. `^` is xor. A shift needs a proven count (below) |
| `~` | Integer types, bitwise not |
| `<` `<=` `>` `>=` | Numbers, `char` and `string` |
| `==` `!=` | Types that implement `Equatable`, compared structurally (section 4) |
| `&&` `\|\|` `!` | `bool` |

- There is no string concatenation operator. Use a template string or `Str.concat(a, b)`. The statement `s += t` joins strings (section 2, Bindings).
- Compound assignment (`+=` and the others) and `x++` / `x--` are statements, not operators, and have no value (section 2, Bindings).
- There is no length operator. Use `Arrays.length(items)`, `Str.length(s)` and `Maps.size(m)`.
- There is no power operator. Use `Float.pow(x, y)` or `Int.pow(x, y)`.
- The vector types of `std/simd` take `+`, `-`, `*`, `/` and the bitwise operators lane by lane (section 4, "Vector types"; not built yet).
- `Int.floorDiv(a, b)` rounds toward negative infinity. It and `Int.divmod(a, b)` (the tuple `#[quotient, remainder]`, truncating like `/` and `%`) return an `Option`, `None` for a zero divisor, so they need no proof. `Int.compare(a, b)` returns an `Ordering`.

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
| `` tag`a${x}` `` calls `tag(strings, ...values)`, `strings.raw` holds the source text, and any expression can be the tag | `tag(["a", ""], [x])`: the values come in one `Array<T>`, the tag picks decoded or raw parts by its first parameter type (`Array<string>` or `Template.Raw`), and the tag is a name (D114, not built yet) |
| `y = x++`, `++x`, `a[i] = b += 1` | Compound assignment and `x++` are statements without a value; `++x` is an error (E0964) |
| `class`, `this`, `new`, `interface` | Not present. Use `struct`, `enum` and `trait` |
| `try`, `catch`, `throw` | Not present. Use `Result` (section 5) |
| Calling an async function without `await` starts it in the background | It starts it too, and the call gives an `Async<T>` handle. A handle left unused is a compile error (E1103): write `await f()`, keep the handle, or let it run with `Async.detach(f())` (section 8) |
| `const o = { a }` builds an object | Not present. `{ a }` in value position is a block expression, which needs `return` (E0933). Write a struct literal: `Point { x: 1 }` |
| `Promise<T>` | `Async<T>` (section 8). The methods are functions: `Async.then(handle, callback)`, `Async.all(handles)` |
| `async function*` returns an object with a `next()` method, and `for await` also takes sync iterables and arrays of promises | `async function*` returns an `AsyncGenerator<T>`, read with `for await` or `Async.next(stream)`. `for await` takes only an `AsyncGenerator<T>` (E1123), and copies of a stream share one position (section 8). `for await` is not built yet (node for-await) |
| A `function` declared inside a function | A compile error (E0930). Declare it at module level as a private function and pass the enclosing values as arguments |

## 3. Modules and imports

- **One file is one module.** Its path from the package root is its name: `src/app/text.nra` is `"app/text"`.
- **Names are private by default.** `export` makes a top-level declaration visible to other modules: `export function`, `export const`, `export type`, `export struct`, `export enum`, `export trait`.
- **Imports are explicit.** A module sees only its own declarations, the names it imports, and the prelude.
- **Prelude.** The prelude holds the types and constructors `Option`, `Some`, `None`, `Result`, `Ok`, `Err`, `Array`, `Map`, `Range`, `Box` and the primitive types, and the functions `print`, `assert`, `panic` and `copy` (section 6, Copies on shared values). Nothing else is in scope without an import. These four are the only global functions.
- **A type in the prelude does not bring its functions.** The functions of `Array`, `Map`, `Range` and `Box` live in `std/array`, `std/map`, `std/range` and `std/box` (D67). `Arrays.push(mutable xs, v)` needs `import * as Arrays from "std/array"`. Without the import the call is a compile error (E0908) whose fix adds the import line. A type the program declares named `Box` or `Map` hides the prelude one.
- **Std aliases are plural by convention.** The std modules of the capitalized prelude types are imported as `Arrays`, `Maps`, `Boxes`, `Ranges`, `Options` and `Results` (D72). Samples and fixes use these names. Any other alias that follows the rules below works too.
- **Standard modules.** `std/array`, `std/box`, `std/build`, `std/bytes`, `std/cell`, `std/char`, `std/cstring` (section 9), `std/env`, `std/f32`, `std/fixedarray` (section 2), `std/float`, `std/fs`, `std/hash`, `std/int`, `std/i8`, `std/i16`, `std/i32`, `std/u8`, `std/u16`, `std/u32`, `std/u64`, `std/io`, `std/json`, `std/lazy` (section 10), `std/map`, `std/math`, `std/option`, `std/path`, `std/process`, `std/ptr` (section 9), `std/random`, `std/range`, `std/result`, `std/shape`, `std/simd` (section 4, not built yet), `std/string`, `std/template` (section 2, Literals; not built yet), `std/test` and `std/time`, and the async modules `std/async`, `std/async/file`, `std/async/net` and `std/async/time` (section 8).
- **Narra parts of builtin modules.** A builtin std module may have functions and constants written in Narra, in `std/src/ext__<name>.nra` (module path `std/ext/<name>`). The compiler loads that module whenever a program imports `std/<name>`, and `Alias.name` and `import { name } from "std/<name>"` find it after the builtins. So far: `Float.PI`; `Int.compare`, `Int.divmod`, `Int.floorDiv`; `Str.find(text, needle)`, the same as `Str.indexOf`; `Arrays.map(items, f)` (the callback takes an effect variable), `Arrays.slice(items, start, finish)` (an `Option`: `None` when the range is outside the array; it copies) and `Arrays.withCapacity(count)` (an empty array that has room for `count` elements; the element type comes from the expected type).
- **`std/shape` and `std/json` (D83).** `std/shape` declares `export enum Value { Null, Bool(value: bool), Int(value: int), Float(value: f64), Text(value: string), List(items: Array<Value>), Record(fields: Map<string, Value>) }`, a dynamic tree for structured data. `std/json` has `parse(text: string): Result<Value, string> uses alloc` and `stringify(value: Value): string uses alloc`. Write `import * as Json from "std/json"` and `import { Value } from "std/shape"`. `parse` follows RFC 8259: a number without fraction or exponent that fits `int` is `Value.Int`, every other number is `Value.Float`; a duplicate key keeps its first place and takes the last value; escapes and surrogate pairs decode, a lone surrogate is an error; nesting deeper than 512 is an error. Errors read `<line>:<column>: <reason>`. `stringify` is compact, writes non-ASCII as UTF-8 and a non-finite `Float` as `null`, and a whole `Float` keeps `.0`.
- **`std/math`, `std/time`, `std/random` and `std/hash`** are Narra source over the primitives of `std/internal/math` (the functions of C's `math.h`), `std/internal/clock` and `std/internal/entropy`.
  - `std/math` has the constants `PI` and `E`; `abs` and `absFloat`; `min`, `max` and `clamp` for any `Ordered` type, so for `int` and `f64`; `pow` and `powInt` (`powInt` returns `None` for a negative exponent, overflow wraps like `*`; use `Int.mulChecked` to detect it); and `sqrt`, `floor`, `ceil`, `round` (half away from zero), `trunc`, `sin`, `cos`, `tan`, `atan2`, `exp` and `log`, all on `f64`. `clamp` returns `None` when `low > high`, and `Some(value)` otherwise.
  - `std/time` has `Instant` and `Duration` (both a count of nanoseconds). `now()` reads the monotonic clock, `elapsed(since)` reads it again, `between(earlier, later)` needs no clock, `unixMillis()` is the wall clock, `sleep(duration)` sleeps. `milliseconds`, `seconds` and `nanoseconds` build a `Duration`, `asMilliseconds` and `asNanoseconds` read it. Reading a clock and sleeping have the `io` effect (section 7).
  - `std/random` is SplitMix64: a `Random` holds one `u64` state, the same seed gives the same numbers on every run and platform, and a value is copied like any struct. `seeded(seed)` and `fromSystem() uses io` (seeded from `/dev/urandom`) make one; `nextInt(mutable random, low, high)` returns `Some(n)` with `n` in `[low, high)` without modulo bias, and `None` when the range is empty; `nextFloat` returns an `f64` in `[0, 1)`; `nextBits` returns 64 random bits; `shuffle(mutable random, mutable items)` is a Fisher-Yates shuffle. It is not for cryptography.
  - `std/hash` has `Hash.of(x)`, the `Hashable` hash of any value (it may change between builds), and `fnv1a(text)`, FNV-1a 64 over the UTF-8 bytes of a string as a `u64`, which never changes. `of` is a keyword, and is accepted as a function name and after `.`, like `from`.
- **`std/process`** (alias `Process`) is builtin: `run(command: string, arguments: Array<string>): Result<#[int, string, string], string>` and `exit(code: int): void`. `exit` has the `io` effect and ends the process with `code` (the operating system keeps the low 8 bits). It flushes pending standard output first, then leaves at once: live values are not released, so the leak check (`NARRA_CHECK_LEAKS`) reports nothing for them, and the code after the call does not run. Like `panic`, a call to `exit` ends a path, so a function that returns a value may end with it. While `run` waits, it forwards SIGTERM, SIGINT and SIGHUP to the child (the child stays in the caller's process group, so Ctrl-C still reaches it). The caller keeps running: `run` returns and reports the child's end as exit code 128 plus the signal number, so a program that calls `Process.run` in a loop must check for it. `narrac run`, `narrac test` and `Test.panics` use the same forwarding for the processes they start, so killing narrac does not orphan the program or a test.
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
- **Fixed-size arrays.** `FixedArray<T, N>` holds exactly `N` elements and is stored inline, like a struct. It needs no heap allocation. `N` is an integer literal or a constant type parameter, and never negative (E0995). One `FixedArray` value takes at most 1 MiB (1048576 bytes), so a local fits well inside a thread's stack; a larger one is E0995, and an `Array` holds more. `[1.0, 2.0, 3.0, 4.0]` builds one when the expected type is a `FixedArray`, and gives all `N` elements. A generic function can take the length as a parameter: `function sum<const N: int>(values: FixedArray<f64, N>): f64`.
- **Fixed-size array indexes are proven.** Indexing a `FixedArray` never panics and has no runtime check. A constant index must be from 0 to `N - 1`, and any other index must be a name that a check proves to be in that range (E0994): an `if` or `while` condition such as `i >= 0 && i < N`, the condition of `for (let i = 0; i < N; i++)` when nothing in the loop assigns `i`, or `for (const i of Ranges.upTo(N))`. An assignment to the name ends the proof. In a generic body the proof is made in each instance, where `N` is a constant. `FixedArrays.get(values, i)` (`std/fixedarray`) returns an `Option<T>` for any index, and `FixedArrays.length(values)` returns `N`. `for (const v of values)` reads a copy of the elements. Equality, hashing, ordering and `Show` follow the element type, as for `Array`.

```ts
import * as FixedArrays from "std/fixedarray"

function sum<const N: int>(values: FixedArray<f64, N>): f64 {
  let total = 0.0
  for (let i = 0; i < N; i++) {
    total += values[i]
  }
  return total
}

function at(values: FixedArray<f64, 4>, index: int): f64 {
  if (index >= 0 && index < 4) {
    return values[index]
  }
  return 0.0
}

function last(values: FixedArray<f64, 4>): Option<f64> {
  return FixedArrays.get(values, FixedArrays.length(values) - 1)
}
```

- **Vector types.** `std/simd` provides fixed-width vectors such as `F32x4`, `F64x2` and `I32x8` (D112; not built yet). Their operators work lane by lane and compile to the target's vector instructions. "Vector types" below has the rules.
- **Maps.** `Map<K, V>` is a hash map. `m[k]` returns `Option<V>`.
- **Maps keep insertion order.** Iteration visits keys in the order they were first inserted. Removing a key and inserting it again moves it to the end.
- **Float keys are not allowed.** `Map<f64, V>` and keys that contain floats are compile errors.
- **Tuples.** `#[int, string]` is a tuple type. It holds a fixed group of at least two values. `#[1, "a"]` builds one, `const #[n, s] = t` (or `let`) destructures one, and `#[0, name]` matches one in a pattern. `pair[0]` reads one element, with that element's type. The index is a constant integer, a literal or a top-level `const`, inside the tuple's range, so the read has no runtime check and cannot panic; any other index is E1010, an index outside the range is E1011. Tuple elements are read-only: `pair[0] = 2` is E0226. The type is written like the literal, in every type position: `Array<#[int, string]>`, `#[#[int, int], string]`. A function that returns several values returns a tuple or a struct.
- **Functions.** `(a: int, b: string) => bool`. A function type may list its effects after the result: `(path: string) => string uses io` (section 7).

### Vector types

Not built yet (D112, nodes simd and simd-library).

- **Twenty types.** `std/simd` has ten lane types at two widths. 128 bits: `I8x16`, `U8x16`, `I16x8`, `U16x8`, `I32x4`, `U32x4`, `I64x2`, `U64x2`, `F32x4`, `F64x2`. 256 bits: `I8x32`, `U8x32`, `I16x16`, `U16x16`, `I32x8`, `U32x8`, `I64x4`, `U64x4`, `F32x8`, `F64x4`. The lanes of `I64x2` are `int`. The names are visible in a module that imports `std/simd` (`import * as Simd from "std/simd"`), whatever the alias.
- **Every type exists on every target** and gives the same results there. A 256-bit type on a machine with 128-bit registers runs as two 128-bit operations.
- **A vector is a plain value,** like a number: assignment copies it, and it needs no heap. It is not a contents type, so it cannot be a `mutable` parameter or the argument of `copy` (section 6). It can be an element of `Array` and `FixedArray`, a struct field, a type argument and the target of a `Ptr`. `==` and `!=` compare whole vectors and give a `bool`. Integer vectors can be map keys; float vectors cannot, like floats. A template shows the lanes as it shows a `FixedArray`. Vectors have no order: `<`, `<=`, `>` and `>=` are compile errors (E1250).
- **Building one.** An array literal with an expected vector type gives every lane, `const unit: F32x4 = [1.0, 0.0, 0.0, 0.0]`; another number of lanes is E1253. `Simd.splat(value)` fills every lane. `Simd.fromFixed(values)` turns a `FixedArray<L, N>` into the vector of `N` lanes of type `L` (E1255 when no vector type has that lane type and count), and `Simd.toFixed(v)` turns it back. `Simd.splat` and `Simd.load` take their vector type from the expected type or a type argument, `Simd.splat<F32x4>(1.0)`; with neither, the call is E1254.
- **Operators work lane by lane.** `+`, `-` and `*` on every vector type, unary `-` on signed and float ones, `/` on float vectors, and `&`, `|`, `^` and `~` on integer vectors. Compound assignment follows (`total += a * b`). Integer lanes wrap on overflow, as integers do. `%`, integer `/`, `<<` and `>>` are compile errors that name the function to use (E1250). Both operands have the same vector type: `v * 2.0` is E1251, whose hint is `v * Simd.splat(2.0)`.
- **Lanes.** `v[i]` reads lane `i`. The index is proven as a `FixedArray` index is (E0994): a constant from 0 to the lane count minus 1, or a name a check proves to be in that range. A vector is a value, so `v[i] = x` is E1252; `Simd.withLane(v, i, x)` returns a copy with lane `i` replaced, with the same proof for `i`.
- **Comparisons give masks.** `Simd.equal`, `Simd.notEqual`, `Simd.less`, `Simd.lessOrEqual`, `Simd.greater` and `Simd.greaterOrEqual` take two vectors of one type and return its mask: the signed integer vector with the same lane width and count (`F32x4` gives `I32x4`, `U8x16` gives `I8x16`), with `-1` in each lane where the comparison holds and `0` elsewhere. `Simd.select(mask, ifTrue, ifFalse)` takes each bit from `ifTrue` where the mask's bit is set and from `ifFalse` elsewhere. `Simd.any(mask)` and `Simd.all(mask)` return whether any lane, or every lane, is not 0.
- **Loads and stores.** `Simd.load(items, start)` reads the lanes `start` onward of an `Array<L>` into a vector and returns `Option<V>`, `None` when they are not all in range. `Simd.store(mutable items, start, v)` writes the lanes and returns `true`, or writes nothing and returns `false` when they do not all fit.
- **Other functions.** `Simd.shuffle(v, [3, 2, 1, 0])` returns a vector whose lane `k` is lane `lanes[k]` of `v`; the list gives one constant per lane, each a lane of `v` (E1260). `Simd.sum(v)` adds the lanes, the high half to the low half until one lane is left, so a float sum is the same on every target. `Simd.min(a, b)` and `Simd.max(a, b)` work per lane. `Simd.abs(v)` takes signed and float vectors; the lowest integer stays itself. `Simd.shiftLeft(v, count)` and `Simd.shiftRight(v, count)` take integer vectors and an `int` count, proven like a shift count against the lane width (E0980); a right shift is arithmetic on signed lanes and logical on unsigned ones. A function given a vector type it does not take is E1261.
- **No effects and no panics.** No `std/simd` function has an effect, allocates or panics.

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
- In a generic struct a field whose type uses a type parameter can have a default (`items: Array<T> = []`, `value: Option<T> = None`). The default is checked once in the declaration with the parameter open, so it must type-check for every type argument (`count: T = 0` is a type error), and it is built for each instance with that instance's types.
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

let current = copy(config)
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

### typeof

`typeof name` in a type position stands for the type of a value in scope: a `const`, `let`, a parameter, a top-level constant or a function (D96). Its use is to keep a type in step with a value, without writing the type twice.

```ts
export const table: Map<string, Array<int>> = Map {}

function fill(mutable target: typeof table, key: string) {
  target[key] = [1]
}
```

For a function it is the full function type, with its effects and `from`:

<!-- context
struct Request {
  path: string
}
struct Response {
  status: int
}
-->
```ts
export function handle(request: Request): Response uses io {
  return Response { status: 200 }
}

export type Handler = typeof handle // (request: Request) => Response uses io
```

- **Names only.** A field path or an expression after `typeof` is a parse error. A parameter can be named by the parameters after it in the same signature: `(request: Request, other: typeof request)`.
- **A name that is not a value is E0982**: a type, a module, a generic function, a function with effect variables, a `mutable` parameter or `requires` lines (none has a function type to write).
- **A `typeof` that refers to its own declaration is E0981**, for example `const count: typeof count = 1` or `function f(handler: typeof f)`.
- **A constant without a written type** has the type of its value, which is known once the constant has been checked: `typeof` of it works in function bodies, but not in a declaration of the same module (a signature, a field, an alias), where it is E1003 and asks for the written type. A constant of another module is always checked first.
- **Nothing reaches the emitted C**: the type is replaced by the type it stands for.

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
- **Arithmetic, bit operators and `<` do not carry over.** `user + 1` and `user < other` are compile errors. The declaring module defines any operations the type needs as functions. `==`, `!=` and map keys work when the base type supports them, through the automatic `Equatable` and `Hashable`; they compare the base value, so `implement Equatable` or `implement Hashable` for a distinct type is an error (E1014).
- **Trait lists.** A distinct type lists traits after its base, as a struct does: `distinct type UserId = int implements Label derives Ordered, Show`. A trait in `implements` needs an `implement` block in the same module or a `derives` entry. A derived trait is the base type's: the derive is an error (E1012) when the base type does not implement the trait, and `Show.show(UserId.from(7))` is `7`. `derives Shape` encodes the base value, also for an `opaque` distinct type, so `Shape.fromValue` builds its values outside the module.
- **No runtime cost.** A distinct type compiles to the same C type as its base.
- **`export opaque distinct type`** hides `from`, so only the declaring module can create values. Other modules can still call the unwrap function.

#### Refinements

A distinct type may restrict its values with `requires` and a list of patterns joined by `&&` (D116). The list follows the base and the trait lists and may start on the next line.

```ts
export distinct type Port = int derives Show requires >= 1 && <= 65535
export distinct type Method = string requires "GET" | "POST"
distinct type Bit = u8
  requires < 8

const DEFAULT_PORT = Port.from(8080)          // proven from the literal

function share(total: int, port: Port): int {
  return total / Port.toInt(port)              // 1 <= n <= 65535, so the divisor is proven
}

function parsePort(raw: int): Option<Port> {
  if (raw >= 1 && raw <= 65535) {
    return Some(Port.from(raw))                // proven by the check
  }
  return Port.check(raw)                       // tested at run time: None here
}
```

<!-- context
distinct type Port = int requires >= 1 && <= 65535
---
-->
<!-- error: E1223 -->
```ts
const port = Port.from(70000)     // compile error: 70000 does not satisfy Port (<= 65535)
```

- **The patterns are `ensures` patterns** (section 13): comparisons `< c`, `<= c`, `> c`, `>= c` and `!= c` with an integer constant that fits the base, literals, and or-patterns of these. `&&` joins patterns only at the top of the list. A malformed list is E1220.
- **The base is an integer type or `string`.** A `string` takes literal patterns only. Any other base, another distinct type included, is E1221. A pattern that does not fit the base (a comparison on a `string`, a bound that is not a constant or does not fit, a u64 bound from 2^63 up, a literal of another type, a wildcard, a name, a constructor, a struct, tuple or array pattern) is E1222.
- **`T.from(value)` is proven at compile time**, as a call to a function with `requires`: the value must match every pattern, proven from a constant, a literal, or an integer name narrowed by a check or by the caller's `requires`. Nothing is checked at run time. An unproven `from` is E1223, and the message names the patterns that failed: `Port.from(8080)` compiles, `Port.from(raw)` compiles inside `if (raw >= 1 && raw <= 65535)`, and `Port.from(70000)` is E1223 (`70000 does not satisfy Port (<= 65535)`). A top-level `const` or a struct field default may hold a `from` call, proven the same way.
- **`T.check(value): Option<T>` tests at run time.** It exists only on a refined type, allocates nothing, and is `None` when a pattern does not match: `if (Port.check(raw) is Some(port))`. `export opaque distinct type` hides it with `from`.
- **Unwrapping gives the fact.** `Port.toInt(port)` has the bounds of the patterns, wherever `port` comes from (a parameter, a field, an array element). It proves divisors, shift counts, `FixedArray` indexes, `requires` arguments and `ensures` returns, and `const n = Port.toInt(port)` starts with it. The fact is the hull of the patterns, as for `ensures`: `Port` gives `1 <= n <= 65535`, `!= 0` gives a non-zero value, and a `string` refinement gives no fact. The unwrapped value is a plain base value: `Port.toInt(port) + 1` is an `int` with no fact.
- **`derives Shape`** decodes the base, then checks the patterns. A value that fails is an `Err` with the usual path: `Json.decode<Config>(text)` gives `Err(".port: 70000 does not satisfy Port (>= 1 && <= 65535)")`, and a string value is quoted. The other derived traits use the base, as for every distinct type.
- **No runtime cost** outside `check` and decoding: a refined type compiles to its base C type, and `from` is the base value.

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
}
```

- Each arm is `pattern => expression` or `pattern => { statements }`.
- Patterns can bind fields, test literals, and add an `if (...)` guard.
- **Alternatives (D98).** An arm may list patterns joined with `|`: `Token.Plus | Token.Minus => ...`. Alternatives nest inside a pattern: `Some(1 | 2)`. Every alternative binds the same names with the same types (E0967 for different names, E0968 for different types), so `Shape.Circle(r) | Shape.Sphere(r) => r` works. A guard covers the whole list: when it fails for one matching alternative, the next alternative that matches is tried with its own bindings, and the arm is taken if any alternative matches with the guard true (as in Rust). A binding keeps the full type of the value: there is no narrowing to the listed constructors and no union type (D73). Exhaustiveness counts each alternative, and an alternative that the arms and alternatives before it already match is an error (E0974). Alternatives are only for `match`; an `is` test and a `const`, `let` or `for ... of` destructuring take one pattern (E0974). `narrac fmt` writes a list on one line when it fits, otherwise one alternative per line, each after the first starting with `|`.
- **Array patterns (D103).** `[]`, `[only]`, `[a, b]`, `[first, ..._]` and `[first, ...rest]` match an `Array<T>` or a `FixedArray<T, N>` in `match` arms and `is` tests. A pattern without `...` matches exactly its number of elements; `...` after `k` elements matches `k` elements or more. `...` appears at most once, as the last element (E0722). Elements are any patterns, literals and alternatives included: `[_, "hash", text, output]`, `[_, "help" | "--help"]`, `Some([head, ...tail])`, `#[[x], y]`. Bound elements are read-only copies, like other pattern bindings. `...rest` on an Array binds a new `Array<T>` with the remaining elements (section 6, Copies on shared values); `..._` copies nothing. On a FixedArray a named rest is an error (E0724), since its type would need `N - k`: write `..._` and index the array. A FixedArray pattern of a length that cannot match is E0723; in a generic body over `N` the length is tested in each instance. An array pattern on any other type (`string`, `bytes`, `Map`, `Range`, a generator, a struct) is E0726, and an array pattern in `ensures` is E0987. For exhaustiveness the lengths are the cases: with `L` the largest element count of any arm, they are the lengths 0 to `L` and "more than `L`", and a pattern with `...` after `k` elements covers every length from `k`. So `[]` with `[_, _, _, ..._]` leaves `[_]` and `[_, _]` missing, and the error lists them as patterns (E0714). A FixedArray with a known `N` has one length; one over a generic `N` needs an arm that covers every length, `[..._]` or `_`. `narrac fmt` prints an array pattern like an array literal, one element per line when it does not fit, with no trailing comma after a rest.
- **Unreachable arms (E0983).** An arm that the unguarded arms before it already match completely is an error: an arm after `_`, a repeated constructor, a repeated literal, or a `_` after arms that cover every constructor. A guarded arm before the same pattern does not cover it. A guarded arm that is already covered is also an error. `narrac apply-fix` removes the arm.
- **`match` as a value.** In value position (after `=`, as an argument, an operand or a `return` value), `match` produces a value. An arm written `pattern => expression` gives the expression. An arm written `pattern => { ... }` is a block expression (see "Block expressions" in section 2): its value comes from `return v`, a `return` inside it ends the arm and not the function, `break` and `continue` cannot leave it, and every path must end in `return`. All arms have the same type. A `match` in statement position keeps block-statement arms, where `return` leaves the function.
- `match` on an `Option<T>` uses `Some(v) =>` and `None =>`.

### Generics

- Type parameters use angle brackets: `function first<T>(items: Array<T>): Option<T>` and `struct Entry<K, V> { ... }`.
- A type parameter can require traits: `<T: Ordered>`, or `<T: Show & Ordered>` for several (see "Traits" below). The type parameters of structs and enums take constraints too: `struct Labeled<T: Show> { ... }`. Field types and implement blocks see them. Every written use of the type with arguments must meet them, in a type, a struct literal or an enum constructor; a failure is a compile error. A type parameter of the enclosing declaration meets a constraint only when it declares the same trait: `function first<T>(pair: Labeled<T>)` is an error, `function first<T: Show>(pair: Labeled<T>)` is not. Constraints are not implied.
- Arithmetic operators do not work on a type parameter. Pass a function or require a trait instead.
- **Structural equality.** `==` compares values structurally. The compiler implements `Equatable` for every type that has no function inside it. Using `==` on a type that contains a function is a compile error.
- **Hashing.** `Hash.of(x)` needs `Hashable`. The compiler implements it for every type that has no function or float inside it.
- Generic code that uses `==` declares `<T: Equatable>`. Code that hashes declares `<T: Hashable>`.
- **Type arguments in a call.** `id<int>(1)`, `Mod.pair<int, string>(a, b)` and `first<Array<int>>(xs)` give the type arguments of a generic function. They fix the type parameters before the arguments are checked, so an argument that does not fit is an ordinary type mismatch. Write one type argument per type parameter or none; a different count is an error, and so are type arguments on a function that is not generic. A `<` after a name starts type arguments only when a balanced `<...>` of type names follows, closed by `>` directly followed by `(`. Every name in it that is not followed by `.` must be a type parameter of the enclosing declaration (so a lowercase `t` works), a primitive type (`int`, `f64`, `bool`, `char`, `string`, `void`), or start with a capital letter. A name of two or more characters with no lowercase letter (`MAX`, `LIMIT_2`) is a constant, not a type, so `a < b > (c)` and `both(x < MAX, MIN > (y))` are still comparisons. The remaining ambiguity is a value that is capitalized but not SCREAMING_CASE, as in `f(x < Limit, Cap > (y))`, which reads as a call with type arguments; parenthesize the second comparison, `(Cap > y)`. Single capital letters (`T`, `K`, `V`) are types. An integer literal (`3`, `-1`) between the brackets is a constant type argument, whatever the name before `<`, as in TypeScript: `both(a < 3, 4 > (c))` is the call `a<3, 4>(c)`, so write the comparison in parentheses, `both((a < 3), 4 > (c))`. Constructors (`Some<int>(1)`) and builtin functions take no type arguments; annotate the variable instead, as in `const x: Option<int> = None()`. Function types are not accepted inside the arguments.
- **Constant type parameters.** `const N: int` in the type parameters of a function or a struct stands for an integer, not a type: `function repeat<const N: int>(text: string): string`, `struct Vector<const N: int> { ... }`. Its argument is an integer literal (`repeat<3>("ab")`, `Vector<-1>`) or a constant parameter in scope (`Vector<N>`). A call infers it from the argument types or the expected type like a type parameter (`const plane: Vector<2> = zeros()`); one that only the body uses must be written. A type argument never computes: `Vector<N + 1>` is a compile error, and so are a local, a top-level constant or a type where a constant belongs, and a constant where a type belongs (`value: N`). In the body `N` is an `int` value, known in each instance: `total / N` needs no check, and an instance with `N` equal to 0 is the usual divisor error, which names the instance and its call. A local or parameter named like a constant parameter is a shadowing error. A constant parameter does not satisfy a `requires` condition yet: a call that passes `N` to a function with `requires` is E0985. Each value is its own instance, so `repeat<2>` and `repeat<3>` are two functions and `Vector<2>` and `Vector<3>` two types. The kind is `int` only for now; the sized integer kinds (`const N: u8`) may come later. Enums and type aliases take no constant parameters.
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

```ts
struct Vector<const N: int> {
  values: Array<f64>
}

function dimension<const N: int>(_vector: Vector<N>): int {
  return N
}

function share<const N: int>(total: int): int {
  return total / N
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
- **Generic types.** A generic struct or enum is implemented once for all its instances: `implement Show for Pair<T> { ... }` names the type with the type parameters of its declaration, by the same names and in order. A concrete type (`Pair<int>`), a missing list, a wrong count or another name is a compile error, so there is no specialization. The block adds no constraints of its own: its functions see the constraints of the type's declaration, so a member that calls `Show.show` on a field of type `T` needs `struct Pair<T: Show>`. The functions take no type parameters of their own.
- **No implementations anywhere else.** A module cannot implement a trait for a type declared in another module. To add a trait to another module's type, wrap it in a `distinct type` or a struct.
- `derives` asks the compiler to write the implementation: `struct User derives Show, Ordered, Shape { ... }`.

**Calling**

- Trait functions are called through the trait name: `Label.label(price)`. There is no `price.label()` form.
- In generic code, `<T: Show>` allows `Show.show(value)` for a `value: T`.
- A trait function may mention `Self` only in its result, for example `function fromValue(value: Value): Result<Self, string>`. A call names the type that `Self` stands for: `Shape.fromValue<User>(value)`. Exactly one type argument is allowed. Without it, the call is an error unless the expected type fixes `Self`. In a generic body `<T: Shape>`, `Shape.fromValue<T>(value)` is resolved when the body is instantiated.

**Trait values**

- `Dynamic<Trait>` holds a value of any type that implements the trait. `Dynamic.of(value)` creates one when the expected type is known.
- `Array<Dynamic<Drawable>>` holds values of different types. `Drawable.draw(item, mutable canvas)` calls the implementation of each item's own type.
- **Which traits.** Every function of the trait takes `Self` as its first parameter, once and not inside another type (`Array<Self>` does not count), and mentions `Self` nowhere else, the result included. Other parameters may be `mutable` (`mutable canvas: Canvas`). So `Show`, `Hashable` and traits like `Drawable` qualify, and `Equatable`, `Ordered` (two `Self` values) and `Shape` (`fromValue` returns `Self`) do not.
- **Layout.** `Dynamic.of(value)` makes one allocation: a reference-counted cell that holds a pointer to the table of the value's type for the trait, then the value. A `Dynamic` is a pointer to that cell, so a copy is one retain and the value is never copied. The value is read-only: no field of it can be written through a `Dynamic`, which is why the copies can share it: `copy(item)` copies nothing and is an error, as for a `string`. A trait function that takes `mutable value: Self` cannot be called through a `Dynamic`.
- A call through a `Dynamic` names the trait as any trait call does, and its effects are the trait function's declared effects. A `Dynamic` has no `==`, cannot be a map key and implements no trait, so it does not meet a bound such as `<T: Show>`: loop over an `Array<Dynamic<Show>>` and call `Show.show(item)` on each. In a generic body `<T: Drawable>`, `Dynamic.of(value)` with `value: T` works. `Dynamic` is a prelude name, like `Box`: a type of the module named `Dynamic` hides it.

**Standard traits**

| Trait | Purpose | Implemented |
| --- | --- | --- |
| `Equatable` | `==` and `!=` | Automatically for types without functions |
| `Hashable` | `Hash.of`, map keys | Automatically for types without functions or floats |
| `Ordered` | `compare`, sorting | By `derives` or `implement` |
| `Show` | Text for logs, tests and errors | By `derives` or `implement` |
| `Shape` | Conversion to and from the `Value` tree of `std/shape`, for serialization | By the compiler for prelude types, by `derives` or `implement` for a struct or enum |
| `Sendable` | Values that may cross threads (section 6, not built yet) | Automatically for types without functions, `Cell`, `Async`, generators and the other types section 6 lists; never by `implement` |

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
- **There is no implicit conversion.** Arithmetic, comparison and bitwise operators need operands of the same type; `u8 + i32` is E0295, and the message names the conversion functions. Unary `-` on an unsigned type is E0296. Shift counts have the type of the value shifted.
- **Divisors and shift counts are proven (D92).** Integer `/`, `%`, `/=` and `%=` compile only when the divisor is a non-zero constant (a literal, a negated literal, or a top-level `const` whose value is one) or a local integer narrowed by a check in the same function: `!= 0`, `> 0`, `< 0`, `>= 1`, `== 3` and the like, joined by `&&`, or by `||` on the failing side, and negated by `!`. A `||` that holds (or an `&&` that fails) gives the hull of its sides when every side compares the same local with a constant: in `if (d < 0 || d > 0)` the `then` branch knows `d` is non-zero; any other `||` gives nothing, and on a u64 a side with a bound from 2^63 up gives nothing. A check between two integer locals of the same type (`if (index < count)`) records a relation between them, which ends when either is assigned, and passes bounds one step: with `index < count` and `count <= 63`, `index` is at most 62. Relations do not chain (`a < b` and `b < c` say nothing about `a` and `c`), and on a u64 an end from 2^63 up is not passed on. Anything else is E0979, and the message says what to add. A shift count (`<<`, `>>`, `<<=`, `>>=`) must be a constant from 0 to the bit width minus 1 (63 for `int`), or a local checked at both ends (`n >= 0 && n < 64`); an unsigned count needs only the upper bound. Anything else is E0980. Float division is unchanged. The emitted C and the runtime have no division or shift panic: `MIN / -1` is `MIN`, `MIN % -1` is 0.
- **What a check proves.** A check narrows in the `then` branch of an `if` (the `else` branch gets the opposite), in the body of a `while`, in the right operand of `&&` (`||` gets the opposite), and after an `if` whose `then` branch ends in `return`, `break`, `continue` or a panic. A narrowing ends at an assignment to the name (`n = ...`, `n += 1`, `n++`), at a call that passes the name as `mutable`, and at the start of a loop that does either; a new check narrows again. A name that an arrow function captures by reference is not narrowed inside the arrow function. Only a plain local name narrows: a field or an element (`size.count`, `counts[i]`) and any other expression (`n + 1`, `Arrays.length(items)`) do not, so copy the value into a local first and check the local. A constant is a literal, a negated literal, a top-level `const` whose value is one, or a `const` local that was set from one (`const items = 7`).
- **Overflow wraps at the type's own width:** `u8` `200 + 100` is 44 and `i8` `-128 - 1` is 127, in every build.
- Conversions are explicit functions named after the target type, in one module per type (`I32.from(x)`, `I32.tryFrom(x)`, `I32.wrap(x)`; modules `std/i8` to `std/u64`; `std/int` serves `i64`), and `Float.fromInt(x)`, `Float.truncate(f)`. `from` is for conversions where every value fits, `tryFrom` returns an `Option`, `wrap` keeps the low bits. Each of the seven modules also has `toString(x)` and `parse(text, base)`, and `std/int` has `from`, `tryFrom` and `wrap`. `X.from(v)` takes any integer type whose every value fits `X`; another integer type is E0940 and a value that is not an integer is E0941. `tryFrom` and `wrap` take any integer type. `parse` is `Int.parse` with the range of the type in its error text. A literal argument is an `int`, so `I32.from(5)` is E0940; write `const x: i32 = 5`.
- **`f32`** is a 32-bit float (C `float`) beside `f64`. A float literal takes `f32` when the context expects it and `f64` otherwise; an integer literal never becomes a float. A literal too large to be finite in `f32` is E0297. `f32` and `f64` do not mix in arithmetic or comparison (E0295). `f32` implements `Equatable`, `Ordered` and `Show` (the shortest text that reads back as the same `f32`) but not `Hashable`, so it is not a map key. `F32.from(v)` takes an `f64` (rounds to nearest) or any integer type, `F32.truncate(v)` gives an `Option<int>` (`None` for NaN or a value outside `int`), `F32.toString(v)` a string (module `std/f32`), and `Float.from(v)` takes an `f32` (exact) or any integer type. Another argument type is E0941.

### Strings

- `string` holds valid UTF-8. Functions that build a string from bytes return `Result` and fail on invalid UTF-8.
- **Positions are byte offsets of type `ByteIndex`**, starting at 0. `ByteIndex` is declared as `export opaque distinct type ByteIndex = u64`, so only `std/string` can create one.
- **Positions come from the string.** `Str.indexOf`, `Str.find` and iterators return `ByteIndex` values that always fall on a character boundary.
- **Slices share memory.** `Str.slice` and `Arrays.slice` return a view into the original buffer, with no copy. The view keeps the whole buffer alive. Use `copy(...)` (section 6, Copies on shared values) to keep a small part of a large buffer without keeping the rest.
- **Slicing checks boundaries.** `Str.slice(s, from, to)` returns bytes `from` up to but not including `to`. It returns `None` when a position is not on a character boundary, is out of range, or `from` is after `to`, and `Some(text)` otherwise.
- `Str.length(s)` is the byte length as `ByteIndex`. `Str.charCount(s)` counts characters and takes time proportional to the length.
- Grapheme clusters, case mapping and normalization are in `std/unicode`.

### Bytes

- `bytes` is an immutable byte sequence with no UTF-8 guarantee. It has the representation of `string`, so `Bytes.fromString(text)` shares the buffer, and `Bytes.toString(data)` shares it too when the bytes are valid UTF-8. Otherwise it returns `Err("invalid UTF-8 at byte N")`, with the offset of the first bad byte.
- `std/bytes` (alias `Bytes`): `fromString(text): bytes`, `toString(data): Result<string, string>`, `length(data): int`, `at(data, index: int): Option<u8>`, `slice(data, start: int, end: int): Option<bytes>`, `concat(left, right): bytes`, `fromArray(items: Array<u8>): bytes` and `toArray(data): Array<u8>`. `at` and `slice` return `None` out of range. `slice` copies the bytes.
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
  const key = Str.slice(line, 0, i)
  if (key is None) {
    return None
  }
  return Some(Pair {
    key: Str.trim(key.value),
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
- **Std functions that cannot handle their input return `Option` (D92).** None of them panics. Each has one way to fail, so `None` says all there is to say.

| Function | Result | `None` when |
|---|---|---|
| `Str.slice(text, from, to)` | `Option<string>` | a position is out of range or inside a character, or `from` is after `to` |
| `Str.split(text, separator)` | `Option<Array<string>>` | the separator is empty |
| `Str.repeat(text, count)` | `Option<string>` | `count` is negative |
| `Bytes.at(data, index)` | `Option<u8>` | `index` is out of range |
| `Bytes.slice(data, start, end)` | `Option<bytes>` | the range is out of bounds |
| `Float.truncate(value)`, `F32.truncate(value)` | `Option<int>` | `value` is NaN or outside `int` |
| `Arrays.filled(value, count)` | `Option<Array<T>>` | `count` is negative |
| `Ranges.stepped(start, end, step)` | `Option<Range>` | `step` is 0 |
| `Math.clamp(value, low, high)` | `Option<T>` | `low` is greater than `high` |
| `Math.powInt(base, exponent)` | `Option<int>` | `exponent` is negative |
| `Random.nextInt(mutable random, low, high)` | `Option<int>` | `low` is not below `high` |

- **Std panics that remain.** `Str.sliceFrom`, `Str.charAt` and `Str.nextChar` still panic on a position that is out of range or inside a character, and array indexing keeps its panic. They are not yet converted.
- **`Maps.update(mutable map, key, initial, change)`.** Sets `map[key]` to `change(old)`, where `old` is the current value, or `initial` when the key is missing. `change` is a function from the value type to the value type.

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
let squares: Array<int> = []
if (Arrays.filled(0, 1000) is Some(zeros)) {
  squares = zeros
}
for (const i of Ranges.upTo(1000)) {
  squares[i] = i * i          // the count is 1, so the write is in place
}
```

- `items[i] = v` and `u.age = v` require the target to be a local `const` or `let`, a `mutable` parameter, or a field or element of one. A plain parameter and a top-level constant cannot be changed. Pattern, `for (const ...)` and destructured bindings are read-only; a `for (mutable item of items)` binding can be changed and changes `items` (section 2, Control flow).

### Copies on shared values

When an update targets a value whose reference count is greater than 1, the value is copied first. The copy is correct, but it costs time proportional to the value's size, and the source does not show it. `copy(x)` makes the copy visible (D94):

- **`copy(x)`** duplicates a value at any depth. At the call it is one reference count increment, so it costs nothing until a write. A later write to `x` or to the copy then copies only the objects on the written path, once each: after `const other = copy(team)`, `other.scores[1][0] = 30` copies the outer array of `other.scores` and one row, not the other rows.
- `copy` takes one argument of a type that a write changes in place: a struct (all-scalar ones too), `Array<T>`, `Map<K, V>`, `FixedArray<T, N>`, `Box<T>`, `StringBuilder`, `Generator<T>` or `Coroutine<In, Out>`, a tuple holding one of them, a distinct type over one, or a type parameter. It also takes an enum whose payload can hold one of them, at any depth (`Option<Array<int>>`, `Result<Team, string>`, a recursive enum with a struct payload). An enum value itself is never changed in place, but a part taken out of its payload can be: the copy retains the payload. A part taken out of the copy still shares the original's part while the copy is held, so changing it takes a `copy(...)` of the part, which then copies only that part. Any other argument (numbers, `bool`, `char`, `string`, `bytes`, enums whose payloads hold only such values, function values) is a compile error whose fix removes the call, and so are a second argument and `copy(mutable x)` (E0975).
- **A named rest copies (D103).** `[first, ...rest]` on an `Array<T>` builds `rest` as a new Array of the remaining elements, each retained: like `Arrays.slice`, it costs time linear in the tail and one allocation each time the pattern matches. A recursive `[first, ...rest] => sum(rest)` is therefore quadratic in the length; walk an index instead when the array is long. `..._` copies nothing.
- `copy` is a prelude function, not a keyword. A local or parameter may be named `copy`, as one may be named `panic`; inside its scope a call `copy(x)` then calls the local. A top-level function named `copy` is never called: the prelude's `copy` is found first.
- **The debug copy check.** A debug build run with `NARRA_CHECK_COPIES` set counts every copy on write of an array, map, `Box`, `StringBuilder` or generator, and at exit reports the ones that no `copy(...)` explains, with exit code 103, like the leak check (D80): `copy check: copies on write that no copy(...) explains: array 2`. Each `copy(x)` gives the objects it shares one credit, a copy on write uses one credit of the object it copies and gives one to each object the two copies now share, and a credit ends when the object's count falls back to 1. So at most one hidden copy per `copy(...)` call goes unreported. The copy that `for ... of` and `yield*` make of a generator they do not own (section 8) counts as explained. The credits a copy of a heap enum gives the parts of its payload end when the last holder of that copy is released, and all of them when the payload box is freed. All-scalar structs live inline and never copy on write; the compile-time shared-write rule covers them. Release builds have no credits and no hooks.

<!-- context
import * as Arrays from "std/array"
-->
```ts
const scores = [[1, 2], [3, 4]]
let other = copy(scores)
other[1][0] = 30            // copies the outer array and one row; scores is unchanged
Arrays.push(mutable other, [5, 6])
```

A copy of an enum with a tracked payload:

<!-- context
struct Row {
  cells: Array<int>
}
enum Slot {
  Empty
  Filled(row: Row)
}
---
-->
```ts
const found: Option<Array<int>> = Some([1, 2])
const kept = copy(found)
if (kept is Some(items)) {
  let list = copy(items)
  list[0] = 9                 // copies this array only; found and kept still hold [1, 2]
}

const slot = Slot.Filled(Row { cells: [3, 4] })
const other = copy(slot)
if (other is Slot.Filled(row)) {
  let edited = copy(row)
  edited.cells[0] = 5         // copies the cells of this row only
}
```

An enum whose payloads hold only scalars and strings has nothing to copy:

<!-- error: E0975 -->
```ts
const count: Option<int> = Some(3)
const same = copy(count)      // E0975: remove the call
```

### Shared writes

A write to a place is a compile error while another name that is still used shares that place, a place that contains it or a place inside it (E0978, D94). Without the rule the write would copy the value, and the source would not show it. Sharing starts where a value is read into a second name and ends when that name is no longer used:

<!-- context
struct Bag {
  items: Array<int>
  count: int
}
---
-->
<!-- error: E0978 -->
```ts
let bag = Bag { items: [1, 2], count: 0 }
const items = bag.items
bag.items[0] = 9              // E0978: items still shares bag.items
const seen = items[0]
```

`copy(...)` around the value that started the sharing is the fix (`const items = copy(bag.items)`), and `narrac apply-fix` writes it. The write then copies nothing at the write:

<!-- context
struct Bag {
  items: Array<int>
  count: int
}
---
-->
```ts
let bag = Bag { items: [1, 2], count: 0 }
const items = copy(bag.items)
bag.items[0] = 9
const seen = items[0] + bag.items[0]    // 1 + 9
```

- **What writes.** An assignment to a field or element, `Arrays.push(mutable items, v)` and any other `mutable` argument. Passing `mutable R` writes `R` and everything below it. A non-escaping closure that writes a captured `const` writes it where the closure is passed. All-scalar structs are checked too, though they live inline and a write never copies them.
- **What shares.** A `const` or `let` initialized from a place, an assignment, a pattern, destructured or `for (const ...)` binding, a struct, array, tuple or map holding the place, the result of a call whose callee has `from` for that argument (above), and a function value made from an arrow function, which shares what it captured: calling it uses those values. A parameter, a top-level constant, a `Cell` and a value read from an identity type count as used for the whole body, so a write to the value they share is an error at once. `copy(x)` shares nothing. Sharing runs in both directions and through chains: after `const b = a; const c = b`, a write to any of the three is checked against the other two.
- **When it ends.** Assigning a local or a field path (`bag = other`, `bag.items = []`) ends its sharing; assigning an indexed place (`rows[0] = x`) does not. Liveness is checked inside one function body: the error names the later use. A `const` or `let` holds its value until its block ends, so a write after its last use is still an error when the write would copy an array, map, `Box`, `StringBuilder` or generator it holds; the error then names the end of the block (D104). A write that copies nothing is accepted after the last use: a field of a struct of numbers, or a `mutable` call on a struct that holds the shared value only inside an enum, whose payload is not a place. An enum built from a place holds that place's value like any other value: after `const held = Some(bag.items)`, `Arrays.push(mutable bag.items, 4)` is an error until `held`'s block ends. A destructuring `const` or `let` holds its fields the same way. A `const` or `let` initialized with `copy(...)` holds its value the same way: a part taken out of it shares the original until its block ends. The bindings of an `is` or a `match` on a value that is not a place (a call result, a constructor) hold it until the `if` block or the arm ends, and so does a `for (const ...)` binding over such a value until the body ends; bindings of a place end at their last use, since the place keeps the value. While a `for ... of` loop other than `for (mutable ...)` runs, its iterator holds the place it walks, so a write to that place in the body is an error; loop over `copy(items)` or write after the loop.
- **Which index places are the same.** Two different integer literal indexes name different places: `const first = rows[0]; rows[1][0] = 4` is accepted. Every other index may name any element, so an identifier or other expression (`rows[i]` and `rows[j]`, `rows[i]` and `rows[0]`) and a map key match every element and the other indexes. The same literal index is one place, and assigning it does not end its sharing (a known false positive: `let row = grid[0]; grid[0] = [9]; row[0] = 5` is reported while `grid[0]` is used later).
- **Known gap: a part of a temporary copy of a heap enum.** `copy(x)` shares nothing, so `let row = takeLeftRow(copy(tree))` is not paired with `tree`. The temporary copy ends with the statement and takes its credits back, so a later write to `row` copies `tree`'s part without a `copy(...)` that explains it. The debug copy check reports it. Take `copy(...)` of the part instead.
- **Known gap: stores into a `mutable` parameter.** A function that assigns a value sharing one parameter into a `mutable` parameter (`into.items = other.items`) leaves the caller's two arguments sharing after the call, and no rule sees it. Store `copy(other.items)`. The debug copy check reports the copy.
- **Known gap: shared elements.** Elements that share one value are not tracked. `Arrays.filled([0, 0], 3)` and `[row, row]` make an array whose elements share one inner array, so a write to `grid[0][0]` through a name that holds it is accepted and copies at run time. The debug copy check reports it.
- **Where it runs.** Every command that checks a package (`build`, `run`, `test`, `emit`, `check`, `apply-fix` and the language server), also when the package has other errors. A write that another error already points at is reported once.

<!-- context
struct Row {
  cells: Array<int>
}
---
-->
```ts
let rows = [Row { cells: [1] }, Row { cells: [2] }]
const first = rows[0]
rows[1].cells[0] = 4          // accepted: rows[0] and rows[1] are different places
const seen = first.cells[0] + rows[1].cells[0]
```

### Shared results

A function whose result shares a parameter says so with `from <parameter>` after the result type (D94). Nothing is inferred across functions: a call's result shares one of its arguments only when the callee's signature or function type has `from`, so a caller sees it in the signature.

<!-- context
import * as Arrays from "std/array"
struct SortItem {
  key: int
}
-->
```ts
function firstItem(items: Array<SortItem>): SortItem from items {
  return items[0]
}

function lastItem<T>(items: Array<T>): T from items {
  return items[Arrays.length(items) - 1]
}
```

- **What shares.** A returned value shares a parameter when it is the parameter, a field, element or narrowed payload of it, a local, pattern or `for ... of` binding that took such a value, a struct, array, tuple, map or constructor holding one (a spread `{ ...base }` holds only the fields of `base` it does not write), or the result of a call whose `from` names an argument that shares it. A local also takes the values assigned to it and the values passed beside it when it is a `mutable` argument. A value whose type holds no struct, array, map, `Box`, `StringBuilder`, generator or type parameter (numbers, `bool`, `char`, `string`, `bytes`, enums without such payloads, function values) shares nothing. `copy(x)` is unshared.
- **The check (E0976).** Returning a value that shares a parameter the marker does not name is an error. Without a marker the fix adds `from <parameter>`; with a marker that names another parameter, the fix wraps the value in `copy(...)`. A `from` that names no parameter is also E0976.
- **Generic results.** `from items` covers the whole parameter and everything reached through it, at any depth, whatever the type arguments are: a result of type `T` taken out of `items: Array<T>` shares `items`. When `T` is a scalar type the marker is allowed and changes nothing.
- **Arrow functions.** An arrow function whose result shares one of its own parameters writes `from` after its result type: `(row: Row): Array<int> from row => row.cells`. An arrow function that writes no `from` takes the `from` of its expected function type, as it takes its parameter types, so `(row) => row.cells` fits a `(row: Row) => Array<int> from row` slot. When the expected type has no `from`, or the arrow function writes no result type, a result that shares a parameter is E0976 with a fix that wraps the value in `copy(...)`. `from` names only a parameter, never a capture: an arrow function whose result shares a captured value is E0976, and the fix wraps the value in `copy(...)`.
- **Function types.** `from` is part of a function type, like `uses` (section 7), and names one of the type's parameter names: `(grid: Grid) => Array<int> from grid uses alloc`. `from` comes before `uses`. A function with a fresh result fits a slot whose type has `from`; a function with `from` does not fit a slot without the same `from` (E0977), unless the result holds no tracked value. The fix adds the marker to the slot's type, naming the parameter `source` when the type writes no name for it. A type printed by the compiler names the shared parameter `source`: `(source: Grid) => Array<int> from source`.
- **Trait functions.** An implementation may leave out the `from` its trait function has, but may not add one the trait function lacks.
- **Generators** are not checked: a generator's result is the generator.

<!-- context
import * as Arrays from "std/array"
struct Grid {
  rows: Array<Array<int>>
}
-->
```ts
struct Picker {
  pick: (grid: Grid) => Array<int> from grid uses alloc
}

function firstRow(grid: Grid): Array<int> from grid {
  return grid.rows[0]
}

function rowSizes(grid: Grid): Array<int> uses alloc {
  return [Arrays.length(grid.rows)]
}

function listPickers(): Array<Picker> uses alloc {
  // A fresh result fits the `from` slot too.
  const second = (grid: Grid): Array<int> from grid => grid.rows[1]
  return [Picker { pick: firstRow }, Picker { pick: rowSizes }, Picker { pick: second }]
}
```

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
- **Cycles through identity types leak** (D80). Identity types (`Shared<T>`, `Sender<T>`, `Receiver<T>`, `Async<T>` and `Cell<T>`; section 8) are handles to one object, so they can form a cycle. There is no cycle collector and no type rule against it. Memory safety holds, but the objects in the cycle are never freed and resources they hold never close. Every cycle passes through an identity object, so the debug leak check (`NARRA_CHECK_LEAKS`, section 13) reports it at exit, and a test that leaks fails. Two patterns leak:
  - A handle stored in its own payload: `Shared.update(state, ...)` that stores `Some(state)` in the value of `state`, or a `Sender` sent into its own channel.
  - A closure stored in a `Cell` that captures the same `Cell`.
- **Breaking a cycle.** Clear the slot before the last outside handle is dropped (`Cells.set(cell, None)`), or keep the graph in `std/arena` and refer to its nodes with handles.
- A later version may add a collector limited to identity objects. It would only free memory that this rule leaks, so no program changes meaning.
- **Graphs** use `Arena<T>` and `Handle<T>` from `std/arena`. A handle is an index plus a generation number. `Arena.get(arena, h)` returns `None` when the slot was freed or reused. Handles can point anywhere in the arena, including in cycles, and the arena frees all its slots at once.

### Values across threads

Not built yet (D111; nodes threads, threads-check).

- **Only `Sendable` values cross threads.** `Sendable` is an automatic trait: the compiler implements it for every type that holds none of these, at any depth: a function type, `Dynamic<Trait>`, `Ptr<T>`, `Cell<T>`, `Async<T>`, `Generator<T>`, `Coroutine<In, Out>`, `Scope`, `Thread<T>`, and the sockets of `std/async/net`. `Shared<T>`, `Sender<T>`, `Receiver<T>` and `Atomic<T>` are `Sendable`. A capture of a thread function, a thread result, and the type argument of `Channel.create`, `Shared.create` and `Atomic.create` must be `Sendable` (E1190). Generic code declares `<T: Sendable>`. `implement Sendable` is a compile error (E1194).
- A value captured by a thread function, sent through a channel or stored in a `Shared` is marked shared. Shared values use atomic reference counts. Unshared values use plain counts. Only a program that imports `std/thread` (or a module built on it) has shared values; any other program uses plain counts throughout.
- An in-place update on a shared value copies first, so threads never see each other's writes.
- Safe code has no shared mutable memory, so no data race is possible. Section 8 covers threads, channels and the few types that are shared on purpose.

## 7. Effects

An effect marks something a function can do beyond computing a value. The set is fixed:

| Effect | Meaning |
| --- | --- |
| `alloc` | May allocate heap memory |
| `io` | May perform input or output, read the clock, or read the environment |
| `unsafe` | Calls code that the compiler cannot check (section 9) |

Rules:

- Exported functions declare their effects after the return type: `function f(s: string): int uses alloc { ... }`.
- The compiler infers effects for private functions and arrow functions. A private function that writes a `uses` list is checked against it, as an exported one is.
- A function's effects include the effects of everything it calls. Calling an `io` function from a function that does not declare `io` is a compile error.
- Panics are not an effect. Any function can panic.
- The entry function (section 15) has every effect. It may be `async` (section 8).
- `async` and `function*` are not effects. They mark the type of function a declaration is (section 8).
- **Function types carry effects** (D86): `(v: int) => int uses io`. A call through a function value has the effects of its type, and a function cannot be stored where the type lists fewer effects than it has. A function value's own type lists none, so a top-level constant that holds a function with effects writes its type: `const step: (v: int) => int uses io = loud`.
- **Generator types carry effects** (D78), as function types do: `Generator<string> uses io` is a generator whose body does I/O when it runs. `Generator<T>` alone is pure. The effects are charged where the body runs: a `for ... of` over the generator, a `yield*` of it, and `Co.resume`. Creating a generator charges only `alloc`, for its frame.
- A generator value may be used where a generator type with more effects is expected, not fewer: passing a `Generator<T> uses io` where `Generator<T>` is expected is a compile error.
- **A `uses` list inside a type** (D117, not built yet) takes one effect, or a comma list in parentheses: `() => int uses io`, `() => int uses (io, alloc)`, `Map<Generator<int> uses (io, alloc), int>`. A `,` after the one effect ends the type, so it can be followed by another type argument, tuple element or parameter, or by a trailing comma. A function declaration's own list, after its result type, is a bare comma list without parentheses: `function f(): int uses io, alloc`. A comma after a result type's one effect is an error (E1240): `function* lines(): Generator<string> uses io, alloc {` must be written `Generator<string> uses (io, alloc)`. An effect name in a type position is an error (E1241). Types in messages and tools are printed in the same form. Until this is built, a `uses` list in a type takes every following `, name` as an effect.

### Effect variables

A function that calls a function it receives has that function's effects. An effect variable states this in a signature.

```ts
export function map<T, U, effect E>(items: Array<T>, f: (item: T) => U uses E): Array<U> uses alloc, E
```

- Effect variables are declared in the type parameter list with the `effect` keyword, and appear only after `uses`.
- **They are bound by a parameter.** An effect variable in the function's own `uses` list must also appear in a parameter's function type or generator type: `export function sum<effect E>(numbers: Generator<int> uses E): int uses E`.
- At each call, the variable takes the effects of the function or generator passed in. `map(names, Str.length)` has `alloc`. `map(paths, File.readText)` has `alloc, io`. A generator result type that names the variable takes them too: with `relay<effect E>(g: Generator<int> uses E): Generator<int> uses (E, alloc)`, `relay(lines(path))` is a `Generator<int> uses (alloc, io)` (the parentheses are D117, not built yet).
- **A function used as a value keeps its variables** (D91). After `const apply = twice`, `apply(loud, 1)` has the effects of `loud`, as `twice(loud, 1)` does, through a top-level or a local constant.

## 8. Generators and concurrency

Narra has three tools. Each covers one type of concurrent work:

| Work | Tool |
| --- | --- |
| Produce values one at a time, on demand | Generators |
| Run computations in parallel on several cores | Scoped threads and channels |
| Wait on many I/O operations at once | Async functions on an event loop |
| Produce values one at a time that need I/O | Async generators |

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
- **Effects** (section 7, D78). An exported generator writes the effects of its body on its return type: `export function* lines(path: string): Generator<string> uses (io, alloc)` (parentheses: D117, not built yet). A private one gets them inferred into its type. A body that uses an effect its declared type omits is a compile error.
- **Calls of generator functions are not recursion.** A generator that calls itself, directly or through another generator, creates a new value, so D37 and tail loops do not apply.
- **`Coroutine<In, Out>`** (D79) is a generator that also receives a value on each resume. It is a separate type from `Generator<T>`: declare it with `function*` and the return type `Coroutine<In, Out>`, as in `function* total(): Coroutine<int, int>`. Inside it, `const input = yield output` yields `output` and, when the coroutine is resumed, binds the value passed to `Co.resume`. The `yield` with a value is valid only in a coroutine, and only as the initializer of a `const` or `let` that names one variable and has no type annotation; a plain `yield output` statement is also valid and drops the input. The input of the first resume is dropped, as in TypeScript, because no `yield` is waiting for it. `yield*` is not valid in a coroutine, and `for ... of` over a coroutine is a compile error (a loop has no input to send). A coroutine type carries effects like a generator type: `Coroutine<string, int> uses io`.
- **`Co.resume`** from `std/coroutine` (`import * as Co from "std/coroutine"`) runs a generator or a coroutine to its next `yield`, changing `g` in place: `Co.resume(mutable g)` for a `Generator<T>` and `Co.resume(mutable g, input)` for a `Coroutine<In, Out>`. It returns a `Step<Out>`, an enum of the same module (`import { Step } from "std/coroutine"`) with the constructors `Step.Yielded(value)`, `Step.Done` and `Step.Panicked(message)`, always qualified (D71). `Step` derives `Show` and `Equatable`. A finished generator or coroutine answers `Step.Done` to every later resume. The call charges the effects carried on the type of `g` (D78), and `alloc` for the `Step`. A shared generator is copied before it resumes (the count above 1), as in `for ... of`. `Panicked` is only reported in test builds (`narra test`): a panic inside the coroutine then ends the coroutine and `Co.resume` returns `Step.Panicked(message)` with the bare message; what the coroutine held at that moment is not freed. In other builds a panic aborts the program.
- **A generator is a value like any other.** Resuming one by hand requires a `mutable` argument. `for ... of` iterates over its own copy. `Generator<T>` has no `==`, no hashing and no `Show`.
- **Only generator bodies become state machines.** A `yield` cannot cross a normal function call, so other functions compile as plain C functions.

### Threads

Threads run in a scope. Every thread started in a scope finishes before the scope returns. Not built yet (D111; nodes threads, threads-check): the compiler has no `std/thread`.

```ts
import * as Thread from "std/thread"

const total = Thread.scope((scope) => {
  const left = Thread.start(scope, () => sum(firstHalf))
  const right = Thread.start(scope, () => sum(secondHalf))
  return Thread.join(left) + Thread.join(right)
})
```

- **`Thread.scope(f)`** calls `f` with a `Scope`. It waits for every thread started in that scope, then returns the result of `f`.
- **`Thread.start(scope, f)`** takes `f` as an escaping parameter and runs it on a new thread and returns a `Thread<T>`. `Thread.join(t)` waits for it and returns its result. A second `join` gives the same value. Each thread has an 8 MiB stack, the size of the main thread's.
- **Effects.** `Thread.start` has the effects of `f` and `alloc`. `Thread.scope` and `Thread.join` add none: joining gives the same result every run.
- **A `Scope` cannot escape.** It can be passed to a function, but it cannot be stored in a field, element, tuple or constructor, returned, or captured by an escaping closure, a thread function included (E1191). Only code running inside `Thread.scope` can start threads in it.
- **Thread functions** are an arrow function written at the call or the name of a top-level function (E1193), so the compiler sees what they capture. They capture only constant values, like every closure that can outlive its call, and each capture and the result must be `Sendable` (section 6, E1190). They cannot be `async` (E1192).
- **Tasks on a thread.** Each thread has its own event loop. A thread function may start tasks; when it returns, its thread cancels and drains what is left, as the program does at exit (Async functions below), before `join` gives the result. An `Async<T>` is not `Sendable`, so a handle never reaches another thread's loop; work moves between loops through channels and `Shared`.
- **A panic on any thread** prints its message and ends the process with code 101.
- **`Parallel.map(items, f)`** from `std/parallel` splits an array into one part per core, runs `f` on each element on those threads and returns the results in order. `f` follows the thread function rules.
- `==`, `!=` and map keys on a `Thread<T>` or a `Scope` are a compile error (E1195).

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

Not built yet (D111; node threads-channels). `std/channel` is a source module, and `Channel.create<int>(16)` is an ordinary call with a type argument (section 4).

- **`Channel.create<T>(capacity)`** returns a `#[Sender<T>, Receiver<T>]`. `T` must be `Sendable` (E1190). A capacity below 1 is taken as 1.
- **`Channel.send(sender, value)`** blocks while the channel is full and returns `true`. It returns `false` and drops the value once the channel is closed or every copy of the receiver is gone.
- **`Channel.receive(receiver)`** blocks until a value arrives and returns `Some(value)`, or `None` once the channel is closed and empty. **`Channel.values(receiver)`** is a `Generator<T> uses io` over the same values. Copies of a `Receiver` share the queue, so several threads can take values from one channel; each value goes to one of them.
- **`Channel.close(sender)`** closes the channel for every copy of the sender. Dropping the last copy also closes it.
- Blocking calls have the `io` effect. In async code they block the thread's event loop, as other blocking calls do.
- `==`, `!=` and map keys on a `Sender` or a `Receiver` are a compile error (E1195).
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
- **A `Cell` cannot cross threads.** It uses plain reference counts and is not `Sendable` (section 6): a `Cell`, or a value that holds one, captured by a thread function or sent through a channel is E1190, with a hint to use `Shared`. Not built yet (node threads-check).
- **A closure stored in the `Cell` it captures leaks** (D80); clear it with `Cells.set(cell, None)` first.

### Shared state

A few types have identity: copying one copies a handle to the same object. They are the only way two threads can see the same changing data.

| Type | Use |
| --- | --- |
| `Sender<T>`, `Receiver<T>` | Channel ends |
| `Shared<T>` | A value guarded by a lock |
| `Atomic<T>` | An integer or `bool` with atomic operations |

Not built yet (D111; node threads-shared).

<!-- context
import * as Shared from "std/shared"
-->
```ts
const hits = Shared.create(0)

Shared.update(hits, (count) => count + 1)

const now = Shared.read(hits, (count) => count)
```

- **`Shared<T>`** (`std/shared`) has the functions of `Cells` and `read`: `Shared.create(value)` (`uses alloc`), `Shared.get(shared)` returns a copy, `Shared.set(shared, value)`, `Shared.update(shared, change)` stores `change(value)`, and `Shared.read(shared, look)` returns `look(value)`. As for `Cells.update`, `change` returns the new value, because a scalar cannot be a `mutable` parameter (D81). `T` must be `Sendable` (E1190).
- **The lock is held only while a call runs.** Each function takes the lock, does its work (running the callback, if any) and releases it. There is no way to take the lock and forget to release it.
- **The callback cannot be `async`** (E1410), so a task never suspends while holding a lock.
- **`Atomic<T>`** (`std/atomic`) holds a `bool` or an integer type up to 64 bits (E1411 for any other `T`). It offers `Atomic.create`, `Atomic.load`, `Atomic.store`, `Atomic.exchange`, `Atomic.add` (integers only; it returns the old value and wraps like `+`) and `Atomic.compareExchange(atomic, expected, replacement)`, which returns `true` when it stored. Every operation is sequentially consistent.
- `==`, `!=` and map keys on a `Shared` or an `Atomic` are a compile error (E1195), as on a `Cell`; compare the values instead.
- **Identity types can form cycles** (D80, section 6 "Cycles"). A `Shared` whose value holds a handle to the same `Shared`, directly or through other identity objects, is never freed. `Async<T>` (Async functions below) and `Cell<T>` follow the same rule. The debug leak check reports the cycle; clear the slot before dropping the last outside handle, or use `std/arena`.
- **Deadlocks are possible** when callbacks nest locks in different orders. Nothing reports nested `Shared.update` calls yet; `narra audit` lists only `unsafe` blocks and `@trusted` functions (section 9).

### Async functions

An async function runs as a task on the event loop of its thread. A task suspends at an `await` that has to wait, and the loop resumes it when the awaited task has finished, a timer has fired or I/O is ready. The syntax is TypeScript's (D93, D95, D109).

```ts
import * as Async from "std/async"
import * as Bytes from "std/bytes"
import * as Net from "std/async/net"
import { Connection } from "std/async/net"

async function echo(connection: Connection): int uses io, alloc {
  let total = 0
  while (true) {
    const data = await Net.read(connection, 4096)
    if (data is Err(_)) {
      return total
    }
    if (Bytes.length(data.value) == 0) {
      return total
    }
    const written = await Net.write(connection, data.value)
    if (written is Err(_)) {
      return total
    }
    total += written.value
  }
}

export async function run(_args: Array<string>): int {
  const listener = Net.listen("127.0.0.1", 8080)
  if (listener is Err(_)) {
    return 1
  }
  while (true) {
    const accepted = await Net.accept(listener.value)
    if (accepted is Err(_)) {
      return 0
    }
    Async.detach(echo(accepted.value))
  }
}
```

Tasks and handles:

- **`async function f(...): T`** declares an async function. Its type is `(...) => Async<T>`: a call starts the task and gives an `Async<T>` handle. The body runs until it first has to wait, then the call returns. `return` in the body gives a `T`.
- **`Async<T>`** is a handle to a result on its way, like a TypeScript `Promise`. It is an identity type (section 6, "Cycles"): copies share one task, and the finished value is kept until the last handle is dropped. `Async` needs no import, and a type named `Async` that the module declares hides it. `==`, `!=` and map keys on a handle are compile errors (E1106).
- **`async` arrow functions and `async { ... }` blocks** make handles too. `async (item) => { ... }` has the type `(item: T) => Async<R>`, so `Arrays.map(items, async (item) => ...)` gives an `Array<Async<R>>`. `async { ... }` is an expression of type `Async<T>` that starts at once; it follows the rules of a block expression (section 2), so its value comes from `return`. Both capture like escaping closures wherever they are written: the task gets each captured value when it starts, and capturing a `let` or a `mutable` parameter is a compile error (E1110). Share changing state through a `Cell`. As for an arrow function, a block that returns a value sharing a captured one is E0976; return a `copy(...)` of it.
- **`await handle`** gives the value of an `Async<T>`. It waits only when the task has not finished, and awaiting the same handle again gives the same value. `await` on another type is E1102.
- **Where `await` is allowed.** Anywhere in an async body (function, arrow or block), in any expression: `if (await ready())` and `total(await first(), await second())` work, and the operands run left to right as in TypeScript. Outside an async body, in a `function*` body and in an arrow function that is not `async` it is E1100.
- **An unused `Async<T>` is a compile error** (E1103), like an unchecked `Result`: a call without `await` as a statement, an `async { ... }` alone, and `const _ = f()`. Await the handle, keep it, pass it on, or let the task run with `Async.detach(handle)`.
- **Effects are charged where the handle is made.** A call of an async function has the function's effects plus `alloc` for the task, and `await` has none. `Async<T>` carries no effects. The type of an async arrow lists the effects of its body: `(id: int) => Async<User> uses io`.
- **Parameters.** An async function or arrow takes no `mutable` parameter (E1104), because the task outlives the call, and a function-type parameter of an async function must be `escaping` (E1111). Its parameters are kept in the task, and a value passed to a task is shared with it: while a local holds the handle, a write to the passed place is a shared write (E0978, section 6), and `copy(...)` at the argument is the fix. The value an async function or arrow returns follows the `from` rules of a plain function or arrow (section 6): a returned parameter needs `from` after the result type, and `await f(x)` of a function with `from x`, called by name or through a function value, shares `x` as a plain call does. The compiler does not yet follow `from` through a stored handle (`const h = f(x)`, then `await h`), so a write to that value copies without an error, and the debug copy check reports it.
- `for (mutable item of items)` is a compile error in an async body (E1105). `async function*` declares an async generator (below).

The functions of `std/async` (`import * as Async from "std/async"`):

- **`Async.then(handle, callback)`** runs `callback` with the value when the task finishes and returns nothing. The callback is an ordinary escaping function with no `await`, and its effects are charged at the call. Callbacks run in the order they were registered; on a handle that has already finished the callback runs on the next loop turn, not inside the `then` call, as in TypeScript. The runtime drops it after it runs or when the handle is cancelled.
- **`Async.detach(handle)`** lets the task run with nobody waiting.
- **`Async.peek(handle): Option<T>`** gives the value once the task has finished, and `None` before that or after a cancel. It never waits.
- **`Async.cancel(handle)`** cancels the task (below).
- **`Async.all(handles): Async<Array<T>>`** gives every value, in order. **`Async.race(handles): Async<Option<T>>`** gives the first value to come, or `None` for an empty array. Neither cancels the handles it was given.
- **`Async<void>`.** A type parameter that is the value of an `Async<T>` is inferred as `void` from an `Async<void>` argument, and a callback parameter of type `void` takes no argument, so the `then` callback of an `Async<void>` is `() => void`: `Async.then(work(), () => print("done"))`. `detach`, `then` and `cancel` take an `Async<void>`; `peek`, `all` and `race` on one are E1150, because `Option<void>` and `Array<void>` are not types. A type parameter inferred as void anywhere else than `Async<T>` or a function-type parameter is E1150 too.
- **`await Async.idle()`** waits until the loop has nothing ready and no timer or I/O wait pending, counting tasks started while it waits. Several idle callers finish one at a time, oldest first. A task stuck in a cycle of awaits does not hold it.
- **`AsyncError`** is the error of every suspending std function: `AsyncError.Cancelled`, or `AsyncError.Failed(message)` with the operating system's message.
- **No blocking wait.** No function waits for a handle outside an async body. A function that is not `async` starts tasks and passes their handles on, detaches them, or registers a `then` callback.
- **No `Thenable` trait.** `await` works only on `Async<T>`, so it never runs user code. A type that can be awaited offers a function that returns an `Async<T>`.

The loop:

- **One event loop per thread,** with no loop value to pass around. The entry function runs as its first task (section 15). It may be `export async function run(args: Array<string>): int`, whose `Async<int>` gives the exit code.
- **The program exits when the entry function finishes.** Tasks still running then are cancelled and run until they return; a task left waiting on a task that can never finish is dropped. To let tasks finish first, end the entry with `await Async.idle()`. A detached task lives at most as long as the program.
- **A cycle of awaits stops the program.** When no task is ready, no timer is set and no I/O is waiting, but the entry has not finished, the tasks wait on each other. The program stops with the message `async: every task waits on another task` and exit code 101, as for a panic.
- **Cancel is cooperative.** After `Async.cancel(handle)`, the suspending std call that the task waits on or makes next returns `Err(AsyncError.Cancelled)` without doing its work, and so does every later one. The body handles it like any error and returns a `T`, so `await` on a cancelled handle still gives a `T`. A cancelled handle's `then` callbacks do not run. A cancel passes down the await chain: when the cancelled task waits on a handle that no other task awaits, that task is cancelled too, and so on, so the innermost timer, I/O or idle wait gives `Err(AsyncError.Cancelled)`. A handle that other tasks also await, that has `then` callbacks or that an unfinished `Async.all` or `Async.race` was given is not cancelled, a cancel of an `Async.all` or `Async.race` handle stops there, and detached tasks and tasks the cancelled task started but does not await keep running. A task cancelled while it runs passes the cancel when it next waits in an `await`. Cancelling a finished task does nothing. A cancel wakes a task waiting in its own `Async.idle()`, and `idle` in a cancelled task finishes at once.
- **Suspending I/O lives in `std/async/*`.** Each suspending function is `async` and returns `Result<T, AsyncError>`, so a body must look at the result (E0229) and sees a cancel.
  - `std/async/time`: `sleep(duration: Duration): Result<Duration, AsyncError>`, with `Duration` from `std/time`. Ok holds the time slept.
  - `std/async/file`: `readText`, `writeText`, `readBytes` and `writeBytes`, with the names and arguments of `std/fs`. Each call reads or writes in one blocking step, then lets the other ready tasks run once.
  - `std/async/net`: `listen(host, port): Result<Listener, AsyncError>` and `localPort(listener): int`, which do not suspend, and `accept(listener)`, `connect(host, port)`, `read(connection, maxBytes): Result<bytes, AsyncError>` (empty at the end of the stream) and `write(connection, data): Result<int, AsyncError>`. `host` is a numeric address, and port 0 picks a free port. One task at a time may read a connection and one may write it; a second gives `Err(AsyncError.Failed(...))`. `read` gives at most 64 KiB per call. `Listener` and `Connection` are handles: copies share the socket, and it closes when the last copy is dropped (section 6).
- **Blocking calls block the loop.** `std/fs`, `Time.sleep`, `Process.run` and the other blocking functions are allowed in async code, and no other task runs while they wait.
- **A task stays on the thread that runs its loop.** An `Async<T>` or a `Cell` cannot be sent to another thread. Pass work between the loops of several threads with channels.
- The loop waits with `kqueue` on macOS and `epoll` on Linux.

### Async generators

An async generator produces values one at a time and may wait between them (D110). The syntax is TypeScript's `async function*` and `for await`. `for await` and `Net.chunks` are not built yet (node for-await); until then a stream is read with `Async.next`, so the sample below is not checked.

```text
import * as Bytes from "std/bytes"
import * as Net from "std/async/net"
import { Connection } from "std/async/net"

async function* sizes(connection: Connection): AsyncGenerator<int> uses io, alloc {
  for await (const chunk of Net.chunks(connection, 4096)) {
    if (chunk is Err(_)) {
      return
    }
    yield Bytes.length(chunk.value)
  }
}

async function total(connection: Connection): int uses io, alloc {
  let sum = 0
  for await (const size of sizes(connection)) {
    sum += size
  }
  return sum
}
```

- **`async function* f(...): AsyncGenerator<T>`** declares an async generator. The result must be written `AsyncGenerator<T>` (E1112); a `function*` that is not async cannot declare it (E1113). `AsyncGenerator` takes one type argument (E1114), needs no import, and a type named `AsyncGenerator` that the module declares hides it. As for a generator, an exported one writes its effects on the result (`AsyncGenerator<T> uses io`), a private one has them inferred, it takes no `mutable` parameter (E1104), and its function-type parameters must be `escaping` (E1111). There is no arrow and no block form.
- **The call runs nothing.** It makes the stream and stores the arguments, as a `function*` call does, and charges `alloc`. The body runs only while a reader waits for a value, so a stream nobody reads does no work.
- **The body** is both an async body and a generator body. `await` stands in any expression, as in an async function. `yield value` is a statement, as in a `function*`, and `yield await f()` works. `return` takes no value (E0292). `const input = yield value` is E1115: there is no async coroutine. `yield* other` passes on every value of another `AsyncGenerator<T>` with the same `T`; any other operand, a `Generator<T>` included, is E1116.
- **`Async.next(stream): Async<Option<T>>`** (`std/async`) asks for the next value. The handle gives `Some(value)` at the next `yield` and `None` once the body has ended, and then for every later call. It is an ordinary handle: `await` it, race it against a `sleep` for a timeout, or keep it to read one value ahead. The call charges the effects of the stream's type and `alloc`. Requests are served in the order they were made, and the body runs until its next `yield` inline in the call when nothing makes it wait.
- **`for await (const x of stream)`** awaits `Async.next` until it gives `None`, binding each value. It is allowed in an async function, an async arrow, an `async { ... }` block and an async generator, and elsewhere it is E1122. The source must be an `AsyncGenerator<T>` (E1123): loop over a `Generator<T>` with `for ... of`, and over an `Array<Async<T>>` with `for ... of` and an `await` on each handle. `for ... of` over an `AsyncGenerator<T>` is E1124. `for await (mutable x of stream)` is E1105. Labels, `break`, `continue` and destructuring heads work as in `for ... of`, and `ensures` on the generator gives the binding the same facts. Not built yet (node for-await).
- **A stream is shared, not copied.** Copies of an `AsyncGenerator<T>` share one body and one position, as copies of an `Async<T>` share one task, so two loops over one stream each get some of its values. `==`, `!=` and map keys on one are E1106. A value passed to a stream is shared with it as with a task: while a local holds the stream, a write to the passed place is E0978.
- **Stopping early.** `break`, `return` or a labeled `break` out of a `for await` drops the loop's reference to the stream, as dropping a copy read with `Async.next` does. When it was the last one, the body is dropped where it waits at its `yield`, and what it holds there is released; a `Connection` closes when its last copy goes. Nothing else of the body runs.
- **Cancel.** A cancel of a task that waits on a value of a stream, or of the `Async.next` handle itself, passes to the stream when it is the only request waiting on it: the wait inside the body gives `Err(AsyncError.Cancelled)`, as do its later suspending calls, and later requests give `None` once the body ends. A request nobody awaits still runs the body to its next `yield`; that value is lost when the handle is dropped.
- **`Net.chunks(connection, maxBytes): AsyncGenerator<Result<bytes, AsyncError>>`** in `std/async/net` yields `Ok(chunk)` for each read that is not empty, ends at the end of the stream, and on an error yields one `Err` and ends. Not built yet (node for-await).

## 9. Unsafe code and C interop

- **`unsafe { ... }`** is a statement, and the only place a call with the `unsafe` effect is allowed: the std/ptr functions that read, write, move or free a pointer, and any function whose effects include `unsafe`. A call outside one is E0996. `unsafe` is a keyword only before `{`, so it stays an effect name. The block is not an expression; assign to a `let` declared before it. Code inside `unsafe` may crash; it is the stated exception to D92.
- **The `unsafe` effect.** A function that holds an `unsafe` block has the `unsafe` effect: an exported one writes `uses unsafe` (E0800), a private one is inferred, and every caller needs an `unsafe` block of its own. A function value whose type lists `uses unsafe` is called the same way.
- **Generators and traits.** Running a generator runs its body, so a `for` loop or `yield*` over a generator whose body has the `unsafe` effect needs an `unsafe` block (E0996). A call through a trait has the effects the trait function declares, so an implementation may use `unsafe` only when its trait function declares `uses unsafe` (E0800); otherwise it wraps the unsafe code with `@trusted`, the one attribute allowed on an implement block member.
- **`@trusted`** on a function states that its `unsafe` blocks were checked by hand. It stops the effect there: the function does not have `unsafe`, so callers need no block and an exported one does not declare it. `@trusted` on a function with no `unsafe` block, or on one that writes `uses unsafe`, is E0997.
- **`Ptr<T>`** is a raw pointer, a plain value with no reference count, compiled to the C pointer `T*`. The type is visible in a module that imports `std/ptr` (`import * as Ptr from "std/ptr"`), and it is written `Ptr<T>` whatever the alias: with `import * as P from "std/ptr"` the functions are `P.read` and the type is still `Ptr<int>`; `P.Ptr<int>` is an error. `T` is a number type, `bool`, `char`, another `Ptr`, a vector type of `std/simd` (not built yet), a `FixedArray` of such values, or a struct whose fields are all such values (tuples count as structs, and a field may be a `FixedArray`). A `FixedArray` of anything else, such as `FixedArray<string, 2>`, is E0998 with a message that names the element. Strings, `bytes`, arrays, maps, boxes, enums, function values, generators, opaque structs and type parameters are E0998, because a pointer to them would bypass reference counting.
- **Pointer values are ordinary data.** A `Ptr` can be stored in a struct, an array or a tuple, copied, passed to and returned from any function, `@trusted` or not, and held across calls. Copying one copies the address. A `Ptr` is not `Equatable`, `Hashable` or printable: `==` and `!=` on it are compile errors, it cannot be a `Map` key, and a struct that holds one cannot derive those traits. Test for null with `Ptr.isNull`.
- **std/ptr.** `Ptr.alloc(count: int): Ptr<T>` allocates `count` zeroed elements (`uses alloc`); `Ptr.free(pointer)` frees what `Ptr.alloc` returned and does nothing for a null pointer; `Ptr.read(pointer): T` and `Ptr.write(pointer, value: T)` read and write one element; `Ptr.offset(pointer, count: int): Ptr<T>` moves by `count` elements. `Ptr.cast(pointer: Ptr<T>): Ptr<U>` reinterprets the address as a pointer to another element type, as a C cast does, with no check (`const bytes: Ptr<u8> = Ptr.cast(words)`); `Ptr.element(pointer: Ptr<FixedArray<T, N>>, index: int): Ptr<T>` is the address of one item of a FixedArray. `Ptr.element` proves its index like a FixedArray read does (E0994 for a constant outside 0 to N - 1 or an unproven index), and a pointer to anything else is E0998. These seven have the `unsafe` effect. `Ptr.null(): Ptr<T>` and `Ptr.isNull(pointer): bool` touch no memory and are safe. `Ptr.alloc`, `Ptr.null` and `Ptr.cast` take the result's element type from the expected type (`const items: Ptr<int> = Ptr.alloc(4)`). There is no `Ptr<void>`, since `void` is not a type that can be written; `Ptr<u8>` stands for C's `void*`. Nothing is checked at run time: a read past the allocation, a use after `Ptr.free` or a double free is undefined behavior, and so is reading or writing through a `Ptr.cast` pointer whose element type does not match the stored data, except through `u8` and `i8`, because programs are built with strict aliasing. Memory from `Ptr.alloc` counts toward the debug leak check (section 13) until it is freed.

```ts
import * as Ptr from "std/ptr"

@trusted
function sumSquares(count: int): int uses alloc {
  let total = 0
  unsafe {
    const items: Ptr<int> = Ptr.alloc(count)
    for (let i = 0; i < count; i++) {
      Ptr.write(Ptr.offset(items, i), i * i)
    }
    for (let i = 0; i < count; i++) {
      total += Ptr.read(Ptr.offset(items, i))
    }
    Ptr.free(items)
  }
  return total
}
```

- **`extern { ... } from "library"`** declares C functions, like an import names a module. The block stands at the top level and holds one or more function signatures, each on its own line, with no type parameters, no `uses` and no body (E1000). `from` names the library as `cc` takes it after `-l` (letters, digits, `_`, `.`, `+` and `-`, not starting with `-`): `"m"` for libm, `"z"` for zlib; `"c"` is libc and needs no `-l`. A member may be written `export`, and is then called from other modules like any exported function.

```ts
import * as Float from "std/float"

extern {
  @extern("abs")
  function absolute(value: i32): i32
} from "c"

extern {
  function cbrt(value: f64): f64
} from "m"

@trusted
export function cubeRootOfMagnitude(value: i32): f64 {
  let result = 0.0
  unsafe {
    result = cbrt(Float.from(absolute(value)))
  }
  return result
}
```

- **Every member has the `unsafe` effect**, so all the rules above apply: a call needs an `unsafe` block (E0996) or a `@trusted` wrapper, a private function that calls one inside a block gets the effect, a generator whose body calls one needs a block where it runs, an implementation may call one only when its trait function declares `uses unsafe`, and a function value of a member has the type `(...) => T uses unsafe`. A member's parameters are never "unused".
- **Types.** Parameters and results are the number types, `bool` and `Ptr<T>`, and a member may return nothing (E1002). In C, `int` is `int64_t`, the sized integers are `int8_t` to `uint64_t`, `f64` is `double`, `f32` is `float`, `bool` is `bool` and `Ptr<T>` is `T*`. So C's `long` is `int` on both supported platforms, C's `int` is `i32`, `size_t` is `u64`, and `char*`, `const char*` and `void*` are `Ptr<u8>`. Nothing checks that the Narra signature matches the C function: a wrong one is undefined behavior, which is why a call is unsafe. A `mutable` parameter is E1002. A C function that returns `int` as a truth value (`isalpha` and the other `ctype.h` predicates) is declared with `i32`, not `bool`: a `bool` result reads only the low byte, and glibc returns values such as 1024.
- **Variadic C functions cannot be declared.** There is no syntax for `...`, and a fixed prototype for one (`printf(format: Ptr<u8>, value: int): i32`) is undefined behavior: variadic arguments are passed differently from fixed ones on some platforms (macOS arm64 puts them on the stack), so the C function reads garbage. Call them through a C-side wrapper with a fixed signature, which waits for C-callable functions and C sources in a package (node c-abi-export).
- **C names.** A member links to the C function with its own name, or with the name `@extern("name")` gives (the only attribute a member takes, written on the line before it). The C name must be a C identifier and must not be one the compiler writes (E1001): `main`, a name starting with `narra_` or `lstr_`, a string literal name (`L` and digits), a name with a module hash (`f_3fa9c1e7_show`), or the generated name of a function or type of the program, alone, followed by `_` and more, or after a word and `_` (the reference counting and trait helpers, such as `release_E_Option_e088c655` and `eq_S_Tuple2_…`). `@extern` on a function with a body is E1000 for now; C-callable Narra functions are step 2 below.
- **How a member's C prototype is written.** The C code keeps calling the member by its generated, hashed name, and declares it once as `extern <result> <generated name>(<parameters>) __asm__("<symbol>")`, with the C types above and the C name as the symbol (with the platform's symbol prefix, `_` on macOS). Because the C identifier is not the C name, the prototype never meets a declaration of the same function in the headers the runtime includes (`string.h`, `stdlib.h`, `math.h` and others): declaring `strlen` with `Ptr<u8>` and `u64` while `string.h` declares `size_t strlen(const char*)`, or `sqrt` twice, compiles with no conflict, and both declarations name the same symbol. Only the prototypes of reached members are written.
- **Linking.** `narrac build` and `narrac run` pass `-l<library>` to `cc` for every library whose block has a member that reached code calls (tree shaking, section 11), after the C file and in declaration order; `-lm` is always passed. A block whose members are never reached, for example because they are called only in a branch a build condition drops, links nothing, so its library need not exist on that target. `[link] search` in `narra.toml` adds `-L<directory>` (section 15).
- **Memory from C.** Memory a C function allocates goes back through the C library's own function (an extern `free`), never through `Ptr.free`, which frees only what `Ptr.alloc` returned.
- **Not built yet** (C interop step 2, node c-abi-export): `@layout(C)`, `@layout(u8)` and C-callable functions, written below under their earlier names.
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
- **C strings.** `std/cstring` provides `CString`, a string with no zero byte, written with `import * as CStrings from "std/cstring"` and `import { CString } from "std/cstring"`. `CStrings.from(text): Option<CString>` is `None` when the text holds a zero byte. `CStrings.pointer(value): Ptr<u8>` is the address of its zero-terminated bytes for a C function that takes `const char*`; it is valid while the CString is alive and must not be written through. `CStrings.toString(value): string` gives the text back, and `CStrings.read(pointer): Result<string, string>` (`uses alloc, unsafe`) copies a zero-terminated string a C function returned, with `Err` for invalid UTF-8.
- `narra audit [path]` lists every `unsafe` block and `@trusted` function in a package, one line each, sorted by file, line and column: `src/x.nra:12:3: unsafe block in function name` and `src/x.nra:4:1: @trusted function name`. A block in a test reads `in test name`, and a trait or implement block member is named with its trait or implementation: `@trusted function measure in implement Measure for Cells`, `unsafe block in function measure in trait Measure`. It reads the source only and exits 0; a package that does not load is an error.

## 10. Top-level declarations

- **`const`** at the top level declares a compile-time constant. It cannot be changed at any depth or passed as `mutable` (D81); a local `const` fixes only its name (section 2). Inside a function, `const` accepts any expression. A top-level value must be one of these:
  - a literal or a constructor
  - a struct or array of constants
  - a named function or another constant
  - a call to an `@const` function, with constant arguments, also inside the constant parts of a value (`const pair = #[square(2), square(3)]`)
- **`@const` functions** are computed by the compiler when a top-level constant calls them (D105). The call is written at the constant, so nothing runs hidden; outside a constant an `@const` function is an ordinary function.
  - Only a function written with `@const` may be called from a constant. Its declaration is checked where it is written, even when no constant calls it, so an edit that breaks the rules fails in the module that made it.
  - Its effects are at most `alloc` (E1051); an exported one declares at most `uses alloc`.
  - Its parameters and result hold only numbers, `bool`, `char`, `string`, `bytes`, `Array`, `Map`, `FixedArray`, `Option`, `Result`, tuples, and structs, enums and distinct types made of these. A function type, `Cell`, `Ptr`, `Shared`, a generator, a trait-constrained type parameter or an effect parameter is E1050, and so is a generator function.
  - Its body may call other `@const` functions, constructors and the functions of `std/string`, `std/array`, `std/map`, `std/int`, `std/float` and `std/range` that take no function, read top-level constants, and use loops, `match`, `is`, patterns, `mutable` parameters and locals. It may not call another function (E1052, which names the function), use an arrow function or call through a function value, or run a user `implement` block, through `==`, a map key or a call. The built-in and derived implementations run: `==` and map keys on numbers, strings and derived types. `std/char`, `std/option`, `std/result`, `std/math`, `std/bytes`, `std/fixedarray` and the sized number modules are not supported yet; `Math.sin`, `cos`, `tan`, `atan2`, `exp`, `log` and `pow` never will be, because their results depend on the C library.
  - Limits, per constant: 10 million evaluation steps (E1054), 64 MiB of values (E1055), 1000 nested calls (E1056), a nesting depth of 5000 (E1056) and a value depth of 1000 (E1056). A step is one expression, statement or loop turn, plus one per 8 bytes or elements a std string, array or map call goes through. Values are counted where they are built: 8 bytes per number, element, field and entry, string lengths rounded up to 8, and 16 per array and map. `copy(...)`, and storing a value read from a place (`const row = rows[0]`, `Arrays.push(mutable rows, row)`, a field or element of a literal), count the outer level of the value again, as the copy its first write makes. A std call that would pass the memory limit (`Arrays.filled`, `Arrays.withCapacity`, `Str.repeat`, `Str.join`, `Str.replace`, string concatenation and templates) is stopped before it runs. A self-call in tail position runs as a loop and does not count toward the calls. The nesting depth counts the calls, blocks, loops and expressions under way in all open calls at once, so a body that nests deeply reaches it before 1000 calls (six nested additions per call allow about 600 calls); it keeps the compiler's own C stack safe. The value depth is how deeply arrays, structs, enums and maps nest in the result.
  - A float literal in an `@const` body is computed only when the compiler reads it exactly: at most 18 significant digits, and a power of ten that reduces to at most 22. An `f32` literal whose value read as `f64` is rounded and lies halfway between two `f32` values is refused too, since rounding it again to `f32` could differ from C's reading. Both are E1052 at the constant.
  - A panic, a failed `assert`, an index out of range and the std functions that would panic (`Str.sliceFrom`, `Str.charAt`, `Str.nextChar`, `Str.byteAt` on a bad position) stop the computation with a compile error at the constant (E1053). Each error of a computation shows the chain of calls from the constant down as notes.
  - A constant that an `@const` function reads while the constant itself is computed is E1057.
  - A scalar result (a number, `bool`, `char`, a constructor without fields) is written at each use, as for any scalar constant. Any other result is built on first use by the constant's getter, from the value; an array of scalars is copied from static C data in one step.
- **`lazy`** defers a computation until first use: `const table = lazy buildTable()`, with `import * as Lazies from "std/lazy"`. The first `Lazies.force(table)` computes the value and later calls return it, borrowed like any top-level constant. Unlike a plain constant, the expression may call functions. The rules:
  - `lazy` is written only on the value of a top-level `const` (E1032). It is a word with this meaning only there; elsewhere `lazy` is an ordinary name.
  - The name has the type of the expression, and it appears only as the argument of `Lazies.force` (E1033). It cannot be passed, stored, compared, used in a template or used in the value of another constant. A lazy value of another module is forced as `Lazies.force(Alias.name)`.
  - The expression may have the `alloc` effect and no other (E1034), because the first force can happen anywhere. `Lazies.force(table)` has the effects of the expression, so a caller of it lists `alloc` when the expression allocates.
  - A lazy value that forces itself, directly or through the functions it calls and the lazy values it forces, is an error (E1035). A function named as a value counts as called.
  - A lazy value that no code forces is not built and has no code in the program. A panic in the expression happens at the first force.
  - Forcing is not thread safe yet: the ready flag is the once guard of one thread, as for other constants that are not scalars (D61). Threads must add the lock.
- Modules have no initialization code. Program startup runs nothing before the entry function.

## 11. Code generation

- **Output is C.** One `.c` file per module, plus a small runtime: allocator, reference counting, panic handler and arenas.
- **Vector types** from `std/simd` compile to C compiler vector extensions (D112; not built yet). Each type the program uses is a typedef such as `typedef float narra_f32x4 __attribute__((vector_size(16), aligned(8)))`. The alignment is 8 because heap blocks start at 8-byte boundaries; loads are unaligned, which costs nothing on x86-64 and arm64. Integer `+`, `-` and `*` go through the unsigned vector type, so they wrap. The build passes no `-march` flag: x86-64 code uses SSE2 and arm64 code NEON. A C compiler without vector extensions (TCC) cannot build a program that imports `std/simd`.
- **Debug builds** use a fast C compiler such as TCC with checks on.
- **Release builds** use Clang or GCC with `-O2`.
- **No fused multiply-add.** The C starts with `#pragma STDC FP_CONTRACT OFF`, so `a * b + c` on floats rounds twice at every optimization level, as the compiler computes it for an `@const` function (D105). Without it, clang contracts the expression on arm64, which changes the last bit of some results.
- **Generators and async functions** compile to a struct and a resume function. The struct holds their live variables and a state number. The resume function switches on that number.
  - Calling a `function*` allocates the struct with a reference count of 1 and state 0, and stores its parameters. It runs nothing of the body.
  - Every local, loop counter, temporary and expression result of the body is a field of the struct, so no C local is alive across a `yield`. Only scratch values used inside one expression stay C locals.
  - `yield v` stores `v` for the caller, records the next state and returns. Ending the body or `return` releases what is still held and sets the finished state; resuming a finished generator returns nothing again.
  - Dropping an unfinished generator releases what is live at its state and frees the struct. Resuming a shared generator first copies the struct and retains what is live in it, so `for ... of` and `yield*` work on their own copy.
  - Calling an async function allocates the struct with a task header and a field for the result, stores its parameters (an async arrow or block stores its captures) and runs the body until it first waits. The struct is the task and the handle: `Async<T>` points to it, and copies of the handle share it, so it is never copied.
  - `await` on a finished handle reads the result field. Otherwise the task joins the handle's waiters, records the next state and returns to the loop, which resumes it once the handle has finished.
  - A finished task has released its locals. The struct and the result are freed with the last handle. At exit the tasks still unfinished are dropped like an unfinished generator, so the leak check sees nothing of them.
- **`mutable` parameters** compile to C pointer parameters.
- **Self-calls in tail position** compile to loops in every function (section 2). Other recursion uses the C stack. Stack overflow aborts the program.
- **Closures.** A non-escaping closure's environment is a C struct on the caller's stack. An escaping closure's environment is a reference-counted heap object. The bodies of `@inline` functions, and of arrow functions passed to them, are written into each call (section 2).
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
- Compile-time evaluation is limited to top-level `const` declarations, `@const` functions and build conditions. A constant that calls an `@const` function is computed inside the compiler, by an evaluator over the checked program, within the limits of section 10; `narrac check`, `build`, `emit` and the language service all compute it, since a failed computation is a compile error. A result depends only on the source and the target: maps keep insertion order, integers wrap the same way, and floats round the same in the compiler and the program (section 11).

### Build conditions

Code that differs by platform branches on constants from `std/build`. The compiler keeps only the matching branch.

```ts
import * as Build from "std/build"
import { Os } from "std/build"
import * as Env from "std/env"
import * as Options from "std/option"

export function configDir(): string uses io, alloc {
  const home = Options.withDefault(Env.get("HOME"), "")
  if (Build.os == Os.Windows) {
    return Options.withDefault(Env.get("APPDATA"), home)
  } else if (Build.os == Os.MacOS) {
    return `${home}/Library/Application Support`
  } else {
    return `${home}/.config`
  }
}
```

- **`Build.os`, `Build.arch` and `Build.debug`** are compile-time constants. `std/build` also exports their types, the enums `Os` (`MacOS`, `Linux`, `Windows`) and `Arch` (`Arm64`, `X86_64`). `Build.debug` is `false` in a `-release` build and `true` otherwise.
- **The host is the target.** The compiler takes `Build.os` and `Build.arch` from the target the C compiler built the compiler itself for (its `Build.hostTarget()`), and only when the package imports `std/build`. The supported targets are `MacOS-Arm64`, `MacOS-X86_64`, `Linux-X86_64` and `Linux-Arm64`. The environment variable `NARRA_HOST_TARGET`, when set, names the target as `<os>-<arch>` in place of the host; it is the only input besides the host that sets the target. A host, or a `NARRA_HOST_TARGET`, with no supported target is a compile error at the module's import of `std/build`.
- **`Build.hostTarget()`** returns `Option<Target>`, where `Target` is a struct of `os: Os` and `arch: Arch`: the target the C compiler builds the program for, read from its predefined macros (`__APPLE__`, `__linux__`, `__aarch64__`, `__x86_64__`) when the program runs. It is `None` for any other combination. It does not run `uname`.
- **`narra emit` writes the C of a debug build.** The C of a release build comes from `narra build -release -emit-c file.c`.
- **An `if` that tests only these constants is resolved at compile time.** The output contains no runtime check, only the matching branch, or nothing when no branch matches. The condition may combine the constants with `==`, `!=`, `is` (a variant without fields), `!`, `&&` and `||`, and each `else if` of a chain is resolved on its own. A condition that mixes the constants with any other value, or a `match` on them, is an ordinary runtime test.
- **Every branch is type-checked on every build.** A Windows branch is checked when building for Linux, so branches for other platforms cannot break unnoticed. The effects of every branch count toward the function's effects, the dropped ones included, so a function has the same signature on every target.
- **`@target(os = Os.MacOS)`** on its own line before a top-level function makes the function exist only on that target. It takes `os =` a variant of `Os`, `arch =` a variant of `Arch`, or both (`@target(os = Os.Linux, arch = Arch.Arm64)`), each once; a part it does not name matches every value. The enum is written the way the module imports it: `Os.MacOS` after `import { Os } from "std/build"`, `Build.Os.MacOS` through `import * as Build`. Narra functions and `extern { } from` blocks take `@target`; other declarations do not.

```ts
import * as Build from "std/build"
import { Os } from "std/build"

@target(os = Os.MacOS)
function macOpener(): string {
  return "open"
}

@target(os = Os.Linux)
function linuxOpener(): string {
  return "xdg-open"
}

export function opener(): string {
  if (Build.os == Os.MacOS) {
    return macOpener()
  } else if (Build.os == Os.Linux) {
    return linuxOpener()
  }
  return "start"
}
```

- **A use outside a branch that guarantees the target is a compile error.** On each target, a call to a `@target` function, or its use as a function value, is an error where the function does not exist, unless that target does not compile the use. A target does not compile:
  - a branch that a build condition (above) drops on it, such as the then branch of `if (Build.os == Os.Windows)` on Linux, or the rest of an `else if` chain after the branch it keeps;
  - the body of a `@target` function whose target it is not.

  So a use is accepted on every target exactly when it sits inside a branch whose build condition holds only on targets where the function exists (`macOpener()` inside `if (Build.os == Os.MacOS)`), or inside a `@target` function with the same or a narrower target. An ordinary `if`, a `match` on `Build.os`, an early `return` and a condition on `Build.debug` alone guarantee nothing.
- **A `@target` function is type-checked on every target**, like a dropped branch, and its effects count the same way.
- **`narra check` checks a debug build for the host; `narra check --all-targets` checks every supported target**, each as a debug and as a release build, since a condition on `Build.debug` can keep a use in one of them only. It reports each diagnostic once, in source order; an E0991 lists every target it fails on, with `(debug)` or `(release)` after a target that fails in one build only. A use that is wrong only on another target shows only with `--all-targets`.
- **`@target` before an `extern { } from` block** declares its C functions, and links its library, only when building for its target. It applies to every member of the block, since a block names one library, and each member follows the rules of a `@target` function above. Only the block takes the attribute; a member does not (E1000).

```ts
import { Os } from "std/build"

@target(os = Os.Windows)
extern {
  function GetTickCount64(): u64
} from "kernel32"
```

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
  if (b == 0) {
    panic("division by zero")
  }
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

- **Contracts.** `requires` lines between a function's signature and its body state what each caller must prove (D95). They are proven at compile time; nothing is checked at run time.

```ts
export function pick(index: int, count: int): int
  requires index >= 0 && index < 64
  requires count != 0
{
  return index % count // count != 0 is known here
}

function pickChecked(n: int): int {
  if (n >= 0 && n < 64) {
    return pick(n, 8)     // a name the `if` checked proves the condition
  }
  return pick(3, 8)       // so do constants
}
```

- **A condition is clauses joined by `&&`** (D102). A clause is one comparison, or comparisons joined by `||`; parentheses group a clause: `requires (index >= 0 || index == -1) && count != 0`. A comparison is `==`, `!=`, `<`, `<=`, `>` or `>=` between an integer parameter and an integer constant, with the parameter on either side, or between two integer parameters of the same type (a relation: `requires low <= high`). Constants are what division proofs accept (section 6): literals and top-level `const` values built from them. Any other condition is E0984: a float or other non-integer parameter, a field path, a call, `!` (the message gives the condition without it: `!(x < 0 || x > 9)` is `x >= 0 && x <= 9`), arithmetic around a parameter in a relation (`low + 1 <= high`, since `+` wraps), a parameter compared with itself, and `&&` inside `||` (split it into clauses). Floats and field paths are a later extension.
- **Each call proves every clause:** one of its comparisons must be proven. A comparison with a constant is proven from the argument for that parameter: an integer constant, or a name narrowed by a check (`if`, `while`, an early return), as for a divisor. A name with no check proves nothing, not even a bound at the end of the int range. A relation is proven from two constants (a u64 constant from 2^63 up is above every other), from intervals whose bounded ends settle the order (`x <= y` when the high end of `x` is at most the low end of `y`), from the same name passed twice, or from a relation the caller holds between the two names, from its own checks (`if (low <= high)`) or its own `requires`. A caller that checked `if (x < 0 || x > 10)` knows only the hull of the sides (below), which proves neither side of `requires x < 0 || x > 10`. The caller's own `requires` count, so `requires count > 0` lets the body pass `count` on to another function with `requires count != 0`. An unproven clause is E0985 at the argument of its first parameter, naming the whole clause.
- **In the body each clause is a known fact** about its parameters, until a parameter is assigned: a body with `requires divisor != 0` may divide by `divisor`. A clause over one parameter gives it the hull of its sides, the smallest interval that holds each: `requires x == 1 || x == 3` gives `1 <= x <= 3`, and `requires d < 0 || d > 0` gives a non-zero `d`. On a u64 a side with a bound from 2^63 up proves nothing, so the hull proves nothing either. A clause over two parameters gives nothing. A relation is a fact too: it proves calls with the same or a weaker relation (`<` proves `<=` and `!=`), and it passes bounds one step (section 4), so `requires low >= 0 && low <= high && high < 64` makes `low` a valid shift count. It proves nothing about `high - low`, which can wrap.
- **A function with `requires` cannot be used as a value** (E0986), and `typeof` of it is E0982: a function type carries no condition, so a call through the value could pass an unproven argument. Functions of an `implement` block cannot have `requires` (E0984), since a call through the trait proves nothing.
- **`ensures` lines** follow the `requires` lines. Each is a pattern every result must match, with no named subject. Every `return` value is proven against each line; in a generator, every `yield` value, and `yield*` is E0988. A pattern holds one bound per place, with no `&&`: write two lines, `ensures Some(>= 0) | None` and `ensures Some(< 64) | None`, and a caller learns both.

```ts
import * as Arrays from "std/array"

function findIndex(items: Array<int>, wanted: int): Option<int>
  ensures Some(>= 0) | None
{
  let index = 0
  while (index < Arrays.length(items)) {
    if (items[index] == wanted && index >= 0) {
      return Some(index)   // index >= 0 is known here
    }
    index++
  }
  return None
}

function slot(index: int): int
  requires index >= 0
{
  return index * 2
}

function firstSlot(items: Array<int>): int {
  if (findIndex(items, 7) is Some(index)) {
    return slot(index)       // the caller knows index >= 0
  }
  return 0
}
```

- **The patterns are the patterns of `match`** (section 5), alternatives included, plus comparison patterns `< c`, `<= c`, `> c`, `>= c` and `!= c`, which are allowed only in `ensures`: `ensures >= 0`, `ensures Ok(> 0) | Err(_)`, `ensures #[>= 0, _]`, `ensures { width: > 0 }`, `ensures "GET" | "POST"`, `ensures != 0`. A comparison needs an integer place and a bound that is an integer constant fitting its type (E0987). The prover orders integers as `int`, so a u64 bound or literal pattern from 2^63 up is E0987 too, and so is a literal pattern `9223372036854775807` or `-9223372036854775807`: the ends of the int range stand for "no fact". An `ensures` pattern binds no names, and a function that returns nothing has no `ensures` (E0987).
- **A value is proven** when it is written to match: a constructor, tuple or struct literal written in place whose parts are proven (a struct field must be written, not taken from a spread or a default), a literal written the same, an integer constant, or an integer name narrowed by a check. Anything else is E0988, including a call: `return g(x)` is not proven from g's own `ensures`; bind it first.
- **A caller learns the pattern** as facts about integer names, the same facts a check gives: `const n = f()` (and `let n = f()` until it is assigned), the names bound by `f() is P`, by `match (f())` arms and by a destructuring of `f()`, the same for a `const r = f()` whose type has no contents that change in place, and the name of `for (const n of g())` for a generator `g`. A name bound by `Some(n)` learns what every alternative that can be `Some` says about its place. `!= 0` makes the name non-zero, so it can be a divisor; `!= c` for another `c` gives no fact.

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
search = ["/opt/homebrew/lib"]

[build]
generics = "compact"    # optional: layout tables everywhere, for smaller binaries
```

- **Entry points are set in `narra.toml`.** Each `[[binary]]` names the binary and its entry function as `"module.function"`. A package with no `[[binary]]` is a library.
- **The entry function** has the signature `export function run(args: Array<string>): int`, or `export async function run(args: Array<string>): int`. The result is the process exit code; for an async entry, the value of its `Async<int>`. It has every effect. Tasks still running when it finishes are cancelled (section 8).
- **`[link]`** has one key, `search`: directories `cc` searches for libraries (`-L`), relative to the package directory unless absolute. The libraries come from the `from` of extern blocks (section 9); any other key in `[link]` is E1000.
- **Version selection is minimal version selection,** as in Go. Each dependency resolves to the highest version that some package in the graph requires as its minimum. The result is the same on every machine and needs no solver.
- **One version per package** in a build. A major version change is a new package name, such as `json2`.
- **`narra.lock`** records the commit hash and content hash of every dependency. `narra build` fails when a fetched dependency does not match its hash.

## 16. Deferred

These are left out of the first version on purpose.

- **Thread cancellation.** A thread runs until it returns. Async tasks have a cooperative cancel (`Async.cancel`, section 8), and a timeout is `Async.race` with a `sleep`; a cancel design for threads is not decided.
- **A cycle collector for identity objects.** Cycles through identity types leak and the leak check reports them (section 6, D80).
- **Functors and module parameters.** See `DECISIONS.md`, D6.
- **A native debug backend.** See `DECISIONS.md`, D10.
- **A bytecode VM.** Possible later for scripting and hot reload. The first target is native code only.
