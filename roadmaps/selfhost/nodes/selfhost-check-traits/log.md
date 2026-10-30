# Log: selfhost-check-traits

- Dump design: `trait`, `impl` (listed, derived, block, holds) and `bound` lines, plus a `function` sub-line for trait and impl signatures. `holds` lines (implements(type, trait) for the four standard traits, for each type of `order`) test `implements` without a body checker. Format is in the doc comment of decldump.go.
- `implements` is a keyword, so the port is `typeImplements`. Struct fields `implements`/`derives`/`impls` are `listed`/`derived`/`blocks` (keywords again).
- Go's pointer-keyed maps become index-keyed: traits by `Checker.traits` index, params by `Checker.params` index. Facts of a struct or enum are in `Checker.typeTraits`, found by "<module>|<name>" (Go's typeKey).
- Tuple trait facts: Go creates the `Tuple` typeTraits when the first tuple is made (tupleOf). Here `typeTraitsOf` creates it on first use, so resolve.nra needs no hook for it. Same answers.
- `setParamType` moved from declare.nra to check.nra, because traits.nra needs it and declare.nra imports traits.nra (no import cycles).
- resolve.nra (not in the brief's shared list) got one line: `checkMapKey(mutable c, te.pos, k)` in the Map case, plus the import. The queue (`checkMapKey`) lives in check.nra, the check (`mapKeyProblem`, `flushMapKeys`) in traits.nra. Recorded as a deviation.
- Go DeclDump now calls `flushMapKeys` before `finish`, so Map key errors in signatures show up in the dump (CheckPackage already flushes). Narra `declarePackage` does the same.
- Dead end: none found. Error ordering matched Go on all inputs the first time; the first-run mismatch risk was substitute() order in sameSignature (Go checks the result before the parameters, and substitute can create instances), kept as is.
- Open doubt: a duplicate type declaration plus `derives` failing: Go uses the declaration's own `d.T`, Narra looks the name up in the scope (the later one). Only matters in programs that already have a duplicate-type error.
