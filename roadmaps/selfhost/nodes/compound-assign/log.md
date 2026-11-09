# Log: compound-assign

- New Stmt.CompoundAssignStmt(pos, target, operator, value); `x++` is operator `++` with a synthetic IntLit 1.
- Parser: parseBinary leaves `op=` and `++`/`--` alone; parseStmtBody and the c-style-for step call parseCompoundTail. Parser.statementStart marks the one place they are allowed (else E0964).
- Lexer: `x++` / `x--` at a line end now ends the statement (no newline token followed a `+`).
- Checker reuses checkAssignment with an operator parameter; emitter evaluates value, then place once into a pointer temp. Map targets look the key up and panic if missing.
- Brief said reading `m[k]` panics; it returns Option, so the Map rule is stated for the compound form only.
- c-style-for landed mid-work; rebased, step accepts `i += 1`, `i++`.
- Guide chapters 2 and 11 still showed `+=` as an error; edited their blocks.
