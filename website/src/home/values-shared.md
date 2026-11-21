```narra error
// error: 15:17: saved would be a second name for the struct in hero; use hero, or copy(hero) for a struct of its own
import * as Int from "std/int"

struct Player {
  name: string
  hp: int
}

function damage(mutable player: Player, amount: int) {
  player.hp = player.hp - amount
}

export function run(args: Array<string>): int uses io, alloc {
  let hero = Player { name: "Ana", hp: 100 }
  const saved = hero
  damage(mutable hero, 30)
  print(`${Int.toString(hero.hp)} ${Int.toString(saved.hp)}`)
  return 0
}
```
