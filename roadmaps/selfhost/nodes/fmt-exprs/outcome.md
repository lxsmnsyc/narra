# Outcome: fmt-exprs

## Built

- `selfhost/src/fmtexprs.nra`: `printExpression`, `printPattern`, `printType` (contract fmt-printer) for every `Expr`, `Pattern` and `TypeExpr` form. Binary chains break after each operator; the last argument that is an arrow with a block or compact body, or a non-empty struct, array or map literal, hugs the call; `match` is one group (flat: `match (x) { A => 1, B => 2 }`, broken: one arm per line, no commas); `c ? a : b` and `if` values are printed back from the parser's encoding; written parentheses are kept and needed ones added; literals (strings, chars, templates, numbers) are copied from the source; struct literals whose first field starts on a new line stay broken.
- `selfhost/src/fmtprinter.nra`: fmt-cli's copy (the stand-in is gone).
- Goldens `compiler/testdata/fmt/{expr,type,pattern}_*.nra` with `.fmt`: binary, call, struct, collections, match, arrow, ternary and `if`, parens, template and literals, comments, patterns, types, function types. `_w<N>` in a name is the print width, and `_none` after it turns trailing commas off (that is only a convention of my test).
- `createPrinter` in `selfhost/src/fmt.nra` uses `Exprs.printExpression`, `printPattern` and `printType`. The stub printers and their header note are deleted. fmt-cli's `TestSelfhostFmt` runs the goldens (prefixes `expr_`, `type_`, `pattern_` added to its list) and its stability checks (idempotence, tokens, comments, tree) over every `.nra` under `testdata/` and `selfhost/src`. The temporary driver and the `selfhost_fmtexprs*_test.go` files are deleted.
- Goldens: `_w<N>` in a name is the print width, `_none` turns trailing commas off (fmt-cli's harness reads the same names). New: `comment_operands`, `expr_nested_hug` (30 nested last-argument callbacks; replaces the exponential-time guard test).

## Deviations

- Edits outside the owned paths, all needed so the two printers fit:
  - `selfhost/src/fmt.nra`: the three printer fields, stubs deleted. `printAssigned` no longer moves a block, `if` or ternary value to the line after `=` (those break inside their own braces or parentheses, so `const x =` followed by an indented `if (...) {` was wrong); a binary or `is` value still does.
  - `selfhost/src/trivia.nra` (`endsLineAfterComments`, new `isOnlySpaceBetween`): a block comment counted as ending its line when a later comment ended the line, even with code in between. `x + /* m */ y /* a */` made `/* m */` trailing of `x` and the printer moved it before the operator. It now looks at the next comment only when blanks alone separate the two.
  - `compiler/internal/narra/selfhost_fmt_test.go`: golden prefixes, and `tokensOf` ignores grouping parentheses (a `(` after an operator, `,`, `(`, `?`, `:` or `return`, with its `)`). The printer drops a redundant pair (`((a))` becomes `(a)`, `Trivia.parenthesized` is a flag) and wraps a conditional that spans lines in parentheses. Call, parameter and condition parentheses stay in the compared sequence, and the tree comparison covers the rest.
- Goldens of fmt-cli that changed because expressions are now formatted: `stmt_basic` (`print( "hi" )` is `print("hi")`), `stmt_ternary` (an over-long `return c(a) ? x : y` breaks the call arguments, the only break the parser accepts there), `decl_long_w40_none` (long chains and calls past 40 columns are broken).
- Generators (`function*`, `yield`, `yield*`, `Generator<T> uses E`) are printed since the story branch took in main (PR #82).

## Stubs left

- None.

## Gotchas

- Narra reads a line break only inside `(` and `[`; a newline after a name, literal, `)`, `]` or `}` ends the statement (lexer `lastEndsStatement`). So: type argument lists never break; a broken ternary needs parentheses (the printer adds them, except directly in call arguments and array elements, `printElement`); the closing `>` can never sit on its own line.
- `printBlock` must print the braces too; the expression printer calls it for arrow bodies, match arm blocks, `if` value branches and plain block expressions.
- The AST keeps no parameter names of function types; `getParameterNames` lexes the type's source (needs `Trivia.ends` for `t<l>:<c>`).
- Number token text in the AST is normalised (`0x1F_F` is `0x1FF`); the printer reads it from the source.
- `Map<K, V>` and struct literal type arguments, import names and type parameter bounds have no comment anchors (fmt-trivia gap); comments there are not attached.
- A broken `return c ? a : b` cannot be parenthesized (the parser would read a value, not the hidden `if`), so a call in its condition breaks inside its own parentheses and the line may stay over the width.
- A comment between an operand and its operator moves after the operator (`x /* c */ + y` becomes `x + /* c */ y`). For `is` the comment goes after `is` and the pattern moves to the next line.
- Layout on the right of `=` is fmt-cli's. The driver used: a binary, `is` or `if` value breaks after `=` and indents; everything else stays after `=`. An arrow body that is a binary chain is indented once, not twice (`printBody`).
- Doc `Text` with a newline inside (a multi-line template) is printed as written; the width of that line is then counted wrongly.
- `Trivia.ends` of a parenthesized expression includes its parentheses. Literals copied from the source cut the text after their last token that is not `)` (`cutClosingParens`); without that `(1)` printed as `(1))` (review finding on PR #81).
- Comments in an empty struct literal (`Point { /* c */ }`) are not printed; `narrac fmt` stops with `comment not printed` and leaves the file unchanged. `f(/* c */)` prints as `f /* c */()` and `a./* c */b` as `a /* c */.b`. Fixed since (fix/fmt-bracket-comments): comments right inside grouping parentheses (`(/* in */ x)`, `(x /* out */)`, `(\n // c\n x\n)`) belong to the expression and its parentheses are dropped when redundant (kept, with the comment inside, when precedence or a multi-line conditional needs them); a comment before the `)` of an `if`, `for`, `while`, `match` or guard header, an index `[ ]` or a `return ( )` stays inside and breaks that pair.
- A Block comment is always printed inline (`/* c */ node`); a Line comment gets a hard line. fmt-cli's `withComments` decides this for statements.

## Proposed learnings

- Test a printer's output by parsing it and printing again, over every source file in the repo | catches output that the parser rejects (line breaks, parentheses) and unstable layouts | selfhost/ | yes (TestSelfhostFmt stable cases)

## Verification

- `cd compiler && go vet ./... && go test -count=1 -timeout 40m ./...`: passes.
- `TestSelfhostFmt` and `TestSelfhostBootstrap` pass.

## Size

- Estimated M, actual M.
