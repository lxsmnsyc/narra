# Wave 15 report

## Merged

- #86 chore/fmt-sources: `selfhost/src` and `compiler/testdata` formatted with `narrac fmt`; `TestSourcesFormatted` runs `narrac fmt --check` in `go test` (`ternary_parens.nra` kept by name in `layoutSamples`).
- #89 generators: `function*` compiles to C state machines in both compilers. Round 2 fixed raw C locals (block, match and map-index results) held across a `yield`; `liveValues` now panics on a non-frame name.
- #88 query-cli: `narrac query` (type-of, def, refs, callers, interface, recursion). Contract `query-output` frozen as v2 from its change request (taken as approved when the maintainer merged #88).
- Story language-guide, on `feat/language-guide` (not yet on main):
  - #87 guide-basics: `TestGuide`, the index and chapters 1 to 5.
  - d8eb59f guide-advanced: chapters 6 to 10, landed as a direct commit.

## Decisions (maintainer, 2026-11-05)

- The guide is standalone: no SPEC, section or D-number citations, no "the author decided" framing, no assumed TypeScript or Rust. Comparisons go to chapter 11 (new node guide-typescript).
- The guide teaches in a tutorial voice: each section starts from a task, builds a program, shows output, then explains. Rules come out of examples. No mechanism before its concept.
- Learner docs are written and reviewed with Opus (`model: opus` in the guide briefs). The Sonnet drafts read as rule lists.
- The maintainer said not to start new nodes until told.

## State

- Active: language-guide story (children guide-basics and guide-advanced done; guide-typescript todo, not started on purpose).
- Waiting for the maintainer's go-ahead:
  - guide-typescript (chapter 11, Opus; also fixes the README's dead link to chapter 11 and adds chapter 10's "Next:" line), then the story PR `feat/language-guide` into main with one fresh Opus review of the whole guide.
  - A fix node for the effect soundness bug in the inbox (calls through function-typed struct fields or call results drop their effects). Proposed as `must`, Opus, run before closure-borrow and coroutines.
  - Ready per `plan.py ready`: std-io-more, unused-names, closure-borrow, coroutines. std-io-more, closure-borrow and coroutines share stdlib.nra, stdcalls.nra, emitstd.nra and check.nra: merge in that order, rebasing each.
- Outlines due for briefs (`plan.py refine`): std-json, identity-cycles.
- Inbox gained: template literal token positions, `function*` to an annotated function type, effects lost through field and result calls, two misleading compiler hints, TestGuide skipping fragments.
- Housekeeping: the merged branches `docs/guide-basics` and `docs/guide-advanced` still exist on GitHub (deleting them was blocked by the permission check). Several agent worktrees under `.claude/worktrees/` are locked by the harness; remove them once no agent runs.

## Process lessons

- The fast check now takes 7 minutes alone and 13 to 24 minutes with four suites running. Two opus workers ran past the 20-minute timeout; use `-timeout 45m` (GOAL.md updated).
- Briefs for docs nodes must state audience and voice. The guide briefs said "for a TypeScript or Rust programmer, linking SPEC and D-numbers", and both workers followed them; the plan was at fault, not the workers.
- Fresh reviews found a blocking or should-fix bug in every M and L node (generators blocking, query-cli four should-fix, both guide halves). The query-cli review also found an effect soundness bug on main and the generators review a checker bug on main; both are in the inbox.
- Agents writing to shared `/tmp/x` overwrote each other's output; reviewer prompts now ask for uniquely named files in their own scratch directory.
