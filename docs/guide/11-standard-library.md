# 11. The standard library

The earlier chapters used a few standard modules along the way: `std/int`, `std/string`, `std/array`, `std/option`. This chapter goes through the modules that larger programs, such as command-line tools, are built on. Each section stands on its own.

## Iterators: `std/iterator`

A generator, written `function*`, gives values one at a time, and a `for ... of` loop asks it for the next one only when it needs it. `std/iterator` holds the usual operations over generators:

```narra
import * as Iterators from "std/iterator"
import * as Int from "std/int"

function* naturals(): Generator<int> {
  let n = 0
  while (true) {
    yield n
    n = n + 1
  }
}

export function run(args: Array<string>): int uses io, alloc {
  const odd = Iterators.filter(naturals(), (value: int): bool => value % 2 == 1)
  const squares = Iterators.map(odd, (value: int): int => value * value)
  const firstFour = Iterators.take(squares, 4)
  for (const square of firstFour) {
    print(Int.toString(square))
  }
  const total = Iterators.reduce(
    Iterators.take(naturals(), 5),
    (sum: int, value: int): int => sum + value,
    0,
  )
  print(`sum ${Int.toString(total)}`)
  return 0
}
```

```text
1
9
25
49
sum 10
```

`naturals` never ends, and the program still stops. `filter`, `map` and `take` are lazy: each returns a new generator and does no work until a loop asks it for a value. The loop asks `take` for a value, `take` asks `map`, `map` asks `filter`, and `filter` asks `naturals`, until four squares have come through. Each stage gets a name with `const`, so the program reads from top to bottom.

The lazy operators are `map`, `filter`, `flatMap`, `take`, `drop`, `takeWhile`, `dropWhile`, `enumerate` (pairs of an index and a value), `zip` (pairs from two generators, ending with the shorter one), `concat`, `chunk` and `window`.

The eager operators run the generator to its end, or to the first answer: `reduce`, `toArray`, `forEach`, `some`, `every`, `find` and `count`. `find` returns an `Option`, since no value may match.

`chunk` groups the values into arrays of a given size, and `window` gives every run of that many values in a row. A size below 1 makes no sense, so both return an `Option`, and `None` for such a size:

```narra
import * as Iterators from "std/iterator"
import * as Int from "std/int"
import * as Str from "std/string"
import * as Arrays from "std/array"

function* letters(): Generator<string> {
  yield "a"
  yield "b"
  yield "c"
  yield "d"
  yield "e"
}

export function run(args: Array<string>): int uses io, alloc {
  if (Iterators.chunk(letters(), 2) is Some(pairs)) {
    for (const pair of pairs) {
      print(Str.join(pair, ""))
    }
  }
  if (Iterators.chunk(letters(), 0) is None) {
    print("no chunks of size 0")
  }
  const found = Iterators.find(letters(), (letter: string): bool => letter != "a" && letter != "b")
  print(match (found) { Some(letter) => letter, None => "none" })
  print(Int.toString(Arrays.length(Iterators.toArray(letters()))))
  return 0
}
```

```text
ab
cd
e
no chunks of size 0
c
5
```

### Callbacks that are kept

A lazy operator keeps its callback inside the generator it returns, and calls it later. So its callback is `escaping` (chapter 3), and it cannot capture a `let`:

```narra error
// error: 12:74: a closure cannot capture "limit" because it is declared with let, unless it is passed straight to a function parameter that is not escaping; to share state that changes with an escaping closure, keep it in a Cell from std/cell
import * as Iterators from "std/iterator"

function* digits(): Generator<int> {
  for (const digit of [1, 2, 3, 4, 5]) {
    yield digit
  }
}

export function run(args: Array<string>): int uses io, alloc {
  let limit = 3
  const small = Iterators.filter(digits(), (value: int): bool => value < limit)
  return Iterators.count(small)
}
```

Use a `const`, or a `Cell` for a value that changes. An eager operator calls its callback before it returns, so its callback is not escaping and may capture a `let`.

