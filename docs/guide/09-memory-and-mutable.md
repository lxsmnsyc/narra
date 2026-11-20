# 9. Memory and `mutable`

You never free memory in Narra, and a change to a value never shows up through another name by surprise. What you decide is where a value is copied and where it may change. This chapter follows a player in a small game. It shows how the game's values are stored and why the compiler refuses some writes. It then covers how a function changes the player and what a closure may capture. It ends with closures that share changing state, and with the loops of shared handles that leak memory.

## Why copies are cheap

Chapters 2 and 4 showed `copy(...)`. After `const snapshot = copy(prices)`, pushing to `prices` leaves `snapshot` as it was, and after `let moreTea = copy(tea)`, changing `moreTea` leaves `tea` alone. So when you pass a value to a function, it comes back unchanged, unless the call marks it `mutable`.

Copying a large array every time would be slow, so Narra copies only when it has to. Arrays, maps, strings, closures, enum variants with fields and `Box` values are kept on the heap. The heap is memory set aside apart from the variable, which lives until nothing uses it. The variable holds a reference to it. Numbers, `bool`, `char` and structs live directly in the variable, and a struct's heap fields are referenced one by one.

Each heap value counts the references to it, which is called reference counting. In chapter 2's example:

1. `const snapshot = copy(prices)` copies nothing yet. Both names refer to the same array, and its count goes up to 2.
2. `Arrays.push(mutable prices, 400)` sees the count of 2. Writing in place would change `snapshot` too, so it copies the array first and writes to the copy.
3. When a value's count is 1, a write changes it in place with no copy.
4. When the count drops to 0, the value is freed.

So a loop that changes an array held by one name writes in place every time. Had a second name held the array, the first write would copy it once. `copy(...)` is where you ask for that second name. When you write it, you choose to pay for the copy.

## A second name for a value

Now take the second name without `copy`. Say the game keeps a backup of the hero to show on the results screen, and the backup is taken by plain assignment:

```narra error
// error: 9:16: backup would be a second name for the struct in hero; use hero, or copy(hero) for a struct of its own
struct Player {
  name: string
  hp: int
}

export function run(args: Array<string>): int {
  const hero = Player { name: "Ana", hp: 100 }
  let backup = hero
  backup.hp = 0
  print(hero.name)
  return 0
}
```

`backup` and `hero` would name one player, and a write through one name would have to copy the player first, so that the other keeps its hp. The source would not show that copy. So Narra rejects the second name at the line that makes it. Arrays, maps, structs, boxes, string builders and generators are the values that change in place, so a name for one of them cannot be given to another name. Numbers, strings, enums and functions never change in place, so `const total = count` is fine.

The fix is the one the message gives: write `copy(...)`. Now the two names hold two players, and the copy happens where you can see it:

```narra
import * as Int from "std/int"

struct Player {
  name: string
  hp: int
}

export function run(args: Array<string>): int {
  const hero = Player { name: "Ana", hp: 100 }
  let backup = copy(hero)
  backup.hp = 0
  print(`${Int.toString(hero.hp)} ${Int.toString(backup.hp)}`)
  return 0
}
```

```text
100 0
```

`narrac apply-fix` writes this change for you.

## Writing to a shared value

A name is not the only way to share. Reading a field or an element into a name shares it too, because that name holds the same array as the field:

```narra error
// error: 13:15: Arrays.push(mutable team.members, ...) may change team.members, which members still uses at line 14; take a copy where it is shared, with copy(...)
import * as Arrays from "std/array"
import * as Int from "std/int"

struct Team {
  name: string
  members: Array<string>
}

export function run(args: Array<string>): int {
  let team = Team { name: "Red", members: ["Ana"] }
  const members = team.members
  Arrays.push(mutable team.members, "Bo")
  print(Int.toString(Arrays.length(members)))
  return 0
}
```

The write would have to copy `members` first, so the compiler reports the write instead of copying silently. The error names the write, the place it may change and the later line that uses the other name. The fix is again `copy(team.members)` where the sharing starts. The rule has these details:

- It looks at both names. Writing through `members` while `team.members` is still used is an error just like the one above.
- It follows chains. After `const members = team.members; const both = [members]`, a write to any of the three is checked against the other two.
- It ends when the other name is no longer used. In the sample above, removing the last `print` line makes the write legal, because nothing reads `members` afterwards.
- It ends when a place gets a whole new value, as in `team.members = []`.
- While a `for` loop walks `players`, a write to `players` in the body is an error. Loop over `copy(players)`, or write after the loop.
- A closure that captures a value shares it, so calling the closure counts as a use of that value.
- Passing `mutable` writes the argument and everything below it, so `damage(mutable hero, 10)` is checked the same way.
- A plain parameter shares its value with the caller, and a top-level constant shares with every user, so a write to something they share is an error at once. A function whose result is part of a parameter says so with `from`, covered later in this chapter, and a caller then treats the result as shared with the argument.

