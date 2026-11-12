# Wave 24 report: batch-11 children landed

Batch-11 runs on the local story branch `feat/batch-11` (worktree `../narra-batch-11`, not pushed). Every child has landed as one direct commit after a worker check and a fresh review, with the full suite run on the rebased tree before landing. Last full run: `test.sh: PASS, 8103 cases` (const-functions, on the batch head).

## Landed

- contract-logic-design (D102) and const-eval-design (D105): approved by the maintainer in session on 2026-11-12. The maintainer chose to build relation facts between two parameters (the worker had proposed deferring them).
- std-function-values: registry std functions as values through a thunk; `enterBodyContext`, `leaveBodyContext` and `findSyntheticCode` in emitstate.nra.
- array-patterns: `[]`, `[x]`, `[first, ...rest]`; no match is judged again in a generic instance body.
- generic-implement: constraints on struct and enum type parameters (E1020), `implement` on generic types (E1021, D106); members named `<block index>_implement_<member>`; `findImplementation` in implementations.nra.
- inline: `@inline` copies bodies and arrow arguments into calls (E1030, E1031).
- contract-logic: `||` clauses with a one-interval hull, relation facts between two parameters, the `!` message.
- distinct-traits: `implements`/`derives` on distinct types (E1012, D108); `implement Equatable`/`Hashable` on a distinct type is E1014 (orchestrator decision, D62 stays).
- lazy-values: `const x = lazy expr` and `Lazies.force` (E1032 to E1035).
- dynamic-traits: `Dynamic<Trait>` and `Dynamic.of` (E1022 to E1025); `Dynamic<Hashable>` allowed (orchestrator decision); shared results through a Dynamic tracked.
- tuple-index: `pair[0]` (E1010, E1011).
- generic-field-defaults: defaults on fields whose type uses a type parameter.
- const-functions: `@const` functions computed by the compiler (E1050 to E1057); limits recorded in D105's status line.

const-lazy is dropped (split into lazy-values and const-eval-design).

## Still to do (end of batch)

1. `sh scripts/seed-update.sh` twice; `sh selfhost/bootstrap.sh` reports the seed current.
2. Mark `Options.map`, `Options.andThen` and `Options.orElse` in `std/src/option.nra` `@inline`, regenerate `stdtext.nra` (`sh scripts/test.sh --suite generated --update`), then one more seed update.
3. `sh scripts/test.sh`, `sh scripts/linux-check.sh`, and a debug build of the seed running `emit selfhost` with `NARRA_CHECK_COPIES=1` (exit 0).
4. Curator pass: remove the inbox entries this batch closed, calibration rows, check `compiler/README.md` "Not supported yet" lists none of the built features, tick the batch-11 acceptance boxes.
5. A fresh review of the whole branch, then the bench row (load under 2), then one pull request from `feat/batch-11` into `main`.

## Inbox from this batch

Fix nodes with a failing test: polymorphic recursion hang, argument evaluation order with an arrow that writes a captured `let`, shared results through a bound `<T: Trait>`, generated Shape names reachable from user code. Follow-ups: const-functions-std, arrays of strings in `@const` results, literal table constants through the `static const` writer, small review nits.