### The source does not move

An operator takes its generator as a value, like any other argument. A loop inside the operator runs on its own copy, so the generator you passed in is still at its start:

```narra
import * as Iterators from "std/iterator"
import * as Co from "std/coroutine"
import { Step } from "std/coroutine"

function* words(): Generator<string> {
  yield "one"
  yield "two"
  yield "three"
}

export function run(args: Array<string>): int uses io, alloc {
  let source = words()
  for (const word of Iterators.take(source, 2)) {
    print(`taken ${word}`)
  }
  for (const word of source) {
    print(`again ${word}`)
  }
  if (Co.resume(mutable source) is Step.Yielded(word)) {
    print(`by hand ${word}`)
  }
  if (Co.resume(mutable source) is Step.Yielded(word)) {
    print(`by hand ${word}`)
  }
  return 0
}
```

```text
taken one
taken two
again one
again two
again three
by hand one
by hand two
```

The second loop starts from `"one"` again, and runs the generator's work a second time. To take a few values and go on from there, resume the generator yourself with `Co.resume(mutable source)`: it moves `source` itself, and each resume goes on where the last one stopped.

### Effects

The result of a lazy operator has the effects of its source and its callback. Mapping over a generator that prints gives a generator that prints, and its type says `uses io`. A pure source with a pure callback gives a pure generator.

## Command lines: `std/args`

`std/args` reads a command line. The program describes its command line as a `Command` value, and one call of `Args.parse` reads the arguments against it. The same value gives the help text and the error messages.

A `Command` has a name, a summary, its flags, its positionals and what to do with extra arguments. A `Flag` with no `value` is a switch such as `--loud`. A flag with a `value` takes one, and the text of `value` names it in the help. `rest` says what happens to arguments after the declared positionals: `Rest.None`, the default, makes them an error, `Rest.Positionals` collects them, and `Rest.Raw` keeps everything after the declared positionals as written, flags included.

`Args.parse` returns a `Result`. On success, `Args.has`, `Args.value`, `Args.values` and `Args.positional` read what was given. `-h` and `--help` work on every command and come back as the error `ArgsError.Help`, so the program prints the help and stops. Any other error has a message from `Args.formatError`, with a hint when a flag or command name is close to a known one:

```narra
import * as Args from "std/args"
import { Command, Flag } from "std/args"

const GREET = Command {
  name: "greet",
  summary: "Say hello.",
  flags: [
    Flag { name: "loud", short: "l", help: "shout the greeting" },
    Flag { name: "times", short: "n", value: "count", help: "a number to show" },
  ],
  positionals: ["name"],
}

function greet(arguments: Array<string>) uses io, alloc {
  const parsed = Args.parse(GREET, arguments)
  if (parsed is Err(Args.ArgsError.Help(path))) {
    print(Args.help(GREET, path))
    return
  }
  if (parsed is Err(error)) {
    print(`error: ${Args.formatError(error)}`)
    return
  }
  const result = parsed.value
  let greeting = "hello"
  if (Args.has(result, "loud")) {
    greeting = "HELLO"
  }
  if (Args.positional(result, "name") is Some(name)) {
    print(`${greeting}, ${name} (${Show.show(Args.value(result, "times"))})`)
  }
}

export function run(args: Array<string>): int uses io, alloc {
  greet(["greet", "--help"])
  greet(["greet", "Ada"])
  greet(["greet", "--loud", "-n", "3", "Ada"])
  greet(["greet", "--lod", "Ada"])
  return 0
}
```

```text
Usage: greet [flags] <name>

Say hello.

Flags:
  -l, --loud           shout the greeting
  -n, --times <count>  a number to show
  -h, --help           print this help

hello, Ada (None)
HELLO, Ada (Some("3"))
error: unknown flag --lod; did you mean --loud?
```

