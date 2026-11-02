# Wave 11: emitter core and wave 2 of the emitter

## Merged

- #46 selfhost-emit-core (262k tokens). Gate: froze `contracts/emitter.md` v2 from the node's change request (cnames takes Checker, keyword parameter renames, stub probes, section stub names).
- #49 selfhost-emit-program (268k). C names, xxHash32, checker facts. Composites with type parameters are dropped at record time. `readEmitSkips` allows a missing skip file.
- #47 selfhost-emit-decls (196k). Types and reference counting parts match Go on all 63 inputs up to the first marker of another stub. Skip file deleted.
- #48 selfhost-emit-bodies (441k). Filled the missing `BodyInfo.instanceRefs` in the checker (`calls.nra`, `funcvalues.nra`). 4 of 59 inputs now byte-identical to Go. Skip file deleted.

## State

- Active: none.
- Ready next (wave 3, in parallel): selfhost-emit-closures, selfhost-emit-match, selfhost-emit-std, selfhost-emit-traits.
- Inputs still blocked per stub (one input counts under several): std 52, match 37, closures 14, traits 14, options 7, maps 2.
- The match, std and options probe calls stay in `emit.nra`; each node removes its own.

## Plan changes

- Added outline node compound-assign (should, m9-adoption, after selfhost-bootstrap): `+=` and the other compound forms, `++` and `--`. Reverses SPEC section 2, needs a decision record.

## Process lessons

- Parallel nodes that depend on each other's facts (program, decls, bodies) could not prove output until program merged. Merge order program, decls, bodies worked; each later node rebased and ran the real comparison before going ready.
- A worker committed a built `narrac` binary at the repo root; `.gitignore` now covers `/narrac` and `/narrac.dSYM/`. Reviewers should check `git diff --stat` for binaries.
- The plain-writing hook blocks commit subjects over 72 characters, so the `<id>: <title>` subject must be shortened; tell workers up front.
- The auto-mode classifier blocks `gh pr merge` and bulk `git worktree remove --force`; the maintainer merges and cleans worktrees.
