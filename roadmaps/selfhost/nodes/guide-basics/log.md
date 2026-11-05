# Log: guide-basics

- Branch docs/guide-basics reset to origin/feat/language-guide. Nothing existed yet (no docs/guide, no guide_test.go).
- Wrote `guide_test.go` (TestGuide) reusing `modes` and `runProgram` from run_test.go and `CompileToC` for error samples.
- Wrote README and chapters 1 to 5. Ran samples through TestGuide; fixed outputs and error texts from the compiler's real messages.
- Compound assignment and C-style `for` (D76) are not in the compiler yet; the guide does not use them and lists them under Planned.
- Hand check: changed `hello, narra` to `hello, narrb` in chapter 1; TestGuide failed at `01-getting-started.md:16/debug` with a want/got diff. Reverted.