Two index places count as the same unless the compiler can tell them apart. A literal index is a place of its own: `rows[0]` and `rows[1]` are different places, so after `const first = rows[0]`, a write to `rows[1][0]` is allowed. An index that is a variable, such as `rows[i]` and `rows[j]`, or a map key, may name any element, so it counts as the same place as every other index. Writing `rows[j][0]` while `first = rows[i]` is still used is an error.

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
// error: 7:23: constant "STARTING_KIT" cannot be changed; copy it into a local first with `copy(STARTING_KIT)`
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
  const heroKit = copy(STARTING_KIT)
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

Every other closure, one that is stored, returned or passed to an `escaping` parameter, can run after the function has returned. It keeps a copy of each value it captures, made when the closure is created, the same way `copy(value)` does. Such a closure cannot change what it captures, since the change would reach only its copy. A `let` is an error there too. After `score = score + 1`, the closure would still see the old score, while the code reads as if it sees the current one. Copy it into a `const` first:

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

## State that closures share: `Cell`

Sometimes a stored closure has to see a score that keeps changing. The game wants a scorer that is handed out and called later, once per kill. A `let` cannot be captured there:

```narra error
// error: 5:5: a closure cannot capture "score" because it is declared with let, unless it is passed straight to a function parameter that is not escaping; to share state that changes with an escaping closure, keep it in a Cell from std/cell
function makeScorer(): (points: int) => int {
  let score = 0
  return (points) => {
    score = score + points
    return score
  }
}

export function run(args: Array<string>): int {
  const addPoints = makeScorer()
  return addPoints(10)
}
```

The last words of the message name the fix. A `Cell` from `std/cell` is one slot that holds a value. The closure captures the `Cell` like any `const`, and what the `Cell` holds can still change:

```narra
import * as Cells from "std/cell"
import { Cell } from "std/cell"
import * as Int from "std/int"

function makeScorer(score: Cell<int>): (points: int) => int {
  return (points) => {
    Cells.update(score, (total) => total + points)
    return Cells.get(score)
  }
}

export function run(args: Array<string>): int {
  const score = Cells.create(0)
  const addPoints = makeScorer(score)
  const resetScore = () => Cells.set(score, 0)
  print(Int.toString(addPoints(10)))
  print(Int.toString(addPoints(5)))
  print(Int.toString(Cells.get(score)))
  resetScore()
  print(Int.toString(Cells.get(score)))
  return 0
}
```

```text
10
15
15
0
```

The module has four functions:

- `Cells.create(value)` makes a `Cell` that holds `value`.
- `Cells.get(cell)` returns a copy of the value.
- `Cells.set(cell, value)` replaces the value.
- `Cells.update(cell, change)` replaces the value with what `change` returns for the old one. `change` returns the new value, because a number cannot be a `mutable` parameter.

A `Cell` is not copied the way a struct is. Passing it to a function, capturing it in a closure or storing it in a struct makes another handle to the same slot. So `addPoints`, `resetScore` and `run` above all see every write. A type like this is an identity type. `Shared`, in [chapter 15](15-threads-and-shared-state.md), is another. `==` on two cells is an error, because it would not be clear whether it compares the handles or the values. Compare `Cells.get(first)` with `Cells.get(second)`.

Reach for a `Cell` when an escaping closure must change something. When the closure is passed straight to a function that only calls it, a `let` is enough, as chapter 3 showed. A `Cell` works on one thread only. Sending one to another thread is an error, and chapter 15 shows what to use there.

## Enums and tuples are replaced

A hero may carry a weapon, so the player gets a field of type `Option<Weapon>`. Every swing wears the weapon down. Reading the weapon with `is` gives a name for it, and writing to that name does not compile:

```narra error
// error: 15:5: "weapon" is bound by a pattern to a copy of part of the value, so it cannot be changed; copy it first with `copy(weapon)`
struct Weapon {
  name: string
  durability: int
}

struct Player {
  name: string
  weapon: Option<Weapon>
}

export function run(args: Array<string>): int {
  const hero = Player { name: "Ana", weapon: Some(Weapon { name: "sword", durability: 10 }) }
  if (hero.weapon is Some(weapon)) {
    weapon.durability = weapon.durability - 1
  }
  return 0
}
```

The fields of an enum value are fixed, and a tuple has no element assignment. So a struct inside either of them is not a place you can write to, and copying `weapon` would only change the copy. To change the weapon, build a changed one and store it back in the field, which is a place:

