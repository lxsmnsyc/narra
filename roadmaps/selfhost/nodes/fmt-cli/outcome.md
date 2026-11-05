# Outcome: fmt-cli

## Built

- `selfhost/src/fmt.nra`: `narrac fmt [--check] [paths]`, `--stdin` / `-` (`runFmt`, `formatSource`), the file, import, declaration and statement printers, `printBlock`, and `createPrinter`, which builds the `Printer` of `contracts/fmt-printer.md`.
- `selfhost/src/fmtprinter.nra`: `Settings`, `Printer`, `withComments`, `printDangling`, `trailingComma` as in the contract. Extra exports: `getOffset`, `isNewlineAfter`, `getBlankKey`, `hasBlankBefore`, `hasDangling`.
- `selfhost/src/fmtconfig.nra`: the `[fmt]` table reader (`printWidth`, `tabWidth`, `useTabs`, `trailingComma`) and the search for the nearest `narra.toml` above a file. An unknown key or a wrong value is `error narra.toml:<line>: <message>`, exit 1.
- `selfhost/src/main.nra`: one dispatch line and import.
- `compiler/internal/narra/selfhost_fmt_test.go`: `TestSelfhostFmt`. Goldens (through `--stdin`, `--check` and in place), config cases, a syntax error case, a directory case, and for every `.nra` file under `testdata/` and `selfhost/src` plus three edge inputs: idempotence, tokens, comment texts and the tree without positions.
- Goldens `compiler/testdata/fmt/{decl_*,stmt_*,comment_*}.nra/.fmt` and cases `config_*/` (`narra.toml`, `input.nra`, `expected.fmt` or `expected.err`). `NARRA_UPDATE_GOLDEN=1` rewrites the `.fmt` and `expected.fmt` files.
- `SPEC.md` section 2 lists the `[fmt]` keys; `compiler/README.md` has a "Formatting" section.

Layout choices (Prettier's TypeScript output where Narra has the same construct):
- Struct fields and enum variants: one per line, no commas (the style of `SPEC.md`). Variant fields, parameter lists, type parameters and import names: one line if they fit, else one per line with the trailing comma of the settings.
- `if (` condition `)`: the condition goes in an indented group that breaks only when too long. `} else if` and `} else {` stay on the closing brace line.
- An empty block is `{}`; a block with only comments holds them.
- Blank lines between items stay, at most one, never first in a block. A blank after a leading comment stays.
- `return c ? a : b`: the parser turns it into an `if`; the printer turns it back (detected by the `?` at the statement position), always on one line, because the parser rejects a conditional that spans lines without parentheses. Parentheses written around the whole conditional or around a nested one stay.
- `for (const pattern of xs)` and `for (let x of xs)`: the parser hides them behind a `__for_L_C` loop and a first declaration; the printer puts them back.
- A parameter list does not take a trailing comma after a function type with `uses`: the parser reads the comma as part of the effect list.
- Right of `=` (const, variable, assignment): a binary, `is` or block value moves to an indented next line when it does not fit; any other value stays on the `=` line. This follows fmt-exprs' driver.
- Type parameter lists of declarations never break (Narra reads a line break only inside `(` and `[`).
- The golden harness reads options from file names like fmt-exprs: `_w<N>` sets printWidth and `_none` sets trailingComma to "none" (case `decl_long_w40_none`).
- Block comments stay as written; trailing blanks of a comment are dropped.

## Deviations

- The stability test ignores `newline` tokens and `,` tokens, and compares comment texts without trailing blanks. The formatter sets line ends, separators and trailing commas, so the strict token sequence of the brief cannot hold.
- `fmtprinter.nra` has extra exported helpers (listed above). The contract interface is unchanged. Function-typed fields have no effects, as in the contract; allocating functions are accepted there.
- `--stdin` finds the config from the current directory (no `--stdin-filepath`).
- Files outside the owned paths: `SPEC.md` and `compiler/README.md` (shared), `selfhost/src/main.nra` (shared).

- After printing, `formatSource` parses the result and compares its comments with the original ones; a comment that is missing exits 1 with `error <path>:<line>:<col>: comment not printed: <text>` (for example a comment inside the empty parentheses of an enum variant). `printBlock` is exported and wraps the block in `withComments`; a trailing comment of a `then` block puts `else` on the next line. Without PWD the config search stays relative and climbs with `..`. The stability test also compares the node kind and placement of each comment.

- Comments before an `else` or `else if` print on their own lines before `else`. Comments between a loop label and its keyword print after the label (a line comment ends the line). Comments inside the braces of an import print as dangling comments, one per line, before the names.

## Stubs left

- `printExpressionStub`, `printPatternStub` and `printTypeStub` in `selfhost/src/fmt.nra` copy the source text of the node (with its parentheses; comments attached to the node are kept). fmt-exprs replaces the three fields in `createPrinter` and deletes the stubs. Until then a multi-line expression keeps its source indentation, and a function type keeps its parameter names.

## Gotchas

- `getOffset` scans the source from the start for every call, so the stub and the comment helpers are quadratic in the file size. Fine for the repo's files (the largest takes a few seconds under AddressSanitizer in the test).
- `Doc.Text` may hold newlines (the stub and block comments rely on it); the printer writes them as they are and counts them in the width.
- A comment attached to a node the statement printer does not visit is lost. The corpus test finds this by comparing comment texts. Known unhandled placements: a comment before `else`, and a comment between `)` and `{`. Comments inside import braces and around enum variants parentheses are hoisted above the first name (import braces) or reported as not printed (empty variant parentheses); they do not stay next to the nearest name. A block comment followed on its line by more comments or by `,` `)` `]` `}` `;` is trailing of the node on its left (`endsLineAfterComments` in `trivia.nra`).
- The parser drops hidden for-loop names, ternary returns and parameter names of function types. fmt-exprs must keep function type parameter names (`(x: int) => int`) from the source; the AST does not hold them.
- `Printer` fields are called as `printer.printExpression(printer, expression)`.

## Proposed learnings

- Compare formatter output on the whole repo by token sequence without separators, comment texts and the tree without positions | the formatter changes separators and line ends by design | compiler/internal/narra | yes (TestSelfhostFmt)
- Keep the stub of a not-yet-built printer in the driver module and swap it in one constructor | the next node deletes it without touching the statement printer | selfhost/src | no

## Verification

- `cd compiler && go test -count=1 -timeout 40m ./internal/narra -run 'TestSelfhostFmt|TestSelfhostBootstrap'`: see the pull request.
- Fast check `cd compiler && go vet ./... && go test -count=1 -timeout 40m ./...`: see the pull request.

## Size

- Estimated M, actual M.
