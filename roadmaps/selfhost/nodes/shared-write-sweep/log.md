# Log: shared-write-sweep

## Baseline (feat/batch-6 e8f0b61f, narrac from `sh selfhost/bootstrap.sh`)

- E0978 (`narrac check selfhost`): 197 total, 141 in checker modules (84 with a fix), 56 in the rest.
  Per checker module: body 3, builtins 1, calls 7, check 6, closures 5, ctors 7, declare 10, declsdump 1,
  deriveshape 2, effects 14, exprs 16, funcvalues 2, generics 1, instances 4, literals 3, match 6,
  resolve 14, scopes 2, sharedresults 1, stdlib 2, stmts 5, traits 27, unused 2.
- E0976 (switch removed locally): 206 total, 104 in checker modules (103 with a fix), 102 in the rest.
  Per checker module: body 2, calls 5, check 17, closures 3, declare 2, deriveshape 1, effects 10,
  exprs 7, fixes 4, generics 3, literals 2, match 2, patterns 8, resolve 12, scopes 1, sharedresults 3,
  sharedwrites 6, stdarray 1, stdcopy 1, stdemithelp 1, stdresult 3, stmts 5, textops 1, traits 3, unused 1.
- `narrac emit selfhost` instructions retired: 2,363,365,381 / 2,360,377,978 / 2,358,781,360; median 2,360,377,978.

## Notes
- `narrac apply-fix` applies one fix per run, and each run is a full check of selfhost. Instead, one
  `narrac check selfhost --json` per round, and a script applied the first fix of every checker-module
  diagnostic (same edits apply-fix writes), deduplicating equal edits and skipping conflicting ones.
  Rounds repeat until no fix is left; then hand edits.
- E0976 first, because a new `from` changes what callers share (more E0976 in callers, new E0978).
- First attempt applied every `from` fix. It cascades through the checker: `checkExpression`,
  `resolveType` and most of the Walker functions return a `Type` that shares `checker` or `want`.
  A result sharing two parameters needs `copy(...)` of the other, and the copy fix E0976 offers for a
  `Type` or `Option<...>` result is rejected by E0975 (copy of an enum copies nothing). Trait
  implementations cannot add `from` (E0520), so the Walker trait would need marks too. And each
  `from checker` on a `Type` result adds E0978 warnings whose held value is a `Type`, which have no
  fix (141 -> 182). Attempt saved, then reset.
- Second attempt: `from` only on results whose type `copy` accepts (structs, arrays, maps). Results of
  all-scalar structs (`Pos`, `CLayout`) get `copy(...)` at the return instead of `from`: the copy
  emits the value itself, and callers do not start sharing. 50 E0976 with enum results left
  (Type 27, Option<Type> 8, other Option 14, Expr 1): need a decision.
- E0978: 3 rounds of fixes (95 + 14 sites), then 20 loop-held loops over `copy(...)`, then hand edits:
  stored parameters copied where they go into the checker (`copy(implDecl)`, `copy(funcDecl)`,
  `copy(typeArguments)`, `copy(pos)`), parts read before the call (`collectExpressionAndArms`),
  `recordValue` helper in effects.nra (the push of an `Expr` into `state.valueLists` paired the
  expression's parts with the list; a helper's `mutable` argument keeps nothing), `finished.loops` and
  `finished.labels` instead of `context` in `checkBlockBody` (nothing writes those fields of a block).
- Save/restore (O9): `resolveAlias` saves the environment after `makeParameters` (no copy);
  `checkArrowParts` clears `labels`/`blocks`/`loops` before `pushArrow` (no copy);
  `checkReferenceFor` keeps `copy(checker.loopBorrows)`: the push needs the outer list and the new
  one at the same time, and Narra has no pop.
- `copy(x)` costs one retain in release builds (`emitCopyCall`), so the loop-over-copy edits cost a
  retain per loop, not a copy of the list.

## Second pass (orchestrator decisions, 2026-11-11)

- Rebased on feat/batch-6 aaee3950 (std-compound-sweep); no conflict.
- copy(x) takes enums with a tracked payload (option a). `canShareType` moved from sharedresults.nra to
  stdcopy.nra; `checkCopyCall` and both fix builders (E0976, E0978) use it. `isCopyTracked` keeps its
  meaning (a write changes the value in place), because the shared-write walk uses it for that.
- Debug copy check: a retain of a heap enum only counts its payload box, so a write to an array taken
  out of a copied `Tree` was unexplained (exit 103). The emitter now writes `credit_<enum>` functions
  (debug builds only) that credit every counted value in a heap payload; `copy(...)` calls them after
  its retain when the type holds a heap enum.
- Seed regenerated after the copy commit (the compiler needs `copy` of `Type`); runtime text unchanged,
  so one update was enough.
- E0976 rounds with the enum skip removed: 50 -> 0 in 9 rounds; the Walker trait got the `from` of its
  one implementation (body.nra). E0978: 19 new sites, 16 + 4 fixes, then 3 by hand (values bound by an
  `is` pattern: compute before the write, or copy the part).
- `recordValue` removed: the effect pass pushes `copy(expression)`.
- Base compiler for the comparisons: built from aaee3950's own source (out/base cannot be used: its
  std text predates std-compound-sweep).
