---
id: contract-logic
title: "Logical operators and relations in requires"
status: done
priority: should
type: feature
story: batch-11
early: false
outline: false
milestone: 
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-11
spent: 570000
owns:
  - roadmaps/selfhost/nodes/contract-logic/**
  - testdata/contracts_logic*
  - testdata/contracts_relations*
  - testdata/errors/contracts_logic*
  - testdata/errors/contracts_relation*
  - testdata/errors/divisor_or_*
  - testdata/errors/shift_relation*
shared:
  - selfhost/src/contracts.nra
  - selfhost/src/check.nra
  - selfhost/src/operands.nra
  - selfhost/src/patterns.nra
  - selfhost/src/diagnostics.nra
  - testdata/errors/contracts_unsupported.nra
  - testdata/errors/contracts_unsupported.err
  - testdata/proven_operands.nra
  - testdata/proven_operands.out
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: array-patterns
    kind: hard
    why: both edit selfhost/src/contracts.nra and selfhost/src/diagnostics.nra; array-patterns lands first in batch-11, and this node rebases on it rather than merging two edits of the same functions
  - on: contract-logic-design
    kind: hard
    why: builds D102, approved by the maintainer on 2026-11-12
---

# Logical operators and relations in requires

## Goal

Build D102 (node contract-logic-design, approved 2026-11-12): `||` clauses in `requires`, relations between two integer parameters, the hull of a one-name `||` and relation facts in D92 checks, and the `!` message. No `ensures` change: `&&` inside a pattern stays a parse error, and several `ensures` lines already give callers both bounds.

### Part 1: `||` clauses

- **Clauses.** A `requires` condition is clauses joined by `&&`. A clause is one comparison, or comparisons joined by `||`. Parentheses group a clause: `requires (index >= 0 || index == -1) && count != 0`. `&&` under `||` (`a > 0 && b > 0 || c > 0`) is E0984 with the detail `split it into clauses joined by &&; && inside || is not supported`.
- **Storage.** `Requirement` (`selfhost/src/check.nra:654`) gets `clause: int` and `other: int` (the second parameter of a relation, -1 for a constant). Comparisons with the same clause index in one function are alternatives; a comparison alone is its own clause. `checker.requirements` keeps its type.
- **Call proof.** `proveArgument` proves per argument today. Prove per clause: a clause holds when one of its comparisons is proven. Report one E0985 per unproven clause, at the argument of the clause's first parameter, naming the whole clause joined by ` || `. One-comparison clauses keep today's text, and failed one-comparison clauses on one argument still join with ` && `. The constant message (`20 does not satisfy ...`) applies when every parameter of the clause got a constant.
- **Body hull.** `checkRequires` refines a parameter by each one-comparison clause as today. A clause whose comparisons all name one parameter narrows it by the hull of the sides. Build each side as `refineFact` would (`operands.nra:350`): on a `u64` name, a side with a bound from 2^63 up is the unknown fact. Join the sides with `joinFacts`, then `narrowFact`. `requires d < 0 || d > 0` lets the body divide by `d`. A clause over two parameters adds no fact.
- **Hull in checks (D92 amendment).** `narrowByBinary` (`operands.nra:301`) narrows nothing for a true `||` or a false `&&`. When every side is a comparison of the same integer local against a constant, narrow that local by the hull of the sides, built as above (the false `&&` negates each side first, with `negateComparison`, exported). Anything else narrows nothing. `if (d < 0 || d > 0) { total / d }` compiles, and `if (x == 9223372036854775813 || x == 1) { y << x }` on a `u64` stays E0980.

### Part 2: relations

- **Requirements.** A comparison of two integer parameters of the same type (`<`, `<=`, `>`, `>=`, `==`, `!=`) is a relation requirement, alone or as a side of a `||` clause. An offset (`low + 1 <= high`) is E0984 with the detail `compare two parameters with no arithmetic; + and - wrap`. The same parameter on both sides is E0984 with the detail `this compares the parameter with itself`. The detail for two parameters of different types stays the type error the comparison already gives.
- **The fact.** A set of possible orders between two integer locals: bit 1 less, bit 2 equal, bit 4 greater, for the local with the smaller index against the other. Stored in `checker.localNames` under `relation|<a>|<b>`, written with `bindName` so scopes undo it like intervals. Narrowing intersects (AND). A set proves a comparison when it is a subset of the comparison's set. An empty set means the branch cannot run; it proves everything, as an empty interval does today.
- **Ending.** Each local with relations keeps a list in `localNames` (`relations|<a>` holds the count, `relations|<a>|<k>` the k-th partner), also through `bindName`. `endNarrowing` (`patterns.nra:119`) ends every relation in that list, on both sides, beside the variant and interval keys. A local captured by an arrow function reads no relation, as `findOperandFact` reads no interval for it.
- **Recording.** `narrowByBinary` records a relation for a comparison of two integer locals of the same type, in every place it narrows today: the `then` and `else` branches, `while` bodies, the right operand of `&&` and `||`, and the rest of a block after an `if` whose `then` branch ends. `checkRequires` records each one-relation clause at the start of the body.
- **Bound step.** When a relation is recorded, and when `refineFact` or `narrowFact` changes a local that has relations, each partner takes the bound the relation passes on: with `a <= b`, `a` takes `b`'s high end and `b` takes `a`'s low end; `<` passes the end less or plus one, without passing an unbounded end; `==` passes both ends. One step only: the partner's own partners are not visited, and relations do not chain. On a `u64` an end from 2^63 up is no bound and is not passed (as `narrowFact` treats it).
- **Call proof.** At `recordContractCall`, for each relation requirement of the callee, record the relation set between the two argument expressions: from two constants (ordered as the type orders them; a `u64` constant from 2^63 up is above every other), from intervals with bounded ends that settle the order, from the same local passed twice (equal), or from the relation fact between the two argument locals. `ContractCall` gets `relations: Array<int>`, one set per relation requirement in order, since `checkContractCalls` runs after narrowing has ended. The proof is a subset test. An unproven relation is E0985, at the first parameter's argument, naming the relation (`requires low <= high of span`).
- **What a body can prove with it.** Calls to other functions with the same or a weaker relation; shift counts, divisors and `FixedArray` indexes through the bound step (`requires low >= 0 && low <= high && high < 64` then `values[low]` on a `FixedArray<int, 64>`); `ensures` returns from those bounds. Not `high - low`: it wraps (D92).

### Part 3: `!`

- `requires !(x < 0)` stays E0984 with the detail `write x >= 0`: negate the comparison (`negateComparison`), and for `!(a || b)` and `!(a && b)` push the negation inward and print the result. When the operand is not made of comparisons, keep today's detail.

### Messages

- The E0984 format (`diagnostics.nra:971`) changes its fixed text to `requires compares integer parameters with integer constants or with each other, joined by && and ||; %s`. The `||` and "two parameters compared" details in `describeUnsupported` go away. E0985 formats keep their shape (`%s` is the clause text). No new code.

## Context

- D102 in DECISIONS.md, with the counts, the `u64` rule and the rejected forms. D92 (checks) and D95 (contract amendments) for the rules this extends.
- `selfhost/src/contracts.nra`: `collectRequirements`, `findRequirement`, `describeUnsupported`, `checkRequires`, `recordContractCall`, `proveArgument`. `selfhost/src/operands.nra`: `narrowByBinary`, `refineFact`, `refineBounds`, `joinFacts`, `narrowFact`, `findOperandFact`, `negateComparison`. `selfhost/src/patterns.nra`: `endNarrowing`, `factKey`, `endKey`.
- The plan review (2026-11-12) asked for the Pattern switches as shared files. D102 adds no `Pattern` variant, so `ast.nra` and the pattern switches of the checker, emitter, `fmt` and `astdump` are not touched.
- `testdata/errors/contracts_unsupported.nra:21` expects E0984 on `requires count < 0 || count > 10`, and its "two parameters compared" case now compiles. Move both to positive samples, put `&&` under `||` and an offset relation in their place, and rewrite the `.err` golden.
- Samples, `testdata/contracts_logic.nra`: a one-parameter `||` clause proven by a constant on either side and by a name narrowed to one side; `requires d < 0 || d > 0` dividing by `d`; `a != 0 || b != 0` proven by either argument; a clause beside a plain comparison.
- Samples, `testdata/contracts_relations.nra`: `requires low <= high` proven by constants, by non-overlapping intervals, by `f(n, n)`, by the caller's `if (low <= high)`, and by the caller's own `requires low < high`; the bound step proving a shift count and a `FixedArray` index; a relation inside a `||` clause.
- `testdata/proven_operands.nra` gets three checks: `if (d < 0 || d > 0) { total / d }`, the false `&&` in `if (d >= 0 && d <= 0) { ... } else { total / d }`, and `if (count > 0 && index >= 0 && index < count) { 1 << index }` with `count` checked `< 64` first (the bound step through a relation from a check).
- Error tests, `testdata/errors/contracts_logic_*.nra`: an unproven `||` clause (E0985 naming the clause), a constant that meets neither side, `&&` under `||` (E0984), `!(x < 0)` with the rewrite in the message, `!(x < 0 || x > 9)` rewritten to `x >= 0 && x <= 9`.
- Error tests, `testdata/errors/contracts_relation_*.nra`: an unproven relation (E0985), a relation ended by an assignment before the call, overlapping intervals that prove nothing, an offset and a self-comparison (E0984). `testdata/errors/divisor_or_u64.nra`: `if (x == 9223372036854775813 || x == 1) { y << x }` with `x: u64` stays E0980. `testdata/errors/divisor_or_two_names.nra`: a `||` over two locals proves no divisor. `testdata/errors/shift_relation_u64.nra`: `if (x <= limit)` with `limit: u64` narrowed only by `limit != 0` passes no bound to `x`, so `y << x` stays E0980.
- SPEC section 13: the condition rule names `||` clauses, relations, the hull, the bound step and the forms E0984 still rejects (`!`, an offset, a self-comparison, `&&` under `||`); the `ensures` paragraph says one bound per line. The D92 checks paragraph in SPEC section 4 ("Divisors and shift counts are proven") gets the hull and the relations. `compiler/README.md` line 161 changes to match.
- About 217 checks in `selfhost/src` and `std/src` compare two names, and each between integer locals now writes a relation, so `emit selfhost` does more work. Measure `narrac emit selfhost` instructions before and after (base 2.41G to 2.43G at batch-11 start) and record both in the outcome. A rise above 1% goes to the orchestrator before landing.
- No compiler code may use the feature until the seed is regenerated at the end of the batch.
- If the node runs long, land part 1 and part 3 first as one commit, then part 2.

## Acceptance

- [ ] `requires` accepts `||` clauses; callers prove one side per clause; the body gets the hull of a one-parameter clause.
- [ ] A true `||` or false `&&` over one local against constants narrows to the hull in `if`, `while` and early returns, and a `u64` side from 2^63 up gives no fact (`testdata/errors/divisor_or_u64.nra` stays E0980).
- [ ] `requires` accepts relations between two integer parameters of the same type; callers prove them from constants, intervals, the same local twice, or relation facts.
- [ ] D92 checks between two integer locals record relations, which end with either name's narrowing, and the bound step passes bounds one step without passing a `u64` end from 2^63 up.
- [ ] `!`, an offset, a self-comparison and `&&` under `||` are E0984 with the details above.
- [ ] The samples and error tests listed in Context pass, clean under AddressSanitizer and the leak check.
- [ ] SPEC sections 13 and 4 and `compiler/README.md` describe the new rules; D102 gets a "Built" line.
- [ ] `narrac emit selfhost` instructions before and after are in the outcome, and the rise is at most 1%.
- [ ] `sh scripts/test.sh` passes.
