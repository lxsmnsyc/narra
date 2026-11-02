# Log: selfhost-emit-bodies

- Read GOAL, CLAUDE, brief, contract v2 and the core outcome. Programs facts are not recorded yet, so the bodies cannot run against real output.
- Wrote emitexprs.nra (expressions, calls, prelude, Array, Box, constRef, lvalues), emitstmts.nra (statements, blocks, for loops) and emitconsts.nra (constant prototypes, getters, functions, tests).
- Kept the skip file: while programRecorded is false, emitConstPrototypes and emitConstBodies report a marker, so U still equals S.
- `narra check selfhost` clean; TestSelfhostEmit passes.
