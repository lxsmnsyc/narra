# Log: block-expressions

- Checked first: no block expression, `BlockExpr` or `if` value existed. `primary` already took `{` as a context-typed struct literal.
- Design: D75. `if` as a value is desugared in the parser (`BlockExpr` of an `if` whose branches return `BlockExpr`s), so only `BlockExpr` is new in the checker and emitter.
- `{` ambiguity: `{}`, `{ ...`, `{ name:` stay struct literals; a label (`name: for`) starts a block.
- Emitter: C block with `rN` and `narra_block_N:;`; `return` stores, unwinds frames and scopes to the block's start depth, `goto`.
- Number of codes: E0931, E0932, E0933 (bare `return` and void `return` reuse E0931).
- First full run of `go vet ./... && go test ./...` after the selfhost port passed, bootstrap included.
