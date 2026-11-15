# Outcome: let-never-reassigned

## Built

- E1510 (`"x" is never reassigned; declare it with `const``) with a fix that rewrites `let` to `const`. `LocalInfo` gets `letLine`, `letColumn` and `assigned` (`selfhost/src/check.nra`); `markAssigned` is called for every assignment target root (`=`, compound, `++`, field and element writes), every `mutable` argument (`checkMutableArgument`) and the source of `for (mutable ...)`. `reportNeverReassigned` runs at the end of each body, after the unused-name check. A `let` that is also unused is reported as unused only.
- Checked: plain `let`, `for (let x of ...)` elements (the parser records the `let` position in `SourceFacts.letPositions`) and `let x = yield ...`. A `_` name is skipped, so the unused fix (`let _total`) stays valid. Not checked, a follow-up gap: destructuring `let` patterns.
- `deriveshape.nra` writes `const __fields` for a struct without fields and an enum without variants.
- SPEC.md rule next to "Unused names are errors"; guide chapter 2 teaches it with a checked example.
- Tests: `testdata/errors/let_const_never_assigned`, `let_const_in_loop`, `let_const_for_of`, `let_const_yield` (the `let_const_*` samples also run as apply-fix cases in `selfhost/tests/src/habits.nra`) and `testdata/let_const_writes` (every kind of write keeps its `let`).

## Proof

Not a fix node.

## Deviations

- Sweep: 19 sites in 15 files, found by the compiler's own E1510 positions and applied by script; every hunk is `let` to `const`. selfhost/src: fixes.nra, fmt.nra, lspnavigation.nra (1 each). bench/particles/particles.nra (2). testdata: arg_order_arrow_writes, box_assign, closure_borrow, copy_enum_credit_kept (1 each), let_narrowing (4), typesdump/patterns/err_narrowing (1), errors/fixed_arrays_shared_write, for_of_sources_range_key, maps_partial_traits, shared_write_reverse, shared_write_scalar_for_const (1 each). No `for (let x of ...)` or yield site needed a change. `deriveshape.nra` also emits `const __fields` for an empty struct or enum. The goldens of maps_partial_traits and for_of_sources_range_key moved two columns.
- Edited outside the owned paths: `selfhost/src/{check,stmts,exprs,deriveshape,diagnostics,ast,parser}.nra`, `selfhost/tests/src/habits.nra`, the sweep files above, SPEC.md and the guide (shared).
- The sweep is not a separate commit; the whole node is one commit, as the orchestrator asked.
- Passing `mutable x` counts as a write, so a `let` array only pushed to stays `let` (the brief's rule). The guide still says a `let` only adds renaming; a stricter rule would be a later node.

## Stubs left

- None.

## Gotchas

- Generated code that declares `let` must pick `const` when it writes nothing (deriveshape did).
- The seed is not regenerated; stage 1 is built by the seed and enforces the rule on selfhost, so a new `let` that is never written breaks the bootstrap.

## Proposed learnings

- Declare a name with `const` unless it is assigned later | the compiler rejects an unwritten `let` | all .nra | yes (E1510)

## Verification

- `sh scripts/test.sh`: see the final line in the report.

## Size

- Estimated M, actual S: the sweep was 8 sites.
