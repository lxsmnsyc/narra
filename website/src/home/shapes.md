```narra
import * as Int from "std/int"

enum Shape {
  Circle(radius: int)
  Rect(width: int, height: int)
}

function area(shape: Shape): int {
  return match (shape) {
    Shape.Circle(radius) => 3 * radius * radius,
    Shape.Rect(width, height) => width * height,
  }
}

export function run(args: Array<string>): int uses io, alloc {
  for (const shape of [Shape.Circle(3), Shape.Rect(3, 4)]) {
    print(`area ${Int.toString(area(shape))}`)
  }
  return 0
}
```
