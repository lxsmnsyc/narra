# Outcome: array-patterns

## Built

- D103 part 1. `Pattern.PatArray(id, pos, elements, rest: Option<Pattern>)` in `selfhost/src/ast.nra`: `[]`, `[x]`, `[a, b]`, `[first, ..._]`, `[first, ...rest]` in `match` arms, `is` tests, `const`, `let` and `for ... of` heads, over `Array<T>` and `FixedArray<T, N>`. Elements nest and take any pattern, literals and or-patterns included.
- Parser (`parseArrayPattern` in `parser.nra`): `[` starts a pattern, and `const`, `let` and `for (const ...)` heads accept it. Anything after the rest is E0722.
- Checker logic in the new module `selfhost/src/arraypatterns.nra` (imports narrac/check only): `resolveArrayShape` (element type; E0723 known length, E0724 named rest on a FixedArray, E0726 other types), `isArrayShapeIrrefutable`, `checkDestructureLengths` (E0723 over an open `N`), `findArrayPattern` (E0987 in `ensures`), `isInstanceBody`, and the length constructors for the matrix (`isLengthColumn`, `findLengthLimit`, `countLengthConstructors`, `listLengthTypes`, `specializeArray`, `describeLength`).
- `patterns.nra`: lengths in `specialize`, `isUseful` (`isLengthUseful`) and `findMissingCases` (`findMissingLengths`); `[..._]` and `[...rest]` count as wildcards; `isIrrefutable` reads the type `resolvePattern` records per pattern id, so nested FixedArray patterns are judged by their own type. Missing cases print as patterns (`[_]`, `[_, _, _, _, ..._]`).
- Emitter helpers in the new module `selfhost/src/emitarraypatterns.nra`; `emitmatch.nra` dispatches. One length test (`->len == k` or `>= k`), element reads `->items[i]` / `.items[i]` without a bounds check, and the named rest as a new Array (`_new(len - k)` plus a retain loop) owned by its local in every position, or-patterns and guarded or-patterns included (a failed guard releases it).
- Cases in `astdump` (`pat-array` with a `.rest` slot), `typesdump`, `idcheck`, `trivia`, `querysites`, `contracts` and `fmtexprs` (printed like an array literal, no trailing comma after a rest).
- Diagnostics rows: E0722, E0723 (two formats), E0724, E0726, and a new E0987 format.
- Samples: `testdata/array_patterns.nra` (lengths, `[first, ...rest]` recursion, string arrays, literal argv dispatch, guard, or-patterns of array patterns, guarded or-pattern with rests, `Some([head, ...tail])`, `#[[x], y]`, nested arrays, `is` with and without bindings, `const [...copied] = items`) and `testdata/array_patterns_instance.nra` (a FixedArray in a struct and a tuple matched inside generic functions) and `testdata/array_patterns_fixed.nra` (FixedArray in `const`, `let`, `for ... of`, inside tuple and struct patterns, `FixedArray<int, 0>` with `[]`, a generic `N` match, literal elements).
- Error tests `testdata/errors/array_pattern_*.nra` (15): missing lengths (two), guarded-only cover, unreachable arm (E0983), rest twice and not last (E0722), FixedArray length in `const` and in `match` (E0723), `const [a, b]` over a generic `N` (E0723), named FixedArray rest (E0724), `string` and `Map` (E0726), `const [a, b] = items` and `for (const [a, b] of rows)` (E0715), `ensures` (E0987).
- Goldens: `testdata/astdump/array_patterns.{nra,ast}`, `testdata/fmt/pattern_array_w40.{nra,fmt}`.
- SPEC.md: section 2 (Array destructuring), section 4 (Array patterns bullet), section 6 (A named rest copies). DECISIONS.md D103 has a "Built" line with four amendments. `compiler/README.md` lists the feature.

## Proof

- base: feat/batch-11 20f9a487, `narrac run` on a `match (items) { [] => ... }` sample: parse error, found "[".
- branch: `testdata/array_patterns.nra` and `array_patterns_fixed.nra` run clean under `-sanitize`, the leak check and the copy check (programs suite).

## Deviations

- Two new files outside the owned and shared paths: `selfhost/src/arraypatterns.nra` and `selfhost/src/emitarraypatterns.nra`, to keep the feature's logic out of the shared files (as the brief's rules ask). No other file outside the lists changed.
- In a generic instance body, no `match` is judged again for unreachable arms or exhaustiveness (`judgesArms = !isInstanceBody(checker)` in `checkMatch`), and a FixedArray length mismatch is not E0723 there. Without this, every instance repeated the generic body's coverage error with its own `N`, and arms for other lengths were E0983 in each instance. The declaration's own body judges every match, against the open length where `N` is a parameter. A first version limited the rule to subjects that hold a FixedArray, read through type arguments only, and rejected a FixedArray field in a struct matched inside a generic function (fresh review; `testdata/array_patterns_instance.nra`). Recorded in D103's Built line.
- E0722 is a parser error, and a trailing comma after the rest is E0722 too.

## Stubs left

- None.

## Gotchas

- `let [...rest] = items` binds a read-only rest (D81 destructured bindings), so `Arrays.push(mutable rest, x)` is the existing "destructured from a copy" error.
- contract-logic (next to change the `Pattern` enum and `contracts.nra`): `PatArray` is `false` in `isProven` and is rejected before resolution in `checkEnsuresPattern`.
- The sweeps the brief names (argv dispatch in `main.nra`, the 15 tail loops) need the seed to have the feature first.

## Proposed learnings

- A check that runs again in generic instance bodies reports again with the instance's types, so the dedupe by message text does not catch it; skip it in instance bodies when the declaration's body already ran it | `corner<const N>` printed E0713 once per instance | selfhost/src | yes

## Verification

- `sh selfhost/bootstrap.sh /private/tmp/claude-501/array-patterns/out`: stage 1 and stage 2 emit identical C (191312 lines); the seed is older than the source (not regenerated, per the batch rule).
- Targeted suites with the built narrac: errors PASS 626 cases; dumps PASS 164; fmt PASS 87; fmt-stable PASS 1533; codes, sources, leaks PASS 1 each; node-ids PASS 1603; trivia PASS 1536; programs `--filter array_patterns` PASS 4 (debug with `-sanitize`, leak check and copy check; release).
- Scratch checks under `-sanitize` with `NARRA_CHECK_LEAKS=1`: a generator binding `[first, ...rest]`, an unguarded or-pattern with rests, `is` on a FixedArray. All clean.
- After the fresh review fix, rebased on feat/batch-11 aa2820e0: `sh scripts/test.sh`: `test.sh: PASS, 7276 cases in 140s`

## Size

- Estimated L, actual L.
