# 9. Memory and `mutable`

You never free memory in Narra, and you never decide whether two names share one object. What you do decide is where a value may change. This chapter follows a player in a small game. It shows how the game's values are stored, how a function changes the player, and what a closure may capture.

## Why copies are cheap

Chapters 2 and 4 showed that assigning an array or a struct gives the new name its own copy. After `const snapshot = prices`, pushing to `prices` leaves `snapshot` as it was, and after `let moreTea = tea`, changing `moreTea` leaves `tea` alone. So when you pass a value to a function, it comes back unchanged, unless the call marks it `mutable`.

Copying a large array at every assignment would be slow, so Narra copies only when it has to. Arrays, maps, strings, closures, enum variants with fields and `Box` values are kept on the heap. The heap is memory set aside apart from the variable, which lives until nothing uses it. The variable holds a reference to it. Numbers, `bool`, `char` and structs live directly in the variable, and a struct's heap fields are referenced one by one.

Each heap value counts the references to it, which is called reference counting. In chapter 2's example:

1. `const snapshot = prices` copies nothing. Both names refer to the same array, and its count goes up to 2.
2. `Arrays.push(mutable prices, 400)` sees the count of 2. Writing in place would change `snapshot` too, so it copies the array first and writes to the copy.
3. When a value's count is 1, a write changes it in place with no copy.
4. When the count drops to 0, the value is freed.

So a loop that changes an array held by one name writes in place every time. Had a second name held the array, the first write would copy it once. To keep large updates fast, avoid holding a second name for a value while you change it.

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

## More of the game with `mutable`

Each task below runs into something the compiler checks at a `mutable` call, and shows the fix.

### Heal the hero

A potion sets the hero's hp back to 100. Forget `mutable` at the call, and the compiler asks for it:

```narra error
// error: 12:8: parameter "player" is mutable; write `mutable` before the argument
struct Player {
  hp: int
}

function heal(mutable player: Player) {
  player.hp = 100
}

export function run(args: Array<string>): int {
  let hero = Player { hp: 5 }
  heal(hero)
  return 0
}
```

Repeating the word at the call means you can tell, without opening `heal`, that `hero` may change.

### Add to the starting kit

Every player starts with the items in a constant, and the hero gets a shield on top. Pushing the shield onto the constant fails, because a `const` never changes:

```narra error
// error: 7:23: constant "STARTING_KIT" cannot be changed; copy it into a `let` variable first
import * as Arrays from "std/array"

const STARTING_KIT = ["sword"]

export function run(args: Array<string>): int {
  Arrays.push(mutable STARTING_KIT, "shield")
  return 0
}
```

A `mutable` argument must be a `let` variable, a `mutable` parameter, or a field or element of one. Following the message, copy the kit into a `let` and change that:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

const STARTING_KIT = ["sword"]

export function run(args: Array<string>): int {
  let heroKit = STARTING_KIT
  Arrays.push(mutable heroKit, "shield")
  print(Int.toString(Arrays.length(heroKit)))
  print(Int.toString(Arrays.length(STARTING_KIT)))
  return 0
}
```

```text
2
1
```

### Swap two players' places

A spell swaps the positions of two players. With two `let` variables, `swapPositions(mutable hero, mutable rival)` works. The game keeps its players in an array, though, and passing two elements of it fails:

```narra error
// error: 20:32: "players" is passed as mutable in this call, so no other argument can use it
struct Position {
  x: int
  y: int
}

struct Player {
  name: string
  position: Position
}

function swapPositions(mutable first: Player, mutable second: Player) {
  const saved = first.position
  first.position = second.position
  second.position = saved
}

export function run(args: Array<string>): int {
  let players = [Player { name: "Ana", position: Position { x: 0, y: 0 } }, Player { name: "Bo", position: Position { x: 5, y: 5 } }]
  swapPositions(mutable players[0], mutable players[1])
  return 0
}
```

While a call changes a variable, no other argument may touch it. Both arguments are parts of `players`. The compiler does not compare the indexes, which could be equal at run time, so it refuses the call. Pass the array once, with the two indexes:

```narra
import * as Int from "std/int"

struct Position {
  x: int
  y: int
}

struct Player {
  name: string
  position: Position
}

function swapPositions(mutable players: Array<Player>, first: int, second: int) {
  const saved = players[first].position
  players[first].position = players[second].position
  players[second].position = saved
}

export function run(args: Array<string>): int {
  let players = [Player { name: "Ana", position: Position { x: 0, y: 0 } }, Player { name: "Bo", position: Position { x: 5, y: 5 } }]
  swapPositions(mutable players, 0, 1)
  for (const player of players) {
    print(`${player.name} at ${Int.toString(player.position.x)},${Int.toString(player.position.y)}`)
  }
  return 0
}
```

```text
Ana at 5,5
Bo at 0,0
```

### Log the hp before and after healing

Now `heal` adds an amount and returns the new hp, and the game logs the change in one line:

```narra error
// error: 15:9: "hero" is passed as mutable in one call, so the rest of the expression cannot use it
import * as Int from "std/int"

struct Player {
  hp: int
}

function heal(mutable player: Player, amount: int): int {
  player.hp = player.hp + amount
  return player.hp
}

export function run(args: Array<string>): int {
  let hero = Player { hp: 5 }
  print(`${Int.toString(hero.hp)} -> ${Int.toString(heal(mutable hero, 10))}`)
  return 0
}
```

The template reads `hero.hp` and also passes `hero` as `mutable`. Whether it prints `5` or `15` first would depend on which part runs first. Split it into statements, so the order is written down:

```narra
import * as Int from "std/int"

struct Player {
  hp: int
}

function heal(mutable player: Player, amount: int): int {
  player.hp = player.hp + amount
  return player.hp
}

export function run(args: Array<string>): int {
  let hero = Player { hp: 5 }
  const before = hero.hp
  const after = heal(mutable hero, 10)
  print(`${Int.toString(before)} -> ${Int.toString(after)}`)
  return 0
}
```

```text
5 -> 15
```

### Keep a snapshot

A knockout sets the hp to 0, and the game wants the player as it was before, to show on the results screen. Returning the `mutable` parameter fails:

```narra error
// error: 8:10: mutable parameter "player" cannot be returned or stored; copy it first with `const copy = player`
struct Player {
  hp: int
}

function knockOut(mutable player: Player): Player {
  player.hp = 0
  return player
}

export function run(args: Array<string>): int {
  return 0
}
```

A `mutable` parameter reaches the caller's variable only while the call runs, so it cannot be returned or stored in anything that outlives the call. Copy it into a `const` first. The copy is an ordinary value, and you can return it:

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

Chapter 3 showed that a closure cannot capture a `let`, and that the fix is to copy the value into a `const` first. Here is what is behind it.

A closure keeps a copy of each value it captures, made when the closure is created, the same way an assignment copies. For a `const` the copy always equals the original, because neither can change. For a `let` the two would drift apart. After `score = score + 1`, the closure would still see the old score, while the code reads as if it sees the current one. Capturing only constants keeps the two the same:

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
// error: 3:16: a closure cannot capture mutable parameter "score" unless it is passed straight to a function parameter that is not escaping; copy it first with `const copy = score`
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
