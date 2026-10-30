# Log: selfhost-check-decls

- Branch feat/selfhost-check-decls from origin/main c8e17a7.
- Go split: `CheckPackage` now calls `newPackageChecker` (prelude) and `declareSourceModule` (scope + declareModule per file), then bodies per module as before. The interleaving of declare and bodies per module is kept, so nothing changes. `Checker.prelude` keeps the parsed prelude for the dump.
- Proof: scratch script runs `narra emit`, `narra check` and `narra check --json` on 187 inputs (every .nra under testdata outside packages/pkgdump, every dir with narra.toml, ../selfhost) before and after; `diff -r` reports no difference.
- Dump format decision: attributes first, the type text last (types contain spaces). Const declared types are resolved by the dump after all modules are declared (the full checker resolves them in constType, a body pass). Instance blocks sorted by header line.
- Harness skip rule: a whitelist of the error formats this node ports. An input whose Go errors include any other format is skipped (trait, import, default-value errors).
- Narra side done in one go: every sample, error file and declsdump case matched on the first full build. Syntax snags: no `=> return` arms, no for-loop destructuring, `Array.filled(value, n)`, mutable alias rule on `Array.push(mutable c.x, Info{.. modPath(c) ..})`.
- resolve/instantiate/buildGeneric are one cycle, so they share resolve.nra; generics.nra has the rest.
- Found: the parser drops the line break after a type ending in `>`, so `distinct type X = Array<int>` followed by a declaration fails (Go and Narra alike). Not fixed here; reported.
- Rebased on 2f52283 (#24 merged); byte-identical comparison rerun on 219 inputs against a binary built from origin/main: identical. Fast check passes.
- Done: all acceptance criteria. Outcome written.
