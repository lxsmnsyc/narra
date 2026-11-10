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
// error: 8:3: parameter "player" is read-only; declare it as `mutable player` to change the caller's value, or copy it first with `copy(player)`
struct Player {
  name: string
  hp: int
}

function damage(player: Player, amount: int) {
  player.hp = player.hp - amount
}

export function run(args: Array<string>): int {
  const hero = Player { name: "Ana", hp: 100 }
  damage(hero, 10)
  return 0
}
```

A plain parameter is read-only: the function gets its own copy, and a change to it would never reach the caller. There are two ways to get the change back to the caller.

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
  const hero = Player { name: "Ana", hp: 100, position: Position { x: 0, y: 0 } }
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

`hero` is a `const`. A `const` fixes the name: `hero` names the same player for the whole function. The player's fields and elements can still change, through `hero.hp = 90` or a `mutable` call. A `let` adds one more thing, giving the name a new value with `hero = ...`.

`mutable` reaches the contents of a value, never the variable that holds it. So the function cannot replace the whole player (`player = other` is an error), and a `mutable` parameter must have a type with contents to change: a struct, `Array`, `Map`, `Box`, `StringBuilder` or `Generator`. A number has no contents, so a function that counts returns the new count instead:

```narra error
// error: 2:17: a mutable parameter needs a type whose contents change in place (a struct, Array, Map, Box, StringBuilder or Generator), found int; return the new value instead
function addOne(mutable _count: int) {
  print("one")
}

export function run(args: Array<string>): int {
  return 0
}
```

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
  const hero = Player { hp: 5 }
  heal(hero)
  return 0
}
```

Repeating the word at the call means you can tell, without opening `heal`, that `hero` may change.

### Add to the starting kit

Every player starts with the items in a constant, and the hero gets a shield on top. Pushing the shield onto the constant fails. A local `const` fixes only its name, but a top-level constant is fixed at every depth: there are no global variables, and a constant that could change would be one.

```narra error
// error: 7:23: constant "STARTING_KIT" cannot be changed; copy it into a local first, as in `let startingKit = STARTING_KIT`
import * as Arrays from "std/array"

const STARTING_KIT = ["sword"]

export function run(args: Array<string>): int {
  Arrays.push(mutable STARTING_KIT, "shield")
  return 0
}
```

A `mutable` argument must be a local `const` or `let`, a `mutable` parameter, or a field or element of one. Following the message, copy the kit into a local and change that:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

const STARTING_KIT = ["sword"]

