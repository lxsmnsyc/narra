---
id: mutable-reference
title: "mutable changes the contents of a struct, array or map, never the binding"
status: active
priority: should
type: spike
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: S
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-04
spent:
owns:
  - roadmaps/selfhost/nodes/mutable-reference/**
shared:
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [decision D81]
provides: []
depends_on:
  - on: mutable-read-conflict
    kind: hard
    why: the soundness fix ships first; the redesign builds on what it learned (done)
---

# mutable changes the contents of a struct, array or map, never the binding

## Goal

Write the maintainer's model of `mutable` as D81 and SPEC text, measure what it changes in `selfhost/src`, `std/src`, the test programs and the guide, and stop at the gate. Node mutable-bindings builds it.

The maintainer (2026-11-07), correcting the first draft of this brief: `mutable` only affects structs, so a `let` cannot be passed as a `mutable` parameter for its own sake. Chosen form, from three offered:

```narra
function damage(mutable p: Player, n: int) {
  p.hp = p.hp - n     // ok
  p = other           // error: cannot replace
}
const hero = Player { hp: 100 }
damage(mutable hero, 10)   // ok, hero.hp is 90

let count = 0
bump(mutable count)        // error: int is not a struct
```

So `const` fixes the name, as in TypeScript, and `let` exists only to reassign the name. `mutable` is a reference to a value's contents, never to the variable.

## Context

- Today (SPEC section 6 "Mutable parameters"): the argument must be a `let`, a `mutable` parameter, or a field or element of one; the callee may replace the whole value; any type is allowed. D50 (pointer passing), E0319, E0929 (`nodes/mutable-read-conflict/outcome.md`), E0915 (`nodes/mutable-escape/outcome.md`) and D77 (closure-borrow) stay.
- In `selfhost/src` and `std/src`, three `mutable` parameters have type `int` (`patterns.nra:358 limit`, `declsdump.nra:118` and `:267 nextImpl`) and 68 have type `Array` or `Map`, mostly for `Arrays.push`.

## Plan (D81)

1. A `mutable` parameter's type must be a struct, an `Array` or a `Map` (types whose contents can change in place). Any other type (numbers, `bool`, `char`, `string`, enums, tuples, function types, `Option`, `Result`) is a compile error at the parameter. A generic `T` parameter is allowed only with a bound that guarantees one of these, otherwise an error (the spike says which bound, or rejects generic `mutable` parameters if none fits).
2. Through a `mutable` parameter the callee may assign fields and elements, at any depth, and pass the value or a part of it on as `mutable`. It may not assign the parameter itself (`p = other`). `Arrays.push`, `Maps.set` and the like keep working, since they change contents.
3. The argument may be any binding that holds such a value: `const`, `let`, a `mutable` parameter, or a field or element path of one. A literal or a call result is still an error.
4. `const` fixes the name only. Fields and elements of a `const` struct, array or map can also be assigned directly (`hero.hp = 1`), so passing it as `mutable` adds no power that direct code lacks. `let` adds reassignment of the name.
5. Value semantics stay: `const copy = hero` is a copy (copy on write, as today); changing `hero` afterwards does not change `copy`. A plain (not `mutable`) parameter is still a read-only copy.
6. Out of this node: the local reference binding `mutable name = place` and `x = f(mutable x)` from the first draft. They go to the inbox, to be weighed again once this model ships.

## Acceptance

- [ ] DECISIONS.md D81: the model above, the maintainer's choice and the two rejected forms (keep `const` deeply immutable and require `let`; a third binding kind `mutable name = value`), reasons, prior art (TypeScript `const`, Swift `inout`, Hylo), and the cost: `const` no longer means deeply immutable.
- [ ] SPEC section 2 "Bindings" and section 6 rewritten to match; the guide chapters that teach `mutable` and `const` (chapter 9 at least) listed for mutable-bindings to rewrite. `compiler/README.md` "Differences from the spec" says the new rules are not built yet.
- [ ] The outcome counts, in `selfhost/src`, `std/src`, `compiler/testdata` and `docs/guide`: `mutable` parameters of other types (and how each is rewritten, for example returning the new value), assignments to a whole `mutable` parameter, and generic `mutable` parameters. Two before and after examples.
- [ ] The outcome answers rule 1's generic bound question.

## Sizing

S. Documents and counts.

## Decision to confirm

The maintainer approved rules 1 and 4 on 2026-11-07 (arrays and maps count; direct field writes on a `const`). The gate stays for the counts in the outcome.


The plan above, in particular rule 4 (direct field writes on a `const`, which follows from "`const` fixes the name") and rule 1's list of allowed types (arrays and maps count, because 68 `mutable` parameters in `selfhost/src` and `std/src` are arrays or maps).
