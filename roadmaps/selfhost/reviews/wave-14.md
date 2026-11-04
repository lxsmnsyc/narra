# Wave 14 report

## Merged

- #60 box-assign, #61 block-arrow-args, #62 mutable-escape, #63 tail-loops, #64 for-destructure: the bug batch from the spec-gap audit.
- #65 labels, #67 for-of-sources, #68 let-narrowing, #69 ts-habit-fixes: spec-gap features.
- #70 nested-function-error: `function` inside a function is E0930 (D74).
- #71 block-expressions, #73 match-block-values, #74 ternary: blocks in value position produce values with `return` (D75), and `c ? a : b`.
- #72 mutable-read-conflict: a variable passed as `mutable` cannot be read elsewhere in the same expression (E0929).
- Closed: #66 local-functions (dropped by the goal debate, D74).
- Full check on main after #74: `go vet ./... && go test ./...` pass (252 s), including TestSelfhostBootstrap.

## Decisions (maintainer, 2026-11-04)

- D74: no local function declarations (debate `roadmaps/selfhost/debate/local-functions/debate.md`).
- D75: block statements and block expressions told apart by position; a block expression's value comes from `return`, and `return`/`break`/`continue` stop at it. `if`/`else` and `match` produce values in value position. Ternary added; a multi-line ternary needs parentheses (the line rule stays).
- Generators keep `function*` (D19 confirmed).
- `mutable` stays restricted (`x = f(mutable x)` is an error); node mutable-reference will rethink `mutable` as a mutable reference.
- Unused locals and parameters are checker errors (SPEC over lint).

## State

- Active: none. Session stopped here at the maintainer's request.
- Ready next: unused-names (runs alone: it cleans every sample and selfhost/src file).
- Gated design: mutable-reference (start with a goal debate).
- Outlines due for briefs: std-source (gated spike), fmt, generators, language-guide, std-json, query-cli, std-io-more, closure-borrow, and the remaining spec-gap outlines (std-spec-functions, sized-numbers, type-aliases, const-lazy, typed-calls-and-effects, spec-samples-test).

## Process lessons

- Fresh reviews found a real bug in most M nodes (let-narrowing blocking, for-of-sources, ts-habit-fixes, block-expressions, ternary blocking). Self-reviewed S nodes were mostly clean, but the orchestrator's own probes caught holes twice (let-narrowing for-of source, mutable-read-conflict match arms). Keep fresh review for M, and probe S nodes that touch soundness.
- Every new expression or statement kind must reach every walker (exclusivity, usesLocal, mutable roots, narrowing ends, effects). Two holes came from walkers that skipped a new kind. A brief that adds a kind should list the walkers.
- Force pushes are refused by auto mode. Workers merge origin/main instead of rebasing after the first push.
- Shared scratch directories caused one reviewer to report another agent's test log as its own. Give each agent its own scratch subdirectory.
- Workers that finish their report but keep a background job wake the orchestrator repeatedly; stop them once they have reported.
- Load from four parallel workers pushes the full suite past Go's 10-minute default; use `-timeout 45m`.
