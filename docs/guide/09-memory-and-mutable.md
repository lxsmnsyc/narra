# 9. Memory and `mutable`

Memory is managed by reference counting, and the compiler inserts the counts. You never free memory by hand. What you do see in the source is where a value can change, because change is marked by three words: `let`, `mutable` and the call-site `mutable`.

## Values, not references

Every value behaves like a value. Assigning or passing a struct, an array or a map gives the receiver its own value, and changing one never changes the other:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

struct Player {
  name: string
  hp: int
}

export function run(args: Array<string>): int {
  let original = Player { name: "Ana", hp: 100 }
  let copy = original
  copy.hp = 40
  print(Int.toString(original.hp))
  print(Int.toString(copy.hp))

  let scores = [1, 2, 3]
  const snapshot = scores
  Arrays.push(mutable scores, 4)
  print(Int.toString(Arrays.length(snapshot)))
  print(Int.toString(Arrays.length(scores)))
  return 0
}
```

```text
100
40
3
4
```

Copying is cheap because of the reference count. Heap values (arrays, maps, strings, closures, enums with fields, `Box<T>`) carry a count. Numbers, `bool`, `char` and structs live inline. `let copy = original` shares the heap parts and bumps their counts. A write checks the count of the value it is about to change:

- count 1 (one owner): the write happens in place, with no copy.
- count above 1 (shared): the value is copied first, then written.

So the program above copies `scores` once, on the `push`, because `snapshot` still holds the old array. A loop that keeps a single owner writes in place all the way.

### Seeing the copies

A copy on a shared value is correct but costs time in proportion to the value's size, and the source does not show it. Tools that report such copies are planned but not available yet.

## `const` and `let`

- `const` is a constant. It can be read, and never changed, including its fields and elements.
- `let` is a variable. Its contents can be changed in place: `value.field = x`, `items[i] = x`, `Arrays.push(mutable items, x)`.

Both are value bindings; `let` is not a reference cell. Changing a field of a `let` variable changes that variable and nothing else.

```narra error
// error: constant "ITEMS" cannot be changed; copy it into a `let` variable first
import * as Arrays from "std/array"

const ITEMS = [1, 2, 3]

export function run(args: Array<string>): int {
  Arrays.push(mutable ITEMS, 4)
  return 0
}
```

## `mutable` parameters

A function can change a caller's variable only through a `mutable` parameter. The keyword appears in the declaration and again at the call, so the call site shows which variable can change.

```narra
import * as Int from "std/int"

struct Vec2 {
  x: int
  y: int
}

struct Player {
  name: string
  hp: int
  position: Vec2
}

function damage(mutable player: Player, amount: int) {
  player.hp = player.hp - amount
}

function moveBy(mutable position: Vec2, dx: int, dy: int) {
  position.x = position.x + dx
  position.y = position.y + dy
}

export function run(args: Array<string>): int {
  let hero = Player { name: "Ana", hp: 100, position: Vec2 { x: 0, y: 0 } }
  damage(mutable hero, 10)
  moveBy(mutable hero.position, 3, 4)
  print(Int.toString(hero.hp))
  print(`${Int.toString(hero.position.x)},${Int.toString(hero.position.y)}`)
  return 0
}
```

```text
90
3,4
```

`mutable` passes a pointer to the caller's variable. No copy is made, and the reference count of the variable is untouched. A parameter without `mutable` is a constant copy: to give new data back, return a new value instead.

| Goal | Form |
| --- | --- |
| Change the caller's data | `damage(mutable hero, 10)` |
| Build new data | `const stronger = withBonus(hero, 5)` |

### What a `mutable` parameter allows

Only a function that declares `mutable` can change its argument, and the call says so with `mutable` too. The function sees the caller's variable, not a copy, for the length of the call, under these rules.

**The argument must be changeable.** A `const`, a literal or a function result is not (see the `ITEMS` error above).

**Access is exclusive during the call.** The same variable cannot be passed as `mutable` twice, or passed as `mutable` and read in another argument:

```narra error
// error: so no other argument can use it
function swap(mutable first: int, mutable second: int) {
  const saved = first
  first = second
  second = saved
}