```narra
import * as Int from "std/int"

struct Weapon {
  name: string
  durability: int
}

struct Player {
  name: string
  weapon: Option<Weapon>
}

function wear(mutable player: Player) {
  if (player.weapon is Some(weapon)) {
    player.weapon = Some({ ...weapon, durability: weapon.durability - 1 })
  }
}

export function run(args: Array<string>): int {
  const hero = Player { name: "Ana", weapon: Some(Weapon { name: "sword", durability: 10 }) }
  wear(mutable hero)
  if (hero.weapon is Some(weapon)) {
    print(Int.toString(weapon.durability))
  }
  return 0
}
```

```text
9
```

`player.weapon = Some(...)` replaces the whole `Option`. This is also why `x is Some(mutable y)` does not exist: an enum payload is replaced, never changed in place. A tuple works the same way: build a new tuple and assign it to the field that holds it.

## Freeing memory

A value is freed when the last name that holds it goes away. That happens when a function returns, or when a variable gets a new value. There is no `free` call and no collector that runs later.

A count can only reach zero if no value points back at itself. For the values in this chapter the rules make sure of that. A write to a value that another name holds copies it first, and `copy(...)` is where you ask for that. A `mutable` parameter cannot be stored, and only a closure that cannot outlive the call reaches a variable directly, so a struct, an array or an enum can never come to contain itself.

### Cycles through a `Cell`

A `Cell` is the exception, because every handle names the same slot. A closure stored in a `Cell` that it also captures forms a loop. The slot holds the closure, and the closure holds the slot. Neither count reaches zero, so the memory is never freed. The compiler accepts the program, and there is no collector to find the loop later:

```narra fragment
type Handler = () => int

const handler: Cell<Option<Handler>> = Cells.create(None)
Cells.set(handler, Some(() => {
  if (Cells.get(handler) is Some(_)) {
    return 1
  }
  return 0
}))
```

`NARRA_CHECK_LEAKS=1` finds it. Run the full program with the variable set, and it reports what is still held when the program exits and stops with exit code 102:

```text
leak check: 2 allocations still live
```

The same can happen with `Shared`, `Sender`, `Receiver` and `Async`, the other identity types. A handle stored inside the object it names makes the same loop.

There are two ways out. The first is to clear the slot before the last outside handle goes away. Setting the `Cell` back to `None` drops the closure, and the loop is gone:

```narra
import * as Cells from "std/cell"
import { Cell } from "std/cell"
import * as Int from "std/int"

type Handler = () => int

export function run(args: Array<string>): int {
  const handler: Cell<Option<Handler>> = Cells.create(None)
  Cells.set(handler, Some(() => {
    if (Cells.get(handler) is Some(_)) {
      return 1
    }
    return 0
  }))
  if (Cells.get(handler) is Some(call)) {
    print(Int.toString(call()))
  }
  Cells.set(handler, None)
  return 0
}
```

```text
1
```

The second is to keep a graph out of identity types altogether. Store the nodes in an arena and let them name each other by handle, as [chapter 13](13-standard-library.md#graphs-stdarena) shows. An arena frees every node when it goes away.

### Copies the compiler does not show

The compile-time rules of this chapter find most places where a write has to copy a shared value, but not all. The compiler does not follow a value into a `mutable` argument of your own function. Here `keep` stores the caller's array in a bag, and the later write copies it:

```narra fragment
struct Bag {
  items: Array<int>
}

function keep(mutable bag: Bag, items: Array<int>) {
  bag.items = items
}

export function run(args: Array<string>): int {
  const first = [1, 2, 3]
  let bag = Bag { items: [] }
  keep(mutable bag, first)
  bag.items[0] = 10
  return 0
}
```

The program is correct, since `first` still starts with 1. It pays for a copy that the source does not show. To find such copies, run the program with `NARRA_CHECK_COPIES=1`. At exit it reports the copies on write that no `copy(...)` explains, and stops with exit code 103:

```text
copy check: copies on write that no copy(...) explains: array 1
```

Write the copy where a reader can see it, in `keep`, and the check passes:

```narra
import * as Int from "std/int"

struct Bag {
  items: Array<int>
}

function keep(mutable bag: Bag, items: Array<int>) {
  bag.items = copy(items)
}

export function run(args: Array<string>): int {
  const first = [1, 2, 3]
  let bag = Bag { items: [] }
  keep(mutable bag, first)
  bag.items[0] = 10
  print(`${Int.toString(first[0])} ${Int.toString(bag.items[0])}`)
  return 0
}
```

```text
1 10
```

Both checks are for debug builds. A release build has neither.

Next: [10. Effects and testing](10-effects-and-testing.md).
