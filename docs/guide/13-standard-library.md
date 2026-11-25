# 13. The standard library

The earlier chapters used a few standard modules along the way: `std/int`, `std/string`, `std/array`, `std/option`. This chapter goes through the modules that larger programs, such as command-line tools, are built on. Each section stands on its own.

## Iterators: `std/iterator`

`std/iterator` holds the usual operations over generators, the `function*` values of [chapter 12](12-generators.md):

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

The second loop starts from `"one"` again, and runs the generator's work a second time. To take a few values and go on from there, resume the generator yourself with `Co.resume(mutable source)` ([chapter 12](12-generators.md#resume-a-generator-by-hand)). It moves `source` itself, and each resume goes on where the last one stopped.

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

## Distinct values: `std/set`

A `Set` holds each value once and remembers the order in which the values were added. The items need the same trait as map keys, `Hashable`. A set is a value like an array: a copy is independent of the original.

```narra
import * as Sets from "std/set"
import { Set } from "std/set"

export function run(args: Array<string>): int uses io, alloc {
  let mine: Set<string> = Sets.create()
  Sets.add(mutable mine, "ada")
  Sets.add(mutable mine, "grace")
  const added = Sets.add(mutable mine, "ada")
  print(`added again: ${Show.show(added)}, size: ${Show.show(Sets.size(mine))}`)

  let yours: Set<string> = Sets.create()
  Sets.add(mutable yours, "grace")
  Sets.add(mutable yours, "linus")

  for (const name of Sets.values(Sets.union(mine, yours))) {
    print(`either: ${name}`)
  }
  for (const name of Sets.values(Sets.intersection(mine, yours))) {
    print(`both: ${name}`)
  }
  for (const name of Sets.values(Sets.difference(mine, yours))) {
    print(`only mine: ${name}`)
  }
  const removed = Sets.remove(mutable mine, "ada")
  print(`removed: ${Show.show(removed)}, has ada: ${Show.show(Sets.has(mine, "ada"))}`)
  return 0
}
```

```text
added again: false, size: 2
either: ada
either: grace
either: linus
both: grace
only mine: ada
removed: true, has ada: false
```

`add` and `remove` return whether the set changed. `values` returns an array, as `Maps.keys` does, so it works in a `for ... of` loop. `union`, `intersection` and `difference` return new sets and leave their arguments alone.

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