export function run(args: Array<string>): int {
  let value = 1
  swap(mutable value, mutable value)
  return 0
}
```

**A variable passed as `mutable` is used nowhere else in the same expression.** Order of evaluation would otherwise decide the result, so the compiler refuses it. Bind the call first, then use it:

```narra error
// error: "value" is passed as mutable in one call, so the rest of the expression cannot use it
function bump(mutable number: int): int {
  number = 10
  return 0
}

export function run(args: Array<string>): int {
  let value = 1
  const total = value + bump(mutable value)
  return total
}
```

The fixed version:

```narra
import * as Int from "std/int"

function bump(mutable number: int): int {
  number = 10
  return 0
}

export function run(args: Array<string>): int {
  let value = 1
  const bumped = bump(mutable value)
  const total = value + bumped
  print(Int.toString(total))
  return 0
}
```

```text
10
```

**A `mutable` parameter cannot escape.** It cannot be returned, stored in a struct or array, or captured by a closure. Copying its value into a `const` is fine, and that copy is yours to keep:

```narra error
// error: mutable parameter "player" cannot be returned or stored
struct Player {
  hp: int
}

function keep(mutable player: Player): Player {
  return player
}

export function run(args: Array<string>): int {
  return 0
}
```

```narra
import * as Int from "std/int"

struct Player {
  hp: int
}

function snapshotHp(mutable player: Player): Player {
  const copy = player
  player.hp = 0
  return copy
}

export function run(args: Array<string>): int {
  let hero = Player { hp: 7 }
  const saved = snapshotHp(mutable hero)
  print(`${Int.toString(saved.hp)} ${Int.toString(hero.hp)}`)
  return 0
}
```

```text
7 0
```

## Why a closure cannot capture `let`

A closure can capture a `const`. It cannot capture a `let` variable or a `mutable` parameter. A `let` is a value owned by one place, which is what makes in-place writes safe. If a closure could hold on to it, there would be two owners of a thing that is written in place, and the closure could outlive the variable.

```narra error
// error: a closure cannot capture "count" because it is declared with let
function apply(f: () => int): int {
  return f()
}

export function run(args: Array<string>): int {
  let count = 1
  return apply(() => count)
}
```

The same holds for a `mutable` parameter, which is even more restricted, since it points into the caller's variable:

```narra error
// error: a closure cannot capture mutable parameter "number"
function bump(mutable number: int): () => int {
  return () => number
}

export function run(args: Array<string>): int {
  let value = 1
  return bump(mutable value)()
}
```

The way out is to capture a copy. Bind the current value to a `const`, and capture that:

```narra
import * as Int from "std/int"

function apply(f: () => int): int {
  return f()
}

export function run(args: Array<string>): int {
  let count = 1
  count = count + 1
  const current = count
  print(Int.toString(apply(() => current * 10)))
  return 0
}
```

```text
20
```

The closure sees the value of `count` at the moment `current` was bound, not later changes. If you need several functions to share a changing value, pass it to them as a `mutable` parameter, or return the new value and thread it through. These are the rules of the current compiler.

## No cycles, no collector

Reference counting cannot free a cycle. Narra avoids the problem by construction. Immutable values cannot point back at themselves, and a `let` or `mutable` parameter can neither be captured nor stored, so a cycle cannot be built in safe code. Graphs that need cycles use handles into an arena (`std/arena`, planned).

## What you do not do

- You never free memory. A value is freed when its last owner drops it, and resources such as files close at that point too.
- You do not choose between copying and sharing. The compiler shares and the count decides when a write must copy.
- Running out of memory aborts the program, the same as a panic.

Next: [Effects and testing](10-effects-and-testing.md).
