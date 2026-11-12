# Log: tuple-index

- Checker: `checkTupleIndex` in exprs.nra, called from `checkIndex`. The element index comes from `findTypedConstant`; E1010 and E1011 are reported at the index expression's position so the void result raises no second error.
- Emitter cannot call `findTypedConstant` (the checker body tables are not current at emit time; a `const` index failed with "unsupported: index"). The checker records the element in a new body table `tupleIndexes`, and `emitIndexRead` writes `(target).f_N`.
- Writes: `findAssignedLocal` returns -1 for an index on a tuple, so assignment, compound assignment, `++` and `mutable` arguments report errors. `update(mutable pair[0])` reports the existing "a mutable argument must be a variable, a field or an element".
- The other IndexExpr readers needed no change (the full suite passes, samples clean under sanitizers and the copy check).
