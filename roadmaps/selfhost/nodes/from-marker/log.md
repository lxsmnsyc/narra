# Log: from-marker

- 2026-11-10: branch feat/from-marker reset to feat/batch-5 (4d50212f). No E0976/E0977 or from marker existed.
- Design: marker stored on the result TypeExpr (resultFrom), TFunc gets a 5th field shared (param index or -1). Check in new sharedresults.nra, run after each non-instance body.
- selfhost/src gets 185 E0976 return sites under the rule (all-scalar structs count); added [package] shared-results = "unchecked" to selfhost/narra.toml so the compiler still builds itself. Orchestrator must accept or replace.
- std/src marked (cell.get, option, math min/max/clamp, coroutine readStep, shape trait + Value impl, tupleItems); test runner marked (4 sites); 16 sample programs marked or copied.
- Full run with the new compiler found guide, package, cnames and stage2 samples that share; marked them. Spread literals now share only through fields that can share (Player/Point spreads were false positives).
- std/src was accidentally reformatted by narrac fmt; restored and reapplied only the marker edits (std/src is not fmt-checked).
- Full check `sh scripts/test.sh` from the seed: PASS in 259 s (5960 cases, bootstrap fixed point holds).
- Rebased onto feat/batch-5 5d40c4a2: conflicts in parser.nra (symbol forms around parseResultType) and emitclosures.nra (cachedTypeId with the five-field TFunc); `from` test now uses SYMBOL_KEYWORD_FROM. Maintainer accepted the shared-results switch.
- Full check after the rebase: PASS in 89 s, 0 failed.
- Review fixes: arrows take the expected type's from (copy fix when the slot is fresh or no result type is written); spread skips written fields; guide chapter 9 section on from; ternary branch ends; helper renames; unknown-name errors point at the name; parser error for `from` with no name.
- Full check after the review fixes: PASS in 90 s, 0 failed.
