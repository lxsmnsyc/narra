```narra
import * as Arrays from "std/array"

export function run(args: Array<string>): int {
  const prices = [250, 320]
  const snapshot = copy(prices)
  Arrays.push(mutable prices, 400)
  return snapshot[0]
}
```
