# Log: for-destructure

- Checked first: `ForStmt` held one `Name`, no `Arrays.entries`, no for-destructure samples. All work was missing.
- Took the parser desugaring in both `parser.go` and `selfhost/src/parser.nra`. The checker and emitter treat the result as an ordinary loop plus a declaration, so pattern errors (arity, missing field) come from the existing declaration checks and point at the pattern.
- Hidden loop name: `__for_<line>_<column>` of the pattern start. Unique per loop (nested loops differ), the same in both compilers, and it appears in the C name as `v___for_L_C_N`.
- `let` heads: `for (let x of ...)` and `for (let #[a, b] of ...)` use the same desugaring with a `let` declaration, so the variable can be changed in the body.
- `Arrays.entries`: checker returns `Array<#[int, T]>` (same check function as `Array.values`). The for-of emitter has a `forArrayEntries` kind that builds `{.f_0 = index, .f_1 = item}`. Outside a for-of it has no emit, like `Array.values`.
- Retain of the loop variable uses the local's type (the tuple), not the array element type.
- Error tests print positions of the pattern: `5:14` (tuple) and `10:19` (the bad field).
