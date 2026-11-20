# Outcome: docs-concept-audit

## Built

- `roadmaps/selfhost/nodes/docs-concept-audit/audit.md`: the concept inventory (185 rows, every subsection of `SPEC.md` sections 1 to 15), the crucial gaps with evidence, 16 proposed writing nodes, the reconciliation of the 9 outline nodes, and chapter order changes.
- Counts: 51 crucial, 107 important, 27 niche. 92 taught, 25 mentioned, 68 missing. 11 crucial rows are not taught (8 missing, 3 mentioned).
- Crucial gaps: async functions and the async entry, C interop (`extern`, `unsafe`, `@trusted`), `std/fs`, `std/io`, `std/env`, the std module catalog, divisor proofs (E0979), `Maps.update`, `Cell`.
- Corrections found: chapter 2 says tuples cannot be indexed and that only a `let` can be passed as `mutable`; chapters 1 and 8 and `skills/narra/SKILL.md` leave `copy` out of the prelude.
- No doc or code changed.

## Deviations

- The coordinator passed maintainer input mid-task: use `SPEC.md` as the baseline and check async and FFI first. The inventory follows it, so it has more rows than a concept-level list would.
- `DECISIONS.md` was used only through the D numbers `SPEC.md` cites.

## Stubs left

- None.

## Gotchas

- The audit's row numbers are referenced by the proposed nodes. A writing node brief should copy the rows it covers rather than cite numbers, in case the table is edited.
- Six outline nodes are proposed for batch-21. `std-reference` depends on `std-interface-query` and `site-reference-section`. `guide-c-interop` links to `manifest-reference` for `[link]`.
- `guide-concurrency` is proposed to be replaced by `guide-async` and `guide-threads`. Its brief needs to be dropped or rewritten by the orchestrator.

## Proposed learnings

- A guide page claims a rule only after the guide suite or a spot check runs it | two chapter 2 sentences contradict the compiler and passed review | `docs/guide` | partly: the guide suite runs samples, not prose claims.

## Verification

- Every claim marked "verified" in `audit.md` was run with `narrac` built from `seed/narrac.c` (see `log.md`).
- `sh scripts/test.sh`: `PASS, 10496 cases in 151s`.
- `check_prose.py` on `audit.md`: only the rank name "crucial" (defined by the brief) and long-sentence warnings remain.

## Size

- Estimated M, actual M.
