# Outcome: contracts

## Built

- `requires <condition>` lines between a function's signature and its body. AST: `FuncDecl.requires: Array<Expr>` (`selfhost/src/ast.nra`, default `[]`). Parser: `parseRequiresLines` in `parser.nra`; `requires` is not a keyword, it counts only at the start of a line after the signature of a function with a body (top-level functions and `implement` members).
- Checking lives in `selfhost/src/contracts.nra`:
  - `collectModuleRequirements(checker, file)` (declaration phase, see the review fixes) normalizes each condition into `Requirement { parameter, comparison, bound, boundText }` (`check.nra`) in `checker.requirements[funcIndex]`.
  - `checkRequires(checker, walker, funcDecl, funcIndex, isInstance)`, called by `checkFunctionBody` (`body.nra`) right after `declareParams`. On the declaration's own check it type-checks each condition as `bool` and reports E0984 for any condition outside the supported form. Then, for the declaration and for every generic instance, it seeds each requirement as a fact on the parameter's local with `refineFact` (`operands.nra`), the same store `if` narrowing writes (`checker.localNames` under `factKey(local)`). So the body can divide by a `requires d != 0` parameter, and the fact ends at an assignment like any narrowing.
  - `recordContractCall(checker, funcIndex, orderedArguments)`, called by `checkFunctionCall` (`calls.nra`, both the plain and the generic path) after the arguments are checked. For a callee with `requires` it stores what is known about each argument right then (`findOperandFact` in `operands.nra`: a constant wrapped to the parameter type, or the narrowed fact of a plain integer name not captured by an arrow) as `ContractCall { file, funcIndex, arguments: Array<ContractArgument { pos, low, high, nonzero }> }` in `checker.contractCalls`. Calls inside generic instance bodies are not recorded (the generic body recorded them).
  - `checkContractCalls(checker)`, called in `checkModules` (`body.nra`) after `checkPending`: proves every recorded argument against the callee's requirements with `isComparisonProven` (`operands.nra`), and reports E0985 at the argument. Deferring the proof makes it independent of the order bodies are checked in (callee later in the file, recursion, other modules).
- Supported conditions: one integer parameter compared (`==`, `!=`, `<`, `<=`, `>`, `>=`, parameter on either side) with an integer constant (`foldConstant`: literals, negation, top-level `const` built with `+ - * <<`), joined by `&&`, on one or more lines. A u64 bound from 2^63 up is rejected.
- Codes: E0984 (unsupported condition: float or other non-integer parameter, field path, call, index, `||`, two parameters compared; also `requires` on a function of an `implement` block), E0985 (unproven argument), E0986 (a function with `requires` used as a value, in `funcvalues.nra` `checkFunctionValue`). `typeof` of such a function is E0982 with its own message (`resolve.nra` `resolveFunctionTypeof`). Rows in `diagnostics.nra`.
- `operands.nra` now exports `Fact`, `findFactLocal`, `findTypedConstant`, `flipComparison`, `isComparisonOperator`, `refineFact`, and adds `findOperandFact` and `isComparisonProven`.
- `fmt` prints each condition on its own line indented one step under the signature, and the body's `{` on the next line at the declaration's indent (`printFunction`, `fmt.nra`). `astdump` prints a `.requires` slot between `.result` and `.body`.
- Review fixes (second commit):
  - Parsing: `isRequiresLine` also accepts a `requires` on a later line than the previous token, since the lexer emits no line break after `>` (`Option<int>`, `Generator<int>` result types).
  - Requirements are now read in the declaration phase: `collectModuleRequirements(checker, file)` runs in `checkModules` right after `declareSourceModule`, for every module including the ones `narrac lsp` skips. It works from names and constants alone (`foldConstant` in `contracts.nra`: literals, `-`, `+`, `*`, `<<`, own and imported top-level constants, `Module.NAME`). The body check type-checks the conditions and reports E0984 with the same collector (`report: true`), so the two cannot disagree. Case `testdata/lsp/incremental_requires`.
  - `isComparisonProven` never proves from the unbounded ends (`UNBOUNDED_LOW`, `UNBOUNDED_HIGH`). A u64 constant from 2^63 up is the fact `[UNBOUNDED_HIGH, UNBOUNDED_HIGH]`, non-zero: above every bound, never `<`, `<=` or `==`.
  - `parseLiteral` (operands.nra) now reads a u64 literal from 2^63 up (it failed in `Int.parse` before). This changes proven-operands too: `x / 18446744073709551615` on u64 was E0979 and is now accepted. The existing proven-operands samples keep their results.
  - E0985: one error per argument, naming every condition it fails joined by `&&`; a constant argument gets `0 does not satisfy \`requires divisor != 0\` of share`. A bound written as a constant name is printed by name (`Requirement.boundText`).
  - fmt: each `requires` line is a node of its own in the trivia (key `q<condition id>`), so a comment above it stays above it. A comment between the last `requires` and `{` is printed at the indent of the `{` (column 0 for a top-level function), like any comment before a block. Case `testdata/fmt/comment_requires`.
