---
id: nested-void-match
title: "A match with no value inside a match arm"
status: active
priority: could
type: fix
story: batch-22
early: false
outline: false
milestone: p6-docs
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
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/nested-void-match/**
shared:
  - selfhost/src/emit/**
  - selfhost/src/check/**
  - testdata/**
  - selfhost/tests/src/**
reserves: []
provides: []
depends_on: []
---
# A match with no value inside a match arm

## Goal

Found by the std-text-more reviewer, confirmed by the orchestrator on feat/batch-22 (2026-11-20). This program passes `narrac check`, but `narrac run` fails in the C compiler with `error: variable has incomplete type 'void'` (`void r8;`):

```narra
function report(value: Result<int, string>, flag: bool) uses io {
  match (value) {
    Ok(_number) => match (flag) {
      true => print("shown"),
      false => print("hidden"),
    },
    Err(message) => print(message),
  }
}
export function run(args: Array<string>): int uses io {
  report(Ok(3), true)
  report(Err("bad"), false)
  return 0
}
```

The same shape with string results works. The emitter declares a result temporary for the inner match even when its type is void. Fix the cause in the emitter (find every place a match, `if` expression or block expression declares a result temporary, and skip it for void), and check the same shape for `? :`, `if` expressions and blocks as arm bodies, and for deeper nesting.

## Acceptance

- [ ] A case with the program above (and the other void shapes found) fails on the batch base and runs on the branch, clean under the sanitizer and leak check.
- [ ] `sh scripts/test.sh` passes.
