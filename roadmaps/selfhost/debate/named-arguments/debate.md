---
goal: Decide whether Narra adds named arguments at the call site, such as foo(a = 1, b = 2), amending D12, and in what form
date: 2026-11-07
rounds: 2 of 2
status: approved
roadmap: roadmaps/selfhost
supersedes: none
decision: DECISIONS.md D87
---

# Debate: named arguments

## Settled structure

Narra adds Kotlin-style named arguments that can be written in any order (Q1). Every argument is still passed; parameter defaults are a separate, later debate (Q2, P9). D12 changes only its "no labeled arguments" clause.

Rules (P6, P8):

- `foo(count: 3, flag: true)`, spelled like a struct literal field (`Point { x: 1, y: 2 }`). An argument that starts with a name and `:` is named. No expression starts that way (an arrow starts with `(`, a conditional's `:` comes after its `?`), and in TypeScript `f(a: 1)` is a syntax error, so the text has no second meaning (O12, maintainer's choice at approval).
- Positional arguments come first, then named ones, in any order. A positional argument after a named one is an error. Each parameter is passed exactly once, by position or by name.
- `mutable` composes: `buf: mutable xs`.
- Labels work only when the callee is a declaration with parameter names: user functions, imports, std functions written in `.nra`. A call through a function value or an unnamed Go-registered builtin (22 registrations) gets one error, "this callee has no parameter names, pass arguments by position". Names are added to builtins only when someone needs them.
- A misspelled name gets an error with an exact fix (D11). No formatter or lint adds labels (P13).

Lowering (P7 in part, O6):

- The checker matches names to parameters and gives the emitter the arguments in declaration order, so each function keeps one C function with a fixed parameter list (D12).
- Arguments are evaluated in the order they are written, as in Kotlin. A reordered argument with side effects is evaluated into a temporary first. This needs the PR #96 evaluation-order defect (inbox) fixed first, or fixed in the same node.
- The AST keeps the written labels and order for astdump, `narra query` and `narrac fmt`. Passes that walk `call.Args` must take one view, which N2 decides (O7).

Nodes, each built in Go, then selfhost (P10, O7):

1. N1 syntax: `Arg` gets a name; parser, astdump, fmt in both compilers. About 150k.
2. N2 checker and lowering: resolve names, reject duplicates, missing and misplaced arguments, permute, written-order temporaries, differential sample test in both compilers. About 400k. Depends on N1 and on the evaluation-order fix.
3. N3 `narra query` (refs, callers) reads labelled calls. About 100k.

## Budget

- tokens: about 650k across three nodes (P10, O7; the 150k for temporaries is an estimate)
- time: three nodes in order, N2 after the evaluation-order fix
- device: not touched

## Rejected

- Optional labels in declaration order only, with `:` (P1, P2): Q1 asks for out-of-order arguments (P11, P12).
- Leaving evaluation order unspecified for reordered calls (P7's second half): Kotlin evaluates in written order, and the inbox already lists the current order as a defect (O6).
- Formatter or lint inserting labels (P4): fmt is a syntax-only pass (O5, P13).
- Editor inlay hints instead of a language change (A1), labels required only on bare booleans (A2), enum refactor of flag parameters (A3): none gives out-of-order arguments (Q1). The enum refactor of about 30 flag-taking callees can still be done on its own (O9).

## Proofs

- None.

## Prior art

- Kotlin uses `foo(a = 1)` with mixing, defaults and written-order evaluation; Swift makes labels part of the name; Python has keyword arguments; C# uses `foo(a: 1)`; TypeScript, Rust and Go have none (J6).

## Open risks

- Parameter names become public API for callers that name them (P5): accepted, as in Kotlin and Swift; unnamed calls stay valid.
- Builtins accept positions only (O10): accepted; the most-called builtins take one or two arguments.
- Two views of one call's arguments (O7): N2 decides one canonical view for every pass; covered by its differential test.
- Defaults would amend D12 a second time (O11): accepted by Q2.

## Inferred choices

- Positional before named, never after (P6), from Kotlin's original rule.

## Answers from the user

- What should named arguments give? Out-of-order assignments, like in Kotlin.
- Should they bring parameter defaults? Later, in a separate debate.
- `=` or `:`? `:`, at approval.

## Overruled

- P6: spelling changed from `=` to `:` by the maintainer, to match struct literal fields and avoid TypeScript's assignment reading of `f(a = 1)` (O12).

## Cost of this debate

- 2 of 2 rounds, about 196k tokens for the two agents
