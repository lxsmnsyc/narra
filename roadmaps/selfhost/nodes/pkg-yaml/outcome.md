# Outcome: pkg-yaml

## Built

`packages/yaml`, a YAML 1.2 reader and writer over `std/shape`, a workspace member (root `narra.toml` lists it) with name `yaml`, version 0.1.0 and a README at `packages/yaml/README.md`.

- Public module `yaml` (`[exports] "yaml" = "src/yaml.nra"`): `YamlError { line, column, reason }`, `formatError`, `ReadOptions { maxDepth (512), maxAliasNodes (100000), mergeKeys (true) }`, `defaultOptions`, `parse`, `parseAll`, `parseWith`, `stringify`, `decode<T: Shape>`, `encode<T: Shape>`, as approved in the brief. Inside the package the same file is `yaml/yaml`.
- Private modules: `scalar` (core schema), `writer`, and the test modules `checks` (entry `run`), `checkkit`, `checksuite`, `checkreader`, `checkround`.
- Reads block and flow collections, plain, single- and double-quoted scalars, `|` and `>` block scalars with chomping and indentation indicators, comments, `%YAML` and `%TAG` directives, `---` and `...`, anchors, aliases, merge keys, and the core tags. Any other tag, a collection key and a repeated key are errors. `maxAliasNodes` caps the nodes that aliases copy in total; an alias to a node that contains it is an error.
- Writes block style with `|` blocks for strings with line breaks, `[]` and `{}`, quotes only where needed.
- `examples/yaml_config` (`narra.toml`, `config.yaml`, `src/main.nra`) reads a YAML file into a `derives Shape` type and writes it back. `scripts/test.sh` builds it as an outside program.
- yaml-test-suite vendored under `packages/yaml/suite/`: release `data-2022-01-17` of https://github.com/yaml/yaml-test-suite (`data` branch commit `6e6c296ae9c9d2d5c4134b4b64d01b29ac19ff6f`), MIT license in `suite/LICENSE`, 402 cases (`===`, `in.yaml`, `in.json`, `error` kept). No network needed.

## Test suite counts

402 cases: 369 passed, 33 skipped, 0 failed. A skipped case that passes fails the check, so `suite/skips.txt` is exact. Skipped cases:

- Complex keys (the keys of a `Record` are text; 15): 4FJ6, 6BFJ, 6PBE, 9MMW, KK5P, LX3P, M2N8/00, M2N8/01, M5DY, Q9WF, RZP5, SBG9, V9D5, X38W, XW4D.
- Tag outside the core schema (16): 2XXW (`!!set`), 565N (`!!binary`), 5TYM, 6CK3, 6WLZ, 7FWL, 9WXW, C4HZ, CC74, CUP7, J7PZ (`!!omap`), M5C3, P76L, UGM3, Z67P, Z9M4 (local tags, `%TAG` handles, other namespaces).
- 2JQS: two empty keys read as the same text key, so the second is a repeated key.
- L24T/01: the expected text ends in a line break that the input lacks (last line of a literal block is spaces only and the input ends right after it); the reference adds the break.

Other checks (5505 in all, 0 failed): scalars, collections, strings, anchors, merge, documents, errors; alias bomb, `maxAliasNodes` and `maxDepth` options, 512/513 nested flow, deep block sequence and mapping, 60000-item inputs; writer exact output; round trip of 3000 random `Value` trees plus awkward strings as values, list items and keys; fuzz of 3000 random and 3000 mutated texts that must not panic and must round trip whatever they read; one `derives Shape` struct decoded from equal YAML, JSON and TOML texts.

## Deviations

- `examples/yaml_config/**` is outside the owned paths (the brief lists `narra.toml` only as shared). The acceptance criteria require the example, so I added it; the `examples/*` glob already covers it in the workspace.
- `parse` of an empty or comment-only text gives `Null` (zero documents), where the brief says "exactly one document". More than one document is an error.
- Module path inside the package is `yaml/yaml` (a file `src/yaml.nra`), so the test modules import that; outside the package it is `yaml`.
- An empty key reads as `""`. Every scalar key, implicit, explicit (`? 0x10`) or alias, keeps its source text (`null: a` is key `"null"`; an alias key has the text of its anchored scalar).
- Flow entries ignore a tag on the `<<` key (a tagged `<<` still merges).
- `maxDepth` is clamped to a ceiling of 512, which is also the default. Measured under `-sanitize`: flow collections overflow the stack at about 640 to 700 levels, block sequences and mappings are fine past 1000. A larger option value is lowered to 512 and a check feeds 100000 levels with `maxDepth: 1000000` and expects a clean error.
- Review fixes: a string starting with a byte order mark is quoted by the writer; foldLines and allDigits use for-of loops.

## Stubs left

- None.

## Gotchas

- A struct literal cannot be qualified (`Yaml.ReadOptions { ... }`); import the type by name. A module alias and a type cannot share a name (`Random` module and type: import the module as `Randoms`).
- Under `-sanitize` frames are large: keep the functions that recurse on nesting small. Big non-recursive logic belongs in helpers that return before the reader goes down (see `readNodeHead`). 512 levels of block mapping overflowed before the split.
- `Float.parse` rejects `inf` and `nan`; build them with `1.0e308 * 10.0` and `inf - inf`.
- `Value` is not `Equatable`; the checks use `sameValue` in `checkkit`.
- Result values that share a parameter need `from <param>`; copy an anchored value (`copy`) when it is handed out of the parser.

## Proposed learnings

- Keep recursive reader functions small and move non-recursive work into helpers | sanitizer frames are several times larger and overflowed the 8 MB stack at the default depth limit | packages | no

## Verification

- `out/narrac run -sanitize packages/yaml`: `yaml-test-suite: 369 passed, 33 skipped, 0 failed` and `5505 checks, 0 failed`.
- `out/narrac fmt --check .`: clean.
- `out/narrac run examples/yaml_config`: prints the decoded config and its YAML.
- `sh scripts/test.sh`: final line `test.sh: PASS, 11028 cases in 336s`; the package step printed `5505 checks, 0 failed` for yaml and built `examples/yaml_config`. After the review fixes (only `packages/yaml` and the outcome changed) the package checks with `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1` print `5516 checks, 0 failed`, and `narrac fmt --check .` is clean.

## Size

- Estimated L, actual L.
- Risk estimated medium, measured medium: the reader is hand written over a large grammar; the suite, round trip and fuzz cover it, and stack use needed a restructure.
