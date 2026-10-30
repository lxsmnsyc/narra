# Log: poly-recursion

- Resumed from an earlier worker's uncommitted `checkPolyRecursion` in `check_types.go`. Judged sound: it walks the generic templates (through other generic types) and errors when a path returns to the type with an argument that contains a type parameter without being one.
- Same-argument recursion (List<T> in Array<List<T>>) is not flagged; the existing testdata still passes.
- The check runs at the end of `buildGeneric`, and the fields are cleared after the error so a later instantiation cannot grow.
- Tests: poly_recursion, poly_recursion_derives, poly_recursion_mutual (A<T> -> B<Array<T>> -> A<..>).
- Open doubt: a cycle that grows only in one step of a longer cycle is caught (the grew flag is carried along the path). Not tested beyond the mutual case.
