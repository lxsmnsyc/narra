# Outcome: std-unicode

## Built

- `std/src/unicode.nra`, module `std/unicode` (alias `Unicode`, Unicode 17.0.0), Narra source:
  - `graphemes(text): Generator<string> uses alloc` (UAX #29 extended clusters, rules GB3 to GB999 with GB9c and GB11; each cluster a `Str.slice` of `text`), `graphemeCount(text): int`.
  - `toUpperCase`, `toLowerCase` (full mappings: UnicodeData simple mappings replaced by the unconditional lines of SpecialCasing; final sigma from Cased and Case_Ignorable; no locale), `foldCase` (CaseFolding status C and F). All `uses alloc`; text with nothing to change comes back as the same string.
  - `displayWidth(text): int`: per cluster, the widest code point (0 for Mn, Me, Mc, Cf, Cc, Zl, Zp, Default_Ignorable and Hangul V/T jamo; 2 for East Asian W/F and Emoji_Presentation; 1 otherwise), and 2 when U+FE0F follows an Emoji=Yes first code point (keycaps, `\u{2764}\u{fe0f}`).
  - Each code point's grapheme properties are looked up once: `clusterEnd` returns the properties of the code point at the cluster end, and the next cluster starts from them; `clusterWidth` reuses the width class of the first code point. On 350,000 CJK characters (1 MB, `-release`, M2): `graphemeCount` 72 ms to 20 ms, `displayWidth` 94 ms to 37 ms; a plain `Str.nextChar` walk is 1.4 ms.
  - ASCII fast paths: two ASCII bytes always break (except CR LF) with no lookup; ASCII letters change case by arithmetic; a one-byte ASCII cluster is 1 column (0 for controls).
- `std/internal/unicode` (`std/src/internal__unicode.nra`, 56.6 KB), generated: `VERSION`, `GRAPHEME` (14-char records: range + 2 hex digits = break + 16 * InCB + 64 * ExtPict), `WIDTH` and `CASE_CONTEXT` (13-char range records), `UPPER_RANGES`/`LOWER_RANGES`/`FOLD_RANGES` (19 chars: range, stride 1 or 2, target of the start) and `UPPER_FULL`/`LOWER_FULL`/`FOLD_FULL` (24 chars: code point + three targets). Lowercase hex, sorted, binary search with `Str.byteAt`. The header names the version, the download URL (https://www.unicode.org/Public/17.0.0/ucd/) and the SHA-256 of the seven inputs.
- `scripts/unicode/generate.nra`: `narrac run -release scripts/unicode/generate.nra <ucd folder> [output]`, run from the repo root (output defaults to `std/src/internal__unicode.nra`). Reads the seven files flat from the folder; version from the DerivedCoreProperties header; SHA-256 written in Narra. 2.9 s. Room for normalization: add files to `INPUTS` and a writer to `writeTables`.
- `selfhost/src/stdunicode.nra`: rows for `std/unicode` and `std/internal/unicode` (`check: "source"`), registered last in `stdlib.nra`; `std/unicode` is last in the unknown-module list. `stdtext.nra` regenerated.
- `selfhost/src/scopes.nra`: the `std/internal/` E0426 check moved up beside the `std/ext/` check, so it covers source modules; the old check in the registry branch is gone (see Deviations).
- Samples: `testdata/unicode_graphemes.nra` (all 766 lines of `testdata/unicode/GraphemeBreakTest.txt` pass, plus a readable sample), `testdata/unicode_case.nra`, `testdata/unicode_width.nra`. Error: `testdata/errors/unicode_internal_import.nra`. `testdata/unicode/LICENSE.txt` is the Unicode License v3.
- Generator determinism: `generated` suite case `unicodetables` (`selfhost/tests/src/unicodetables.nra`) runs the generator twice on the hand-written UCD-format files in `testdata/unicode/mini/` and compares both runs with each other and with `testdata/unicode/mini/tables.out` (`--update` rewrites it).
- SPEC section 3 (std modules list; D118 bullet as built) and section 4 "Strings"; `compiler/README.md`; guide section "Text as people see it" in `docs/guide/12-standard-library.md`, after the regex section, with the width edges (ambiguous width 1, controls and tab 0, U+FE0E after a wide emoji still 2, U+115F 0).

## Proof

- base: 9d3fe64b (the same on ae877384), `narrac run proof.nra` (imports `std/unicode`, calls `toUpperCase`): exit 1, `unknown module "std/unicode"`.
- branch: the three samples print their `.out` files, clean with `-sanitize`, the leak check and the copy check (programs suite, debug and release).

## Deviations

- The tables live at `std/internal/unicode`, file `std/src/internal__unicode.nra`, not the owned `std/src/unicode__tables.nra`: the brief asks to try the internal path first, and it works. The loader serves any source path; only the E0426 check needed a change.
- `selfhost/src/scopes.nra` (not in owned or shared paths): the `std/internal/` check ran only for registry modules after `importPackageModule`, so a user program could import a source module under `std/internal/`. The check now runs first for every import. Builtin internal modules give the same error as before (`terminal_internal_import.err` unchanged).
- `selfhost/tests/src/generated.nra` (3 lines of dispatch) and the new `selfhost/tests/src/unicodetables.nra`, for the determinism acceptance box. `testdata/unicode/mini/*.txt` are hand-written test files in UCD format, not UCD files.
- No codes from E1320-E1324 used.

## Stubs left

- `normalize`, `isNormalized` and `NormalForm` are node std-unicode-normalize.

## Gotchas

- The generator needs the UCD files flat in one folder (`auxiliary/` and `emoji/` files beside the others).
- Std sources must be ASCII (the stdtext generator rejects other bytes), so samples and comments write `\u{df}`, not the letter.
- Each record table is one string literal of up to 23 KB on one line; `narrac fmt` leaves it as is.
- Display width of a cluster is its widest code point; a cluster that starts with a zero-width code point and holds a wide one is 2.
- narrac grows by the embedded text: stage 1 binary 4,430,792 to 4,497,128 bytes (+66 KB) on ae877384.

## Proposed learnings

- A check on import paths goes before the package and source module lookup | E0426 was skipped for source modules under std/internal | selfhost/src/scopes.nra | yes (an error test per internal source module)

## Verification

- After the review fixes and the rebase on ae877384: `--filter unicode`: PASS, 41 cases. `--suite guide`: PASS, 313 cases. `--suite generated`: PASS, 22 cases. `--suite errors`: PASS, 769. `--suite dumps`, `sources`, `spec`, `codes`: PASS.
- Every code point through the three case functions against Python 3 `unicodedata` 13.0.0: no differences except two Unicode 16 upper mappings (U+019B, U+0264).
- `sh scripts/test.sh` on ae877384 plus this change: `test.sh: PASS, 8741 cases in 152s`.
- `narrac emit selfhost` from the base compiler and the branch compiler: byte-identical (224,350 lines, ae877384 base compiler against the branch compiler). Bootstrap fixed point holds.
- `sh scripts/linux-check.sh` not run.

## Size

- Estimated L, actual M to L.
