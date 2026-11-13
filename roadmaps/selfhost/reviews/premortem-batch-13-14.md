# Premortem: batches 13 and 14

It is three weeks from now and the batches are late or reverted. Why?

1. **Gates piled up.** Nine design spikes (D110 to D118) reach the maintainer within hours, and every build waits on its gate. Change: spikes write one-page decisions with a recommended answer per question, so the maintainer can answer them in one sitting; a gate not answered moves its build out of the batch, and the batch ships the rest.
2. **threads and regions grew past L.** Atomic counts, the allocator and copy-out touch every runtime path. Change: their spikes split them into children and measure the cost first; if a child would raise `emit selfhost` instructions above the noise in code without threads or regions, it waits behind a flag or leaves the batch.
3. **Parser and diagnostics conflicts.** Five syntax builds edit `parser.nra`, `ast.nra`, `astdump.nra`, `fmt.nra` and `diagnostics.nra`. Change: logic in each feature's module, dispatch rows only in shared files, codes reserved per node, and builds land as they finish with a rebase before review.
4. **The two batches broke each other.** batch-13 changes syntax that std sources use (uses lists), and both regenerate `stdtext.nra` and the seed. Change: whichever merges second rebases, regenerates, and runs the full and Linux checks again before its pull request.
5. **The feature review came too late to matter.** Its proposals land after the builds. Change: it runs after the design gates (not after every build), and its accepted proposals become the next batch; removing a feature just built is still cheap, since the language is unreleased.

Accepted: DECISIONS.md and SPEC.md are edited by nine spikes; each appends its own record and section text, so rebases stay mechanical. uses-list-design and uses-list-syntax stay two S nodes, because the syntax choice needs the maintainer's gate before the parser change.
