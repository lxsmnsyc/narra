---
goal: Replace the enum keyword with ML-style sum type syntax (type SchoolPerson = | Teacher | Student(name: string)), keeping today's semantics
date: 2026-10-31
rounds: 1 of 1
status: approved
roadmap: none
supersedes: none
decision: DECISIONS.md D73
---

# Debate: ML-style sum type syntax instead of `enum`

A short debate: one round, written and judged by the orchestrator. The question is syntax only. Semantics stay as they are. The ledger is `ledger.json`, and the measurements are under `evidence/`.

## Settled structure

1. Keep `enum`. The ML form would be a regression for Narra's target users. (O1, J3)
2. Add a TypeScript-habit error for `Red = 1` inside an enum, with a fix pointing to `@repr` and `Enum.fromInt`. Today it is a bare parse error. (O3, J2)

## Budget

- device: none.
- time: the error with its fix is one small compiler change plus a `testdata/diagnostics` case. The rejected switch would have rewritten 68 declarations and both parsers. (O2, J6)
- tokens: one small node.
- language: no change to how existing code reads.

## Rejected

- `type X = | A | B` (P1): in TypeScript, `type Shape = Circle | Square` is a union of existing types. In Narra the same text would declare new constructors. That is the "same text, two meanings" D2 forbids, and it is the form TypeScript users and models write most. (O1, J3)
- "`enum` already breaks D2" (P2): every TypeScript enum that Narra accepts keeps its meaning. Fieldless members are read as `Color.Red`, as in TypeScript (J2, J7). Payloads are new syntax with no TypeScript meaning, the same path Rust and Swift took. (J5)
- One `type` keyword for everything (P3): `type` would mean alias or new nominal type depending on the first token after `=`. `struct`, `enum` and `trait` each keep their own keyword. (J4)

- Moving structs to `type Point = { x: f64, y: f64 }` as well (user follow-up, judged inline): in TypeScript that text is a structural alias, so two types with the same fields are interchangeable. Narra structs are nominal, so this is D2's "same text, two meanings" again. `type P = Point` would stay an alias while `type P = { ... }` makes a new type, so the parser and the checker still split on the right-hand side. No case in the checker or emitter goes away.

## Proofs

- none

## Prior art

- `enum` with payloads: Rust, Swift, Haxe, Kotlin (`enum class` and `sealed`). `type ... = | A | B`: OCaml, F#, Haskell, Elm, Gleam, ReScript. ReScript reaches JavaScript users with ML syntax throughout, not with TypeScript syntax. (J5)

## Open risks

- none

## Answers from the user

- "Variant types" means ML-style syntax with today's semantics. It does not mean variants as their own types or open polymorphic variants.

## Overruled

- none

## Cost of this debate

- 1 round, about 30k tokens
