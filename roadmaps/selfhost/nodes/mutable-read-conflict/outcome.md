# Outcome: mutable-read-conflict

## Built

- E0929 "%q is passed as mutable in one call, so the rest of the expression cannot use it", at the other use (the read, the assignment target or the second `mutable` argument). Within one statement's expression (initializer, assignment, expression statement, return value, `if`/`while` condition, `for` iterable), a local passed as `mutable` cannot be read, assigned or passed again at any depth, including `p.x` and `xs[0]`.
- Go: `checkMutableConflicts`, `useCollector`, `hoistFix` in `compiler/internal/narra/check_call.go`, called from `Checker.stmt` in `check_stmt.go`.
- Selfhost: `checkMutableConflicts`, `collectUses`, `collectRootUse` in `selfhost/src/stmts.nra`, called at the end of `checkStatement`. No fix there (selfhost has no fixes).
- Fix (Go, `--json`): binds the call to a fresh `const <callee>Result` on the line before the statement. Offered for var, assignment, expression and return statements, when the call is on one line and returns a value and is not on the right of `&&`/`||` or in a match arm.
- The scan descends into match arms (expression and block bodies, statement by statement). A read in an arm of another argument of the call is E0929 (checkExclusive does not look into arms). Exempt inside an expression: sibling arms, a match subject against its arms, an `if`/`while`/`for` header against its body, `then` against `else`, and statements one after the other in a block. Test: `errors/mutable_read_conflict_arm.nra`; the legal sibling-arm case is in `mutable_read_ok.nra`. Not done: `BlockExpr` of #71 (not on main).
- Exempt: the two sides of `&&`/`||` and different match arms (the emitter sequences them), pairs inside one call's arguments (E0319), a callee against a mutable argument inside it (E0609).
- SPEC.md section 6 states the rule with an example.
- Tests: `compiler/testdata/errors/mutable_read_conflict_{operand,nested,option,field,template,twice,loop,match}.nra`, `testdata/diagnostics/mutable_read_conflict_fix.nra` (fix applied, file must compile), `testdata/mutable_read_ok.nra` (+ `.out`: different variables, `&&`, match arms).
- First commit: `selfhost/src/parser.nra` bound two `newId(mutable parser)` calls to consts (the only existing hit).

## Deviations

- `compiler/internal/narra/ast.go` (`CallExpr.End`, `File.Src`), `parser.go` (sets them) and `clone.go` (copies `End`) are outside the brief's paths. The fix needs the call's text range.
- `selfhost/src/parser.nra` (the first commit) is covered by the brief's `selfhost/src/**`.

## Stubs left

- No fix in selfhost output.

## Gotchas

- The option case (`add(x.value, clear(mutable x))`) reports the narrowing error first and E0929 after it, because #68 ends narrowing before the statement is checked. Its tests expect that order.
- The scan covers statement expressions only. The expression body of an arrow function is not scanned.
- Assignment `x = f(mutable x)` is now an error (the target counts as an assignment).
- Template string parts all carry the template's position, so a read inside a template is reported at the template.
- `uses` is a keyword in Narra, so it cannot be a field name.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test -count=1 ./...`: pass (201 s), including `TestSelfhostBootstrap`, `TestCompileErrors`, `TestHabitFixes` and the `TestSelfhost*` harnesses.
- `bench/` programs with `mutable` (interp, particles) and the other bench programs check clean.
- Self-review tries (nested call two levels deep, template string, match subject against an arm read, `if` conditions, calls in a `while` body): the expected lines were reported, `&&` and separate arms stayed legal.

## Size

- Estimated S, actual M. The exemptions and the fix needed more than the rule.
