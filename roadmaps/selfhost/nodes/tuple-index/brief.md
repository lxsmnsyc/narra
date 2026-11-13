---
id: tuple-index
title: "Reading a tuple element by index"
status: done
priority: should
type: feature
story: batch-11
early: false
outline: false
milestone:
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-12
spent: 177000
owns:
  - roadmaps/selfhost/nodes/tuple-index/**
  - testdata/tuple_index*
  - testdata/errors/tuple_index*
shared:
  - selfhost/src/exprs.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/stmts.nra
  - selfhost/src/sharedwrites.nra
  - selfhost/src/sharedresults.nra
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E1010, diag E1011]
provides: []
depends_on: []
---

# Reading a tuple element by index

## Goal

`pair[0]` reads the first element of a tuple `pair: #[int, string]`, with that element's type. The index is a constant (E1010 otherwise) inside the tuple's range (E1011 otherwise), so the read needs no runtime check and cannot panic. Tuple elements stay read-only: `pair[0] = 2` is E0226.

Error on the current compiler (built from the seed at `57d55087`):

```ts
const pair = #[1, "a"]
export function run(_args: Array<string>): int uses io, alloc {
  print(pair[0])
  return 0
}
```

`tuple1.nra:3:13: only arrays can be indexed, found #[int, string]` (E0210, `checkIndex` in `exprs.nra`). `pair.1` is a parse error (E0012 `expected a name, found "1"`).

Plan: the parser, AST, `fmt` and `astdump` already handle `target[index]` as `Expr.IndexExpr`, so there is no new node kind. The checker gives an `IndexExpr` on a tuple the element's type, and the emitter writes it as a field read of the tuple's C struct.

## Context

- SPEC section 4 "Composite types", the Tuples bullet (line 412), and section 2 "Literals" (line 252). SPEC does not say how to read one element today; the README lists it under "Not supported yet" (line 210). D53 (tuple syntax) and D81 rule 7 (tuple elements are not places, SPEC line 1239).
- A tuple type is a `Type.TStruct` whose `StructInfo.tuple` is true, with fields named `"0"`, `"1"`, ... (`types.nra`, the comment above `StructInfo`). The C fields are `f_0`, `f_1` (`emitStructLiteral` in `emitexprs.nra` writes `.f_${name}`; `tupleFieldName` in `emitmatch.nra` reads the names).
- Checker: `checkIndex` in `exprs.nra` (around line 1442) checks the index as `int`, then handles `TArray` and `TFixedArray`. Add the tuple case before the E0210 report. Read the index with `findTypedConstant(checker, index, Type.TInt)` (`operands.nra:162`), the helper `checkFixedIndex` uses for a constant FixedArray index, so an integer literal and an integer `const` both work.
- Emitter: `emitIndexRead` in `emitexprs.nra` (around line 345) dispatches on the target type. Add the tuple case and emit it like a struct field read, with the same retain and borrow rules.
- Writes: the assignment, compound assignment, `x++` and `mutable` argument checks in `stmts.nra` accept an `IndexExpr` as a place. A tuple target there is E0226 (`only variables, fields and array elements can be assigned`), an existing code. Check `pair[0] = 2`, `pair[0] += 1`, `pair[0]++` and `update(mutable pair[0])`.
- Sharing: `sharedwrites.nra` and `sharedresults.nra` treat an `IndexExpr` as a part of its target (two different literal indexes are different places, SPEC line 1087). A tuple element read shares its target as a field read does. Make sure nothing there assumes the target is an `Array`, `Map` or `FixedArray`.
- Other `Expr.IndexExpr` users that may assume an array or map target: `operands.nra`, `effects.nra`, `emitstmts.nra`, `emitmaps.nra`, `stdcopy.nra`, `closures.nra`, `contracts.nra`, `querysites.nra` (`grep -n "Expr.IndexExpr" selfhost/src/*.nra`). Each needs nothing or a tuple case.

## Acceptance

- [ ] Samples (`testdata/tuple_index*`): read each element of a pair and of a triple with mixed types (`int`, `string`, a struct, an `Array`); a nested read `grid[0][1]` on a tuple of tuples; an index from a top-level `const`; a read of a tuple returned by a call; a read inside a generic function over `#[A, B]`; a read used as a `match` subject and passed to a function. Strings and arrays read from a tuple stay alive after the tuple goes out of scope.
- [ ] Every sample is clean under the leak check and sanitizers.
- [ ] Error tests (`testdata/errors/tuple_index*`): an index from a `let` and from a parameter (E1010); `pair[2]` on a pair and `pair[-1]` (E1011); `pair[0] = 2` and `pair[0] += 1` (E0226).
- [ ] E1010 and E1011 are rows in `diagnostics.nra`; the `codes` suite passes.
- [ ] SPEC.md, the Tuples bullet of section 4, states the read, the constant index rule and that elements are read-only. D53 gets a note line for the syntax (see Notes). The README line "Reading a tuple element by index." is removed.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

S: one checker case, one emitter case, and a check that the place and sharing walks accept a tuple target. No parser, AST, fmt or astdump change.

## Notes for the worker

- Open design questions, with the recommended answer:
  - **Syntax: `pair[0]`, not `pair.0`.** `#[...]` comes from the JavaScript Records & Tuples proposal (D53), where `tuple[0]` reads an element, and TypeScript reads tuple elements the same way. It needs no parser change. Rejected: `pair.0` (Rust, Swift), new syntax that also makes the lexer split `pair.0.1`. Record the choice and the rejected form as a note line on D53, not as a new decision.
  - **What counts as constant.** Whatever `findTypedConstant` returns: an integer literal or an integer `const`. A local whose facts pin it to one value does not count, so an element's type never depends on flow facts.
- Messages for the code table: E1010 `a tuple index must be a constant integer; use a pattern such as \`const #[first, second] = pair\` to read the elements`, E1011 `index %d is out of range for %s, whose indexes are 0 to %d`.
- No deprecation paths: `pair.0` stays a parse error like any other unknown syntax.

## Plan review (2026-11-12)

- Eight `Expr.IndexExpr` readers may need a tuple case (operands, effects, emitstmts, emitmaps, stdcopy, closures, contracts, querysites). List any you edit under Deviations in the outcome.
