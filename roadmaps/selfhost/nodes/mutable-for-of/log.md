# Log: mutable-for-of

- 2026-11-09: branch feat/mutable-for-of from origin/feat/batch-1 (7d12e22). The `git checkout -b` hit a `.git/config` lock held by the other worker; the branch was made and the index matched it, so HEAD was pointed at it with `git symbolic-ref`. The brief is on main (7784361), not on the story branch; read it from there.
- AST: one field on `ForStmt`, `reference: Option<LoopReference>`, instead of a new statement kind, so every walker keeps visiting the body. The entries form needs a second local, so `LoopReference` carries the index name and node id.
- Parser: `parseReferenceFor` takes `mutable name` or `const #[name, mutable name]` (found by `isMarkedEntries`, 5 tokens of lookahead) and does not desugar into a `VarStmt`.
- Checker: the binding is a new `BindingKind.Reference` local with `isMutable` true and `param` false. `isMutable` gives the capture-by-reference and stack-closure rules for free; `canChangeContents`, E0915 (escape), E0613 (heap closure) and E0958 (whole assignment) check the kind. The source is checked like a `mutable` argument: E0290 for a constant, E0954 for no place, `checkContentsChange` (E0315 for a plain parameter, E0950 for a copying binding).
- Borrow: a stack of `LoopBorrow` in the checker; `checkExpression` compares each whole place (root name plus field path, `[]` for any index) with it. The root ident id is recorded so the inner parts of the same place are not compared again (their shorter paths would overlap falsely, `game` inside `game.round`).
- Emitter: `emitReferenceFor` takes `emitLvaluePointer` of the source, `*p = A_unique(*p)`, then `T* v = &a->items[i]`. No frame entry for the array, no retain or release of the element. `isReference` returns true for the kind.
- After a form error the element type is still taken from the source (array element, or map value), so the body gets no follow-on "void has no fields".
- Orchestrator change mid-task: no full test.sh; ran the touched suites only.