Long flags take two dashes and short flags one: `--times 3`, `--times=3`, `-n 3` and `-n3` are the same. Short switches can be bundled, so `-ab` is `-a -b`. `--` ends the flags, and a lone `-` is a positional. A command can have subcommands in `commands`: the first positional picks one, and `Args.commandPath` returns the names, such as `["narrac", "build"]`.

## Graphs: `std/arena`

A graph with cycles cannot hold its nodes directly: a value that holds itself would never be freed. Keep the nodes in an `Arena` instead, and let each node name its neighbours with `Handle` values. A handle is a slot number and a generation, and it works as a map key.

```narra
import * as Arenas from "std/arena"
import { Arena, Handle } from "std/arena"

struct Node {
  name: string
  next: Option<Handle<Node>>
}

export function run(args: Array<string>): int uses io, alloc {
  let graph: Arena<Node> = Arenas.create()
  const first = Arenas.insert(mutable graph, Node { name: "first", next: None })
  const second = Arenas.insert(mutable graph, Node { name: "second", next: Some(first) })
  Arenas.update(mutable graph, first, (node) => Node { name: node.name, next: Some(copy(second)) })
  if (Arenas.get(graph, first) is Some(node)) {
    print(`${node.name} is followed by a node`)
  }
  Arenas.remove(mutable graph, second)
  print(`second is there: ${Show.show(Arenas.contains(graph, second))}`)
  return 0
}
```

```text
first is followed by a node
second is there: false
```

`first` and `second` point at each other, and the arena frees both when it goes. Removing a node raises its slot's generation, so a handle that is left over finds nothing: `get` returns `None`, and `set` and `update` return `false`. `clear` does the same for every slot. A handle from another arena is not detected, so keep one arena per kind of node.

## Configuration files: `std/toml`

`std/toml` reads and writes TOML, the format of `narra.toml`. Like `std/json`, it works with the `Value` of `std/shape`: a table is a `Value.Record`, an array a `Value.List`, and numbers, strings and booleans are `Int`, `Float`, `Text` and `Bool`. Dates and times have no `Value` of their own, so they come back as `Text` holding the date as written.

`Toml.decode` reads a document straight into a type that derives `Shape`. The struct's fields are the table's keys, and an array of tables (`[[binary]]`) is an `Array` of structs. A syntax error is placed at a line and a column. `Toml.encode` writes a value back, and `Toml.stringify` writes a `Value`:

```narra
import * as Toml from "std/toml"

struct Binary derives Shape, Show {
  name: string
  entry: string
}

struct Package derives Shape, Show {
  name: string
}

struct Manifest derives Shape, Show {
  package: Package
  binary: Array<Binary>
}

const TEXT = `[package]
name = "hello"

[[binary]]
name = "hello"
entry = "main.run"
`

export function run(args: Array<string>): int uses io, alloc {
  const manifest = Toml.decode<Manifest>(TEXT)
  if (manifest is Ok(value)) {
    print(value.binary[0].entry)
    if (Toml.encode(value) is Ok(text)) {
      print(text)
    }
  }
  const broken = Toml.parse("[package\nname = 1\n")
  if (broken is Err(error)) {
    print(Toml.formatError(error))
  }
  return 0
}
```

```text
main.run
[package]
name = "hello"

[[binary]]
name = "hello"
entry = "main.run"

1:9: expected ]
```

`Toml.parse` returns a `TomlError` with `line`, `column` and `reason` fields, so a tool can report the error in its own format. `Toml.stringify` writes the plain keys of each table before its sub-tables and drops comments, so a tool that rewrites a file changes its layout. It returns an error for `Value.Null`, which TOML cannot write.

## Terminals: `std/terminal`

`std/terminal` adds colors and styles to text, and tells a program whether it is writing to a terminal. The style functions do not look at the terminal. They always add the codes, so a test can compare their output:

```narra
import * as Terminal from "std/terminal"
import * as Str from "std/string"

function show(text: string) uses io, alloc {
  print(Str.replace(text, "\u{1b}", "<ESC>"))
}

export function run(args: Array<string>): int uses io, alloc {
  show(Terminal.red("error"))
  show(Terminal.bold(`${Terminal.dim("note")} and more`))
  print(Terminal.stripStyles(Terminal.green("done")))
  return 0
}
```

```text
<ESC>[31merror<ESC>[39m
<ESC>[1m<ESC>[2mnote<ESC>[22m<ESC>[1m and more<ESC>[22m
done
```

Each function closes with its own reset code. When the text already holds that close, as `dim` inside `bold` does, the outer style is opened again after it, so `and more` stays bold. `Terminal.stripStyles` removes the codes again. `Terminal.style(text, Style { ... })` applies several at once: `bold`, `dim`, `italic`, `underline`, `inverse`, `strikethrough`, and a `foreground` and `background` such as `Color.Red`, `Color.Ansi256(21)` or `Color.Rgb(255, 128, 0)`. Both start as `Color.Default`, which leaves the color as it is.

The caller decides when to style. `Terminal.useColor(Stream.Stderr)` answers for one stream: `false` when `NO_COLOR` is set and not empty, `true` when `FORCE_COLOR` is set to anything but `""` or `"0"`, `false` when `TERM` is `dumb`, and otherwise whether the stream is a terminal. `Terminal.isTerminal(stream)` is that last test alone, and `Terminal.columns(stream)` is `Some(width)` for a terminal and `None` for a pipe or a file.

## Patterns: `std/regex`

`std/regex` finds text that matches a pattern. The syntax is JavaScript's, without back references (`\1`) and lookaround (`(?=...)`): those can take exponential time, and this engine runs in time linear in the text. Flags go inside the pattern: `(?i)` ignores the case of ASCII letters, `(?m)` makes `^` and `$` match at each line, and `(?s)` lets `.` match a line break.

`Regexes.compile` returns the compiled pattern or a `RegexError` with the byte offset of the mistake, and `Regexes.formatError` writes it as `<position>: <reason>`. `find` returns the first `Match`, `findAll` is a generator over every match, and `group` and `namedGroup` read the groups. Offsets are byte offsets, as `Str.indexOf` returns them:

```narra
import * as Regexes from "std/regex"
import * as Int from "std/int"

export function run(args: Array<string>): int uses io, alloc {
  const compiled = Regexes.compile("(?<key>\\w+)=(?<value>\\d+)")
  if (compiled is Err(error)) {
    print(error.reason)
    return 1
  }
  const pairs = compiled.value
  for (const found of Regexes.findAll(pairs, "width=80 height=24 depth=x")) {
    if (Regexes.namedGroup(found, "key") is Some(key)) {
      print(`${key} at ${Int.toString(found.start)}`)
    }
  }
  print(Regexes.replaceAll(pairs, "a=1, b=2", "?"))
  const bad = Regexes.compile("(a)\\1")
  if (bad is Err(error)) {
    print(Regexes.formatError(error))
  }
  return 0
}
```

```text
width at 0
height at 9
?, ?
3: back references are not supported: the engine is linear-time
```

The function that only answers yes or no is `Regexes.matches`, since `test` is a keyword in Narra. `replace` and `replaceAll` take the replacement as plain text: `$1` is not expanded. To build each replacement from its match, pass a function to `replaceWith`. `split` cuts a string at every match.

## Text as people see it: `std/unicode`

A `string` is UTF-8 bytes, and `Str.length` counts bytes. What a reader sees as one character can be several code points: an accent after its letter, a flag made of two regional indicators, a family emoji joined with zero width joiners. Such a group is a grapheme cluster. `std/unicode` splits text into clusters, maps case, and counts terminal columns:

