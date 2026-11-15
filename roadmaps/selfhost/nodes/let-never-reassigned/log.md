# Log: let-never-reassigned

- Rule built in `check.nra` (`reportNeverReassigned`, called from `reportUnusedLocals`), marks in `stmts.nra` (assignment, `for (mutable ...)` source) and `exprs.nra` (`checkMutableArgument`).
- Measured with the seed-built compiler: 3 sites in selfhost/src, 4 more in testdata and bench, derive_shape generated code for empty structs. 8 sites in total, far under 500.
- A write through a closure goes through the same assignment check, so it counts.
