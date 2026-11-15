```narra
import * as Int from "std/int"

export function describe(celsius: int): string uses alloc {
  return `${Int.toString(celsius)} C`
}
```
