# Outcome: scope-map

## Built

- `Checker.localNames` (one `Map<string, int>`) replaces the per-depth `localScopes` maps in `selfhost/src/check.nra`. `findLocal` is one lookup.
- Undo log: `ScopeBinding`, `Checker.scopeUndo`, `scopeUndoCount`, `scopeStarts`, and `bindName`. `pushScope` records where the scope's bindings start; `popScope` restores each replaced name, last first, so the rebinding after E0103 still works. `BodyState` saves and restores the new fields.
- Narrowing (`selfhost/src/patterns.nra`) uses the same map with `narrow|N` keys. `endNarrowing` sets -1 and patches logged previous values to -1, which gives the old "-1 in every scope" result; it does nothing when the key is absent or already -1.
- `LocalInfo.depth` is the depth that last bound the name; `isDeclaredBelow` (`selfhost/src/funcvalues.nra`) compares it with the stop depth.

## Proof

- base: 48823a59, `/usr/bin/time -l narrac emit selfhost`, instructions retired. Quiet runs (load 3 to 4): 2.622e9 to 2.635e9, spread 0.5%. Paired runs under load 10: 2656 to 2716 M, median 2672 M.
- branch: paired runs under load 10: 2620 to 2688 M, median 2632 M. Lower than base in 7 of 7 pairs, by 14 to 52 M (median 36 M, about 1.4%). Load adds spread (up to about 2% within one side), so the gain is real but close to the loaded noise; it exceeds the 0.5% quiet spread. Matches the profile's 2% estimate.

## Deviations

- None. No diagnostics changed.

## Stubs left

- None.

## Gotchas

- `popScope` now needs `uses alloc` (it sets map entries).
- A scope's names are restored only when it is popped; code that changes `scopeDepth` directly would skip that. Only `pushScope`, `popScope` and `startBody`/`finishBody` do so.
- `LocalInfo.depth` is set by `declareLocal` and `bindPatternLocals`.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh`: fixed point holds with the committed seed.
- `sh scripts/test.sh`: PASS, 0 failed.

## Size

- Estimated M, actual M.
