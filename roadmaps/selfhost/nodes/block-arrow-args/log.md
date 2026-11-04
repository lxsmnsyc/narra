# Log: block-arrow-args

- Cause: the lexer counted `(` and `[` depth and dropped line breaks at depth above 0, also inside a `{ }` nested in `( )`.
- Fix: on `{` the lexer saves the depth and sets it to 0; on `}` it restores it. Same in `lexer.go` and `selfhost/src/lexer.nra` (a saved-depth array with a count, since Narra std has no `pop`).
- Sample `compiler/testdata/block_arrow_args.nra`: one arrow, nested call, last of several arguments, a multi-line struct literal inside the block.
