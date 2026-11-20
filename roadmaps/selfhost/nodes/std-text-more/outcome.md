# Outcome: std-text-more

## Built

- `std/float` (`std/src/ext__float.nra`): `parse(text: string): Result<f64, string> uses alloc`. Like `Int.parse` it returns a `Result` with a message. Grammar: optional sign, digits with an optional point (at least one digit in all, so `.5` and `5.` pass), optional exponent. Spaces, `inf`, `nan` and hex are errors. A number that reads as infinity is an error. It checks the text, then calls `FloatText.parse` (`strtod`).
- `std/float`: `toFixed(value: f64, digits: int): Option<string> uses alloc`, a registry row in `selfhost/src/std/stdfloat.nra` over `narra_float_to_fixed` in `selfhost/runtime/runtime.h`. It rounds as C's `%.*f` does, on the exact binary value: `1.005` with 2 gives `1.00`, `2.5` with 0 gives `2`, `3.5` gives `4`, `-0.001` with 2 gives `-0.00`. `None` for a digit count outside 0 to 20. NaN and the infinities print as `Float.toString` does.
- `std/char` (`std/src/ext__char.nra`, new): `isDigit`, `isLetter`, `isAlphanumeric`, `isUpper`, `isLower`, `isWhitespace`. ASCII only. `isWhitespace` is the six characters `Str.trim` removes.
- `std/string` (`std/src/ext__string.nra`): `lastIndexOf(text, needle): Option<int>`, `trimStart`, `trimEnd` (same white space as `Str.trim`), `padStart(text, width, fill: char)`, `padEnd(text, width, fill: char)`. Width counts bytes, as `Str.length` does. Padding is whole copies of `fill` only, so a multi-byte fill never splits and the result may be shorter than `width`. A text at or past `width` comes back unchanged.
- Samples with expected output: `testdata/std_text_float.nra`, `std_text_char.nra`, `std_text_string.nra` (parse and toFixed edge cases, every class on boundary characters and non-ASCII, trim, pad, lastIndexOf with empty needle, empty text and multi-byte text).
- Guide chapter 02: three runnable samples (trim, pad and `lastIndexOf` with the char classes; fixed decimals and `Float.parse`). Std reference pages regenerated; module descriptions in `scripts/std/generate.nra` updated.
- Hand-written copies replaced: `testdata/strings2.nra` `isDigit`, `scripts/errors/generate.nra` `isDigit`, `selfhost/tests/src/guide.nra` `lastIndexOf`.

## Deviations

- The survey says letters and white space come from the `std/internal/unicode` tables. Those tables hold case mapping, grapheme, width and normalization data, with no general category, so a Unicode `isLetter` needs a new generated table. Classes are ASCII only, in `std/char`, and the doc says so. A Unicode letter class belongs in `std/unicode` with its own table, a later node.
- `padStart` and `padEnd` take a `char` fill, not a string, so the fill is never cut mid-character.

## Stubs left

- None.

## Gotchas

- Replacing compiler call sites waits for the seed (the seed rule). Left for after the seed regeneration: `selfhost/src/syntax/lexer.nra` (`isLetter`, `isDigit`, `isHexDigit`, over bytes), `syntax/trivia.nra` `isSpace`, `driver/manifest.nra` `isLetter`, `tools/fmtprinter.nra` `trimEnd`, `syntax/triviadump.nra` `padNumbers`, `consteval.nra` `parseFloatText`. Most take a byte `int`, so they need `Char.fromCode` or a byte variant first. `std/src/json.nra` and `toml.nra` `isDigit` take bytes and stay. `scripts/unicode/generate.nra` `paddedByte` and `selfhost/tests/src/fmtstable.nra` `trimSpaces` (trims spaces only) and `isDigit` (byte) were left: different semantics.
- `Str.lastIndexOf` returns `Option<int>`; the old test helper returned -1.
- A division in a std source part needs a zero check on a local name first (`makePadding`).

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh --suite generated --update`: std text, runtime text and reference pages regenerated.
- Dump and query goldens that list std/string, updated with `--update` (`testdata/declsdump`, `typesdump`, `pkgdump`, `query/std_modules`).
- Full `sh scripts/test.sh`: `PASS, 10841 cases in 228s`.
- `pnpm --dir website build`: passes.

## Size

- Estimated M, actual M.
