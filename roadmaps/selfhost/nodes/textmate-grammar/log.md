# Log: textmate-grammar

- Branch renamed to feat/textmate-grammar, reset to feat/batch-16. Nothing existed under editors/.
- Keywords taken from KEYWORDS in selfhost/src/lexer.nra; contextual words (requires, ensures, typeof, opaque, distinct, extern, unsafe, in) added from parser.nra.
- Tokenizer for the check: vscode-textmate 9 with vscode-oniguruma, installed in /private/tmp/claude-501/textmate-grammar (not in the repo). The check is editors/textmate/check.cjs, run with NODE_PATH pointing at that folder.
- The lexer has no numeric literal suffixes (`42u8` is not Narra), so the grammar has none; the type comes from context.
- First run failed on two guide samples: a template after `:` in a conditional expression, and `=> value < limit` after a return type opened a `<`. Fixed by letting a type annotation hold templates and end at `=>` and `=`.
- Result: 220 narra fences of docs/guide, 283 with SPEC.md, 187 selfhost/src files: no invalid scope, no open rule.
- SPEC.md: all 63 `ts` fences were Narra code (none is a deliberate TypeScript comparison; the comparison table is a markdown table), so all became `narra`.
- selfhost/tests/src/spec.nra selected fences by the info string `ts`; changed to `narra` (3 lines, one is the test, two are comments).
