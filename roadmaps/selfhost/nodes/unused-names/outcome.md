# Outcome: unused-names

## Built

- An unused local or parameter is a compile error. E0921: a `const`, `let`, loop variable or pattern binding is never read. E0922: a parameter (function or arrow) is never read. The fix renames to `_name`; a struct pattern shorthand `{ host }` becomes `{ host: _host }`.
- A name that is `_` or starts with `_` is unused on purpose. The hidden `__for_L_C` names of for-destructuring start with `__`, so they never report.
- A plain assignment `x = v` does not read `x`: a `let` that is only assigned is unused, and its fix renames the assignments too. A `mutable` parameter that is only assigned counts as used (an output parameter). The rule is SPEC.md line 64.
- The error is reported at the end of each body (function, test, constant, field default), in source order, at `Local.Pos` (the `const`, `let` or `for` keyword, or the name of a parameter or pattern binding), and only when the body has no other error.
- Entry exemption: only the `args` parameter of the entry function of a binary (`Checker.isEntryFunc` in Go; `Checker.entryTargets` in selfhost). A `run` in a library module or an `implement` block reports.
- Guide: chapter 3 ("A name you do not use") explains the `_` rule with an error sample and its fix; chapter 7 says an unused parameter in an `implement` function starts with `_`.
- The fix of a write-only struct-pattern shorthand (`{ host }` assigned later) rewrites the declaration to `host: _host` and renames the assignments (`errors/unused_written_pattern.nra`). Selfhost has no fixes. `x = x + 1`, `p.x = 1` and `e[0] = 3` count as reads of `x`, `p` and `e` (README).
- New tests: `errors/unused_written.nra`, `testdata/unused_mutable_output.nra`, `testdata/packages/errors/unused_library_run/`.
- Go: `compiler/internal/narra/check_unused.go` (`reportUnused`), `Checker.declared`, `Local.NamePos/Shorthand/Exempt`, `Param.NamePos`, `VarStmt.NamePos`, `ForStmt.NamePos`, `Pattern.Shorthand` (AST fields not shown in the AST dump; copied in `clone.go`), codes in `diagnostic_codes.go`.
- selfhost: `reportUnusedLocals` in `selfhost/src/check.nra`, called from `finishBody`; `BodyInfo.errorCount`. It matches Go's messages and positions. It has no fixes (selfhost has no fix output).
- Tests: `compiler/testdata/errors/unused_{const,let,param,pattern,loop}.nra` (added to `TestHabitFixes`, which applies the fix and recompiles), `compiler/testdata/unused_names.nra` + `.out` (names in the `_` form compile clean, including for-destructuring).
- Cleanup commit (first commit of the branch): applied the fix to every hit in `compiler/testdata`, `selfhost/src`, `bench` and the guide samples (about 215 renames). Goldens (`typesdump/`, `declsdump/`, fmt goldens, `astdump`) were regenerated; the diffs are only renames and column shifts.

## Deviations

- The `args` parameter of a binary's entry function is exempt, and a write-only `mutable` parameter counts as used (orchestrator decisions; the maintainer confirms later). Both are recorded in `compiler/README.md` under "Differences from the spec" (they narrow the rule, so they are differences, not clarifications). The same rules are in selfhost.
- Edited outside the owned paths: `compiler/internal/narra/diagnostic_test.go` (glob of `errors/unused_*.nra` in `TestHabitFixes`), `cnames_test.go` and `std_test.go` (inline sources with unused names), `compiler/internal/narra/{ast,clone,parser,types,check,check_stmt,check_match,check_const,check_types,check_closure}.go` (position and tracking fields), `docs/guide/{05,06,10}-*.md` (three samples used `_`).
- Implement-block function parameters are not exempt (the spec has no exemption).

## Stubs left

- None.

## Gotchas

- Unused imports and private functions (SPEC line 64) are not checked.
- Error tests for an error that is reported outside a function body (signature errors) also get the unused errors of the same file, and `TestCompileErrors` takes the first error: write such tests without unused names.
- In an error test, a rename before the error column on the same line shifts the `// error:` column.

## Proposed learnings

- A parameter or local that nothing reads must start with `_` | the checker rejects it (E0921, E0922) | compiler/testdata, selfhost | yes

## Proposed follow-ups

- Unused imports and private functions as errors (SPEC line 64), with an `Unused import` fix that removes the line.

## Verification

- `cd compiler && go vet ./... && go test -timeout 45m ./...`: pass (734 s), including `TestSelfhostBootstrap` and every `TestSelfhost*` harness.

## Size

- Estimated M, actual M (the rename cleanup was large but mechanical, done with the rule's own fixes).
