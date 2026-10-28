# Wave 1

Nodes: split-compiler (PR #2, merged by the user), bench-harness (PR #1, auto-merged).

## Changes to the plan

- Contracts `diagnostic` v1 and `builtins` v1 are frozen. No change requests.
- json-diagnostics: added `driver.go` to its shared files, because `Diagnostic.File` must be filled there.
- strings-std, io-std, maps: their briefs now say to put runtime C in `Module.Runtime`.

## Learnings promoted

- `compiler/internal/narra/CLAUDE.md`: how to add std functions and report errors, and the rule that refactors keep output byte-identical.
- `bench/CLAUDE.md`: how to report benchmark numbers.

## Follow-ups

- `bench/BASELINE.md` cites commit 587b654 and was recorded on a busy machine. Re-record it on an idle machine.
- The new error tests match only part of the message. This is accepted.
- In agent worktrees, the rtk wrapper refused `git`. Workers used `/usr/bin/git`.
