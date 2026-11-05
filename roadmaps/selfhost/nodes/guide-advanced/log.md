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

## 2026-11-05: review fixes

- Fresh review of 994f6b5: nothing blocking, 8 should-fix and 7 nits, mostly reteaching chapters 1 to 5 (guide-basics now b73a6e7). Read those chapters first.
- Chapter 9: copy section cut to a recap, heap defined once, `mutable` section rebuilt as game tasks (heal, starting kit, swap in an array, log hp, snapshot), closure section builds on chapter 3, `squares` sample dropped.
- Chapter 8: one sentence on the `mutable` parameter of `addWord`, package layout points to chapter 1, `uses alloc` written in the fragment with a pointer to chapter 4 (error block dropped). Package rebuilt from fragments.
- Chapter 6: opens from chapters 2 and 5, `findSetting` shown once and later samples use a `Map`, `orElse` needs `withDefault` after, panic example is an impossible state.
- Chapter 7: equality opens from chapter 4, adds `countOf<T: Equatable>`.
- Chapter 10: leak failure needs `NARRA_CHECK_LEAKS=1`, `void` explained, bare `return` removed, refers to chapter 4.
- TestGuide ok, prose check ok.
