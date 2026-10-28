# Log: patterns

- Branch renamed to feat/patterns, based on origin/main 3d87ad0. No log existed; fresh start.
- Checked existing work: README lists nested patterns, struct patterns, spread and destructuring as not supported; the parser rejects `...`. Nothing of the brief exists yet.
- Plan: recursive `Pattern` (kinds wildcard, name, ctor, struct, tuple, literal). A bare name is a constructor when the type is an enum with that variant, else a binding. Exhaustiveness by the usefulness algorithm with witnesses. Tuples are `StructType` with `Tuple: true`, fields "0".."n-1", interned per checker.
- Parser, AST, clone, checker (check_match.go, new check_patterns.go) and emitter (emit_match.go, new emit_patterns.go) done. Shared files got a few case lines each.
- `from` is a keyword: sample field names avoid it.
- Tuple show needed a case in emit_traits.go (outside the brief): `Tuple(1, "a")`.
- Else-branch narrowing after `is` with a refutable nested pattern was wrong; fixed in check_stmt.go with argsIrrefutable.
- Samples, package test and 8 error tests pass; full fast check passes.
