# Outcome: fmt-exprs

## Built

- `selfhost/src/fmtexprs.nra`: `printExpression`, `printPattern`, `printType` (contract fmt-printer) for every `Expr`, `Pattern` and `TypeExpr` form. Binary chains break after each operator; the last argument that is an arrow with a block or compact body, or a non-empty struct, array or map literal, hugs the call; `match` is one group (flat: `match (x) { A => 1, B => 2 }`, broken: one arm per line, no commas); `c ? a : b` and `if` values are printed back from the parser's encoding; written parentheses are kept and needed ones added; literals (strings, chars, templates, numbers) are copied from the source; struct literals whose first field starts on a new line stay broken.
- `selfhost/src/fmtprinter.nra`: written from the contract text because fmt-cli had not landed (`Settings`, `Printer`, `withComments`, `printDangling`, `trailingComma`). Take fmt-cli's copy on rebase. `trailingComma` returns an `IfBreak`; the expression printer wraps it in an `IfBreak` itself, so a plain `,` from fmt-cli's version also works.
- Goldens `compiler/testdata/fmt/{expr,type,pattern}_*.nra` with `.fmt`: binary, call, struct, collections, match, arrow, ternary and `if`, parens, template and literals, comments, patterns, types, function types. `_w<N>` in a name is the print width, and `_none` after it turns trailing commas off (that is only a convention of my test).
- Test driver: `compiler/testdata/fmt/expr_driver.txt` (a replacement `main.nra` for a copy of `selfhost/`; it formats top-level `const` declarations and has a `--check` mode), `compiler/internal/narra/selfhost_fmtexprs_test.go` (`TestSelfhostFmtExprs`: goldens plus a second pass for idempotence; `FMT_UPDATE=1` rewrites the `.fmt` files) and `selfhost_fmtexprs_stable_test.go` (`TestSelfhostFmtExprsStable`: every statement expression of every `.nra` under `testdata/` and `selfhost/src`, at width 30 with commas and width 100 without, is printed, parsed back and printed again; the text must be equal).

## Deviations

- Two new Go test files and the driver `.txt` are outside the brief's owned paths (`compiler/internal/narra/selfhost_fmtexprs*_test.go`), because no CLI existed to run the goldens. fmt-cli's `TestSelfhostFmt` should take over the goldens; the two test files can then be deleted.
- Not done: replacing the stub fields in `fmt.nra` (fmt.nra does not exist yet; after fmt-cli lands set `printExpression: Exprs.printExpression` etc.). The generators note does not apply (generators are not on this branch).
- The full AST and token checks of the story are fmt-cli's; I only check text stability of expressions.

## Stubs left

- The real driver wiring in `fmt.nra` (fmt-cli plus a rebase). My `fmtprinter.nra` is a stand-in for fmt-cli's.

## Gotchas

- Narra reads a line break only inside `(` and `[`; a newline after a name, literal, `)`, `]` or `}` ends the statement (lexer `lastEndsStatement`). So: type argument lists never break; a broken ternary needs parentheses (the printer adds them, except directly in call arguments and array elements, `printElement`); the closing `>` can never sit on its own line.
- `printBlock` must print the braces too; the expression printer calls it for arrow bodies, match arm blocks, `if` value branches and plain block expressions.
- The AST keeps no parameter names of function types; `getParameterNames` lexes the type's source (needs `Trivia.ends` for `t<l>:<c>`).
- Number token text in the AST is normalised (`0x1F_F` is `0x1FF`); the printer reads it from the source.
- `Map<K, V>` and struct literal type arguments, import names and type parameter bounds have no comment anchors (fmt-trivia gap); comments there are not attached.
- Layout on the right of `=` is fmt-cli's. The driver used: a binary, `is` or `if` value breaks after `=` and indents; everything else stays after `=`. An arrow body that is a binary chain is indented once, not twice (`printBody`).
- Doc `Text` with a newline inside (a multi-line template) is printed as written; the width of that line is then counted wrongly.
- A Block comment is always printed inline (`/* c */ node`); a Line comment gets a hard line. fmt-cli's `withComments` decides this for statements.

## Proposed learnings

- Test a printer's output by parsing it and printing again, over every source file in the repo | catches output that the parser rejects (line breaks, parentheses) and unstable layouts | selfhost/ | yes (TestSelfhostFmtExprsStable)

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostFmt'` passes (the goldens, idempotence, stability over the repo).
- Fast check `cd compiler && go vet ./... && go test -count=1 -timeout 40m ./...`: see the report of the worker run.

## Size

- Estimated M, actual M.
