# Log: nested-void-match

- Base feat/batch-22 (5c2b15ab), built with `sh selfhost/bootstrap.sh <scratch>/base`.
- Reproduce: `<base>/narrac run a.nra` (the brief's program) printed `error: variable has incomplete type 'void'` at `void r8;`, then `C compiler failed: exit status 1`.
- Cause: `emitMatchValue` in `selfhost/src/emit/emitmatch.nra` declared a result temporary of the match type before emitting arms, also for a void match. An arm body that is a match reaches it through `emitExpression`.
- Other shapes: `? :` with void branches and `if` as an expression are rejected by the checker ("both branches of a conditional expression must have a value"; no `if` expression syntax). `const x = <void match>` is rejected ("this expression has no value"). A block arm `{ ... }` of a void match works. A block arm inside the inner match in expression position is rejected by the checker. So only the nested match was affected, at any depth.
- Fix: `emitMatchValue` emits the arms as statements and returns empty code when the type is void.
- Case: testdata/nested_void_match.nra (the brief's program, 3-deep nesting, a block arm around a void match, a non-void nested match). Base: three `void rN;` errors. Branch: runs, output matches the .out.
