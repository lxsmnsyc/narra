```narra
import * as Fs from "std/fs"
import * as Int from "std/int"

export function total(prices: Array<int>): int {
  let sum = 0
  for (const price of prices) {
    sum = sum + price
  }
  return sum
}

export function receipt(prices: Array<int>): string uses alloc {
  return `total: ${Int.toString(total(prices))}`
}

export function save(path: string, prices: Array<int>): Result<int, string> uses io, alloc {
  return Fs.writeText(path, receipt(prices))
}
```
