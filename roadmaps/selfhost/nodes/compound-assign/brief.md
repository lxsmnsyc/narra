---
id: compound-assign
title: Compound assignment, increment and decrement
status: todo
priority: should
type: feature
outline: true
milestone: m9-adoption
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-02
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes after bootstrap go into the Narra compiler only
---

# Compound assignment, increment and decrement

## Goal

`x += e`, `x -= e`, `x *= e`, `x /= e`, `x %= e` and the bitwise and shift forms (`&=`, `|=`, `^=`, `<<=`, `>>=`) assign to a `let` binding, a `mutable` parameter, a field or an index, with the same meaning as `x = x op e`. String `+=` concatenates. `x++` and `x--` add or subtract 1 from an `Int` target. The target is evaluated once, so `items[next()] += 1` and `items[next()]++` call `next` once.

## Notes

Outline. This reverses SPEC section 2 ("There are no compound assignments (`+=`, `++`)"), so it needs a decision record and a SPEC update. Rules to settle there: the statement form only (no value, as with `=`), so `x++` and `++x` mean the same and the decision record picks one spelling, or allows both; overflow, division by zero and shift checks as for the binary operator; Map index targets (`counts[key] += 1`) and what happens when the key is missing; reference counting of the old value for string `+=`. Add samples and compile-error tests, then use it in the selfhost compiler where `x = x + 1` and `index = index + 1` are common.