`Toml.decode` reads a document straight into a type that derives `Shape`, the trait that [chapter 7](07-traits.md#saving-and-loading-values) explains. The struct's fields are the table's keys, and an array of tables (`[[binary]]`) is an `Array` of structs. A syntax error is placed at a line and a column. `Toml.encode` writes a value back, and `Toml.stringify` writes a `Value`:

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

`std/regex` finds text that matches a pattern. The syntax is the common one of `\d`, `[a-z]`, `(...)`, `*`, `+` and `?`, without back references (`\1`) and lookaround (`(?=...)`): those can take exponential time, and this engine runs in time linear in the text. Flags go inside the pattern: `(?i)` ignores the case of ASCII letters, `(?m)` makes `^` and `$` match at each line, and `(?s)` lets `.` match a line break.

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

Each cluster from `Unicode.graphemes` is a slice of the text, so the loop copies nothing. `toUpperCase` and `toLowerCase` use the full mappings, where one letter can become two (`ß` is `SS`), and a capital sigma at the end of a word lowers to `ς`. They take no locale: the result does not depend on the language of the user. To compare two texts without case, compare their `foldCase` results.

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

## Addresses: `std/url`

`std/url` takes a URL apart, puts it back together, and resolves a relative reference against a base. `Url.parse` returns a `Url` whose parts stay as written: percent escapes are not decoded, and a part that is missing is `None`, so `a` and `a?` stay different. The port is text, because `http://host:` and `http://host:0080/` must come back unchanged. Read it with `Int.parse(port, 10)` when you need a number. `parse` gives an `Err` with a reason for a space, a bad percent escape, a port above 65535 and similar mistakes.

```narra
import * as Url from "std/url"
import { QueryParam } from "std/url"

export function run(args: Array<string>): int uses io, alloc {
  const parsed = Url.parse("https://ada@example.com:8080/docs/intro?lang=en&tag=a%20b#top")
  if (parsed is Err(message)) {
    print(message)
    return 1
  }
  const address = parsed.value
  print(Show.show(address.scheme))
  print(Show.show(address.host))
  print(Show.show(address.port))
  print(address.path)
  if (address.query is Some(text)) {
    const params = Url.parseQuery(text)
    if (params is Ok(items)) {
      for (const item of items) {
        print(`${item.name} = ${item.value}`)
      }
    }
  }
  if (Url.parse("../guide?page=2") is Ok(reference)) {
    if (Url.resolve(address, reference) is Ok(joined)) {
      print(Url.toString(joined))
    }
  }
  print(Url.encode("a b&c/d"))
  print(Show.show(Url.decode("a%20b")))
  print(Show.show(Url.decode("100%")))
  print(Url.queryToString([QueryParam { name: "q", value: "hello world" }, QueryParam { name: "n", value: "1" }]))
  return 0
}
```

```text
Some("https")
Some("example.com")
Some("8080")
/docs/intro
lang = en
tag = a b
https://ada@example.com:8080/guide?page=2
a%20b%26c%2Fd
Ok("a b")
Err("a bad percent escape")
q=hello+world&n=1
```

`Url.resolve(base, reference)` joins a relative reference to an absolute base, with the rules of RFC 3986, so `..` and `.` segments are removed. It gives an `Err` when the base has no scheme. `Url.toString` writes a `Url` back as text.

`Url.encode` escapes every byte that is not a letter, a digit or one of `-._~`, and `Url.decode` reverses it. `decode` leaves `+` as it is, and it gives an `Err` for a bad escape or a result that is not UTF-8. `Url.parseQuery` reads the text after `?` as `name=value` pairs, in order, with `+` as a space, and repeated names stay as separate entries. `Url.queryToString` writes them again. The host rules of a browser, such as lower-casing or default ports, are not applied.

## Versions: `std/semver`

`std/semver` reads, compares and changes version numbers of the form `MAJOR.MINOR.PATCH`, with an optional prerelease after `-` and build metadata after `+`, as Semantic Versioning 2.0.0 defines them. `Semver.parse` returns `None` for text that is not a version. A `Version` has the three numbers and two arrays of identifiers, `prerelease` and `build`.

```narra
import * as Semver from "std/semver"
import { Part } from "std/semver"

export function run(args: Array<string>): int uses io, alloc {
  const parsed = Semver.parse("1.4.2-rc.1+build.7")
  if (parsed is Some(version)) {
    print(Semver.toString(version))
    print(Show.show(version.prerelease))
    if (Semver.parse("1.4.2") is Some(release)) {
      print(Show.show(Semver.compare(version, release)))
    }
    print(Semver.toString(Semver.bump(version, Part.Minor)))
    print(Show.show(Semver.satisfies(version, "1.4.0")))
    print(Show.show(Semver.satisfies(version, "1.4.2")))
  }
  print(Show.show(Semver.parse("1.4")))
  return 0
}
```

```text
1.4.2-rc.1+build.7
["rc", "1"]
-1
1.5.0
true
false
None
```

`Semver.compare` returns -1, 0 or 1 and follows the precedence rules: a prerelease is below its release, and build metadata is ignored. `Semver.bump(version, part)` raises the major, minor or patch number, sets the lower numbers to 0 and drops the prerelease and the build. `Semver.satisfies(version, requirement)` is true when the version is at least the minimum named by `requirement`, which must be written as three numbers such as `1.4.0`. Any other form of requirement, such as `^1.4.0` or `1.4`, gives `false`.

## Hashes: `std/digest`

`std/digest` computes SHA-256 and SHA-1 hashes, HMAC, and PBKDF2 with HMAC-SHA-256, all in Narra without a C library. Every input and result is `bytes`. Print a result as text with `Bytes.toHex` (see [chapter 11](11-files-and-the-system.md)). SHA-1 is only there for formats that still use it. Choose SHA-256 for new work.

```narra
import * as Bytes from "std/bytes"
import * as Digest from "std/digest"
import { Algorithm } from "std/digest"

export function run(args: Array<string>): int uses io, alloc {
  const message = Bytes.fromString("abc")
  print(Bytes.toHex(Digest.sha256(message)))
  print(Bytes.toHex(Digest.sha1(message)))
  let hasher = Digest.create(Algorithm.Sha256)
  Digest.update(mutable hasher, Bytes.fromString("a"))
  Digest.update(mutable hasher, Bytes.fromString("bc"))
  print(Show.show(Digest.equal(Digest.finish(hasher), Digest.sha256(message))))
  print(Bytes.toHex(Digest.hmac(Algorithm.Sha256, Bytes.fromString("key"), message)))
  if (Digest.pbkdf2Sha256(Bytes.fromString("password"), Bytes.fromString("salt"), 1, 16) is Some(derived)) {
    print(Bytes.toHex(derived))
  }
  return 0
}
```

```text
ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
a9993e364706816aba3e25717850c26c9cd0d89d
true
9c196e32dc0175f86f4b1cb89289d6619de6bee699e4c378e68309ed97a1a6ab
120fb6cffcf8b32c43e7225256c4f837
```

`Digest.sha256` and `Digest.sha1` hash a whole value. For data that arrives in pieces, `Digest.create(algorithm)` makes a `Hasher`, `Digest.update` feeds it, and `Digest.finish` returns the hash. `finish` leaves the hasher as it was, so it can take more input afterward, and `copy(hasher)` makes an independent one. `Digest.hmac(algorithm, key, message)` proves that someone who holds `key` wrote `message`.

`Digest.pbkdf2Sha256(password, salt, iterations, length)` derives a key of `length` bytes from a password. It returns `None` when `iterations` is below 1 or `length` is negative or too large. Each iteration allocates, so a large count is slow: 80000 iterations take about a second. To compare two hashes or two message codes, use `Digest.equal`. It looks at every byte before it answers, so the time it takes does not tell an attacker where the first difference is.

## Log lines: `std/log`

`std/log` writes one line per event to standard error, as text or as JSON. `Log.create(level, format)` makes a `Logger`. The levels are `Debug`, `Info`, `Warn` and `Error`, and a logger drops lines below its level. A line has a message and fields, each field a `#[name, value]` pair whose value is a `Value` from `std/shape`. `Log.with(logger, fields)` returns a logger that adds those fields to every line, which suits a request or a task. No timestamp is written, so the output is the same on every run.

```narra
import * as Log from "std/log"
import { Level, Format } from "std/log"
import { Value } from "std/shape"

export function run(args: Array<string>): int uses io, alloc {
  const logger = Log.create(Level.Info, Format.Text)
  const request = Log.with(logger, [#["request", Value.Int(42)]])
  print(Log.render(request, Level.Info, "started", [#["path", Value.Text("/a b")]]))
  print(Log.render(Log.create(Level.Info, Format.Json), Level.Warn, "slow", [#["seconds", Value.Float(1.5)]]))
  print(Show.show(Log.enabled(logger, Level.Debug)))
  return 0
}
```

```text
level=info msg="started" request=42 path="/a b"
{"level":"warn","msg":"slow","seconds":1.5}
false
```

`Log.render` returns the line as a string, here to print it. The functions `Log.debug`, `Log.info`, `Log.warn` and `Log.error` write the line to standard error when the level allows it, and `Log.write` takes the level as an argument. `Log.enabled(logger, level)` tells whether a line at that level would be written, so a program can skip building an expensive field. In the text format a value is bare unless it is empty or holds a space, a quote, an `=` or a control character, and then it is quoted as JSON. The message is always quoted. A field whose name repeats an earlier one, or `level` or `msg`, replaces that value and keeps its place.

Next: [14. Async](14-async.md).
