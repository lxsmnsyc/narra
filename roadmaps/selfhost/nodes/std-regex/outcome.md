# Outcome: std-regex

## Built

- `std/src/regex.nra` (alias `Regexes`), pure Narra over `std/string`, `std/array`, `std/int` and `std/char`. Types: `Regex`, `RegexError { position, reason }`, `Match { start, end, text, groups, names }` (`groups[0]` is the whole match, `Option<string>` per group). Functions: `compile`, `matches` (the D118 `test`, see Deviations), `find`, `findFrom`, `findAll` (`Generator<Match>`), `group`, `namedGroup`, `replace`, `replaceAll`, `replaceWith` (non-escaping callback with an effect variable) and `split`.
- Parser: JavaScript syntax without back references and lookaround, RE2 inline flags `(?i)`, `(?m)`, `(?s)`, `(?flags:...)` and `-` to turn off; `{` that is not a repeat is a literal. Errors as values: unclosed group or class, unmatched `)`, bad class range, nothing to repeat, back references and lookaround ("... is not supported: the engine is linear-time"), repeat count above 1000, program above 10,000 instructions, instructions times slots above 1,000,000, bad escapes and group names, duplicate names, nesting deeper than 1000.
- Engine: the pattern becomes a node array, then a program of parallel arrays (`ops: Array<Op>`, `firsts`, `seconds`, `ranges`). Classes are sorted range pairs, negated and ASCII-folded at compile time. Pike VM over code points (own UTF-8 decoding with `Str.byteAt`), leftmost-first with captures, an explicit stack for splits and saves (no recursion on the text or program), JavaScript's repeat rules (each pass clears its groups with a `Reset` instruction; a pass beyond the lower count that matches empty fails at a `Progress` instruction, with one position slot per repeat), surrogate pair escapes as one code point, a literal prefix searched before the machine runs, and a start-anchor shortcut.
- `selfhost/src/stdregex.nra` (module row), one import and one `registerAll` line in `stdlib.nra`, regenerated `stdtext.nra`, module-list goldens.
- Samples: `testdata/regex.nra` (106 pattern and text pairs, the repeat and surrogate cases checked against `node`, plus findAll, findFrom and matches cases), `regex_linear.nra`, `regex_errors.nra` (37 bad patterns), `regex_replace.nra`; error test `testdata/errors/regex_alias_clash.nra`.
- SPEC section 3 `std/regex` bullet (the `Regexes.checked` line stays under "Not built yet"), `compiler/README.md` line, guide chapter 12 section "Patterns: `std/regex`".

## Proof

- base: 183eec94 (before the rebase on 9d3fe64b), `/private/tmp/claude-501/std-regex/base/narrac run testdata/regex.nra`: `unknown module "std/regex"`, exit 1.
- branch: `NARRAC=<branch> sh scripts/test.sh --filter regex`: PASS, 28 cases (ASan, leak and copy checks included).

## Deviations

- `test` is named `matches`: `test` is a keyword, so `Regexes.test(...)` does not parse. Allowing keywords after `.` is a parser change outside this node. Decision for the maintainer: keep `matches`, or open a node that allows keyword member names.
- The literal-prefix search scans bytes with `Str.byteAt`, not `Str.indexOf`: `Str.indexOf` has no start offset, and `Str.sliceFrom` copies the rest of the text, which makes `findAll` quadratic.
- `replaceWith` replaces every match (D118 does not say first or all); `replace` is the first-match form.
- `split` does not add capture groups to the result, unlike JavaScript.
- Most Annex B forms that JavaScript accepts without the `u` flag stay errors: `\8` and `\9` (back reference error), unknown letter escapes such as `\a`, a class range with a set escape (`[\w-z]`), `\c` without a letter. A lone `]` or `}` and a `{` that does not start a repeat are literals, as in Annex B.
- Thread state memory is capped, not shared: a program whose instructions times slots (two per group, one per repeat) is above 1,000,000 is a `RegexError`, so a machine holds at most 16 MB of slots. Copy-on-write capture sharing between threads was not built.
- The alias error test shows the existing message (`the import alias Regex is also a type name; ...`), not the generic clash message the brief quoted; the hint suggests `Regexs`, which is what the existing error prints.
- `std/char` is imported (for `Str.appendChar` in the prefix). std-regex-const must add it to the evaluator.
- No diagnostic codes used (E1350-E1359 stay free).
- `docs/guide/12-standard-library.md` on feat/batch-14 (9d3fe64b) had the std/terminal section merged into the middle of the std/toml sample, so the guide suite failed on the base (`12-standard-library.md:274`: a `narra` block with no `text` block). The branch puts the toml sample, its output and paragraph back together before the terminal section.

## Stubs left

- None. `Regexes.checked` is node std-regex-const.

## Gotchas

- For std-regex-const: `compile` has no closures, no traits beyond derives, and uses `copy`, `Str.builder`/`append`/`appendChar`/`build`, `Str.byteAt`, `Str.length`, `Str.slice`, `Char.fromCode`, `Arrays.push`, `Arrays.length`, `Arrays.filled` (machine only, not in `compile`), and index writes. `Regex` holds strings, ints, bools, `Array<int>`, `Array<string>` and `Array<Op>` (an enum without payloads).
- Throughput: about 150 ns per thread step in a `narrac run` build; `a{0,1000}a{1000}` over 1,000 characters (4 million steps) took 0.64 s, so that case is not in the timed sample. Larger texts with long programs need a faster VM (one-pass or DFA) later.
- `Match.names` is the regex's array, shared: `find`, `findFrom` and `buildMatch` return `... from regex`.
- The Pike VM keeps one state per instruction and list, so a `Progress` slot of a thread that reaches an instruction after a higher-priority thread is dropped with it; the samples match `node` on the repeat cases tested.

## Proposed learnings

- `test` and other keywords cannot be std function names | `Regexes.test` from D118 does not parse | std/src, DECISIONS | yes (a sources check of exported std names against the keyword table)

## Verification

- `NARRAC=<branch> sh scripts/test.sh --filter regex`: PASS, 28 cases. `--suite guide`: PASS, 309. `--suite sources`: PASS, 1. `--filter unknown_module`: PASS, 15.
- The seed compiler (`cc seed/narrac.c`) compiles and runs `regex_replace.nra` against the module in a scratch package.
- `narrac emit selfhost` of the base tree (`git archive feat/batch-14`) with the base and the branch compiler: byte-identical, 224,243 lines (base 9d3fe64b).
- Full check `sh scripts/test.sh` after the review fixes and the rebase on 9d3fe64b: PASS, 8715 cases in 304s.

## Size

- Estimated L, actual L (1,705 lines of module after `narrac fmt`).
