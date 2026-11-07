---
goal: Decide whether Narra adds async blocks and generator blocks (block expressions that suspend), and in what form
date: 2026-11-07
rounds: 1 of 2
status: approved
roadmap: roadmaps/selfhost
supersedes: none
decision: DECISIONS.md D88
---

# Debate: async blocks and generator blocks

## Settled structure

Narra adds neither block now (A1, O7, Q1). Both stay parked until a program in `selfhost/`, `std/` or an example needs one.

1. Keep today's forms. An inline async body is an async arrow, `Async.spawn(loop, async () => { ... })`. An inline generator is a top-level `function*` that takes the values it needs as arguments (D74).
2. Build the `async` node with async functions and async arrows only.
3. Replace the 2026-11-07 inbox entry with a pointer to this debate and the revisit conditions below.

When a program needs one of them, rerun this debate from these designs:

- Async block (P1): `async { ... }` means exactly `async () => { ... }`. It ships only after the `async` node (O3), and it is a second spelling of an async arrow (O4).
- Generator expression (P2): TypeScript's anonymous `function* (): Generator<T> { ... }`, capturing constants only (D77). `yield` stays banned in arrows and block expressions. Expect size L, 450k tokens or more (O2). Share the capture work with `cell-capture` if both are built.
- Either form keeps every suspension point visible (P5): nothing runs until `Async.spawn`, `Async.run`, `for ... of` or `Co.resume` starts it.

The second round was skipped. The maintainer's answer to Q1 decided the structure, and the claims still open (P3, O1) only matter if the blocks are built.

## Budget

- tokens: 0 now. About 450k or more for the generator expression and a small node after `async` for the async block, if revisited (O2, O3)
- time: nothing on the roadmap changes
- device: not touched

## Rejected

- Async block now (P1, A2): it cannot ship before the XL `async` node, and it only saves `() =>` (O3, O4).
- Generator expression now (P2): no program needs one: 0 `function*` declarations in `selfhost/src` and `std` (O5, J4).
- Ship order async, then P2, then P1 (P4): nothing ships now (Q1).
- O6 (an inline generator hides an allocation) was rejected: an escaping arrow already allocates its environment where it is written (SPEC.md line 113), and J11 is about hidden calls.

## Proofs

- None.

## Prior art

- Rust has async blocks (stable 1.39) and async closures (1.85); `gen` blocks (RFC 3513) are unstable. Kotlin builds generators with `sequence { yield(x) }`. TypeScript has neither block and uses anonymous `function*` expressions and async IIFEs. C# lambdas cannot be iterators (J7).
- This repo's local-functions debate (D74) rejected lifting nested functions by renaming at parse time (O1).

## Open risks

- Lowering a block by lifting it to a private top-level function (P3, O1): accepted for now, because nothing is built. Capture analysis already exists for arrows (`check_closure.go`, `selfhost/src/funcvalues.nra`), so PR #66's defects do not carry over. Enclosing type parameters and effects at the lifted function are a spike if the debate is rerun.

## Inferred choices

- Roadmap priorities: a spec feature with no current user is `could` (J9), from `roadmaps/selfhost/GOAL.md`.
- TypeScript syntax where the meaning is the same (J10), from SPEC.md line 41.
- No hidden implementations (J11), from the maintainer's statement on 2026-11-07.

## Answers from the user

- Build async and generator blocks now, or park them until a program needs them? Park both.

## Overruled

- none

## Cost of this debate

- 1 of 2 rounds, about 90k tokens for the two agents
