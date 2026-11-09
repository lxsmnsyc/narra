# Log

## Proof on the base (commit 6478acd, origin/feat/batch-2)

Compiler: `sh selfhost/bootstrap.sh <dir>` on the base.

Command: `<dir>/narrac testdata/template_nested.nra` (the sample added by this node; lexes the file).

Result: `error 4:15: unterminated literal`. Line 4 is `` print(`a${"`"}b`) ``: `lexTemplate` ends the template at the backtick inside the nested string. The `${"}"}` case (line 2) lexes but `parseTemplate` counts the `}` in the string and ends the interpolation early. Reproduced.

## Fix

- lexer.nra: `lexTemplate` now uses `skipTemplateBody`, `skipInterpolation` and `skipQuoted`, which skip strings, character literals and nested templates inside `${...}`. An unterminated string reports E0009 at the string.
- parser.nra: `findInterpolationEnd`, `findQuoteEnd`, `findTemplateEnd` replace the brace counters in `parseTemplate` and `collectTemplateNames`.
- Branch result: `testdata/template_nested_strings.nra` runs with the expected output; `errors/template_nested_unterminated` reports `2:13: unterminated literal` at the string.
- Goldens changed: `parse/template_nested.ast` recorded the bug (`error 1:33: unexpected character '$'`) and now parses; `parse/template_lex.ast` now `error 1:19: unterminated literal` (was reported at the template); `parse/template_open.ast` now `error 1:22: unterminated template string` (the backtick after `${x` opens a nested template).
- Note: the first sample name `template_nested` collided with the parse sample by basename, so the sample is `template_nested_strings`.
- The full run used NARRAC built from the branch source (`sh selfhost/bootstrap.sh`), because the old seed cannot compile the new samples.
