```narra error
// error: 7:16: + needs two numbers of the same type, found Money and Money
struct Money {
  cents: int
}

export function total(price: Money, tax: Money): Money {
  return price + tax
}
```
