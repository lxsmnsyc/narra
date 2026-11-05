# 9. Memory and `mutable`

You never free memory in Narra, and you never decide whether two names share one object. What you do decide is where a value may change. This chapter follows a player in a small game. It shows what happens when the player is copied, changed by a function, and captured by a closure.

## Assigning makes a copy

Make a player, assign it to a second name, and change the second one:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

struct Player {
  name: string
  hp: int
}

export function run(args: Array<string>): int {
  const hero = Player { name: "Ana", hp: 100 }
  let clone = hero
  clone.hp = 40
  print(`${hero.name} ${Int.toString(hero.hp)}`)
  print(`${clone.name} ${Int.toString(clone.hp)}`)

  let scores = [10, 20]
  const before = scores
  Arrays.push(mutable scores, 30)
  print(Int.toString(Arrays.length(before)))
  print(Int.toString(Arrays.length(scores)))
  return 0
}
```

```text
Ana 100
Ana 40
2
3
```

`clone` is a player of its own. Changing its `hp` leaves `hero` alone. The same holds for the array: `before` keeps two scores after `scores` grows to three. Every value works this way, whether it is a number, a struct, an array or a map. Assigning it or passing it to a function gives the receiver its own value.

So when you read a function, you know that a value you pass in comes back unchanged. The one exception is a `mutable` argument, which the call marks (see below).

## Why copies are cheap

Copying a large array at every assignment would be slow, so Narra copies only when it has to. Arrays, maps, strings, closures, enum variants with fields and `Box` values live on the heap, and each one counts how many names hold it. This is reference counting. Numbers, `bool`, `char` and structs live directly in the variable, and a struct's heap fields are counted one by one.

- `const before = scores` copies nothing. Both names point at the same array, and its count goes up to 2.
- `Arrays.push(mutable scores, 30)` sees the count of 2. Writing in place would change `before` too, so it copies the array first and writes to the copy.
- When a value's count is 1, a write changes it in place with no copy.
- When the count drops to 0, the value is freed.

That is why a loop that builds an array touches the same memory each time:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"
import * as Ranges from "std/range"

export function run(args: Array<string>): int {
  let squares = Arrays.filled(0, 1000)
  for (const i of Ranges.upTo(1000)) {
    squares[i] = i * i
  }
  print(Int.toString(squares[999]))
  return 0
}
```

```text
998001
```

`squares` has one owner the whole time, so each of the 1000 writes is in place. Had another name held the array during the loop, the first write would have copied it once. The other writes would then go to that copy. The result is the same either way. To keep large updates fast, avoid holding a second name for a value while you change it.

## Changing a player in a function

A trap hits a player for some damage. The first try changes the parameter:

```narra error
// error: 8:3: "player" is constant; declare it with `let` to change it
struct Player {
  name: string
  hp: int
}

function damage(player: Player, amount: int) {
  player.hp = player.hp - amount
}

export function run(args: Array<string>): int {
  let hero = Player { name: "Ana", hp: 100 }
  damage(hero, 10)
  return 0
}
```

A parameter is a constant: the function gets its own copy and cannot change it. The hint about `let` in the message applies to local names, not to parameters. For a parameter, there are two ways to get the change back to the caller.

The first is to return a new player. `{ ...player, hp: ... }` copies `player` with one field changed (chapter 4):

```narra
import * as Int from "std/int"

struct Player {
  name: string
  hp: int
}

function damaged(player: Player, amount: int): Player {
  return { ...player, hp: player.hp - amount }
}

export function run(args: Array<string>): int {
  const hero = Player { name: "Ana", hp: 100 }
  const hurt = damaged(hero, 10)
  print(`${Int.toString(hero.hp)} -> ${Int.toString(hurt.hp)}`)
  return 0
}
```

```text
100 -> 90
```

The second is a `mutable` parameter. The function then works on the caller's variable itself, and the call writes `mutable` too:

```narra
import * as Int from "std/int"

struct Position {
  x: int
  y: int
}

struct Player {
  name: string
  hp: int
  position: Position
}

function damage(mutable player: Player, amount: int) {
  player.hp = player.hp - amount
}

function moveBy(mutable position: Position, dx: int, dy: int) {
  position.x = position.x + dx
  position.y = position.y + dy
}

export function run(args: Array<string>): int {
  let hero = Player { name: "Ana", hp: 100, position: Position { x: 0, y: 0 } }
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

`damage(mutable hero, 10)` passes the variable `hero`, not a copy, so the function changes `hero` in place. `moveBy(mutable hero.position, 3, 4)` passes one field. Neither call copies the player.

Return a new value when the caller may want to keep the old one. Use `mutable` when the caller is done with the old one and the value is large or changes often. `Arrays.push(mutable items, value)`, which you have used since chapter 2, is a function with a `mutable` parameter.

## What a `mutable` call allows

The rest of this section covers the mistakes the compiler stops at a `mutable` call, and how to fix each one.

### The call must say `mutable`

Leave the word out at the call, and the compiler asks for it:

```narra error
// error: 12:8: parameter "hero" is mutable; write `mutable` before the argument
struct Player {
  hp: int
}

function heal(mutable hero: Player) {
  hero.hp = 100
}

