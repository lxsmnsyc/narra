# Log

- Added selfhost/src/deadcode.nra (checkDeadCode, E1500) with one call in Stmts.checkBlock; pathends.nra exports findWhileOpenPath and findCForOpenPath.
- Swept dead returns in selfhost/src, selfhost/tests/src/codes.nra, std/src (json, toml, unicode, async__net) and testdata samples.
- narrac apply-fix prints the fixed file and fixes only the first diagnostic per run, and unrelated fixes can ride along; check the diff after use.
