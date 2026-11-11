# Outcome: const-generics

## Built

- Constant type parameters `const N: int` on functions and structs (D95, fixed arrays part 1). Logic in `selfhost/src/constgenerics.nra`: `markConstantParameters`, `findConstantParameter`, `checkConstantName`, `findConstantValue`, `isOpenConstant`, `rejectConstantType`, `resolveConstantArgument`.
- Syntax (`parser.nra`): `const N: int` in `parseTypeParams` (the name joins `names`, `constants` and `parser.typeNames`); an integer literal or `-` literal where a type is written is `TypeExpr { name: "3", isConstant: true }` (`parseConstantType`); `rejectTypeArithmetic` after each type argument; both type-argument lookaheads accept integers (`isConstantArgumentToken`), so `f<3>(x)` and `Vector<2> { ... }` parse. The rule is syntactic, as in TypeScript (maintainer, review round 1): an integer literal between the brackets is a type argument whatever the head is, so `both(a < 3, 4 > (c))` is the call `a<3, 4>(c)`; the comparison is written `both((a < 3), 4 > (c))`. The other lookahead rules stay as they were: a name between the brackets must be a type name (type parameter or constant parameter in scope, primitive, or capitalized), and the all-caps-is-a-value rule still applies, so `both(x < MAX, MIN > (y))` stays two comparisons. E0606 (`"a" is a int, not a function`) on such a call adds `if \`<\` is a comparison, wrap it in parentheses, as in \`(a < 3)\`` (`closures.nra`). AST: `FuncDecl.constantParameters`, `StructDecl.constantParameters` (default `[]`, each name also in `typeParameters`), `TypeExpr.isConstant`. astdump prints `constParams=N` and a `const` flag on a constant type; fmt prints `const N: int`.
- Types: `ParamInfo.isConstant`; `Type.TConst(value: int)` is what an instance binds a constant parameter to. Key `K<value>` (`types.nra buildKey`), C id `N<value>` / `Nm<abs>` (`cnames.nra buildConstantId`), folded into the hashed instance and struct names. `typeString` prints the value (`Vector<3>`).
- Resolution: `resolveTypeArguments(checker, written, parameters)` in `resolve.nra` resolves each argument for its slot (both generic branches of `resolveType`/`qualifiedType`, and `takeTypeArguments` in `calls.nra`, which now takes the parameter list). `resolveType` starts with `rejectConstantType`. Inference needs no change: `unify` binds `TParam(N)` to `TConst(3)` through struct arguments and the expected result type.
- Bodies: `checkName` (`exprs.nra`) resolves a constant parameter after locals: type `int`, `Ref.ConstantParameter(value, open)`. The emitter writes `INT64_C(value)` (`emitexprs.nra emitIdent`, `writeConstantParameter`). A local, parameter or pattern binding named like a constant parameter is E0103 (`isNameTaken` and `isConstantParameter` in `check.nra`, used by `declareLocal` and the pattern binder in `match.nra`). The prover reads it as a constant (`operands.nra findNamedConstant`), so divisors, shift counts and comparison bounds work per instance; in the open generic declaration `checkDivisor`/`checkShiftCount` skip a bare `N` (its instances check it).
- E0220 for a constant parameter (`describeUnknownParameter`): `write it, as in f<3>(...)` for a call, `write it, as in Vector<3> { ... }` for a struct literal, `a function with a constant parameter cannot be a value` for a function value.
- An E0979 or E0980 on a constant parameter in an instance body names the value, the instance and its first call: `...; this constant is zero; N is 0 in share<0>, called at src/main.nra:7:45` (`describeConstantOperand`; call sites in `Checker.instanceCalls`, noted by `noteInstanceCall` in `checkGenericCall`).
- Codes: E0992 (kind other than `int`, constant parameter on an enum or alias, arithmetic in a type argument; parser), E0993 (constant where a type belongs, local/type/top-level constant where a constant belongs, literal out of int range; checker). Rows and explanations in `diagnostics.nra`.
- Docs: SPEC.md section 4 "Generics" bullet and sample, D95 amendment in DECISIONS.md, compiler/README.md line.
- Tests: `testdata/const_generics_comparisons.nra` (+ `.out`): comparisons in parentheses beside constant arguments. `testdata/packages/const_generics`: `repeat<3>(x)`, `Text.repeat<2>(x)` and `Grid<4>` from another module. `testdata/const_generics.nra` (+ `.out`): inference from the expected type, explicit `repeat<3>`, `N` as divisor and shift count without a check, several instances, a struct with a constant parameter (also mixed with a type parameter), a negative argument, a constant parameter passed on (`repeat<N>`). Errors `testdata/errors/const_generics_{arithmetic,kind,enum,not_constant,not_type,mismatch,zero_divisor,unknown,shadow,value,comparison}`. AST dump `testdata/astdump/const_parameters`. Emit check `const-generics/distinct-instances` (`selfhost/tests/src/emitchecks.nra`): the C of the sample has 3 names for `repeat`, 3 for `Vector`, 2 for `share`.

### Adding a Type variant (for fixed-arrays and unsafe-ptr)

Every match below is exhaustive now (no `_` fallback), so the compiler stops at each one until the new variant has a case:

