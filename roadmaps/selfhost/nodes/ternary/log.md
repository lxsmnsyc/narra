# Log: ternary

- Nothing existed (`?` was "unexpected character").
- Chose the parser-only desugar suggested by block-expressions: `BlockExpr{ IfStmt }`. Tail loops handled by rewriting a returned ternary into an `if` of returns in the parser.
- First test run failed: `&&` with an `is` binding is not allowed in selfhost (`is` can bind only as the whole condition); nested the ifs.
