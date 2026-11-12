# Log: lazy-values

- Branch feat/lazy-values reset to feat/batch-11. Nothing of the node existed before.
- A lazy name has type `T`; the declaration carries `ConstDecl.lazyPos`. No `Lazy<T>` type.
- Parser: `lazy` is read after `=` of a top-level `const` when a non-punctuation token follows on the same line (`startsLazyValue`). A local `const`/`let` with it reports E1032. A trivia key `k` at the word's position holds comments around it.
- Checker: `Lazies.force(name)` marks the argument node with `Ref.Const(-1)` before checking it; the const-resolving paths in `exprs.nra` call `checkLazyUse`, which reports E1033 unless the node is marked. A busy lazy const (a type cycle through force) reports E1035 instead of E0191.
- Effects: `walkConstants` now also runs in the non-reporting passes for lazy values, storing their effect sets (`lazyConsts`, `lazyEffects`). A force adds those effects (`addForceEffects`). E1034 in the reporting pass. The cycle check (E1035) builds a graph of bodies (calls, function names used as values, forces) in `lazyEdges` and walks it in `lazyvalues.nra`.
- Emit: `Lazies.force(name)` emits its argument, which reads the D61 getter. Reachability needed no change: `emitReachableBodies` reaches the getter through the reference `constRef` notes, so an unforced lazy value emits no C and `emit.nra` is untouched (no conflict with inline).
- Samples cannot count builds with a print: the expression may use only `alloc` (E1034). The loop sample shows reuse by the result; the leak check shows each value freed once.
- `std/lazy` is registered last in `stdlib.nra`, so the module list in the unknown-module message gains `std/lazy` at the end (two goldens updated).
- Multi-module sample lives in `testdata/packages/lazy_values_modules` because package samples are found only there.
