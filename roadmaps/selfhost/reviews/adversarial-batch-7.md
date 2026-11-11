# Adversarial review of batch-7 (fresh-context agent), with dispositions

Scope: the batch-7 briefs, checked against the code for claims, shared files and edges. All eight findings were checked against the code and accepted.

1. tree-shake and tree-shake-types had no closing `---` in their frontmatter (an edit script removed it). Fixed.
2. bench-check-suite listed `scripts/**`; suites are Narra modules registered in `selfhost/tests/src/registry.nra` (`listSuites`). Fixed: suite `bench`, shared paths `registry.nra` and `bench.nra`.
3. contracts promised more than the prover holds: `Fact { low, high, nonzero }` in `operands.nra` covers integer locals only (`findFactLocal`), so float and field conditions cannot be proven; parameter facts are seeded in `body.nra`; function values are formed in `funcvalues.nra`; the brief mixed E0982 with a reserved code. Fixed: integer parameter conditions only, others an error at the declaration; flow described; shared paths added; one reserved code for function values, E0982 stays for `typeof`. contracts-ensures got a design paragraph: caller facts land on bound integer locals.
4. build-conditions missed the std module registration files (`stdtext.nra`, `runtimetext.nra`, a `stdbuild.nra` registry, `fixes.nra`) and the release flag plumbing (`driver.nra` passes `release` only to `cc`). Fixed: paths added; the target (os, arch, debug) lives on the Checker, set by the driver and by the language server.
5. return-trailer: the missing-return rule is in `pathends.nra`, and the trailer is in `emitclosures.nra` and `emitconsts.nra`, not `emitstmts.nra`. Fixed.
6. Edges: tree-shake must skip branches build-conditions folds; copy-check-macros' hard edge on return-trailer was not needed. Fixed: tree-shake hard on build-conditions and copy-check-macros (shared `emitdecls.nra`); copy-check-macros has no edge.
7. copy-check-macros listed files without copy-check code. Fixed: `emitdecls.nra`, `emitgenerators.nra`, `stdcopy.nra`, `runtimecore.nra`.
8. shared-results-walk: `checkSharedResults` is called from `body.nra`. Added. No edge with place-ids needed (the imported helpers keep their signatures).
