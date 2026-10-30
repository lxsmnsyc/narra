# Log: std-option-result

- No `Array.map` builtin exists (it is a user generic in the spec), so callback checking follows `genericCall` + `hintArrow`: hint with parameter types, `Stack = true`.
- Builtin effects are static; added `Builtin.Callbacks` and a loop in `effectWalker.call` rather than a per-builtin hack.
- Dead end: an `(x) => {` block body inline in call parentheses does not parse (newlines ignored inside parens); used a const in the effect test.
- Open doubt: E0909-E0913 may collide with codes taken by parallel branches.
