---
goal: Decide whether Narra keeps local function declarations, and if so whether they capture enclosing locals
date: 2026-11-04
rounds: 2 of 2
status: approved
roadmap: roadmaps/selfhost
supersedes: none
decision: DECISIONS.md D74
---

# Debate: local function declarations

## Settled structure

Narra has no local function declarations for now. In order:

1. Record PR #66 before closing it (P12, O11). Commit `roadmaps/selfhost/nodes/local-functions/outcome.md` to main as a roadmap commit. It holds the B design, the six defects (J9, O1, O4), and two preconditions that any later debate on local functions must meet: names resolve in the checker with scope, not by parse-time renaming, and the enclosing type parameters carry over to the lifted function. Then close PR #66 with a comment linking this debate.
2. New S node `nested-function-error`, after PR #69 (ts-habit-fixes) merges (O7 in its O10 form). `function` inside a function body is a dedicated compile error with no automatic fix. The message says to declare the function at module level as a private function and pass enclosing values as arguments. Both compilers, the TestHabitFixes harness, and the next free diagnostic code. SPEC.md line 62 says so, and "Differences from TypeScript" gets a row. On main a nested `function` already fails to parse, so no program changes meaning (P10's measurement).

Turning this error into local functions later only adds programs, which shipping B and then C would not (O1).

## Budget

- tokens: about 65k: 5k for the record, about 60k for the S node (O10, O11, CALIBRATION.md S rows)
- time: one S node after PR #69 merges (O9, O10)
- consistency: no fix that changes meaning. A move-to-module-level fix is out until scope analysis exists (O8)
- device: not touched

## Rejected

- Keep local functions as PR #66 builds them (B, P2, P3, P7): six defects from parse-time renaming, about 150k to fix (O2). B is also not an error-only subset of capture: a local function reading a name that a local const shadows prints 2 where TypeScript prints 11 (O1, measured), the same "same text, two meanings" defect the lab rejected in earlier debates (O3).
- Capturing local functions, option C (P4, P5): about 220k with no user in selfhost or the samples (O5, J4).
- Self-recursive const arrows (P6): no user, and more cost than the error (P9 concession).
- A fix that moves the function to module level (P10): it changes the meaning and breaks generic helpers (O8, measured).
- Folding the error into ts-habit-fixes (P11): PR #69 is in review and has used its reserved codes; about 85k against 60k (O9).

## Proofs

- None. Measurements settled every dispute (J3, O1, O4, O6, O8).

## Prior art

- Go and Zig have no named nested functions; Rust's nested `fn` cannot capture; TypeScript, Swift and Kotlin local functions capture and may recurse (J7).
- This repo's qualified-enum-constructors and sum-type-syntax debates rejected designs where the same text gets a second meaning (O3).

## Open risks

- None open. If a recursive helper inside a function shows up in real Narra code (selfhost, samples), rerun this debate from the PR #66 record and P6.

## Inferred choices

- Narra keeps TypeScript text's meaning or rejects it (J8, from SPEC.md section 2).
- Capture has no consumer while the goal is self-hosting (O5, from GOAL.md and J4).

## Answers from the user

- The maintainer asked for this debate (2026-11-04) instead of choosing between "no capture" and "capture later".

## Debate cost

- 2 of 2 rounds, about 149k tokens
