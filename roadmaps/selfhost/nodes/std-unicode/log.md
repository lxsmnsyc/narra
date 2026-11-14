# Log: std-unicode

## Proof on the base

- Base: feat/batch-14 at 9d3fe64b, compiler built with `sh selfhost/bootstrap.sh` (release flags).
- Command: `narrac run proof.nra`, a program that imports `std/unicode` and prints `Unicode.toUpperCase("stra\u{df}e")`.
- Result: exit 1, `proof.nra:1:1: unknown module "std/unicode" (this compiler supports ... std/toml and std/terminal)`.

## Inputs

- UCD 17.0.0 from https://www.unicode.org/Public/17.0.0/ucd/ (UnicodeData, SpecialCasing, CaseFolding, auxiliary/GraphemeBreakProperty, emoji/emoji-data, EastAsianWidth, DerivedCoreProperties, auxiliary/GraphemeBreakTest) into `/private/tmp/claude-501/std-unicode/ucd`, flat.
- The generator's SHA-256 (Narra, FIPS 180-4 over `int` masked to 32 bits) matches `shasum -a 256` on UnicodeData.txt and CaseFolding.txt.
- `narrac run -release scripts/unicode/generate.nra <ucd>`: 2.9 s wall. Output 56,572 bytes: GRAPHEME 22,860 chars (1,632 records), CASE_CONTEXT 8,194, WIDTH 7,615, UPPER/LOWER/FOLD ranges 3,792/3,545/3,810, full mappings 2,476 (upper), 24 (lower, U+0130 only), 2,523 (fold).

## Module path

- `std/internal/unicode` works as a source module: the loader finds any path that `stdSourceFor` knows, so the file is `std/src/internal__unicode.nra` (the generated suite maps `__` to `/`). The registry needs a row for it.
- The E0426 check for `std/internal/` sat inside the registry branch of `importModules`, after `importPackageModule`, which handles source modules. A user program could import `std/internal/unicode`. Moved the check up beside the `std/ext/` one (`scopes.nra`), so it runs for every import.

## Checks

- `testdata/unicode_graphemes.nra`: 766 of 766 lines of GraphemeBreakTest-17.0.0 pass on the first run.
- Every code point through toUpperCase, toLowerCase and foldCase compared with Python 3 (`unicodedata` 13.0.0) for code points assigned in 13.0: 0 differences in lower and fold; 2 in upper (U+019B to U+A7DC and U+0264 to U+A7CB, mappings added in Unicode 16).

## Review fixes (after rebase on ae877384)

- Final sigma skips Case_Ignorable before testing Cased (ICU, V8). Probes in unicode_case: `\u{391}\u{3a3}\u{2b0}` is final, `\u{2b0}\u{3a3}` and `\u{391}\u{2b0}\u{3a3}\u{2b0}\u{391}` are not.
- Properties carried between clusters. 350,000 CJK chars, -release: graphemeCount 71.5 ms to 20.0 ms, displayWidth 93.6 ms to 37.3 ms, Str.nextChar walk 1.4 ms.
- Rebase conflicts: SPEC, README, stdlib (kept both), goldens and stdtext regenerated, guide taken from feat/batch-14 (std-regex repaired the order) with the unicode section appended after regex.
