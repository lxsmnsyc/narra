---
id: nested-function-error
title: "A function inside a function is a dedicated compile error"
status: done
priority: should
type: feature
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: S
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestCompileErrors|TestHabitFixes|TestErrorCodes|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent: 72917
owns:
  - compiler/testdata/errors/nested_function_*
  - roadmaps/selfhost/nodes/nested-function-error/**
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/internal/narra/diagnostic_test.go
  - selfhost/src/parser.nra
  - SPEC.md
reserves: [diag E0930]
provides: []
depends_on:
  - on: ts-habit-fixes
    kind: hard
    why: same parser paths and TestHabitFixes harness; PR #69 is in review, so this lands after it instead of reopening it (debate O9, O10)
---

# A function inside a function is a dedicated compile error

## Goal

Settled by the goal debate `roadmaps/selfhost/debate/local-functions/debate.md` (approved 2026-11-04) and DECISIONS.md D74. On main, `function` inside a function body fails to parse with the generic `expected an expression, found "function"`. Make it a dedicated diagnostic, E0930: a function inside a function is not supported; declare it at module level as a private function and pass the enclosing values as arguments. No automatic fix: moving the declaration to module level changes the meaning when the body reads a name that a local shadows, and breaks helpers that use the enclosing type parameters (debate O8, measured).

## Acceptance

- [ ] `function` as a statement inside any function, arrow, implement-block function or test body gives E0930 at the `function` keyword, in both compilers, with the same message, position and exit code (the `TestSelfhost*` harnesses cover the new error tests).
- [ ] No fix is offered (`--json` has `fixes: []`); `narra explain E0930` explains why and shows the module-level form.
- [ ] `function*` and `async function` inside a body get the same error (they are not supported there either), unless SPEC says otherwise.
- [ ] SPEC.md line 62 says local function declarations are not supported, with the reason, and "Differences from TypeScript" gets a row (nested function declaration -> declare it at module level).
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap`.

## Notes

- The PR #66 design and its six defects are recorded in `roadmaps/selfhost/nodes/local-functions/outcome.md`, for any later debate.
