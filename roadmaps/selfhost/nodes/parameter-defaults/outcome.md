# Outcome: parameter-defaults

## Built

- D115: constant defaults on the trailing parameters of declared functions (`function`, `function*`, `async function`, `@inline`, `@const`, exported or private). A call leaves a default out by passing fewer positional arguments or by naming later ones.
- Syntax: `Param.default: Option<Expr> = None` in `selfhost/src/ast.nra`, parsed after the type in `parseFuncHeader` (`parser.nra`). astdump writes a `.default` slot under the param, fmt prints `name: T = value`, declsdump flags the param `default`.
- `selfhost/src/paramdefaults.nra` holds the feature:
  - `checkDeclarationDefaults` (called from `checkDefaults` in `body.nra`, before the module's bodies): E1210, E1211, E1214 (trait and implement blocks), E1215 (extern blocks); each default checked as a body against the parameter type (type parameters open), E1212 through `isConstantValue`. A default that holds gets the body key `paramdefault|func|<funcIndex>|<parameter index>`.
  - `checkFilledDefaults` (from `checkFunctionCall` in `calls.nra`): each default a call leaves out is checked again as a body keyed `paramdefault|<caller body>|<call id>|<parameter index>`, in the callee's scope, with the call's parameter types (the instance's for a generic callee). Only defaults whose declaration check held are filled, so a broken default is reported once.
  - Helpers for the later passes: `findPassedParameters`, `fillArguments` (every argument in parameter order, defaults included), `findFilledParameters`, `findWrittenIndexes`, `findCallDefaultBody`, `leavesOutDefaults`, `countRequired`, `writeDefaultValue` (constant text for hover and query).
- Calls (`calls.nra`): `matchNamedArguments` accepts a missing parameter with a default and reports E0316 with the new row `this call passes %d arguments, the function takes %d to %d` for a function with defaults (positional too few or too many); E0331 stays for a named call that skips a required parameter. `argumentOrders` keeps one entry per written argument, so `orderArguments` and every existing reader are unchanged. Typing (`checkArguments`, `checkGenericCall`) sees only the passed parameters, so a left-out argument gives no inference evidence. `recordContractCall` gets the filled list, so `requires` sees the default.
- Emitter (`emitDefaultsCall` in `emitexprs.nra`): each filled default is emitted from its call body into `emitter.heldArguments`, then the plain call, the instance call or the `@inline` copy emits the full list and picks the held code up. A nested call that fills the same default saves and restores the outer entry.
- Effects (`walkFilledDefaults` in `effects.nra`): a filled default is walked in its call body and its effects (`alloc` for an array literal) are charged at the call, via "the default of <name>".
- Constant evaluator (`consteval.nra`, `fillDefaultValues`): an `@const` call evaluates its left-out defaults from the call bodies.
- Parser errors: E1213 for a default on an arrow function parameter and in a function type, E1216 for `name?: T` with the fix `name: Option<T> = None` (three edits, in `check --json`; `apply-fix` gives a file that checks clean). `name?: T = value` gets E1216 with no fix, and an arrow's `name?: T` gets the second E1216 row, `an arrow function parameter cannot be optional; write "name: Option<T>"`.
- Hover and `narrac query type-of` show `name: T = value` for a parameter with a default; `narrac query interface` writes the same.
- Function values: `const draw = render` has the full type with no defaults; a call through it that leaves an argument out is E0316 (unchanged code path, covered by a test).
- SPEC: "not built yet" dropped from the bullet. Guide chapter 3 has "Arguments a caller may leave out"; chapter 11's sentence now says defaults exist and are constants. `compiler/README.md` lists the feature.
- Tests: `testdata/param_defaults.nra` (positional, named skip, reordered named, string, bool, `None`, `Some(1)`, enum constructor, `[]`, unary, generic with and without the argument, two instances of a generic default in one body, a nested call filling the same default, `@inline`, `@const` at compile and run time, a generator, a self-call that fills a default, `requires` on a default, a function value), `testdata/param_defaults_effects.nra`, `testdata/packages/param_defaults` (an exported function with defaults called from another module, by namespace and by named import), 15 error tests `testdata/errors/param_defaults_*`, the tail-call sample `testdata/param_defaults_tail.nra`, goldens `testdata/astdump/param_defaults`, `testdata/declsdump/param_defaults`, `testdata/fmt/decl_param_defaults`.

## Proof

- base: 82e54d0b (feat/batch-13), compiler built from `git archive feat/batch-13`; `narrac run testdata/param_defaults.nra` reports `param_defaults.nra:9:42: expected ")", found "="`, exit 1.
- branch: the same command prints the 24 lines of `testdata/param_defaults.out`, exit 0, also with `-sanitize`.

## Deviations

- Files outside the brief's owned and shared paths:
  - `selfhost/src/consteval.nra`: `callFunction` fills defaults for `@const` calls (`fillDefaultValues`), and `Param` is imported. Without it an `@const` call that leaves out a default gets too few values.
  - `selfhost/src/querysignature.nra`: `writeParameter` adds ` = value`, for `narrac query interface`.
  - `selfhost/src/emittail.nra`: a tail self-call that leaves out a default stays a jump (review fix; SPEC section 2 makes every tail self-call a jump).
  - `testdata/packages/param_defaults/`: the cross-module sample needs a package; a single-file sample cannot import another module.
  - `testdata/fmt/decl_param_defaults.{nra,fmt}`: the fmt suite only collects goldens named `decl_*` and the like, so `testdata/fmt/param_defaults*` would never run.
- E1213 and E1216 are reported by the parser, not the checker (the brief's step lists them under the checker). Arrow `mutable` parameters are a parser error already (E0028), a function type has no `Param` to hold a default, and `name?: T` has no AST form; the parser keeps the first error only, which the one-error test files allow.
- No side table by call id for the filled parameters: they follow from the written arguments and the callee's parameters (`findFilledParameters`), so `Checker` (in `check.nra`, not a shared file) is unchanged. The per-call default bodies play that role for the later passes.
- `param_defaults_not_constant` reports `unknown name "width"` beside E1212 for a default that reads a parameter: the default body has no locals, as for a struct field default that reads a name.

## Stubs left

- None.

## Gotchas

- A default body exists for each call that leaves the default out (kind "default", name `<function>.<parameter>`); typesdump and other body listings show them for programs that use defaults.
- A tail self-call that leaves out a default is a jump like any other: `emittail.nra` matches it on the filled argument list and builds each left-out default from the call's default body after the written arguments (`testdata/param_defaults_tail.nra`, 50,000,000 calls).
- The emitter relies on `heldArguments`: anything that emits a call's arguments must go through `EmitWalker.emitExpression` for the default to appear. A new call path that reads `orderArguments` instead of `ParamDefaults.fillArguments` sees only the written arguments.
- Confirmed separate bug (not fixed here): an omitted struct field default with an array literal is not charged `alloc`. `export function make(): Bag { return {} }` with `struct Bag { items: Array<int> = [1, 2] }` compiles without `uses alloc`, while `return { items: [1, 2] }` is rejected. `Expr.StructLit` in `effects.nra` walks only the written fields; the fix can copy `walkFilledDefaults` (walk the default in the `default|...` body).
- The compiler cannot use defaults until the batch's seed update; the adoption sweep (`checkExpression(..., None)`, `argumentText(..., 0)`) is a later node.

## Proposed learnings

- none

## Verification

- After the review fixes and the rebase on feat/batch-13 (4b7a6dfa):
  - `NARRAC=<release branch compiler> sh scripts/test.sh --filter param_defaults`: PASS, 107 cases (samples under AddressSanitizer, the leak check and the copy check).
  - `testdata/param_defaults_tail.nra`: overflows the stack under ASan (`f_..._down`) with the c825d80e compiler; passes on the branch.
  - `narrac emit selfhost` with the base and the branch compilers: byte-identical, for the branch sources and for the base sources.
  - `sh scripts/test.sh` (built from the seed): PASS, 8905 cases in 348s.

## Size

- Estimated M, actual M.
