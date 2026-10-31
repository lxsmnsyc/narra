---
goal: Enum constructors are written qualified (SchoolPerson.Teacher) instead of bare (Teacher) after importing the enum
date: 2026-10-31
rounds: 1 of 1
status: approved
roadmap: none
supersedes: none
decision: DECISIONS.md D71, D72
---

# Debate: qualified enum constructors

A short debate: one round, written and judged by the orchestrator, with the user's answers on the two open points. The ledger is `ledger.json`, and the measurements are under `evidence/`.

## Settled structure

1. A user enum's constructors are always written qualified, in expressions and in patterns: `SchoolPerson.Teacher`, `SchoolPerson.Student("x")`, `SchoolPerson.Teacher =>`. Importing `SchoolPerson` is enough to use them. (P1, amended by Q1)
2. `Some`, `None`, `Ok` and `Err` stay bare everywhere, because they are prelude names. (P1, O5)
3. A bare user constructor and `import { Teacher }` are compile errors with an exact fix: write `SchoolPerson.Teacher`, or import `SchoolPerson`. (P3)
4. The ambiguity rule goes away. A constructor always names its enum, so a local can no longer hide one, and two enums can share a constructor name. (P4, J2, J3)
5. An import alias may not equal any type name visible in the module (prelude, declared or imported), as in TypeScript. `enum Shape` with `import * as Shape` is an error, so `Shape.Circle` always names a constructor. This replaces D67's rule that an alias may equal a type name, and its "the module wins for `Box.new`" case. (Q3)
6. The std modules whose type is capitalized take plural aliases, and their paths stay as they are: `import * as Arrays from "std/array"`, `Maps`, `Boxes`, `Ranges`, `Options`, `Results`. The error from item 5 suggests the plural as its fix. `Int`, `Str`, `Float` and `Char` are unaffected because their types are lowercase. (Q4)
7. Constructors with the same name in different enums, of any shape, never clash. A user `Maybe.None` and the prelude `None` coexist, and a bare `None` is always the prelude one. Duplicate constructor names within one enum stay an error. (Q3)
8. Named imports accept renaming, as in TypeScript: `import { map as mapArray } from "std/array"`. The new name follows the same rules as any other name in the module: it may not clash with a declaration, another import or a type name. This amends D25, which allowed only `{ a, b }`. Named imports from std modules already work (`import { push } from "std/array"`). Today `as` is a parse error, `expected "}", found "as"`. The plural aliases in item 6 are only the convention for docs and fixes, not a requirement. (Q5)
9. The migration applies the fixes from items 3 and 5 to selfhost and testdata in one mechanical pull request. (P6, J7)

## Budget

- device: none of note.
- time: about 700 std alias call sites and 107 imports (`Array` alone is 52 imports and about 550 calls), plus about 1,100 constructor sites (about 190 user-enum expressions and 500 patterns in selfhost, about 400 in testdata), all by the compiler's own fix. The selfhost part must land between selfhost waves, because 21 selfhost nodes are still todo. (J1, J6, J7, O1)
- tokens: one compiler node (checker change, diagnostic, tests) plus the migration pull request. No per-site agent work.
- language: longer names, most of all where enums already prefix their constructors (`Type.TStruct`, `Token.TokIdent`). Those prefixes can be dropped later (`Type.Struct`). (O2, P7)

## Rejected

- Bare constructors in patterns (P2): the user chose one rule over shorter match arms. (Q1, O3)
- Keeping bare constructors with better errors and a written ambiguity rule (P5): a local can silently change what a bare constructor means, which D2 forbids. (J3)
- Swift-style `.Teacher` shorthand (O4): no TypeScript precedent, and the user ruled it out for good. (Q2)

## Proofs

- none

## Prior art

- TypeScript and Kotlin always qualify enum members. Rust requires `Enum::Variant` unless `use Enum::*`, and keeps `Some`/`None`/`Ok`/`Err` bare through its prelude. OCaml, Haskell and Gleam use bare constructors. (J5)

## Open risks

- Migration collides with selfhost work (O1, J6): accepted. Land it between waves, and rebase open selfhost branches with the same fix.
- Stuttering prefixes in selfhost (O2): accepted. Renaming them is an optional follow-up node, not part of this change.

## Answers from the user

- Must match arms and `is` patterns qualify too? Yes, qualified everywhere.
- Add `.Teacher` shorthand? Never.
- May an import alias equal a type name (D67)? No. TypeScript rejects that, so it is an error.
- Which alias do the std modules take? Plural (`Arrays`, `Maps`, `Boxes`, `Ranges`, `Options`, `Results`), with paths unchanged.
- Named imports with renaming (`{ map as mapArray }`)? Yes, added to the settled structure. Plural aliases are only a convention.

## Overruled

- P2: the judge upheld bare names in patterns, and the user rejected them (Q1).
- O3: the judge left it open, and the user upheld it (Q1).
- O4: the judge left it open, and the user rejected it (Q2).

## Cost of this debate

- 1 round, about 60k tokens
