# Outcome: c-style-for

## Built

- `for (let i = 0; i < n; i = i + 1) { ... }` and `for (;;)` with TypeScript's meaning (D76), as the AST node `Stmt.CForStmt(pos, label, init, condition, step, body)` in `selfhost/src/ast.nra`.
- Parser: `parseCStyleForHead` and `parseCStyleFor` in `parser.nra`. The `for` head switches to the C-style form when it starts with `;`, or with `let`/`const` and a name followed by `=` or `:`. Labels work (`parseLabeledStmt`).
- Lexer: `;` is a token only inside brackets (`lexer.depth > 0`). Elsewhere it is still E0003.
- Checker: `checkCFor` in `stmts.nra`. The `init` local lives in a scope around the loop. The narrowing-end rule covers names the condition, body and step assign. Walkers in `exprs.nra`, `effects.nra`, `idcheck.nra`, `closures.nra`, `typesdump.nra`, `querysites.nra`, `emittail.nra`, `trivia.nra`.
- Emitter: `emitCFor` in `emitstmts.nra`. An unlabeled `continue` jumps to the step through a `goto` (`LoopInfo.hasStep`, `emitStepJump` in `emitstate.nra`).
- `narrac fmt`: `printCFor` in `fmt.nra`. One line when it fits, otherwise one part per line. `for (;;)`, `for (; cond; step)` and `for (init; cond; )` for empty parts.
- `astdump`: node `cfor` with slots `.init`, `.cond`, `.step`, `.body` (absent parts are omitted).
- Codes (table in `diagnostics.nra`): E0959 non-bool condition, E0960 `const` in `init`, E0961 step that is not an assignment or a call. Using the `init` name after the loop gives the ordinary unknown-name error.
- Tests: `testdata/c_style_for.nra` and `.out`, `testdata/errors/c_style_for_{condition,const,step,semicolon,scope}`, `testdata/fmt/stmt_c_style_for`, `testdata/astdump/c_style_for`, `testdata/lex/c_style_for_semi`.
- Docs: SPEC.md section 2 (statements, control flow, differences table), D45 and D76 in DECISIONS.md, `compiler/README.md`.

## Deviations

- `init` takes one `let` declaration with an optional type: no destructuring, no comma list.
- Absent `init` and `step` are omitted from the astdump, like other absent slots.
- A `;` inside `(`, `[` or `#[` that is not a `for` head now reaches the parser, which reports a generic "expected" error instead of E0003.

## Stubs left

- `++`, `--` and compound assignment as the step: node compound-assign. `parseCStyleFor` reads the step as an expression, then `=`; compound-assign must extend it.
- `selfhost/src` does not use the new syntax (seed rule). The seed was not regenerated.

## Gotchas

- `test` is a keyword; it cannot name a binding.
- A new statement kind needs cases in every walker. Several use `_ => {}`, so a missing case compiles and silently skips the loop. Search for `Stmt.WhileStmt`.
- `continue` in a C-style loop must not become C `continue`; it would skip the step.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh`: stage 1 and stage 2 emit identical C.
- `--filter c_style_for`: programs (debug and release, sanitizers and leak check), errors, fmt, astdump and lex pass.
- `sh scripts/test.sh`: passes (see the commit message and the final report).

## Size

- Estimated M, actual M.
