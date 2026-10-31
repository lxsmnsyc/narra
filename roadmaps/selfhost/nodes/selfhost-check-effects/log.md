# Log: selfhost-check-effects

- Baseline on main: TestSelfhostBodies 341 compared, 37 skipped (30 does not load, 7 unsupported).
- Removed the 7 unsupported stubs: duplicate constants get their own Checker.consts entry (exprs.findConstInfo), and E0425 for a constructor import is reported at the name (ImportName.pos added to ast.nra and parser.nra).
- Rebased on main after typesdump-targs-spacing landed (8957591); no conflicts.
- Ported check_effects.go to selfhost/src/effects.nra. Go: checkEffects returns the state, TypesDump runs it and prints `effects` lines.
- Goldens regenerated with NARRA_UPDATE=1 (only effects lines appended). New cases in testdata/typesdump/effects/.
- Full check passes.
