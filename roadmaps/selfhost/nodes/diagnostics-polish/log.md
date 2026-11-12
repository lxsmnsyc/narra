# Log

Base commit 1956faac (feat/batch-9). Base compiler built from a clean `git archive` of it.
Repro files live in /private/tmp/claude-501/diagnostics-polish/p.

## Proof on base (`base/narrac check FILE`)

1. Mutual typeof cycle (`function f1(x: typeof g1)`, `function g1(x: typeof f1)`):
   `cycle.nra:1:16: typeof g1 refers to its own declaration; write the type out` (wrong: g1 is not self-referencing).
2. Unannotated constant in a signature: `constsig.nra:3:23: the type of constant "limit" is not known here ...` carries E0982, shared with "typeof of a type".
3. narrowing_shadowed_panic: `8:16 ... may hold any of its variants` then `8:16: expected int, found void`.
4. A token at line 1048581 (or a column past 1048576) is accepted silently; both pack to the same key as an earlier position.

## Plan

- 1: keep the chain of typeof names being resolved (`Checker.typeofChain`) and print `g1 -> f1 -> g1`.
- 2: new code E1003 (next free after E1002) for the constant case.
- 3: `expectType` drops a void mismatch at a position that already has an error (`hasErrorAt`).
- 4: lexer rejects a token at line or column >= 1048576 with new code E1004; KEY_STRIDE kept (widening cannot remove aliasing).
