---
id: mutable-read-conflict
title: "A variable passed as mutable cannot be read elsewhere in the same expression"
status: active
priority: should
type: fix
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
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestHabitFixes|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - compiler/testdata/errors/mutable_read_conflict_*
  - roadmaps/selfhost/nodes/mutable-read-conflict/**
shared:
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/internal/narra/diagnostic_fixes.go
  - selfhost/src/calls.nra
  - selfhost/src/exprs.nra
  - selfhost/src/stmts.nra
  - SPEC.md
  - compiler/testdata/**
  - selfhost/src/**
reserves: [diag E0929]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
  - on: let-narrowing
    kind: hard
    why: let-narrowing (PR #68) ends narrowing per statement for mutable arguments; this node widens the same rule and must build on it
---

# A variable passed as mutable cannot be read elsewhere in the same expression

## Goal

On main, `let x = 1 ⏎ x + bump(mutable x)` (where `bump` sets its parameter to 10 and returns 0) prints 10, where TypeScript prints 1: the emitter runs `mutable` calls into temporaries before the rest of the expression, so reads of the same variable see the changed value. `pair(y, bump(mutable y))` reads `y` after `bump` too. SPEC.md (section 6, "Access is exclusive during the call") already rejects `move(mutable p, p.x)`, a variable passed as `mutable` and read in another argument of the same call, but not a read in an enclosing call or an operand. The maintainer chose a compile error over left-to-right evaluation (2026-11-04).

## Approach

Widen the exclusive-access rule: within one expression (a statement's expression, a condition, a return value, an initializer), a variable passed as `mutable` anywhere cannot also be read, assigned or passed again anywhere else in that expression, at any depth. This includes fields and elements rooted at the variable (`p.x`, `xs[0]`). The error names the variable and has a fix that binds the `mutable` call to a fresh `const` on the line before. Reuse the existing E-code of the `move(mutable p, p.x)` rule if its message fits, otherwise use E0929. Update SPEC.md section 6 to state the widened rule with one example.

## Acceptance

- [ ] Error tests cover `x + bump(mutable x)`, `pair(y, bump(mutable y))`, `add(x.value, clear(mutable x))` on an Option, a field read `p.x + reset(mutable p)`, and a read inside a template string, each with the position in the `// error:` header and the fix in `--json` output.
- [ ] Two `mutable` calls on different variables in one expression stay legal; the same variable passed as `mutable` to two nested calls is an error.
- [ ] Every existing sample, `selfhost/src` and `bench/` still compile; fix any that hit the new rule in a separate first commit.
- [ ] The Go compiler and `selfhost/` change together; the self-hosted compiler matches Go on the new error tests through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap`.

## Notes

- Found by the fresh review of PR #68 (let-narrowing), where the same ordering let `x.value` read a `None` at run time.
- Node mutable-reference may later redesign `mutable` parameters as mutable references; this fix stands until then.