1. `types.nra buildKey`: a key with a head no other variant uses (taken: S, E, D, P, A, B, M, F, G, K, the primitive names, `Range`, `module`), for example `X(${element},${length})` for `TFixedArray(element: Type, length: Type)` and `R(${element})` for `TPtr(element: Type)`.
2. `check.nra typeString`: the source spelling (`FixedArray<f64, 3>`, `Ptr<u8>`).
3. `cnames.nra typeId`: the C id, built with `internedName(kind, head, [ids])` so arguments are hashed (`internedName("X", "FixedArray", [typeId(element), typeId(length)])`); `typeId(TConst(3))` gives the length part, so two lengths get two names.
4. `cnames.nra cType`: the C type of a value.

Not exhaustive (they have `_`), but a composite needs a case: `cnames.nra isCachedType` and `cachedCType` (to go through the name cache), `generics.nra unify` (to infer `T` and `N` from an argument), `resolve.nra substitute` (instance bodies and signatures), `check.nra containsTypeParameter` (else a generic signature counts as closed and makes an instance with open types), `exprs.nra containsOpen`, and the layout functions `layoutValue`/`measureValue` in `check.nra`. A length written `FixedArray<T, N>` resolves with `resolveConstantArgument(checker, typeExpr, parameter)`; for a builtin type with no `ParamInfo` slot, call `rejectConstantType` for the element and read the length with `findConstantParameter` or a `TypeExpr.isConstant` literal.

## Deviations

- `selfhost/src/emitexprs.nra` (not in `shared:`): one `emitIdent` case and `writeConstantParameter`. The emitter must write the value of `N`; there is no other way to emit it.
- `selfhost/tests/src/emitchecks.nra` (not in `shared:`): the `const-generics/distinct-instances` case and the `Ranges` import.
- `Ref.ConstantParameter(value: int, open: bool)` instead of carrying a `Type`: a `Type` in `Ref` makes `Ref` a contents type, and every `getRef` caller would need `from checker` (files outside the node).
- `testdata/parse/type_bad.nra` changed from `x: 1` to `x: "a"` (see Verification).
- `selfhost/src/funcvalues.nra`, `selfhost/src/match.nra` and `selfhost/src/closures.nra` (not in `shared:`): the E0606 hint in closures.nra;: the E0220 text through `describeUnknownParameter`, and the shadowing test through `isNameTaken`.
- E0220 got three more message formats for constant parameters (an existing code, not a new one).

## Stubs left

- None. `FixedArray<T, N>` is node fixed-arrays.

## Gotchas

- A top-level constant cannot be a constant argument (`repeat<SIZE>(x)`): SCREAMING_CASE names fail the call lookahead, which stays as it was, and in a type position it is E0993.
- In the open generic declaration `N` is unknown. Only a bare `N` as divisor or shift count is deferred to the instances; `i < N` gives no fact there. fixed-arrays must defer index proofs the same way (`isOpenConstant`), or check them in instance bodies only.
- Only E0979 and E0980 name the instance; other errors in an instance body are reported at the generic body's position as before (appending the instance to every error would undo the de-duplication with the declaration's check).
- `a < 3, 4 > (c)` in call arguments changed meaning: it is now a call with type arguments (TypeScript and C# read it the same way). No migration error, per the no-deprecation rule; E0606 carries the parentheses hint.
- A `requires` callee called with `N` from a generic body is recorded from the open body (contracts skip instance bodies), so it is not proven from the instance's value (E0985; SPEC says so).
- `narrac query` signatures (`querysignature.nra`) print `<N>`, not `<const N: int>`.
- A struct field default that reads `N` is not supported (no instance value there).

## Proposed learnings

- Make every match on `Type` that names or keys a type exhaustive | a `_` fallback gave every new variant the C id "x", so `Vector<3>` and `Vector<4>` would share one C name and one would be dropped silently | selfhost/src/types.nra, cnames.nra | yes (the compiler rejects a non-exhaustive match)

## Verification

- `sh selfhost/bootstrap.sh /private/tmp/claude-501/const-generics/build`: stage 1 and stage 2 emit identical C (178564 lines); seed older than the source, as expected mid-batch.
- Targeted runs with that compiler: `--filter const` (all const-related cases, 298 passed), `--filter distinct-instances` (pass), `--suite sources`, `--suite codes`, `--suite spec`: pass.
- `sh scripts/test.sh` (log /private/tmp/claude-501/const-generics/full2.log): PASS in 262 s, exit code 0, 8 shards (841 to 842 cases each) and the bootstrap job. The first run failed one case, `edge/parse/type_bad`: `x: 1` is now a constant type (E0993 in the checker), so the sample now writes `x: "a"` to keep testing the parse error.

- Review round 1 (syntactic type-argument rule, findings 2 to 6): `sh scripts/test.sh` (log /private/tmp/claude-501/const-generics/full4.log): PASS in 175 s, exit code 0, 8 shards (845 to 846 cases each) and the bootstrap job. The review probe `ambig` now gives E0606 with the parentheses hint; `shadow` reports E0103; `xmod` names `share<0>` and its call.

## Size

- Estimated M, actual M.
