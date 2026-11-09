# Log: c-style-for

- Reset onto origin/feat/batch-2, branch feat/c-style-for. Nothing of the feature existed (the lexer rejected every `;`).
- New AST variant `Stmt.CForStmt(pos, label, init, condition, step, body)`; added cases in parser, checker, effects, idcheck, emitter, fmt, astdump, typesdump, querysites, closures, emittail, trivia and the expression walkers in exprs.
- Lexer emits `;` as a punct token only when `lexer.depth > 0`; at depth 0 it is still E0003. Braces inside parentheses reset the depth, so a block in a call argument keeps the error.
- Emitter: `while (1) { cond; body; continue-label; step }`. An unlabeled `continue` in such a loop is a `goto` to the continue label (LoopInfo.hasStep), because C `continue` would skip the step.
- Codes: E0959 condition not bool, E0960 `const` init, E0961 bad step. Use of the init name after the loop is the ordinary unknown-name error.
- `test` is a keyword in Narra; it cannot name a binding.
