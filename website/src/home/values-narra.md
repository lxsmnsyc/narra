```narra
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
  const saved = copy(hero)
  damage(mutable hero, 30)
  print(`${Int.toString(hero.hp)} ${Int.toString(saved.hp)}`)
  return 0
}
```