- Review round 2 (third commit): `foldConstant` takes the function's parameter names, so a parameter named like a module constant or an import alias is the parameter (E0984 in both phases), and a bound that does not fit the parameter's type (`v < 255 + 1` on a u8) is E0984 "the bound does not fit the parameter's type" instead of wrapping. Samples `testdata/errors/contracts_names`, `testdata/packages/errors/contracts_alias_parameter`.
- Docs: SPEC.md section 13 contracts paragraph rewritten to D95 with an integer example, and the typeof E0982 bullet; D95 records what was built; compiler/README.md line; guide chapter 3 section "Conditions a caller must meet".
- Tests: `testdata/errors/contracts_extremes` (int range ends), `testdata/contracts_requires.nra` (u64 constants from 2^63 up, `Option<int>` and generator results, a callee declared after its caller and calling itself, constants, named constant bound, narrowed names, the caller's own `requires`, named arguments, u32 shift by a `requires` count, body dividing by a `requires d != 0` parameter); `testdata/errors/contracts_unproven`, `contracts_unsupported` (float, field, `||`, two parameters, call), `contracts_value` (value and typeof), `contracts_implement`; `testdata/astdump/requires_lines`.

## Deviations

- `selfhost/src/idcheck.nra` (not in `shared:`): `collectFuncIds` collects the condition ids. Without it the body's id range misses them and the checker's side tables cannot hold their types and refs.
- `check.nra` gained two fields (`requirements`, `contractCalls`) and three small structs (`Requirement`, `ContractCall`, `ContractArgument`). They must sit there because `contracts.nra` imports `check.nra`.
- `selfhost/tests/spec-skip.txt` (not in `shared:`): removed the skip line of the old SPEC contracts sample, as the file asks. build-conditions removes the two lines just above it, so the two branches may conflict there on merge; keep both deletions.
- `requires` in an `implement` block is an error (E0984), not in the brief: a trait call carries no condition, the same reason as function values.

## Stubs left

- None. `ensures` is node contracts-ensures.

## Gotchas

- For contracts-ensures: facts are the `Fact { low, high, nonzero }` of `operands.nra`, written with `refineFact(checker, local, comparison, bound)` and read with `findOperandFact(checker, expression, type)`. A result fact would be seeded on the local that receives a call result; the caller side of `requires` is in `recordContractCall`/`checkContractCalls`.
- An argument passed as `mutable n` loses its fact before the call is recorded (passing as mutable ends narrowing), so such a parameter can only be proven by a constant.
- E0985 is reported after every body, so it comes after the other errors of a file in the output.

## Proposed learnings

- None.

## Verification

- `NARRAC=<stage 1> sh scripts/test.sh --suite errors|dumps|programs|spec|guide`: pass.
- `sh scripts/test.sh`: PASS after the review fixes (520 s, machine shared with other workers; first run 314 s); stage 1 and stage 2 emit identical C, the seed is older than the source as expected.

## Size

- Estimated M, actual M.
