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

Done on 2026-11-12: the seed updates, `@inline` on `Options.map`, `andThen` and `orElse`, both checks and the debug copy check (the Linux check found the `@const` nesting limit past the gcc stack; fixed in eb21b81b, limit 5000), the curator pass, and a fresh review of the whole branch (no blocking or should-fix findings).

1. Bench row, on an idle machine (1-minute load under 2): `bash bench/run.sh -n 9`, then an "end of batch-11" section at the top of `bench/BASELINE.md` in the batch-10 format. Instructions retired by `narrac emit selfhost` (median of 5, `/usr/bin/time -l`, compilers built from their seeds with `cc -O2 -DNARRA_RELEASE`): on the batch-10 source, 2.408G with the batch-10 compiler and 2.453G with the batch-11 compiler (+1.9%); the batch-11 compiler on its own source 2.902G. The source grew from 65,405 to 75,144 lines; the seed is 215,778 lines. A per-commit bisect puts the one lasting step at array-patterns (+1.3%), within the 2% code layout spread (inbox entry). Rerun the instruction numbers on the idle machine too.
2. Tick the last two acceptance boxes in `nodes/batch-11/brief.md`.
3. Push `feat/batch-11` (the repo is public) and open one pull request into `main`, marked ready: the review was clean before it opened.

## Inbox from this batch

Fix nodes with a failing test: polymorphic recursion hang, argument evaluation order with an arrow that writes a captured `let`, shared results through a bound `<T: Trait>`, generated Shape names reachable from user code. Follow-ups: const-functions-std, arrays of strings in `@const` results, literal table constants through the `static const` writer, small review nits.
