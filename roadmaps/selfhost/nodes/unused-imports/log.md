# Log: unused-imports

- 2026-11-09: branch feat/unused-imports from origin/feat/batch-1. E0946 and E0947 were free in `selfhost/src/diagnostics.nra`.
- Fixes exist in selfhost now (`narrac apply-fix`, `fixes.nra`), so both codes carry a fix.
- Chose a token rule for imports (parser fills `SourceFacts.referenced`) and a ref rule for private functions (`Ref.Function` in another body). Tracking every name lookup in the checker (about 70 sites over scopes, types, traits, modules) would have spread the diff across hot files that bytes-type also edits.
- The check runs at the end of `checkPackage`, after `finish`, and only when there are no errors. Placing it before `finish` let it fire ahead of "type S contains itself" (packages/errors/contains_itself_imported).
- First stage 1 build: 60 import names and 1 function in `selfhost/src`, plus the 26 std/shape provider functions (`intToValue`, ...). Those are called by name from `setShapeRef`, so they count as used now.
- packages/errors/missing_entry names a private `start` as entry: a function an entry target names counts as used, so the entry error stays.
- Cleanup with a scratch script that applies the first fix of each E0946/E0947 in a loop. Cleanup in `selfhost/src`: removes import names, and deletes `getPatternPosition` in fmt.nra (no rename).
- Habit samples: the fix of throw_call, throw_result, ts_throw and ts_number left a private function unused, so `run` now calls it. ctor_import: after its fix `SchoolPerson` was imported but not used, so main.nra has an exported `describe(_person: SchoolPerson)`; expected.err gets the second line, `unknown type "SchoolPerson"`.
- One run of the dumps/query/lsp/std-tests/guide group took 360 s, against 14 s the time before. It passed. The other worker was probably loading the machine at the time.
