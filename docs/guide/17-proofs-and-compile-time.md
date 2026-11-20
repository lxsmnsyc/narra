# 17. Proofs and compile-time constants

[Chapter 2](02-values-and-types.md#dividing-by-a-name) showed that the compiler accepts a division only when it can see that the divisor is not zero. [Chapter 3](03-functions-and-closures.md#conditions-a-caller-must-meet) showed `requires`, which moves that proof to the caller. This chapter uses the same proofs for more. A function can require an order between two of its parameters, and it can promise something about its result. A type can carry a rule that every value of it follows. An array of fixed length can be read with no runtime check at all. These proofs happen while the program compiles, so they cost nothing when it runs.

The examples work with a week of working hours. Hours of the day go from 0 to 24, and days of the week from 0 to 6.

## Conditions between two parameters

A shift starts and ends at an hour of the day, and it must not end before it starts. A `requires` condition can compare two integer parameters of the same type, as `start <= end` does here:

```narra
import * as Int from "std/int"

function hoursBetween(start: int, end: int): int
  requires start >= 0 && start <= end && end <= 24
{
  return end - start
}

function shiftLength(start: int, end: int): int {
  if (start >= 0 && start <= end && end <= 24) {
    return hoursBetween(start, end)
  }
  return 0
}

export function run(args: Array<string>): int {
  print(Int.toString(hoursBetween(9, 17)))
  print(Int.toString(shiftLength(13, 22)))
  print(Int.toString(shiftLength(22, 13)))
  return 0
}
```

```text
8
9
0
```

A comparison between two parameters is called a relation. A caller proves it in one of these ways:

- **Two constants**, such as `hoursBetween(9, 17)`.
- **Its own check** of the same two names, such as `if (start <= end)` in `shiftLength`.
- **Its own `requires`** with the same relation, or a stronger one: `start < end` proves `start <= end` and `start != end`.
- **Bounds that settle the order**, such as a name checked to be below 10 and another checked to be above 20.

Bounds that leave the order open prove nothing. This caller checks that `start` is at least 0 and `end` at most 24, which says nothing about which is larger:

```narra error
// error: 10:25: this argument is not proven to meet `requires start <= end` of hoursBetween; pass a constant, or a name checked first with an `if`
function hoursBetween(start: int, end: int): int
  requires start >= 0 && start <= end && end <= 24
{
  return end - start
}

function shiftLength(start: int, end: int): int {
  if (start >= 0 && end <= 24) {
    return hoursBetween(start, end)
  }
  return 0
}

export function run(args: Array<string>): int {
  return shiftLength(9, 17)
}
```

In the body, a relation is a known fact, and it passes a bound on to the other name. With `start <= end` and `end <= 24`, the body knows that `start` is at most 24 too. A relation passes a bound one step only: `a < b` and `b < c` say nothing about `a` and `c`. A relation cannot have arithmetic around a parameter, so `requires start + 1 <= end` is an error. The reason is that `+` wraps around at the largest `int`, so `start + 1` is not always larger than `start`.

### One condition or another

A shift that has not ended yet has no end hour. Say the program writes `-1` for it. A condition can list several comparisons joined by `||`, and the caller must prove one of them:

```narra
import * as Int from "std/int"

function describeShift(start: int, end: int): string
  requires start >= 0 && start <= 24
  requires end == -1 || end >= start
{
  if (end == -1) {
    return `from ${Int.toString(start)}, no end yet`
  }
  return `from ${Int.toString(start)} to ${Int.toString(end)}`
}

function describeChecked(start: int, end: int): string {
  if (start >= 0 && start <= 24 && end >= start) {
    return describeShift(start, end)
  }
  return "not a shift"
}

export function run(args: Array<string>): int {
  print(describeShift(9, 17))
  print(describeShift(22, -1))
  print(describeChecked(13, 20))
  print(describeChecked(13, 8))
  return 0
}
```

```text
from 9 to 17
from 22, no end yet
from 13 to 20
not a shift
```

`describeShift(22, -1)` proves the clause with the constant `-1`, and `describeChecked` proves it with its check of `end >= start`. Each line of `requires`, and each part joined by `&&`, is one clause. A clause is one comparison, or comparisons joined by `||`. Use parentheses to put a `||` clause inside an `&&` list: `requires (end == -1 || end >= start) && start >= 0`. An `&&` inside a `||` is an error. Split it into more clauses.

A caller must prove one side of the `||`. Checking the whole `||` in an `if` is not enough:

```narra error
// error: 13:33: this argument is not proven to meet `requires end == -1 || end >= start` of describeShift; pass a constant, or a name checked first with an `if`
import * as Int from "std/int"

function describeShift(start: int, end: int): string
  requires start >= 0 && start <= 24
  requires end == -1 || end >= start
{
  return `${Int.toString(start)} ${Int.toString(end)}`
}

function describeChecked(start: int, end: int): string {
  if (start >= 0 && start <= 24 && (end == -1 || end >= start)) {
    return describeShift(start, end)
  }
  return "not a shift"
}

export function run(args: Array<string>): int {
  print(describeChecked(13, 20))
  return 0
}
```

A check with `||` keeps only the smallest range that holds every side. It keeps that range when every side compares the same name with a constant. After `if (step < 0 || step > 0)` the compiler knows that `step` is not zero, but not which side holds. When the sides compare different things, as here, the check keeps nothing. Check one side at a time instead, with two `if` statements.

## Promise something about the result

`requires` states what the caller must prove. An `ensures` line states what the function proves about its result, and the caller learns that fact. Here `parseHour` reads an hour from text and promises that a `Some` holds an hour of the day:

```narra
import * as Int from "std/int"

function parseHour(text: string): Option<int>
  ensures Some(>= 0) | None
  ensures Some(<= 24) | None
{
  if (Int.parse(text, 10) is Ok(hour)) {
    if (hour >= 0 && hour <= 24) {
      return Some(hour)
    }
  }
  return None
}

function hoursBetween(start: int, end: int): int
  requires start >= 0 && start <= end && end <= 24
{
  return end - start
}

function shiftLength(startText: string, endText: string): Option<int> {
  if (parseHour(startText) is Some(start)) {
    if (parseHour(endText) is Some(end)) {
      if (start <= end) {
        return Some(hoursBetween(start, end))
      }
    }
  }
  return None
}

export function run(args: Array<string>): int {
  print(Show.show(shiftLength("9", "17")))
  print(Show.show(shiftLength("9", "25")))
  print(Show.show(shiftLength("noon", "17")))
  return 0
}
```

```text
Some(8)
None
None
```

`shiftLength` checks only `start <= end`. The bounds of `start` and `end` come from the `ensures` lines of `parseHour`.

An `ensures` line is a pattern, as in `match` (chapter 5), and every result must match it. It can also use the comparisons `< c`, `<= c`, `> c`, `>= c` and `!= c` with an integer constant `c`. A comparison fits any place that holds an integer: `ensures >= 0` for an `int` result, `ensures Ok(> 0) | Err(_)`, `ensures #[>= 0, _]` for a tuple, or `ensures { width: > 0 }` for a struct. Each place takes one comparison, so a range needs two lines, as `parseHour` has. A function with no result has no `ensures`.

The compiler proves every `return` against every line. A returned value is proven when it is written to match. That is a constant, a name checked first with an `if`, or a constructor such as `Some(hour)` around such a name. Anything else is an error, including a value that only looks right:

```narra error
// error: 9:10: this value is not proven to match `ensures <= 24`; write a constant, a constructor, or a name checked first with an `if`
function clampHour(value: int): int
  ensures >= 0
  ensures <= 24
{
  if (value < 0) {
    return 0
  }
  return value
}

export function run(args: Array<string>): int {
  return clampHour(30)
}
```

`value` is known to be at least 0 at the last `return`, but nothing says it is at most 24. Add `if (value > 24) { return 24 }` before it. A call is never proven, even a call to a function with its own `ensures`: `return parseHour(text)` is an error. Bind the result to a name first, check it, and return the name.

A caller learns the facts wherever it gives the result a name: `const hour = f()`, `f() is Some(hour)`, the arms of a `match`, and a destructuring. A generator ([chapter 12](12-generators.md)) can have `ensures` too. The compiler proves each `yield` value, and the loop name of `for (const n of g())` learns the facts.

## Arrays with a fixed length

A week always has seven days. `FixedArray<int, 7>` holds exactly seven `int` values. Its length is part of its type, so it needs no heap memory and is stored in place, as a struct is. An array literal builds one when the type says `FixedArray`:

```narra
import * as FixedArrays from "std/fixedarray"
import * as Int from "std/int"
import * as Ranges from "std/range"

function hoursOn(week: FixedArray<int, 7>, day: int): int {
  if (day >= 0 && day < 7) {
    return week[day]
  }
  return 0
}

export function run(args: Array<string>): int {
  let week: FixedArray<int, 7> = [8, 8, 6, 8, 4, 0, 0]
  week[5] = 3
  let total = 0
  for (const day of Ranges.upTo(7)) {
    total += week[day]
  }
  print(Int.toString(total))
  print(Int.toString(hoursOn(week, 2)))
  print(Int.toString(hoursOn(week, 9)))
  print(Show.show(FixedArrays.get(week, 9)))
  print(Int.toString(FixedArrays.length(week)))
  return 0
}
```

```text
37
6
0
None
7
```

Reading `items[i]` of an `Array` checks the index when the program runs and stops the program when the index is out of range. A `FixedArray` index is proven instead, and a read never stops the program. The index must be one of these:

- **A constant** from 0 to the length minus 1, such as `week[5]`.
- **A name checked at both ends**, such as `day` inside `if (day >= 0 && day < 7)`.
- **The name of a counting loop**: `for (const day of Ranges.upTo(7))`, or `for (let i = 0; i < 7; i++)` when nothing in the loop assigns `i`.

These are the same checks that prove a divisor, and the same rules end them: an assignment to the name, or passing it as `mutable`. A check of one end only is an error, and the message shows the check to add:

```narra error
// error: 4:17: the index of FixedArray<int, 7> must be proven to be 0 to 6; check it against 7 first, as in `if (day >= 0 && day < 7) { ... }`, or read it with FixedArrays.get
function hoursOn(week: FixedArray<int, 7>, day: int): int {
  if (day >= 0) {
    return week[day]
  }
  return 0
}

export function run(args: Array<string>): int {
  const week: FixedArray<int, 7> = [8, 8, 6, 8, 4, 0, 0]
  return hoursOn(week, 2)
}
```

When the index may be out of range, `FixedArrays.get(week, day)` from `std/fixedarray` takes any index and returns an `Option`. `FixedArrays.length(week)` gives the length. `for (const hours of week)` visits each element, and the patterns of [chapter 5](05-pattern-matching.md) take a `FixedArray` apart. A `FixedArray` cannot grow: it has no `push`. Use an `Array` when the length changes. One `FixedArray` value can take at most 1 MiB.

## A type with a rule

`hoursOn` above takes any `int` as the day and checks it every time. A refinement type checks it once. A distinct type (chapter 4) can list the values it accepts after `requires`:

```narra
import * as Int from "std/int"

distinct type Day = int derives Show requires >= 0 && <= 6

const MONDAY = Day.from(0)

function hoursOn(week: FixedArray<int, 7>, day: Day): int {
  return week[Day.toInt(day)]
}

function parseDay(text: string): Option<Day> {
  if (Int.parse(text, 10) is Ok(number)) {
    return Day.check(number)
  }
  return None
}

export function run(args: Array<string>): int {
  const week: FixedArray<int, 7> = [8, 8, 6, 8, 4, 0, 0]
  print(Int.toString(hoursOn(week, MONDAY)))
  print(Int.toString(hoursOn(week, Day.from(2))))
  print(Show.show(parseDay("4")))
  print(Show.show(parseDay("9")))
  if (parseDay("4") is Some(day)) {
    print(Int.toString(hoursOn(week, day)))
  }
  return 0
}
```

```text
8
6
Some(4)
None
4
```

A `Day` is always from 0 to 6, so `Day.toInt(day)` is a proven index, with no check in `hoursOn`. There are two ways to make a `Day`:

- **`Day.from(value)`** is proven while the program compiles, like a call to a function with `requires`. The value must be a constant or a name checked first with an `if`. Nothing runs when the program does.
- **`Day.check(value)`** tests the value while the program runs. It returns `Some(day)` when the value matches and `None` when it does not.

A `from` that the compiler cannot prove is an error, and the message names the rule that failed:

```narra error
// error: 5:27: 7 does not satisfy Day (<= 6)
distinct type Day = int requires >= 0 && <= 6

export function run(args: Array<string>): int {
  const sunday = Day.from(7)
  return Day.toInt(sunday)
}
```

The rule after `requires` is a list of patterns joined by `&&`. Each is a comparison with a constant (`< 8`, `>= 1`, `!= 0`), a literal, or several of these joined by `|`. The base type is an integer type or `string`, and a `string` takes literals only: `distinct type Method = string requires "GET" | "POST"`. The rule goes after `derives` and may start on the next line.

Unwrapping gives back the fact, wherever the value comes from: a parameter, a field or an array element. `Day.toInt(day)` is known to be from 0 to 6, so it proves an index and a `requires` condition. With a rule such as `!= 0` or `>= 1`, it proves a divisor. The fact covers only the unwrapped value: `Day.toInt(day) + 1` is a plain `int` again. A `string` rule gives no fact.

`derives Shape` (chapter 13) decodes a refinement type and then tests its rule, so a `Day` read from JSON is always from 0 to 6. `export opaque distinct type` hides `check` together with `from`, so only the declaring module can make values.

## A length as a parameter

A function that adds up a `FixedArray` should work for any length. A constant type parameter, written `const N: int` among the type parameters, stands for a number instead of a type:

```narra
import * as Int from "std/int"

function averageHours<const N: int>(days: FixedArray<int, N>): int {
  let total = 0
  for (let i = 0; i < N; i++) {
    total += days[i]
  }
  return total / N
}

function ruler<const N: int>(mark: string): string {
  let line = ""
  for (let i = 0; i < N; i++) {
    line = `${line}${mark}`
  }
  return line
}

export function run(args: Array<string>): int {
  const week: FixedArray<int, 7> = [8, 8, 6, 8, 4, 0, 0]
  const weekend: FixedArray<int, 2> = [3, 5]
  print(Int.toString(averageHours(week)))
  print(Int.toString(averageHours(weekend)))
  print(ruler<5>("-"))
  return 0
}
```

```text
4
4
-----
```

The compiler makes one copy of the function for each value of `N` the program uses, so `averageHours<7>` and `averageHours<2>` are two functions. In each copy `N` is a constant. That is why `i < N` proves the index and `total / N` needs no check. A call finds `N` from the types of its arguments, as `averageHours(week)` does. When no argument says it, write it, as in `ruler<5>("-")`. A struct can take one too: `struct Schedule<const N: int> { hours: FixedArray<int, N> }`, built with `Schedule<14> { hours: ... }`.

Each copy is proven on its own. A copy where `N` is 0 divides by zero, and the error names the copy:

```narra error
// error: 7:18: the divisor must be a non-zero constant or a name checked against zero first; this constant is zero; N is 0 in averageHours<0>
function averageHours<const N: int>(days: FixedArray<int, N>): int {
  let total = 0
  for (let i = 0; i < N; i++) {
    total += days[i]
  }
  return total / N
}

export function run(args: Array<string>): int {
  const none: FixedArray<int, 0> = []
  return averageHours(none)
}
```

A type argument is an integer literal or a constant parameter in scope. It never computes, so `FixedArray<int, N + 1>` is an error:

```narra error
// error: 2:76: a type argument cannot compute with `+`; write an integer literal or a constant parameter
function longer<const N: int>(days: FixedArray<int, N>): FixedArray<int, N + 1> {
  return days
}

export function run(args: Array<string>): int {
  return 0
}
```

`N` does not prove a `requires` condition yet. Passing `N` to a function with `requires index >= 1` is an error, even when every copy has `N` above 0. Only `int` constant parameters exist, and enums and type aliases take none.

## Values computed while compiling

The proofs above run while the program compiles. The compiler can also compute values. A top-level `const` may call a function marked `@const`, and the compiler runs that call and keeps the result. The program starts with the value already built, and nothing runs for it at startup.

Here is a table of the hours planned up to the end of each day, with 8 hours on each of the five weekdays:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"
import * as Ranges from "std/range"

@const
function plannedTotals(): Array<int> uses alloc {
  let totals: Array<int> = []
  let running = 0
  for (const day of Ranges.upTo(7)) {
    if (day < 5) {
      running += 8
    }
    Arrays.push(mutable totals, running)
  }
  return totals
}

const PLANNED = plannedTotals()

export function run(args: Array<string>): int {
  print(Int.toString(PLANNED[2]))
  print(Int.toString(PLANNED[6]))
  print(Int.toString(Arrays.length(plannedTotals())))
  return 0
}
```

```text
24
40
7
```

The call is written at the constant, so the program text shows where the work happens. Anywhere else, `plannedTotals()` is an ordinary function call and runs when the program does, as the last `print` shows. Only a function marked `@const` can be called from a constant. The compiler checks each `@const` function where it is written, so a mistake in one fails even when no constant uses it yet.

An `@const` function has these limits:

- **Its effects** are at most `alloc`, so it cannot print or read a file.
- **Its parameters and result** are numbers, `bool`, `char`, `string`, `bytes`, arrays, maps, `FixedArray`, `Option`, `Result` and tuples. Structs, enums and distinct types made of these are fine too. A function type, `Cell`, `Ptr`, `Shared` and a generator are not allowed.
- **Its calls** go to other `@const` functions, constructors, and the functions of number, text and collection modules (`std/string`, `std/array`, `std/map`, `std/int`, `std/range` and more). A function that takes a function is out. So is an arrow function, and so is another function of yours.
- **Its math** leaves out `Math.sin`, `cos`, `tan`, `atan2`, `exp`, `log` and `pow`. The C library decides their last digits, and the compiler must give the same result as the program.

A call outside the list is an error at the call:

```narra error
// error: 8:13: @const function paddedHours calls pad, which is not @const
function pad(hours: int): int {
  return hours + 1
}

@const
function paddedHours(hours: int): int {
  return pad(hours)
}

const PADDED = paddedHours(8)

export function run(args: Array<string>): int {
  return PADDED
}
```

Mark `pad` with `@const` and the program compiles.

A panic during the computation is a compile error at the constant, with the chain of calls under it. Here an index is out of range:

```narra error
// error: 7:19: constant BAD panicked while the compiler computed it: index 5 out of range for length 3
@const
function hourOn(week: Array<int>, day: int): int {
  return week[day]
}

const BAD = hourOn([8, 8, 6], 5)

export function run(args: Array<string>): int {
  return BAD
}
```

The computation is limited so that a compile cannot hang. One constant may run 10 million steps, where a step is one expression, statement or loop turn. It may build 64 MiB of values and nest 1000 calls. Passing a limit is an error that names the constant. A loop that adds up the numbers below 100 million stops with `constant ALL needs more than 10000000 evaluation steps`. Compute a value that large in a function at run time.

### Patterns checked while compiling

`Regexes.compile` ([chapter 13](13-standard-library.md#patterns-stdregex)) returns a `Result`, because the pattern may be wrong. A pattern written in the source is usually fixed text, and a mistake in it is a mistake in the program. `Regexes.checked(pattern)` is an `@const` function that compiles the pattern while the compiler runs. It returns a `Regex`, with no `Result` around it, and a bad pattern is a compile error:

```narra
import * as Regexes from "std/regex"

const CLOCK = Regexes.checked("^(\\d\\d):(\\d\\d)$")

function describeTime(text: string): string uses alloc {
  if (Regexes.find(CLOCK, text) is Some(found)) {
    if (Regexes.group(found, 1) is Some(hour)) {
      return `hour ${hour}`
    }
  }
  return "not a time"
}

export function run(args: Array<string>): int uses io, alloc {
  print(describeTime("09:30"))
  print(describeTime("9:30"))
  return 0
}
```

```text
hour 09
not a time
```

The compiled pattern is part of the program, so the program does not compile the pattern again. A pattern with a mistake stops the compile, and the message gives the place in the pattern:

```narra error
// error: 4:30: constant CLOCK panicked while the compiler computed it: bad pattern "(\d\d:(\d\d)" at 0: unclosed group
import * as Regexes from "std/regex"

const CLOCK = Regexes.checked("(\\d\\d:(\\d\\d)")

export function run(args: Array<string>): int {
  return 0
}
```

`checked` is only for the value of a top-level `const`. Inside a function it would have to panic at run time for a bad pattern, so it is an error there. Use `Regexes.compile` and handle the `Err`:

```narra error
// error: 5:32: Regexes.checked can only be the value of a top-level const; use Regexes.compile
import * as Regexes from "std/regex"

function matchesClock(text: string): bool uses alloc {
  const clock = Regexes.checked("^\\d\\d:\\d\\d$")
  return Regexes.matches(clock, text)
}

export function run(args: Array<string>): int uses alloc {
  if (matchesClock("09:30")) {
    return 0
  }
  return 1
}
```

### Tagged templates and `Template.raw`

The pattern above doubles every backslash, because a string decodes `\\d` to `\d`. A tagged template avoids that. Write a name before the opening backtick of a template, and the compiler calls that function with the pieces. ``tag`a${x}b` `` is the call `tag(["a", "b"], [x])`. The first argument holds the text between the holes, always one more piece than there are values, with `""` for an empty piece. The second holds the values of the holes. The tag has two parameters, neither `mutable`, and the values must have one type. A tag can build any result:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"
import * as Ranges from "std/range"
import * as Str from "std/string"

function hours(parts: Array<string>, values: Array<int>): string {
  let builder = Str.builder()
  for (const index of Ranges.upTo(Arrays.length(parts))) {
    Str.append(mutable builder, parts[index])
    if (index < Arrays.length(values)) {
      Str.append(mutable builder, `${Int.toString(values[index])}h`)
    }
  }
  return Str.build(builder)
}

export function run(args: Array<string>): int {
  const start = 9
  print(hours`from ${start} to ${start + 8}, break ${1}`)
  print(hours`no holes`)
  return 0
}
```

```text
from 9h to 17h, break 1h
no holes
```

Each hole is checked against the type the tag takes, so a hole of another type is an error:

```narra error
// error: 10:22: value 1 of the tagged template is string, but the tag takes values of type int; convert it inside the ${}
import * as Arrays from "std/array"
import * as Int from "std/int"

function hours(parts: Array<string>, values: Array<int>): string {
  return `${parts[0]}${Int.toString(Arrays.length(values))}`
}

export function run(args: Array<string>): int {
  print(hours`from ${"nine"} to`)
  return 0
}
```

The tag must be on the same line as the backtick, and a template after `)`, `]` or a literal is an error.

A tag can take `Raw` from `std/template` as its first parameter. Then the pieces keep their source text: `\d` is the two characters `\` and `d`, and no escape is decoded. `Raw.toArray(parts)` gives the pieces as an array of strings. `Template.raw` is a tag that joins the pieces and values as written. Its result is the text with no escape decoded, which is what a pattern needs:

```narra
import * as Regexes from "std/regex"
import * as Template from "std/template"

const CLOCK = Regexes.checked(Template.raw`^(\d\d):(\d\d)$`)
const PLAIN = Template.raw`\d\d:\d\d`

export function run(args: Array<string>): int uses io, alloc {
  print(PLAIN)
  print(Template.raw`tab\there`)
  print(Show.show(Regexes.matches(CLOCK, "09:30")))
  return 0
}
```

```text
\d\d:\d\d
tab\there
true
```

`Template.raw` is an `@const` function, so a constant can call it, and `Regexes.checked` can take its result. The compiler computes both while it compiles, and the pattern is written once, with single backslashes.

Next: [18. Tools](18-tools.md).
