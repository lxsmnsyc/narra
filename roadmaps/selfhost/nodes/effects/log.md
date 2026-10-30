# Log: effects

- Nothing existed (parser kept `uses` on declarations only). Built a post-check pass over the checked AST, so types and C are unchanged.
- Function types stay effect-free in the type checker (identity of `FuncType` and C names untouched). Effects of function values are tracked syntactically in check_effects.go.
- A `uses` list inside a parameter list ends at `, name:` (paramAfterComma).
- Newlines inside call parentheses are ignored by the lexer, so block-bodied arrows with several statements cannot be written as call arguments; samples bind them to a const first.
