# Outcome: local-functions

## Built

- What the spec says: SPEC line 62 only says a function declared inside another function is local to it. It gives no rule for capture, use before declaration or recursion. Closures (SPEC "Closures") capture `const` and parameters, but that section is about arrow functions.
- What I built: non-capturing local functions, by parser desugaring (as for-destructure). `function name(...)` directly in a block becomes a private module-level function `__local_<name>_<line>_<column>` (position of the name). Uses of `name` in the block (and nested blocks, arrows, templates, and the local function bodies) are renamed at parse time. The declarations of a block are found by a pre-scan of its tokens, so a call before the declaration works, recursion and mutual recursion between siblings work, and a local function is usable as a value.
- Capture is a compile error (E0919): an unknown name inside a hoisted function reports `unknown name "x"; local function "f" cannot use the locals of its enclosing function, pass them as arguments`.
- Shadowing errors (E0919): a local `const`/`let`/parameter/for variable named like a local function in scope; a local function named like a module function, constant, struct, enum or import; two local functions of one name in a block (`function "f" is declared twice`, existing code).
- Go: `parser.go` (`block`, `localFuncDecl`, `declareLocalFuncs`, `findLocalFunc`, `notLocalFunc`, `file`, template sub-parser), `check_closure.go` (`checkLocalFuncNames`), `check.go` (call), `check_expr.go` (unknown-name message), `diagnostic_codes.go` (E0919, three formats).
- selfhost: `parser.nra` (same functions, plus exported `localFuncSource`), `declare.nra` (`checkLocalFuncNames`), `exprs.nra` (unknown-name message).
- Tests: `compiler/testdata/local_functions.nra` and `.out`; `compiler/testdata/errors/local_functions_{capture,shadow_module,shadow_local,twice}.nra`.

## Deviations

- `selfhost/src/exprs.nra` and `compiler/internal/narra/check_expr.go` (outside the brief's lists): the capture message hooks the unknown-name report.
- E0919 has three message formats (one code), because the table maps a format to a code.
- Spec gap, no spec change made: capture of locals by a named local function is rejected, not supported. A follow-up node can make it a closure bound to a `const` if the maintainer wants it.

## Stubs left

- Capturing local functions.

## Gotchas

- Error messages about a local function's own signature (wrong arity, type mismatch) show the hidden name `__local_f_L_C`.
- Hoisting is per block: a local function in an `if` block is visible only there. Inside a local function, an error about an unknown name in an arrow within it gives the plain message.
- Hidden names are recognised by prefix and trailing `_<line>_<column>`; a user function named `__local_x_1_2` would be treated as a local function.
- A statement-level `function` is consumed in `block()`, not `stmt()`.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (includes TestPrograms, TestCompileErrors, TestSelfhost* and TestSelfhostBootstrap; 290 s on a loaded machine).

## Size

- Estimated M, actual S.
