# Log: traits

- 2026-10-28: started on feat/traits from f84c1ef. Nothing of the node exists yet: no trait syntax, `==` rejects non-primitive types, constraints fail in the parser.
- Plan: standard traits and `Ordering` declared in the Narra prelude. Constrained generics stay compiled once per type, so trait calls resolve to direct calls; no hidden table. All trait logic in check_traits.go and emit_traits.go; shared files get a few case lines and calls.
- Function types `(x: T) => U` get a minimal parse so the `==` error test can declare a struct with a function field. Closures owns the full feature.
- Built parser, checker (check_traits.go) and emitter (emit_traits.go). Shared-file edits: AST fields, parser decl cases and helpers, 5 calls in check.go run(), 1 case in check_call.go, checkBounds call in genericCall, `==` check in check_expr.go, FuncType in resolveType, 3 lines in the emitter.
- Wrote contracts/trait-model.change-request.md: the contract's name functions cannot make the emitter define the generated functions, so added `e.useEq` and friends.
- Samples traits.nra and traits_derive.nra pass under ASan and the leak check. Nine error tests under testdata/errors/traits_*.nra.
- Fast check passes.
- Coordinator asked to make constructor lookup deterministic. `ctor` now uses `ctorOwners`: declaration order, program enums hide prelude enums, and two program enums sharing a name give an ambiguity error (SPEC section 3). Added traits_ctor_shadow.nra and errors/traits_ctor_ambiguous.nra. 20 runs each were stable. Fast check passes.
- Rebased onto modules (c8a5730). Squashed my two commits first. Conflicts were in check.go (Checker fields, Check) and README.md. Took main's Check and added the traits field; the trait hooks merged into declareModule and checkBodies on their own.
- Ported to modules: traits visible per module, trait export and import, typeKey by module and name, ownsType by module, impl names through funcCName, generated names `<mangle>__op`. The ctor ranks are own, then imported, then prelude, with ambiguity within a rank. Added packages/traits and two error packages. Updated the change request.
- Review of PR #3 fixed, in a new commit. Trait and trait function names must pass cSafeSegment, so impl names are injective. FuncType compares structurally and mangles as F<n>__params__result. Duplicate bounds are rejected. The no-value cascade is gone after an earlier error. Show escapes \n, \t and \r. README notes the new keywords and prelude names. The other nits are recorded in outcome.md Gotchas.
