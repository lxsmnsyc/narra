# Log: shape-derive

- 2026-11-09: Read shape-trait and std-json outcomes. Chose generated Narra text over AST building or C: the checker and emitter then need no new paths, and the prelude mechanism (`setShapeRef` pointing at a function) carries over to user types.
- Prototyped the generated code by hand in a scratch file (aliased named imports `Shape as __Shape`, `Value as __Value`, narrowing after `is None`, `let` struct with field assignment, generic struct literal `Box<T> { ... }`): it compiled and ran clean under `-sanitize` with the existing compiler.
- `findFunction` (body.nra) and `ProgramFunc` match declarations by position, so the generated text starts after the file's last line (blank lines prepended) to keep positions unique.
- Unused locals are errors, so a struct with no fields binds `__Value.Record(_)` and a struct whose fields are all required uses `const __result`.
- A derive in error (E0531, E0532, or a user trait named Shape) would make the generated bodies report errors at positions past the end of the file; `shapeDeriveHolds` lets `checkModuleBodies` skip them.
- `declsdump/traits/err_unknown_trait` and `err_derive` goldens changed: E0502 is gone and E0503 now lists Shape.
- The orchestrator asked for suites only, no full run. Ran: dumps, errors, codes, programs, packages, sources, std-tests, bootstrap, query, lsp. All pass.
