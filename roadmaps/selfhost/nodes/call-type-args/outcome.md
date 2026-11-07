# Outcome: call-type-args

## Built

- Explicit type arguments in calls in both compilers: `f<int>(x)`, `Mod.f<int, string>(x)`, `f<Array<int>>(xs)`, `Trait.f<T>(x)`.
- Go: `CallExpr.TypeArgs` (`ast.go`, `clone.go`, `astdump.go`), `callTypeArgsAhead` and `callArgs` in `parser.go`, `call`/`takeTypeArgs`/`genericCall` in `check_call.go`, `traitCall` in `check_traits.go`.
- Selfhost: `Expr.CallExpr` has `typeArguments` after `callee` (parser, astdump, fmtexprs, calls, traits, exprs, check); every other `CallExpr` pattern got a `_`.
- Codes: E0326 (type arguments on a non-generic callee), E0327 (wrong count), E0529 (trait call with Self only in the result and no type argument), with explanations.
- Trait functions may now have `Self` only in the result.
- Tests: `compiler/testdata/call_type_args.nra/.out`, `testdata/astdump/call_type_args.*`, five `testdata/errors/call_type_args_*.nra`.
- Docs: SPEC "Generics" and "Traits", README (F21 removed), guide chapters 3 and 11 (chapter 11 sample is tested).

## Deviations

- Brief says a trait call without `<T>` is an error. It is an error only when the expected type does not fix `Self` (`const dog: Dog = Describe.make(3)` works), the same as `none()` with an annotation.

## Stubs left

- None.

## Gotchas

- `a < b > (c)` stays a comparison only because a lowercase name that is not a primitive type (`int`, `f64`, `bool`, `char`, `string`, `void`) and not followed by `.` rules out type arguments. A new lowercase primitive type needs adding to `isPrimitiveTypeName` (Go) and `isPrimitiveTypeName` (parser.nra); node sized-numbers must do this.
- Builtin functions (std modules written in Go) take no type arguments: E0326.
- The selfhost passes type arguments through `Checker.callTypeArguments` (by call node id), not through `Walker.checkCall`. Whoever takes them empties the entry.
- The E0512 message still says "a parameter ... must use Self"; it now fires only when Self is in neither a parameter nor the result.
- The query site index does not index type arguments (same as literal type arguments).

## Decisions

- Disambiguation by name shape (capital letter or primitive) instead of TypeScript's always-generic reading.
- `Self` inferable from the expected type is accepted without `<T>`.
- E0326 is reported after the callee's own errors, in both compilers.
- Error text for E0327 names the function (`pair`) or `Trait.f`, not the module qualifier.

- Review round: a SCREAMING_CASE name (two or more characters, no lowercase letter) is a constant, not a type, in the lookahead; single capitals stay types. The type parameters of the declaration being parsed count as types, so lowercase `t` works (`Parser.typeNames`, reset per top-level declaration; template-string sub-parsers do not see them, in both parsers).
- The remaining ambiguity (a capitalized, non-SCREAMING value such as `Cap`) is documented in SPEC with the workaround `(Cap > y)`.
- E0326 is not reported when the call already failed with another error. Its format gained a suffix argument: a constructor callee (`Some<int>(1)`) adds a hint to annotate the variable.
- E0512 text now reads "%s.%s must use Self in a parameter or its result, ...".

## Follow-ups (skipped on purpose)

- Cascading `void` errors after E0327 (for example `const #[a, b] = pair<int>(1, 2)` reports more errors).
- Function types inside call type arguments are not accepted (the call parses as comparisons).
- Type argument lists that span lines.
- Grammar: "takes 1 type arguments".
- The builtin that builds channels needs an exception to E0326 when channels land (noted in SPEC).

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test -timeout 45m ./...`: pass (6308 tests in 2 packages).

## Size

- Estimated M, actual M.
