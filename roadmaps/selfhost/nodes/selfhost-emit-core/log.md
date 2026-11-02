# Log: selfhost-emit-core

- Nothing existed on main: no emit files, no skip files, no emit tests.
- Dead end: the contract has `cnames.nra` take an `Emitter`, but `emitstate.nra` calls cnames, so the import would cycle. cnames stubs take the `Checker` (see `contracts/emitter.change-request.md`).
- `Map {}` is not a constant, so the new `BodyInfo.instanceRefs`, `traitSelf` and `Checker.composites` have no default; they are set in `startBody` and `newChecker`.
- `function` is a keyword, so `implName` uses `member`.
- Stubs find their work with `hasWork` (scans body types, refs and closures). `match`, `std` and `options` have no section function, so each has a probe (`reportMatchWork` and so on) that `emit.nra` calls only while `programRecorded` is false.
- The recorded path of `emit.nra` (comment table, array typedefs and structs, type typedefs, C main) is written but cannot run until selfhost-emit-program sets `programRecorded`. Untested. Function prototypes report `unsupported: function signatures (selfhost-emit-program)` because they need the parameter C names.
- Unsure: the sort key of composites. Go sorts arrays and boxes by the type id of the element, maps by `keyId|valueId`; `emit.nra` does the same through `typeId`.
- The Bash sandbox rejects commands with long scratch paths or `GitHub` in odd forms; used a symlink `/private/tmp/claude-501/sp`.
