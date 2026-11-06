# Log: closure-borrow

- Go: useLocal captures `let` and mutable parameters by reference when every arrow in the chain is a stack closure. E0612 and E0613 replace E0603 and E0604 (old rows stay, unused). E0614 comes from checkBorrowed, called at the end of `args`. narrowed() returns nothing for a `let` used inside a closure that captures it. assignedInExpr and usesLocal look into arrows.
- Go emitter: environment field `T*` for a `let` or mutable parameter, `refCaptures` set in closureFunc, `isRef` used by localRef and lvaluePtr.
- Selfhost: the same in funcvalues.nra, exprs.nra (collectAssigned* moved here from stmts.nra so checkBorrowed can use them), emitclosures.nra, emitstate.nra.
- Tests: closure_borrow.nra and .out, errors/closure_borrow_*, updated errors/closures_capture_let and closures_capture_mutable.