export function run(args: Array<string>): int {
  let ana = Player { hp: 5 }
  heal(ana)
  return 0
}
```

Repeating the word at the call means you can tell, without opening `heal`, that `ana` may change.

### The argument must be a variable you can change

A `mutable` parameter writes to what it is given. So the argument must be a `let` variable, a `mutable` parameter, or a field or element of one. A `const` cannot change:

```narra error
// error: 7:23: constant "STARTING_ITEMS" cannot be changed; copy it into a `let` variable first
import * as Arrays from "std/array"

const STARTING_ITEMS = ["sword"]

export function run(args: Array<string>): int {
  Arrays.push(mutable STARTING_ITEMS, "shield")
  return 0
}
```

Following the message, copy the constant into a `let` and change that: `let items = STARTING_ITEMS`, then `Arrays.push(mutable items, "shield")`.

### One use of the variable per expression

While a call changes a variable, nothing else in the same expression may use it. Passing it twice is an error:

```narra error
// error: 10:16: "score" is passed as mutable in this call, so no other argument can use it
function swap(mutable first: int, mutable second: int) {
  const saved = first
  first = second
  second = saved
}

export function run(args: Array<string>): int {
  let score = 1
  swap(mutable score, mutable score)
  return 0
}
```

Reading it elsewhere in the expression is also an error. In `score + bump(mutable score)`, the result would depend on whether `score` is read before or after `bump` changes it:

```narra error
// error: 9:17: "score" is passed as mutable in one call, so the rest of the expression cannot use it
function bump(mutable number: int): int {
  number = number * 10
  return 1
}

export function run(args: Array<string>): int {
  let score = 2
  const total = score + bump(mutable score)
  return total
}
```

Split it into two statements, so the order is written down:

```narra
import * as Int from "std/int"

function bump(mutable number: int): int {
  number = number * 10
  return 1
}

export function run(args: Array<string>): int {
  let score = 2
  const bonus = bump(mutable score)
  const total = score + bonus
  print(Int.toString(total))
  return 0
}
```

```text
21
```

### A `mutable` parameter stays inside the call

A `mutable` parameter reaches the caller's variable only while the call runs. It cannot be returned, or stored in a struct, array or other value that outlives the call:

```narra error
// error: 7:10: mutable parameter "player" cannot be returned or stored
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

To give back what the player looked like, copy it into a `const` first. The copy is an ordinary value, and you can return it:

```narra
import * as Int from "std/int"

struct Player {
  hp: int
}

function knockOut(mutable player: Player): Player {
  const before = player
  player.hp = 0
  return before
}

export function run(args: Array<string>): int {
  let hero = Player { hp: 7 }
  const before = knockOut(mutable hero)
  print(`${Int.toString(before.hp)} -> ${Int.toString(hero.hp)}`)
  return 0
}
```

```text
7 -> 0
```

## Closures and `let`

The game wants a function that reports the current score. A closure that captures `score` looks like the natural way to write it, but the compiler refuses it:

```narra error
// error: 8:21: a closure cannot capture "score" because it is declared with let; closures capture only constants
function show(report: () => int): int {
  return report()
}

export function run(args: Array<string>): int {
  let score = 1
  return show(() => score)
}
```

A closure keeps a copy of each value it captures, made when the closure is created, the same way an assignment copies. For a `const` that copy always equals the original, because neither can change. For a `let` the two could drift apart. After `score = score + 1`, the closure would still return the old score, while the code reads as if it returns the current one. So the compiler accepts only constants, and you make the copy yourself, which shows which value the closure sees:

```narra
import * as Int from "std/int"

function show(report: () => int): int {
  return report()
}

export function run(args: Array<string>): int {
  let score = 1
  score = score + 1
  const current = score
  print(Int.toString(show(() => current * 10)))
  score = score + 1
  print(Int.toString(show(() => current * 10)))
  return 0
}
```

```text
20
20
```

The second call prints 20 too, since `current` was bound when `score` was 2. When a closure must see each new score, give it a parameter, and pass the score at each call:

```narra
import * as Int from "std/int"

export function run(args: Array<string>): int {
  const report = (value: int): int => value * 10
  let score = 2
  print(Int.toString(report(score)))
  score = score + 1
  print(Int.toString(report(score)))
  return 0
}
```

```text
20
30
```

`report` captures nothing, and each call reads `score` at that moment.

A `mutable` parameter cannot be captured either, for a stronger reason. It is the caller's variable, reached only while the call runs, and a closure can still be called after the function has returned:

```narra error
// error: 3:16: a closure cannot capture mutable parameter "score"; copy it first with `const copy = score`
function reporter(mutable score: int): () => int {
  return () => score
}

export function run(args: Array<string>): int {
  let points = 1
  return reporter(mutable points)()
}
```

The fix is the same: copy it into a `const` (`const current = score`) and capture `current`.

## Freeing memory

A value is freed when the last name that holds it goes away. That happens when a function returns, or when a variable gets a new value. There is no `free` call and no collector that runs later.

A count can only reach zero if no value points back at itself, and the rules in this chapter make sure of that. Values cannot change once shared, and a `let` or `mutable` parameter cannot be captured or stored, so a value can never come to contain itself. Run a program with `NARRA_CHECK_LEAKS=1` to check: it reports any memory still held when the program exits.

Next: [10. Effects and testing](10-effects-and-testing.md).
