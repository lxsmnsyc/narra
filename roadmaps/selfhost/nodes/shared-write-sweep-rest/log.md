# Log: shared-write-sweep-rest

## 2026-11-11, measure (base feat/batch-6 25cd3092)

- Base compiler: `sh selfhost/bootstrap.sh` from the seed (stage 1 = stage 2, 182,642 lines).
- Counts (E0978 with the switch on, E0976 with it removed locally; a package with errors shows no warnings):
  - rest modules: E0978 62, E0976 103 (parser 25, querysites 13, emitexprs 12, emitstate 10, emittraits 9, lspcache 9, ...). Checker modules: 0 and 0.
  - `selfhost/tests`: E0978 3, E0976 0.
  - `testdata` (every program and package outside `errors/`, `emit-errors/`): E0978 37 in 26 programs (36 + the `shared_write_fix` sample), E0976 14 (5 `diagnostics/from_*_fix` samples, which are fix samples on purpose, and 9 `declsdump` samples).
- `narrac emit selfhost`, instructions retired, base: 2,380,458,720 / 2,378,739,849 / 2,378,579,127, median 2,378,739,849.

## selfhost/src

- Fix script (`out/fix.py`, not committed): one `check --json selfhost` per round, applies the first fix of every E0976/E0978 diagnostic, deduplicated, one fix per overlapping range per round. `from` fixes on all-scalar results (`Pos`, `Code`, `Token`, `LoopInfo`, `Location`) are left for hand edits: `copy(...)` at the return instead, as the first half did.
- E0976: 9 rounds to 0. The marks then pushed E0978 in the rest modules to 139 (callers of the new `from` functions); 3 rounds of fixes, then 20 sites with no fix or loop-held, by hand.
- Hand edits: loops over `copy(emitter.checker...)` (emittraits x7, emitmatch, emitexprs, emit); `emitClosure` clears frames/scopes/loops/blocks right after saving them (nothing in between reads them), so the saves need no copy; `copy(readNodeType(...))` where an `is` binding of the type is used after a `mutable emitter` call; `Arrays.push(mutable emitter.blocks, copy(info))`; `copy(bodyBlock(...))`; struct literals the fix wrapped whole (`copy(Parser {...})`, `copy(Collector {...})`, `copy(Walk {...})`) now copy the one field that shares (`tokens: copy(tokens)`, `modules: copy(modules)`, `templateTokenEnds: copy(...)`).
- The fmt printer's function fields (`printExpression`, `printPattern`, `printBlock`) are `from printer` (E0977 after the marks; the printer is never written).
- Switch removed from narra.toml, manifest.nra, modules.nra, check.nra, body.nra, SPEC.md, DECISIONS.md, compiler/README.md. No other package used it.
- First measurement after removal: `emit selfhost` 3.96 G instructions (+66%): the E0976 walk now runs on every compiler body, and `collectHeldTypes` walked the whole Checker type graph for every `mutable checker` argument. Cached the held-type set per type key (`Checker.heldTypeSets`, moved into the walk's facts and back): 2.48 G. With the switch put back on the branch source the branch compiler measures 2.39 G, so the edits cost about +0.5% and the check about +3.6%.
- Base source emitted by base and branch compilers: identical C (raw and release). Bootstrap: stage 1 = stage 2.
- Targeted suites (errors, habits, codes, sources, dumps, fmt, spec, query, generated): PASS, 1416 cases.

## selfhost/tests and testdata

- `selfhost/tests`: the 3 `newCommand(...)` results take the fix (`copy(Shell.newCommand(...))`).
- `testdata`: the fix script, then hand edits. Locals named `copy` renamed to `copied` first (const, coroutine_heap, loop_unique_generator, mutable_contents, mutable_for, strings2, strings2_builder), so `copy(...)` can be written. Array literals the fix wrapped whole (`copy([words, words])`) copy their elements instead. No-fix sites: `Arrays.push(mutable history, copy(items))` (loop_unique_hoist), loops over `copy(loop)` (maps) and `copy(growing)` (perceus_loop_items).
- `copy_report`: the report is the point, so the hidden copies now come from a case the static rule does not follow (a `mutable` argument of a user function keeps the other argument): same `.out`, same `.copies` line, checks clean.
- declsdump samples with E0976 tried and reverted: they also have unused-name, missing-entry or trait errors on purpose; they are dump samples, not checked programs.
- Programs suite with the branch compiler: PASS, 321 cases.

## End

- `JOBS=3 sh scripts/test.sh` on 7c21a345: PASS in 341 s, 6327 cases + bootstrap.
- Final instructions (load 32): base 2,376,661,564, branch 2,455,517,939 (+3.3%).
