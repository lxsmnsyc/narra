# Log: patterns

- Branch renamed to feat/patterns, based on origin/main 3d87ad0. No log existed; fresh start.
- Checked existing work: README lists nested patterns, struct patterns, spread and destructuring as not supported; the parser rejects `...`. Nothing of the brief exists yet.
- Plan: recursive `Pattern` (kinds wildcard, name, ctor, struct, tuple, literal). A bare name is a constructor when the type is an enum with that variant, else a binding. Exhaustiveness by the usefulness algorithm with witnesses. Tuples are `StructType` with `Tuple: true`, fields "0".."n-1", interned per checker.
- Parser, AST, clone, checker (check_match.go, new check_patterns.go) and emitter (emit_match.go, new emit_patterns.go) done. Shared files got a few case lines each.
- `from` is a keyword: sample field names avoid it.
- Tuple show needed a case in emit_traits.go (outside the brief): `Tuple(1, "a")`.
- Else-branch narrowing after `is` with a refutable nested pattern was wrong; fixed in check_stmt.go with argsIrrefutable.
- Samples, package test and 8 error tests pass; full fast check passes.
- Review round 1 (user decisions): tuples use `#[...]` for literals, destructuring, patterns and types. `Tuple(...)` and the expected-type `[a, b]` path are gone; `Tuple` is a plain name. Show prints `#[1, "a"]`. SPEC.md and DECISIONS.md (D23, D44, new D53; D51 and D52 are taken by closures) updated.
- Review fixes: tuple C name is `S__Tuple__...` (a module named "Tuple" collided; `testdata/Tuple.nra`). Witnesses print `Name(_)` for variants with fields. `const Only(s) = u` parses. Failed patterns bind their names as void, and the match skips the arm body and the exhaustiveness check, so one error does not cascade.
- Rebased onto closures (9f0fa89). Conflicts only in README and DECISIONS (kept D51, D52, then D53). Go files merged cleanly.
- Re-review blocker: nested tuples of different arity shared a type and C name. Name and intern key are now `S__Tuple<n>__...` (`tupleCName`). Samples `patterns_tuple_shapes` and `patterns_closures` (closures inside tuples, structs and enums), error test for `==` on a tuple holding a function. D23 bullet reworded.
