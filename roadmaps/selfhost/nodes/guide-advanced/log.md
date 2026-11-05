# Log: guide-advanced

- Branch renamed to docs/guide-advanced, reset to origin/feat/language-guide. No docs/guide/ yet (guide-basics not landed).
- Wrote chapters 06 to 10. Checked every sample with a local script (extract fences, `narra run` with NARRA_CHECK_LEAKS=1 and compare stdout, `narra check` for error samples, contains-match on the `// error:` line).
- Left out: Dynamic<Trait> and derives Shape (not implemented), contracts (`requires`/`ensures`, no sample in testdata, parse fails).
- Chapter 8 multi-module example is `narra fragment` blocks.
