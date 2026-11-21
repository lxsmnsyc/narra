```narra error
// error: 4:16: * needs two numbers of the same type, found int and f64
// error: 4:16: expected f64, found int
export function withTax(price: int): f64 {
  return price * 1.12
}
```