export function run(args: Array<string>): int {
  const heroKit = STARTING_KIT
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

A spell swaps the positions of two players. With two local variables, `swapPositions(mutable hero, mutable rival)` works. The game keeps its players in an array, though, and passing two elements of it fails:

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
  const players = [Player { name: "Ana", position: Position { x: 0, y: 0 } }, Player { name: "Bo", position: Position { x: 5, y: 5 } }]
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
  const players = [Player { name: "Ana", position: Position { x: 0, y: 0 } }, Player { name: "Bo", position: Position { x: 5, y: 5 } }]
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

### Heal the whole party

A fountain heals every player. A loop that writes to its loop variable does not compile:

```narra error
// error: 10:5: "player" is a copy of an element of players, so changing it would not change players; write `for (mutable player of players)`, or copy it first with `copy(player)`
struct Player {
  name: string
  hp: int
}

export function run(args: Array<string>): int {
  const players = [Player { name: "Ana", hp: 3 }, Player { name: "Bo", hp: 7 }]
  for (const player of players) {
    player.hp = player.hp + 5
  }
  return 0
}
```

`for (const player of players)` gives each element as a copy, as a pattern does. Write `mutable` in place of `const`, and `player` reaches each element of `players` itself, the way a `mutable` parameter reaches the caller's value:

```narra
import * as Int from "std/int"
import * as Arrays from "std/array"

struct Player {
  name: string
  hp: int
}

export function run(args: Array<string>): int {
  const players = [Player { name: "Ana", hp: 3 }, Player { name: "Bo", hp: 7 }]
  for (mutable player of players) {
    player.hp = player.hp + 5
  }
  for (const #[index, mutable player] of Arrays.entries(players)) {
    player.name = `${Int.toString(index + 1)}. ${player.name}`
  }
  for (const player of players) {
    print(`${player.name}: ${Int.toString(player.hp)}`)
  }
  return 0
}
```

```text
1. Ana: 8
2. Bo: 12
```

Inside such a loop the body cannot use `players` at all, not even to read it: a `push` there could move the elements `player` points at. Reach the elements through `player`, or write a loop over the indexes when the body needs other elements too. The element type needs contents to change, so `for (mutable n of numbers)` over an `Array<int>` is an error; write `numbers[index] = ...` in an `Arrays.entries` loop.

### Log the hp before and after healing

Now `heal` adds an amount and returns the new hp, and the game logs the change in one line:

```narra error
// error: 15:25: "hero" is passed as mutable in one call, so the rest of the expression cannot use it
import * as Int from "std/int"

struct Player {
  hp: int
}

function heal(mutable player: Player, amount: int): int {
  player.hp = player.hp + amount
  return player.hp
}

export function run(args: Array<string>): int {
  const hero = Player { hp: 5 }
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
  const hero = Player { hp: 5 }
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
// error: 8:10: mutable parameter "player" cannot be returned or stored; copy it first with `copy(player)`
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

A `mutable` parameter reaches the caller's variable only while the call runs, so it cannot be returned or stored in anything that outlives the call. Copy it into a `const` first with `copy(player)`. The copy is an ordinary value, and you can return it:

```narra
import * as Int from "std/int"

struct Player {
  hp: int
}

function knockOut(mutable player: Player): Player {
  const before = copy(player)
  player.hp = 0
  return before
}

export function run(args: Array<string>): int {
  const hero = Player { hp: 7 }
  const before = knockOut(mutable hero)
  print(`${Int.toString(before.hp)} -> ${Int.toString(hero.hp)}`)
  return 0
}
```

```text
7 -> 0
```

### Return part of a parameter

The results screen also shows the party's leader. `leader` returns a player that is still stored in the party, so the result and the parameter share one value. The signature says so with `from party` after the result type. Then every caller can see, from the signature alone, that the player it gets back is part of the party it passed. Without `from` the compiler reports the shared result and offers to add it.

`from` names one parameter. A function that may return either of two parameters names one of them and copies the other with `copy(...)`, so the result never shares the unnamed one. A copy costs nothing until one of the two values changes, as described at the start of this chapter.

```narra
struct Player {
  name: string
  hp: int
}

struct Party {
  players: Array<Player>
}

function leader(party: Party): Player from party {
  return party.players[0]
}

function healthier(first: Player, second: Player): Player from first {
  if (second.hp > first.hp) {
    return copy(second)
  }
  return first
}

export function run(args: Array<string>): int {
  const party = Party { players: [Player { name: "Ana", hp: 60 }, Player { name: "Bo", hp: 80 }] }
  print(leader(party).name)
  print(healthier(party.players[0], party.players[1]).name)
  return 0
}
```

```text
Ana
Bo
```

An arrow function can write `from` too, as in `(party: Party): Player from party => party.players[0]`. One that is passed to a function whose parameter type has no `from` returns a copy instead: `(party) => copy(party.players[0])`.

## Closures and `let`

A closure captures what it uses in one of two ways, and the way depends on where the closure goes.

A closure passed straight to a parameter that is not `escaping` can only run during that call. It works directly on the `let` variables and `mutable` parameters around it, and on each `const` whose fields or elements it changes, with no copy. So it sees every assignment, and its own changes are visible afterwards ([chapter 3](03-functions-and-closures.md#use-values-from-around-a-function)). Nothing is copied or counted, because the variable outlives the closure.

Every other closure, one that is stored, returned or passed to an `escaping` parameter, can run after the function has returned. It keeps a copy of each value it captures, made when the closure is created, the same way `const saved = value` copies. Such a closure cannot change what it captures, since the change would reach only its copy. A `let` is an error there too. After `score = score + 1`, the closure would still see the old score, while the code reads as if it sees the current one. Copy it into a `const` first:

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

A `mutable` parameter cannot be captured by a closure that outlives the call, for a stronger reason. It is the caller's variable, reached only while the call runs, and such a closure can still be called after the function has returned:

```narra error
// error: 7:16: a closure cannot capture mutable parameter "score" unless it is passed straight to a function parameter that is not escaping; copy it first with `copy(score)`
struct Score {
  points: int
}

function reporter(mutable score: Score): () => int {
  return () => score.points
}

export function run(args: Array<string>): int {
  const total = Score { points: 1 }
  return reporter(mutable total)()
}
```

The fix is the same: copy it into a `const` (`const current = score.points`) and capture `current`.

## Freeing memory

A value is freed when the last name that holds it goes away. That happens when a function returns, or when a variable gets a new value. There is no `free` call and no collector that runs later.

A count can only reach zero if no value points back at itself, and the rules in this chapter make sure of that. A write to a shared value copies it first. A `mutable` parameter cannot be stored, and only a closure that cannot outlive the call reaches a variable directly, so a value can never come to contain itself. Run a program with `NARRA_CHECK_LEAKS=1` to check: it reports any memory still held when the program exits.

Next: [10. Effects and testing](10-effects-and-testing.md).
