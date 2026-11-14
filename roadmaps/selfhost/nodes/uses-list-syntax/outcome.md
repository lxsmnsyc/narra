# Outcome: uses-list-syntax

## Built

- D117. In a type (function type, `Generator`, `Coroutine`, `AsyncGenerator`), `uses` takes one effect or a parenthesized list with an optional trailing comma: `parseTypeUses` in `selfhost/src/parser.nra`. A declaration's own list stays the bare loop `parseUsesList`; `uses (io)` there is the plain `expected a name, found "("`. `isParamAfterComma` is gone.
- E1240 (parser, `reportTypeUsesComma`): a `,` right after `uses name` of a type, followed by a name, after a declaration's result type or before a parameter name with no `:`. No fix.
- E1241 (checker, `resolve.nra`): `reportEffectArgument` at the top of `resolveType` reports an effect name (`io`, `alloc`, `unsafe`, an effect variable of the function being checked, only when no type has that name) that follows a type ending in a `uses` list, with ``write `uses (io, alloc)` ``. `reportUnknownType` (exported) gives E1241 without the hint for any other effect name used as a type, else E0118; `traits.nra` uses it for implement targets.
- Display: new module `selfhost/src/typeuses.nra`, `writeTypeUses(effects)`: `""`, ` uses io`, ` uses (io, alloc)`. Used by `check.nra` (generator and function types, so every message), `declsdump.nra`, `query.nra` (type-of of named functions and arrows), `querysignature.nra` (`writeType`), and fmt's `printTypeBody` and `printFunctionType`. `usesSuffix` is removed: all its callers were types. Declaration lists (interface line, hover `uses a, b`, query `effects:`) never used it and stay bare.
- fmt: `endsWithUses`, `canEndWithComma` and the trailing-comma parameter of `printParameterList` are gone, so a broken list always ends with a comma. `uses (io)` prints as `uses io`.
- E1242 (parser, `reportRepeatedEffect`, called from `parseTypeUses` and `parseUsesList`): an effect named twice in a `uses` list, in a type (`uses (io, io)`) or a declaration (`uses io, io`). No source in the repo had one.
- E0803's hint (`effects.nra`) writes the type form through `writeTypeEffects`: ``write `uses (alloc, io)` in the type`` (`testdata/errors/uses_list_argument_hint.nra`).
- Codes rows and `narrac explain` texts for E1240, E1241 and E1242 in `diagnostics.nra`.
- Tests: `testdata/uses_list_types.nra` (+ `.out`), `testdata/parse/uses_list_forms.nra`, `testdata/errors/uses_list_{result_comma,param_comma,effect_type,effect_variable,decl_parens,duplicate,decl_duplicate,argument_hint}.nra`, `testdata/fmt/type_uses_parens_w40.nra`, `testdata/query/uses_lists/`.
- Rewritten: 44 type lists in 34 files before the rebase, plus 6 after it (`std/src/async__net.nra`, `testdata/for_await_*.nra`) (`registry.nra` 2, the rest in `testdata`), found with base `narrac parse` (any `type` node whose `uses=` holds a comma). Docs: SPEC section 7 and the generator section lose the "not built yet" notes; `compiler/README.md` Effects and Generators lines; guide chapter 10 gets a paragraph and a checked example.

## Proof

- base: `feat/batch-13` a2c7f01d, `narrac run testdata/uses_list_types.nra` with a release build of the base: `uses_list_types.nra:8:52: expected a name, found "("`, exit 1.
- branch: the same command prints the checked output (`uses_list_types.out`), exit 0; clean in the programs (sanitizers), leaks and copies suites.

## Deviations

- The sample cannot hold `Map<Generator<int> uses (io, alloc), int>`: a generator is not hashable, so it is not a map key. The sample puts the generator first in a user `Pair<Generator<int> uses (io, alloc), int>`, and the Map form is in the parse test. SPEC's example of the form keeps the Map; it parses but does not type-check.
- `std/src/async.nra`, `coroutine.nra`, `ext__array.nra`, `option.nra` and `thread.nra` (not in the brief's paths) are reformatted, and `async__net.nra` has one type list rewritten to `uses (io, alloc)`: fmt now writes the trailing comma after a last parameter ending in `uses E`, and the sources suite checks them. This is safe for the seed: the seed uses its own embedded std text and never parses `std/src`; `stdtext.nra` was regenerated with `--suite generated --update`. `selfhost/src` needed no such change.
- `selfhost/tests/src/fmtstable.nra` (not listed): the fmt-stable token comparison now treats the parentheses after `uses` as grouping, since fmt drops them for one effect.
- `roadmaps/selfhost/inbox.md`: one line for the comment that moves inside the parentheses, as the brief asked.

## Stubs left

- None.

## Gotchas

- The branch also fixes the trailing comma after `uses E` in a parameter list that threads-check hit.

- E1240 looks only at the tokens `uses name , name`; `uses (io), alloc` after a result type is the ordinary `expected "{"` parse error.
- E1241's effect-variable test sees only the effect parameters of the function whose signature or body is being checked (`checker.functionIndex`), the same binder rule as generator types.
- A comment inside `uses ( ... )` moves to the type arguments, like the old `uses /* c */ io` bug (inbox).

## Proposed learnings

- A list inside a type is closed by its own delimiter when the type can sit in another comma list | `uses io, int` inside `Map<...>` was read as one list | selfhost/src/parser.nra | yes (parse dumps)
- The seed never parses `std/src`; it uses its own embedded std text, so std source may use new syntax once `stdtext.nra` is regenerated | the trailing commas fmt now writes in std/src | std/src | no

## Verification

- `NARRAC=<branch release build> sh scripts/test.sh`: PASS, 9086 cases (before the rebase).
- `narrac emit selfhost` on the same tree with the base and the branch compilers: byte-identical (232953 lines).
- `sh scripts/test.sh` after the rebase on feat/batch-13 9284717b with the review fixes (narrac built from the seed): PASS, 9344 cases. The emit comparison was made before the rebase; the rebase brought no change to the files this node touches beyond merged rows.

## Size

- Estimated M, actual M.
