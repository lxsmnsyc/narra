# yaml

A YAML 1.2 reader and writer for Narra. It reads YAML text into the `Value` of `std/shape`, the same tree that `std/json` and `std/toml` use, so one `derives Shape` type decodes from YAML, JSON or TOML. It has no other dependency than `std`.

## Using the package

Add this line to the `[dependencies]` table of your `narra.toml`:

```toml
[dependencies]
yaml = { git = "https://github.com/lxsmnsyc/narra", version = "0.1.0" }
```

It needs narrac 0.2.0 or later. The next `narrac build` downloads the package and records it in `narra.lock`. Then import the one module:

```narra fragment
import * as Yaml from "yaml"
```

The package exposes the module `yaml`. Its scalar rules and its writer are private.

## Reading

`parse` reads one document into a `Value`. A mapping is a `Record` that keeps the order of the text, a sequence is a `List`, and a scalar is a `Null`, `Bool`, `Int`, `Float` or `Text`.

```narra fragment
import * as Yaml from "yaml"

export function run(_args: Array<string>): int uses io, alloc {
  const text = "name: demo\nports: [80, 443]\ndebug: false\n"
  const value = Yaml.parse(text)
  if (value is Err(error)) {
    print(Yaml.formatError(error))
    return 1
  }
  print(Yaml.stringify(value.value))
  return 0
}
```

The functions are:

- `parse(text)` reads exactly one document. An empty text, or a text with only comments, is `Null`. Two documents are an error.
- `parseAll(text)` reads every document of a stream, such as texts separated by `---`.
- `parseWith(text, options)` is `parse` with the options below.
- `stringify(value)` writes a value as block-style YAML.
- `decode<T>(text)` reads a type that implements `Shape`, and gives `Result<T, string>`.
- `encode<T>(value)` writes any type that implements `Shape`.
- `formatError(error)` writes a `YamlError` as `line:column: reason`. Line and column start at 1.
- `defaultOptions()` gives the options that `parse` uses.

An error is a `YamlError` with the `line`, the `column` and the `reason`. The reader never panics on bad input.

### What the reader accepts

- Block mappings and sequences, flow mappings and sequences, and comments.
- Plain, single-quoted and double-quoted scalars, with all the escapes of double quotes and the folding of scalars that span lines.
- Block scalars with `|` and `>`, the chomping indicators `-` and `+`, and an indentation indicator.
- Documents with `---` and `...`, a `%YAML` directive, and a byte order mark at the start.
- Anchors and aliases, and the `<<` merge key.
- Tags of the core schema.

### Scalars

A plain scalar follows the core schema:

| Text | Value |
| --- | --- |
| `null`, `Null`, `NULL`, `~`, or nothing | `Null` |
| `true`, `True`, `TRUE`, `false`, `False`, `FALSE` | `Bool` |
| `42`, `-7`, `+5`, `0o17`, `0x1F` | `Int` |
| `1.5`, `-2e3`, `.5`, `5.`, `.inf`, `-.inf`, `.nan` | `Float` |
| anything else | `Text` |

A decimal number too large for an `Int` is a `Float`. A quoted scalar and a block scalar are always `Text`.

The tags `!!str`, `!!int`, `!!float`, `!!bool`, `!!null`, `!!seq` and `!!map` set the type, and the text must fit it: `!!int "7"` is `Int(7)`, and `!!int abc` is an error. `!` alone makes a scalar `Text`. Any other tag is an error, including `!!set`, `!!binary`, a local tag such as `!point`, and tags that a `%TAG` directive defines.

### Keys

The keys of a `Record` are text. A scalar key becomes its text as written, so `1: a` has the key `"1"`, `? 0x10` has the key `"0x10"`, an alias key has the text of its anchored scalar, and an empty key is `""`. A key that is a sequence or a mapping is an error. A repeated key is an error. Keys keep the order in which they appear.

### Anchors, aliases and merge keys

An alias gives a copy of the node that its anchor marks. An anchor belongs to its document and can be defined again later. An alias to an anchor that is not defined, or to a node that contains the alias, is an error.

