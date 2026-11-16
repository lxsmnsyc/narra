# Outcome: restrict-renames

## Built

- E0971 in the shared-write walk (`selfhost/src/check/sharedwrites.nra`, `reportRename`): `const b = a`, `let b = a` and `b = a` with a bare name of a tracked type. Message as in D124, with an `apply-fix` that wraps the name in `copy(...)`. A reported rename starts no sharing, so E0978 does not repeat it. Row added in `selfhost/src/driver/diagnostics.nra`.
- Not checked: destructuring declarations, type parameters, parser-made `__` names.
- The sites in `selfhost/src` (68 with the three from workspace-manifest), `selfhost/tests/src/lsp.nra` and `std/src/args.nra` are rewritten without a second value where possible (helpers, conditional expressions, in-place rows); the rest keep `copy(...)`. Seed regenerated.
- Samples: `testdata/errors/rename_tracked.nra` (every binding form and tracked type), `testdata/rename_allowed.nra`. About 30 older samples reached E0978 through a rename; they now use a helper `share<T>(value: T): T from value`.
- SPEC "Renames" section, guide chapters 2, 9 and 13, D124 "Built" line.

## Deviations

- The brief counted about 13 typed sites; the check also finds `x = binding` after a pattern, so there were about 70. A `from` marker names one parameter, so a helper whose result shares two parameters cannot replace a loop swap; those keep `copy(...)` (it costs a reference count, not a clone).
- `testdata/lsp/incremental_shared_writes`: the helper is in every text variant; one line of the stderr golden changed (`no module changed` became a recheck of three modules). The diagnostics comparison with `--full` passes.
- Edited `selfhost/tests/src/lsp.nra` and `std/src/args.nra` (shared paths).

## Gotchas

- A rename site in std shows only after the seed is rebuilt with the new `stdtext.nra`.
- The tuple rule follows D94: a tuple holding an array is tracked, so renaming it is E0971.

## Verification

- `sh scripts/test.sh`: PASS, 10189 cases.
- `pnpm --dir website build`: passes.

## Size

- Estimated M, actual M.
