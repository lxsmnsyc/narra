# Outcome: closure-borrow

## Built

- An arrow function passed straight to a non-escaping function-type parameter captures `let` variables and `mutable` parameters by reference, in the Go compiler and in `selfhost/`. Every `let` and `mutable` it uses is by reference, also when only read. Constants keep by-value capture.
- Go checker: `useLocal` in `check_closure.go` finds the chain of arrows that capture a local and requires every one to be a stack closure for a `let` or mutable parameter. `checkBorrowed` in `check_call.go` (called at the end of `args`) is rule 3. `usesLocal` counts a closure's captures as uses, `assignedInExpr` (`check_stmt.go`) looks into arrow bodies, and `narrowed` (`check.go`) returns nothing for a `let` inside a closure that captures it.
- Go emitter: stack environment field is `T*` for a `let` or mutable parameter. It is initialized with `&x`, or with the pointer itself when the variable is already a pointer (`captureValue`). `Emitter.refCaptures` (set in `closureFunc`, saved and restored) makes `localRef` and `lvaluePtr` go through the pointer (`isRef`). Drop functions skip these fields.
- Selfhost: `useLocal` and `isCapturedByArrow` in `funcvalues.nra`; `checkBorrowed`, `findStackClosure`, `capturesLocal` and the `collectAssigned*` family (moved from `stmts.nra`, now exported) in `exprs.nra`; `checkBorrowed` call in `closures.nra`; `captureValue`, `refCaptures` and the pointer declarations in `emitclosures.nra`; `isReference` and `Emitter.refCaptures` in `emitstate.nra`; `emitexprs.nra` uses `isReference`.
- Codes: E0612 (escaping closure captures a `let`), E0613 (same for a `mutable` parameter), E0614 (borrowed variable used in another argument). The E0603 and E0604 rows stay in `diagnostic_codes.go`, unused. E0602 is unchanged.
- Docs: SPEC.md sections 2 (Bindings, Closures, TypeScript differences) and 6 (Mutable parameters, Cycles); DECISIONS.md D77; `compiler/README.md` closures line.
- Tests: `testdata/closure_borrow.nra` and `.out`; `testdata/errors/closure_borrow_*.nra` (escaping const, returned, escaping param, function value, mutable param, stack in escaping, escaping in stack, other arg, mutable arg, field arg, two closures, narrowing after the call, narrowing inside). `errors/closures_capture_let` and `closures_capture_mutable` updated to the new texts (the let test now passes the closure to an `escaping` parameter). `typesdump/closures/err_closures.dump` regenerated.
- `selfhost/src` has no higher-order functions, so there was no loop workaround to convert and none was added.

## Deviations

- `compiler/testdata/typesdump/closures/err_closures.dump` regenerated (outside owned paths, but it holds the old E0603/E0604 texts).
- `docs/guide/03-functions-and-closures.md`, `09-memory-and-mutable.md` and `11-narra-for-typescript-programmers.md` (outside owned paths): the three `narra error` samples carry the new E0612/E0613 text, and the chapter 3 sample uses an `escaping` parameter so it still errors. `TestGuide` checks them. The prose around them ("a closure cannot capture a `let`", chapter 9 "Closures and `let`") is now stale for stack closures; language-guide should revise it.
- Rule 4 is conservative: `assignedInExpr` now enters arrow bodies, so the narrowing of a variable a closure assigns ends before the whole statement that holds the call, and a loop with such a call ends it before the loop.
- `usesLocal` (E0319, E0609) now treats a closure's captures as uses of those variables.

## Stubs left

- None.

## Gotchas

- The E0929 whole-statement scan still does not look into arrow bodies.
- Rule 3 is checked only for calls to Narra functions (`args`). Std builtins with callbacks (Option.map and similar) do not run `checkBorrowed`.
- A `let` captured by reference is `T*` in the closure body; a new emitter path that reads `l.CName` directly for a local must use `localRef`.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test -timeout 45m ./...`: pass after the guide fix (first full run: only the three TestGuide samples failed; `-run TestGuide` then passed; the rest of the suite, including TestSelfhostBootstrap, TestPrograms and TestCompileErrors, passed in that run).

## Size

- Estimated M, actual M.
