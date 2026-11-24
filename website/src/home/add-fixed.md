```narra
struct Money {
  cents: int
}

function add(first: Money, second: Money): Money {
  return Money { cents: first.cents + second.cents }
}

export function total(price: Money, tax: Money): Money {
  return add(price, tax)
}
```