```narra
import * as Unicode from "std/unicode"
import * as Int from "std/int"
import * as Str from "std/string"

export function run(args: Array<string>): int uses io, alloc {
  const word = "cafe\u{301}"
  print(`${Int.toString(Str.length(word))} bytes, ${Int.toString(Unicode.graphemeCount(word))} clusters`)
  for (const cluster of Unicode.graphemes("a\u{1f1ef}\u{1f1f5}b")) {
    print(`[${cluster}]`)
  }
  print(Unicode.toUpperCase("stra\u{df}e"))
  print(Unicode.toLowerCase("\u{3a3}\u{391}\u{3a3}"))
  print(Show.show(Unicode.foldCase("Stra\u{df}e") == Unicode.foldCase("STRASSE")))
  print(Int.toString(Unicode.displayWidth("ab\u{4e2d}\u{6587}")))
  return 0
}
```

```text
6 bytes, 4 clusters
[a]
[🇯🇵]
[b]
STRASSE
σας
true
6
```

Each cluster from `Unicode.graphemes` is a slice of the text, so the loop copies nothing. `toUpperCase` and `toLowerCase` use the full mappings, where one letter can become two (`ß` is `SS`), and a capital sigma at the end of a word lowers to `ς`. They take no locale, like `toUpperCase()` in JavaScript. To compare two texts without case, compare their `foldCase` results.

`displayWidth` counts the columns a terminal gives the text: wide and fullwidth characters, such as Chinese, Japanese and Korean, and emoji take 2, combining marks and zero width characters take none, and the rest take 1. It works per cluster, so a flag or a family emoji is 2 columns. A tool can pad a table column with it where `Str.length` would count bytes. It counts the bytes of escape codes too, so pass styled text through `Terminal.stripStyles` first: `Unicode.displayWidth(Terminal.stripStyles(text))`.

A few edges follow from those rules. Characters of ambiguous East Asian width, such as `±` and Greek letters, are 1 column, as in a terminal set to a Western locale. Control characters, the tab included, are 0: a terminal moves the cursor for them, and the distance depends on where it stands. U+FE0F after an emoji that is text by default, such as `❤`, makes it 2, but U+FE0E after an emoji that is wide by default, such as `😀`, leaves it 2. The Hangul filler U+115F is 0, since it is default-ignorable.

### Normalization

One text can be written with different code points: `é` is one code point, U+00E9, or an `e` followed by the combining accent U+0301. The two look the same and are different strings, so `==` says they differ. `Unicode.normalize` writes text in one of the four normal forms of Unicode, and two texts in the same form are equal when a reader would call them equal:

```narra
import * as Unicode from "std/unicode"
import { NormalForm } from "std/unicode"
import * as Int from "std/int"
import * as Str from "std/string"

export function run(args: Array<string>): int uses io, alloc {
  const typed = "caf\u{e9}"
  const pasted = "cafe\u{301}"
  print(Show.show(typed == pasted))
  print(Show.show(Unicode.normalize(typed, NormalForm.NFC) == Unicode.normalize(pasted, NormalForm.NFC)))
  print(Int.toString(Str.length(Unicode.normalize(typed, NormalForm.NFD))))
  print(Unicode.normalize("\u{fb01}le \u{2460}", NormalForm.NFKC))
  print(Show.show(Unicode.isNormalized(pasted, NormalForm.NFC)))
  return 0
}
```

```text
false
true
6
file 1
false
```

NFC composes: a letter and its accents become one code point where Unicode has one. NFD decomposes: every accent becomes a code point of its own, after its letter. NFKC and NFKD do the same and also replace compatibility characters with plain ones: the ligature `ﬁ` becomes `fi`, and the circled digit `①` becomes `1`. The K forms lose formatting, so they suit search and identifiers, not text to show. NFC is the form most text on the web is in already.

`isNormalized` tells whether `normalize` would change the text. Both look up each code point first and stop early, so text already in the form, ASCII text among it, costs one pass and no allocation: `normalize` gives back the same string. Hangul syllables are composed and decomposed by arithmetic, as Unicode defines them, not from the tables.

Next: [12. Narra for TypeScript programmers](12-narra-for-typescript-programmers.md), optional reading for TypeScript programmers.
