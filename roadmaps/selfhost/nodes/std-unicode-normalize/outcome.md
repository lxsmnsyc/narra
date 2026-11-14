# Outcome: std-unicode-normalize

## Built

- `std/unicode` (`std/src/unicode.nra`): `export enum NormalForm derives Show, Equatable { NFC, NFD, NFKC, NFKD }`, `normalize(text, form): string uses alloc` and `isNormalized(text, form): bool uses alloc` (UAX #15, Unicode 17.0.0).
  - Quick check first: one pass over the text that reads the combining class and the form's quick check of each non-ASCII code point. A Yes answer returns `text` itself, with no allocation (ASCII text always gets Yes). Otherwise `normalize` keeps the bytes before the last starter with a Yes answer and normalizes the rest: full decomposition (recursive over one-step tables), canonical ordering on insert, then canonical composition for NFC and NFKC. `isNormalized` answers Yes and No from the quick check and normalizes and compares only on Maybe.
  - Hangul syllables are decomposed and composed by arithmetic (UAX #15 section 16); no table holds them.
- `std/internal/unicode` (`std/src/internal__unicode.nra`, 56.6 KB to 148.8 KB), new tables, all with 5-digit lowercase hex code points: `NORMAL_PROPERTIES` (14-char ranges: combining class + 256 * quick check bits), `CANONICAL` (15: code point and its one-step canonical decomposition of one or two code points), `COMPOSITION` (3-digit indexes of the CANONICAL records that compose, sorted by pair), `COMPATIBILITY_RANGES` (15: runs of one-code-point compatibility decompositions), `COMPATIBILITY` (9: code point + offset) and `COMPATIBILITY_DATA` (5 per code point). The header lists the two new inputs with their SHA-256.
- `scripts/unicode/generate.nra`: reads `CompositionExclusions.txt` and `DerivedNormalizationProps.txt` too (nine inputs; NormalizationTest.txt is not an input). Derives Full_Composition_Exclusion from CompositionExclusions plus singletons and non-starter decompositions and panics when it differs from the derived property. `writeRanges` takes the key width (6 for the old tables, output unchanged). Field writers panic when a value does not fit its digits. 2.4 s.
- Canonical ordering sorts each finished run of non-starters (`orderMarks`): insertion for runs up to 8 marks, a stable bottom-up merge sort for longer ones, so a long run of alternating marks is not quadratic.
- Samples: `testdata/unicode_normalize_linear.nra` (80,000 alternating U+0301/U+0316 after `a` in NFD and NFC, right answer and under a 1 s bound), `testdata/unicode_normalize.nra` (exits 1 when a line fails) (every line of `testdata/unicode/NormalizationTest-sample.txt`: each of the five columns in all four forms, through `normalize` and `isNormalized`; takes another path as its first argument), `testdata/unicode_normalize_basic.nra` (the four cases of the brief and more). `testdata/unicode/NormalizationTest-sample.txt`: the 17.0.0 header, all of part 0 and every ninth line of part 1 (1,990 lines, 211 KB; 1,943 test lines).
- `testdata/unicode/mini/`: `CompositionExclusions.txt` and `DerivedNormalizationProps.txt` added and `UnicodeData.txt` extended (canonical, singleton, non-starter, excluded and compatibility decompositions), so the `unicodetables` case covers the new writers; `tables.out` regenerated.
- SPEC section 3 (`std/unicode` bullet, now built) and section 4 "Strings"; guide section "Normalization" under "Text as people see it" in `docs/guide/12-standard-library.md`; `compiler/README.md` paragraph on `std/unicode`.

## Proof

- base: 87bba23c (feat/batch-14), `narrac run testdata/unicode_normalize_basic.nra`: exit 1, `module "std/unicode" has no export "NormalForm"`, then `no export "normalize"` / `"isNormalized"` at each call.
- branch: both samples print their `.out` files (`1943 of 1943 lines pass`), clean under `-sanitize`, the leak check and the copy check (`--filter unicode`: PASS, 51 cases).

## Deviations

- The tables are in `std/src/internal__unicode.nra` (where std-unicode put them), not the `std/src/unicode__tables.nra` the brief lists.
- Files outside the owned and shared paths: `testdata/unicode/mini/*` (new mini inputs, for the determinism case) and `compiler/README.md` (its std/unicode paragraph said normalization was not built).
- The sample of part 1 is every ninth line rather than the first lines, so it covers Hangul, CJK compatibility ideographs and the supplementary planes, not only Latin.
- Size: the tables are 92 KB, not the 70 KB D118 estimated (5-digit fields, one-step decompositions; 133 KB with fully expanded 6-digit records).
- No codes from E1325-E1329 used.

## Stubs left

- None.

## Later work

- Speed, for a later perf-pass node: on 1 MB of text `normalize` is 20 to 90 times slower than ICU (fresh review). Two causes: each lookup is a binary search that parses hex digits at every probe (`readHex` over `NORMAL_PROPERTIES`, `CANONICAL`, `COMPOSITION`), and after the first code point without a Yes quick check answer, `normalizeFrom` decomposes and composes the whole rest of the text instead of going back to the quick check fast path at the next safe starter.

## Gotchas

- Every table field is 5 hex digits for normalization and 6 for the older tables; `findNormalRange` and `findNormalKey` read the 5-digit ones.
- `isNormalized` is `uses alloc`: a Maybe answer normalizes to compare.
- The generator needs the nine UCD files flat in one folder (17.0.0 from https://www.unicode.org/Public/17.0.0/ucd/).
- Sizes: narrac stage 1 binary 4,497,128 to 4,612,712 bytes (+115.6 KB, the embedded tables). A `-release` program that calls `graphemeCount` only: 51,256 bytes before and after (tree shaking drops the new tables; its C differs only in local id numbers and source line strings). The same program calling `normalize` too: 167,608 bytes.

## Proposed learnings

- None.

## Verification

- Full NormalizationTest.txt (17.0.0, all parts) through `testdata/unicode_normalize.nra`: `20034 of 20034 lines pass`. Every code point not listed in part 1 through the four forms: 1,094,978 checked, 0 changed (scratch program).
- Regenerating the tables with the branch compiler gives the committed file byte for byte.
- `--suite generated`: PASS, 22. `--suite guide`: PASS, 315. `--suite spec`: 62, `sources`, `codes`: PASS.
- `narrac emit selfhost` from the base compiler and the branch compiler, same source: byte-identical (224,350 lines). Bootstrap fixed point holds.
- `sh scripts/test.sh` on 1379554c plus this change: first run failed only `programs/async_api_all_observed.nra/debug` (event order of timers under load; passes 3 of 3 alone, no std/unicode use); second run `test.sh: PASS, 8793 cases in 218s`.
- `sh scripts/linux-check.sh` not run.

## Size

- Estimated M, actual M.
