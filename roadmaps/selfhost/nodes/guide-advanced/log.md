# Log: guide-advanced

- Branch renamed to docs/guide-advanced, reset to origin/feat/language-guide. No docs/guide/ yet (guide-basics not landed).
- Wrote chapters 06 to 10. Checked every sample with a local script (extract fences, `narra run` with NARRA_CHECK_LEAKS=1 and compare stdout, `narra check` for error samples, contains-match on the `// error:` line).
- Left out: Dynamic<Trait> and derives Shape (not implemented), contracts (`requires`/`ensures`, no sample in testdata, parse fails).
- Chapter 8 multi-module example is `narra fragment` blocks.

## 2026-11-05: rewrite after maintainer review

- New branch guide-advanced-rewrite from origin/docs/guide-advanced, rebased on origin/feat/language-guide.
- Feedback: chapters read as rule lists, leaned on spec and decision references and on TypeScript knowledge. Rewrote all five chapters from scratch as one thread each (settings reader, shop catalog, word counter package, game player, temperature module). Kept only sample code that still fit the thread.
- Every restriction is now an error sample with the real message and position, followed by the fix. Bullet lists of rules are gone. Lists left are real lists (helper tables, effects table, numbered steps).
- Built the chapter 8 package and the chapter 10 test file in a scratch folder. Fragments and quoted output come from those runs.
- Found: `narra build app` from the parent folder fails when the binary name equals the folder name (linker writes over the directory). Chapter 8 uses `narra build .` inside the folder. Chapter 1 shows the failing form.
- Found: the constant-parameter message suggests `let`, and the missing-implement message suggests `derives` for user traits. Both noted in outcome.md.
- Dropped from the old draft: `export type` re-exports (not implemented), "planned" and "not yet" notes, the tools for showing copies, arenas.
- TestGuide (harness copied from origin/docs/guide-basics, removed before commit): ok. check_prose.py: ok.
- Fast check (vet + go test -timeout 20m, harness copied in): ok, 801 s.
