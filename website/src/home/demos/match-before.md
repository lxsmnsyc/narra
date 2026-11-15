```narra
enum Shape {
  Circle(radius: int)
  Rect(width: int, height: int)
  Triangle(base: int, height: int)
}

function area(shape: Shape): int {
  return match (shape) {
    Shape.Circle(radius) => 3 * radius * radius,
    Shape.Rect(width, height) => width * height,
  }
}
```
