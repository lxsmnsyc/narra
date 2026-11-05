# Log: fmt-cli

- Branch feat/fmt-cli reset to origin/feat/fmt (d016f57). Read fmt-doc, fmt-trivia, fmt-printer contracts from origin/main.
- Wrote fmtprinter.nra (Printer, withComments, printDangling, trailingComma, plus helpers), fmtconfig.nra, fmt.nra, the `fmt` dispatch in main.nra.
- The parser rewrites `for (const pattern of xs)` and `return c ? a : b`; the statement printer reverses both (hidden `__for_L_C` name; an `if` whose position is a `?`).
- A ternary cannot span lines without parentheses (parser rule), so `return c ? a : b` is always one line.
- Function-typed fields of Printer match the contract (no effects); allocating functions are accepted there, so no change request.
- Stub expression, pattern and type printers copy the source text by range (fmt.nra, `printExpressionStub` and the two after it).
