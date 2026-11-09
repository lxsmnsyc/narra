# Outcome: template-nested-strings

## Built

- `lexTemplate` in `selfhost/src/lexer.nra` skips strings, character literals and nested templates inside `${...}` (`skipTemplateBody`, `skipInterpolation`, `skipQuoted`).
- `selfhost/src/parser.nra` finds the end of `${...}` with `findInterpolationEnd`, `findQuoteEnd` and `findTemplateEnd`, in `parseTemplate` and `collectTemplateNames`.
- Samples: `testdata/template_nested_strings.nra` and `.out`, `testdata/errors/template_nested_unterminated.nra` and `.err`.

## Proof

- base: 6478acd, `narrac testdata/template_nested_strings.nra` (sample was first named `template_nested`) gives `error 4:15: unterminated literal`; the existing golden `parse/template_nested.ast` also recorded the bug (`unexpected character '$'`).
- branch: the sample lexes, runs and prints the expected output under the test runner (debug and release, sanitizers); the error sample reports `2:13: unterminated literal` at the string.

## Deviations

- Three parse goldens changed (see log.md): `template_nested` (bug fixed), `template_lex` (error now at the string), `template_open` (error now `unterminated template string` at the nested backtick, E0040 is no longer reached by that input).

## Stubs left

- None.

## Gotchas

- Sample basenames must be unique across `testdata/` and `testdata/parse/`; the edge suite pairs by basename.
- The full run needs a compiler built from the branch source; the old seed cannot compile the new samples.

## Proposed learnings

- None.

## Verification

- `NARRAC=<branch narrac> sh scripts/test.sh`: 2 failures (`parse/template_open`, `parse/template_lex`), both expected golden changes. Both goldens were updated after that run; `--filter template` then passes.

## Size

- Estimated S, actual S.