`<<: *a` copies the fields of a mapping in, and `<<: [*a, *b]` copies several. A mapping's own keys win over merged ones, and an earlier entry of a list wins over a later one. A merge value that is not a mapping or a list of mappings is an error. A quoted `"<<"` is an ordinary key. Merge keys are a YAML 1.1 feature, on by default; set `mergeKeys` to `false` and `<<` becomes an ordinary key.

### Options

`ReadOptions` has three fields:

- `maxDepth` (default 512) is the deepest nesting of sequences and mappings. A deeper text is an error. 512 is also the ceiling: a larger value is lowered to 512, because the reader recurses on nesting and a deeper limit could overflow the stack. Under the sanitizer, 512 levels of flow collections fit with a margin (they overflow at about 640 to 700), and block collections fit past 1000.
- `maxAliasNodes` (default 100000) is the most nodes that aliases may copy in total. A file that expands exponentially, such as nested aliases of aliases, fails with an error instead of using all the memory.
- `mergeKeys` (default `true`) turns merge keys on or off.

## Writing

`stringify` writes block style, with the fields of a mapping in order and sequences indented under their key. An empty list is `[]` and an empty record is `{}`. A string is written plain unless it would read back as another type or break the syntax, such as `"true"`, `"123"`, `""`, a string with leading or trailing spaces, or one that holds `: ` or ` #`. Those are double-quoted. A string with line breaks is a `|` block. Floats keep a `.0` when they are whole, and `.inf`, `-.inf` and `.nan` are written by name. The writer makes no anchors.

`parse(stringify(value))` gives `value` back.

## Decoding a type

`decode` and `encode` work through `Shape`, as `Json.decode` and `Toml.decode` do:

```narra fragment
import * as Yaml from "yaml"

struct Server derives Shape {
  host: string
  port: int
}

export function run(_args: Array<string>): int uses io, alloc {
  const server = Yaml.decode<Server>("host: 127.0.0.1\nport: 8080\n")
  if (server is Err(message)) {
    print(message)
    return 1
  }
  print(Yaml.encode(server.value))
  return 0
}
```

A read error comes back as `line:column: reason`, and a type error as `<path>: <reason>`, for example `.port: expected int, found Text`.

## Limits

- The reader takes YAML 1.2 with the core schema. It does not take the other schemas, so `yes` and `no` are text, and timestamps are text.
- Tags outside the core schema are an error, so `!!set`, `!!omap` and `!!binary` do not read.
- A mapping key must be a scalar. YAML allows a sequence or a mapping as a key; a `Value` cannot hold one.
- Two empty keys in one mapping are the same key and so a repeated key.
- A document is read whole. There is no streaming of events or of a large file.
- Comments are dropped, and the writer does not keep the layout of the text it was read from.
- Only UTF-8 text is read.

## Tests

`narrac run -sanitize packages/yaml` runs the checks from the repository root, and `scripts/test.sh` runs them as part of a full run. They hold:

- The [yaml-test-suite](https://github.com/yaml/yaml-test-suite), vendored under `suite/` (see below), each case checked against the expected values or the expected error.
- Checks of scalars, collections, strings, anchors, merge keys, documents, errors and limits, including an alias bomb, deep nesting and long inputs.
- A round trip of random values through `stringify` and `parse`, and a fuzz of random and mutated text that must not panic and must round trip whatever it reads.
- A check that one `derives Shape` type decodes from the same data in YAML, JSON and TOML.

### The vendored test suite

`suite/` holds the data of the yaml-test-suite, release `data-2022-01-17` (commit `6e6c296ae9c9d2d5c4134b4b64d01b29ac19ff6f` of the `data` branch), under the MIT license in `suite/LICENSE`. Each case is a folder with `in.yaml`, the name in `===`, and `in.json` or `error`. The cases that the reader leaves out are in `suite/skips.txt`, one per line with its reason: a mapping key that is a collection, a tag outside the core schema, and one case whose expected text depends on a quirk of the reference implementation. A listed case that starts to pass fails the checks, so the list stays true.
