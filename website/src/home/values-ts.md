```ts
type Player = { name: string; hp: number };

function damage(player: Player, amount: number) {
  player.hp = player.hp - amount;
}

const hero: Player = { name: "Ana", hp: 100 };
const saved = hero;
damage(hero, 30);
console.log(`${hero.hp} ${saved.hp}`);
```
