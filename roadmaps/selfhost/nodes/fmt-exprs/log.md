# Log: fmt-exprs

- fmt-cli had not landed, so `selfhost/src/fmtprinter.nra` was written from the contract text. Take fmt-cli's copy on rebase.
- Narra reads a line break only inside `(` and `[` (the lexer ends a statement at a newline after a name, literal, `)`, `]` or `}`). So type argument lists never break (a break before `>` would end the statement), and a broken `c ? a : b` is wrapped in parentheses (the parser rejects a conditional that spans lines otherwise). Inside call arguments and array elements the parentheses are left out.
- The AST drops the parameter names of a function type; `getParameterNames` lexes the type's source text (from `Trivia.ends`) to get them back.
- Number literals are copied from the source: the AST keeps the normalised token text (`0x1F_F` became `0x1FF`).
- Test driver: no CLI exists, so `testdata/fmt/expr_driver.txt` is a replacement main.nra for a copy of `selfhost/` that `selfhost_fmtexprs_test.go` builds. It formats top-level `const` declarations; a block is printed by copying its statements. `FMT_UPDATE=1` rewrites the `.fmt` goldens.
- `--check` mode of the driver prints every statement expression of a file, reads it back as `const x = <text>`, prints again and requires equal text (run on all `.nra` under testdata/ and selfhost/src at width 30 with commas and width 100 without).
