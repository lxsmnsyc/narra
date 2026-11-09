# batch-2 log

## Review fixes (branch fix/batch-2-review)

Each finding has a test that fails on feat/batch-2 and passes after its commit.

1. `Cells.update` read the slot without holding it, so a `Cells.set` inside `change` freed the value (`testdata/cell_update_set_inside`). It now holds `const old = get(cell)`, as D90's cost line says. `stdtext.nra` regenerated.
2. Types that differ only in a generator's effects or origin have one C name since `generatorId` dropped them, and arrays, Options, tuples and generic instances over them were emitted twice (`testdata/generator_effects_composites`). Normalizing the type before `checker.composites` would not cover enum and struct instances (Option, tuples), which get one checker index per effect set and must keep it for the effect pass. So `dropDuplicateTypes` (`cnames.nra`), called once at the start of `emitProgram`, keeps the first function, ordered type and composite of each C name.
3. `parser.statementStart` was overwritten by statements in a block-bodied closure (`testdata/compound_closure_index`). `parseStatementExpr` saves and restores it.
4. `CompoundAssignStmt` arms in `collectStatementUses`, `checkMutableConflicts` and `findStatementPos` (`testdata/errors/mutable_conflict_compound_assign`).
5. A call through a function value whose result type is a generator charges `alloc` (`testdata/errors/generator_value_call_alloc`). README "Differences from the spec" and a paragraph in D91 record it.
6. A `;` inside brackets: the parser reports E0003 with the remove fix when its first error is at a `;` outside a C-style for head (`Parser.forHead`) (`testdata/errors/semicolon_in_call`). The remove fix does not give a clean file inside a call (`add(1 2)`), so no `diagnostics` fix test was added.
