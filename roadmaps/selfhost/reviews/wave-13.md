# Wave 13 report

## Merged

- #56 selfhost-emit (67k tokens). No skip files left; the harness fails on any `unsupported:`. The selfhost package (95,033 lines of C) is an emit input and matched Go on the first run. narrac emits it in 0.19 s against Go's 0.09 s (release).
- #57 selfhost-bootstrap (124k). Stage 1 (Go-built) builds stage 2; stage 2 reaches the fixed point and matches Go on every emit input. `TestSelfhostBootstrap` takes 21 s; `selfhost/bootstrap.sh` needs only `go` and `cc`. Every must node of the roadmap is done.
- #58 spec-gap-audit (164k). 21 confirmed gaps between SPEC.md and the compiler, two of them bugs in supported features (F4 block arrows in calls, F8 Box assignment).
- #59 spec-fixes (60k). Four SPEC.md text fixes.

## State

- Active: block-arrow-args, box-assign, mutable-escape, tail-loops (the bug nodes, started first by maintainer decision).
- Ready next: for-destructure (also adds `Arrays.entries`).
- Refine next: std-source, fmt, generators, language-guide, std-json, query-cli, std-io-more, closure-borrow (all within two waves).

## Plan changes

- New milestone m12-spec-gaps with 19 nodes from the audit (5 full briefs, 14 outlines), plus option-inline (later, m6-performance).
- Maintainer decisions 2026-11-04: unused locals and parameters are checker errors (SPEC wins over lint; lint brief updated); `Arrays.entries` goes into for-destructure; spec text fixes in their own docs node; F15 (`@unique`, copy tracing) folded into perceus.
- spec-samples-test added from the audit's learning: every SPEC.md sample checked by a test with a skip ratchet.
- Should work now ends by wave 9 (was 6).

## Process lessons

- Bugs hid in features marked done because no test ran the spec's own samples. spec-samples-test closes that.
- An S docs worker filled a gap in the spec with an invented API name. Briefs that edit SPEC.md should say: never introduce names; mark planned parts as planned.
- zsh does not split `set -- $p`; write shell helpers as functions with explicit arguments.
