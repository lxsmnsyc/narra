# Outcome: textmate-grammar

## Built

- `editors/textmate/narra.tmLanguage.json`: standard TextMate JSON grammar, scope `source.narra`.
- `editors/textmate/check.cjs` and `package.json`: tokenizes every `narra` fence of `docs/guide/*.md` and `SPEC.md` (and any `.nra` file given) with `vscode-textmate` and `vscode-oniguruma`; fails on an `invalid` scope or a sample that ends inside an open rule.
- `editors/textmate/README.md`: loading in VS Code, Shiki and twinkleplop, and running the check.
- `SPEC.md`: all 63 `ts` fences are now `narra`; none held TypeScript on purpose.

## Proof

Not a fix node.

## Deviations

- `selfhost/tests/src/spec.nra` (not owned): the spec suite picked fences by the info string `ts`; changed to `narra` (3 lines). Without it the suite would test no SPEC sample.
- No numeric literal suffixes in the grammar: the lexer has none.
- The README names twinkleplop only generally; its loading API was not read.

## Stubs left

- None.

## Gotchas

- The check needs `npm install` in `editors/textmate/`; it is not in `scripts/test.sh`.
- Type annotations after `:` are matched by shape, so a value after `:` in a struct literal or map entry is colored like a type name when capitalized.

## Proposed learnings

- A fence info string in SPEC.md or the guide is read by the spec and guide suites | changing it needs the suite's reader changed too | selfhost/tests | yes

## Verification

- `NODE_PATH=<scratch>/node_modules node editors/textmate/check.cjs`: 283 samples (guide plus SPEC), 0 problems. Tokenizer: vscode-textmate 9 with vscode-oniguruma. Also 187 `selfhost/src` files, 0 problems.
- `sh scripts/test.sh`: `PASS, 9781 cases`.

## Size

- Estimated M, actual M.
