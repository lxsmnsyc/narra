# Outcome: std-toml

## Built

- `std/toml` (alias `Toml`, D118 item 5), Narra source in `std/src/toml.nra`, registry row in `selfhost/src/stdtoml.nra`, one `registerAll` line in `selfhost/src/stdlib.nra`, regenerated `selfhost/src/stdtext.nra`.
- Public names: `struct TomlError { line: int, column: int, reason: string }` (derives `Show`, `Equatable`); `parse(text): Result<Value, TomlError>`, `stringify(value): Result<string, string>`, `decode<T: Shape>(text): Result<T, string>`, `encode<T: Shape>(value): Result<string, string>`, `formatError(error): string` (`<line>:<column>: <reason>`).
- Parser: bare, quoted and dotted keys; basic, literal and both multi-line strings with every TOML 1.0 escape (`\uXXXX`, `\UXXXXXXXX` checked as scalar values, line-ending backslash); decimal, `0x`, `0o`, `0b` integers with underscores and a 64-bit range check; floats with exponents, `inf` and `nan`; booleans; offset and local date-times, dates and times, checked (leap years, seconds up to 60) and kept as `Value.Text` of their text as written; arrays (mixed types, newlines, comments, trailing comma); inline tables; `[table]` and `[[array]]` headers; the TOML 1.0 redefinition rules; nesting deeper than 512 is an error.
- `stringify`: plain keys before sub-tables, `[[name]]` for a non-empty `List` of only `Record`s, inline tables for other nested `Record`s, quoted keys when not bare, no header for a table with only sub-tables, `-0.0` for negative zero, and `Err` (`<path>: nesting is deeper than 512`) for tables and arrays nested deeper than `parse` accepts. `Err` for `Null` (`<D84 path>: TOML has no null`) and for a non-`Record` root (`the root must be a Record`).
- Samples: `testdata/toml.nra` (every value kind, `Json.stringify` of it, `stringify`, re-parse, same JSON text, plus the two `stringify` errors, the negative-zero round trip, and 512 and 600 levels of nesting through `stringify` and `parse`), `testdata/toml_errors.nra` (34 invalid documents, most from toml-test `tests/invalid`, named per case), `testdata/toml_manifest.nra` (the `selfhost/narra.toml` form with `[package]`, `[link]`, `[[binary]]` through `Toml.decode` and derived `Shape`, then `Toml.encode`, a D84 path error and a syntax error). Error test `testdata/errors/toml_decode_no_shape.nra`.
- SPEC section 3: a `std/toml` bullet (the D118 "not built yet" line removed). `compiler/README.md`: one line. Guide: "Configuration files: `std/toml`" section in `docs/guide/12-standard-library.md` with a tested program, and the README line.

## Proof

- base: 7569f176 (feat/batch-14), `narrac run testdata/toml.nra` with a release build of the base: `toml.nra:3:1: unknown module "std/toml" ...`.
- branch: the same command prints the document as JSON, its TOML text and `round trip ok`.

## Deviations

- None in owned paths. No diagnostic code used (E1340-E1349 stay reserved; E1340 is for manifest-toml).
- Dates keep the text as written, including a space separator or a lowercase `t`/`z` (D118 says "its RFC 3339 text"; RFC 3339 allows the space). No normalization.
- A table created as the parent of a header (`[a.b.c]` makes `a` and `b`) cannot take dotted keys later (`[a]` then `b.x = 1` is an error), as toml-test `append-with-dotted-keys` expects.

## Stubs left

- None. narrac does not use the module yet; manifest-toml moves `manifest.nra` after the batch seed update.

## Gotchas

- A parse error in a std source is a panic of narrac (`std source std/toml: expected an expression, found "return"`): `return` is not an expression in a match arm; use a block.
- To test a std source change without a 9-minute compiler rebuild, copy it into a scratch package as a local module, with `std/internal/float` replaced by a stub (only std modules may import `std/internal/`). That is also how the seed check was run.
- Memory: `addNode` makes a `Node` (a `Map` and an `Array`) for every value, about 80 bytes of parse tree per input byte (240 MB for a 2.8 MB document; review note, left as is). Fine for manifests; a large-file reader would build `Value`s directly for leaves.
- `stringify` reorders: plain keys of a table come before its sub-tables, so a `Record` with a sub-table between plain keys reads back in another order.

## Proposed learnings

- None.

## Verification

- `NARRAC=<branch narrac> sh scripts/test.sh --filter toml`: PASS, 28 cases.
- Seed: `narrac0` from `seed/narrac.c` runs the three samples with `toml.nra` as a local module (floats stubbed); all exit 0.
- Emit: `narrac emit selfhost` of the base tree with the base and branch compilers is byte-identical (223,942 lines).
- `sh scripts/test.sh`: `test.sh: PASS, 8636 cases in 376s`.

## Size

- Estimated M, actual M. About 1,460 lines after `narrac fmt` (the brief said about 900): the formatter splits long calls, and the date, number and string checks are written out byte by byte.
