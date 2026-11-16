# Log: restrict-renames

- Check lives in `reportRename` of `selfhost/src/check/sharedwrites.nra`, called from the `VarStmt` and `AssignStmt` arms of the walk. The report is stored in the walk (`renames`) and sent from `reportFound`, because the walk has a read-only checker. `canWriteInStatement` now lets a body with a bare-name rename through, so such bodies are walked.
- A reported rename returns no shares, so E0978 does not repeat it.
- Skipped: destructuring declarations, names starting with `__` (parser-made loop names), hidden loop locals, type parameters.
- Found 65 sites in `selfhost/src`, 1 in `selfhost/tests/src/lsp.nra`, 2 in `std/src/args.nra` (the brief's 13 counted only typed declarations, not assignments of a pattern binding). All take `copy(...)`.
- About 30 samples under `testdata` tested E0978 through a rename. Each now calls a helper `share<T>(value: T): T from value`, which shares the argument without a rename.
- Seed regenerated twice (compiler sites, then the std text of `args.nra`).

## Review round
- Rebased onto batch-18 (workspace-manifest brought 3 more renames: manifest.nra, workspace.nra x2). Rewrote them by pushing items, a `dropLast` helper in `driver/textops.nra` and an `expandSegment` helper.
- Replaced `copy(...)` where no second value is meant: named-argument matches (`const named`, then `named.value`), generator effect lists (return the payload directly), `is Some` defaults (conditional expressions or helpers), edit-distance rows (rewritten in place), loop swaps (helpers `replaceCharacters`, `mergeRuns`, `findTargetDiagnostics`, `foldTerms`, `dropLast`).
- Kept `copy(...)` where a helper would need a result that shares two parameters (`from` names one), where the copy starts a value that is then written (`fixes.nra` result, `workspace` sorted names), and where an AST or printer cursor starts from a parameter (fmt, fmtexprs).
- `describeTracked` says tuple and coroutine; `m = m` is skipped; a top-level alias and a destructuring declaration are not checked (SPEC says so).
- `testdata/lsp/incremental_shared_writes/src/text.nra` now has `share`, so closing with identical text rechecks nothing again.
