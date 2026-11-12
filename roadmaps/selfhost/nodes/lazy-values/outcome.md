# Outcome: lazy-values

## Built

- `const name = lazy expression` at the top level, forced by `Lazies.force(name)` from `std/lazy`. The name has type `T` with `ConstDecl.lazyPos` on the declaration; no `Lazy<T>` type.
- `selfhost/src/lazyvalues.nra`: E1033 use check, the forcing-cycle graph (E1035). `selfhost/src/stdlazy.nra`: the `std/lazy` rows, check and emitter. Wiring: parser (`startsLazyValue`, E1032 for a local), `exprs.nra` (skips the constant-expression rule, E1033, E1035 for a busy lazy value), `effects.nra` (effect sets of lazy values in the fixed point, E1034, cycle report), fmt, astdump, declsdump, `diagnostics.nra`.
- Emit: `Lazies.force(name)` emits the D61 getter call. An unforced lazy value emits no C (reached only through the getter reference).
- Tests: `testdata/lazy_values*.nra` (5 samples), `testdata/packages/lazy_values_modules`, `testdata/errors/lazy_*` (8), astdump, declsdump and fmt goldens (one with comments), emit check `tree-shake/lazy-unforced`.
- SPEC section 3 and 10, `compiler/README.md` updated.

## Proof

Not a fix node.

## Deviations

- Files outside the owned and shared lists: `testdata/packages/lazy_values_modules/**` (package samples are found only under `testdata/packages`), `testdata/errors/unknown_module.err` and `testdata/declsdump/imports/err_unknown_module.dump` (the module list in the unknown-module message gains `std/lazy`).
- No sample counts the builds with a print: E1034 allows only `alloc` in the expression, so a builder cannot print. Reuse is shown by the results and the leak check.
- `emit.nra` is unchanged: reachability already follows the getter reference, so there is no conflict with inline.
- Fmt: a comment before `lazy` moves it to the line after `=`, as for other values.

## Stubs left

- Forcing is not thread safe (the ready flag), as for D61 constants; the threads node adds the once guard.

## Gotchas

- `lazy` is read when a non-punctuation token follows it on the same line, or `(`, `[` or `!` after a space, or a `-` touching its operand; `lazy(x)`, `lazy[0]` and `lazy - 1` stay ordinary expressions.
- Passing a lazy name to `Lazies.force` is marked with `Ref.Const(-1)` on the argument node before it is checked; `checkLazyUse` reads the mark. A set of forced node ids would need a new `Checker` field (check.nra is not in this node's paths), so the mark stays in the ref table.
- `EffectState` fields need constant defaults (`[]`); two `mutable` fields of one struct cannot go in one call, so the edge list is one array of pairs.
- A new std module changes the module list in the unknown-module error goldens.

## Proposed learnings

- Register a new std module last in `stdlib.nra` | the unknown-module message lists modules in registration order | selfhost/src | no

## Verification

- `sh scripts/test.sh`: `test.sh: PASS, 7616 cases in 239s`
- Targeted: programs 391, errors 623, packages 36, emit 19, dumps 165, fmt 86, codes 1, all PASS.

## Size

- Estimated M, actual M.
