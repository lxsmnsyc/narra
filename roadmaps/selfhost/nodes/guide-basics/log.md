# Log: guide-basics

- Branch docs/guide-basics reset to origin/feat/language-guide. Nothing existed yet (no docs/guide, no guide_test.go).
- Wrote `guide_test.go` (TestGuide) reusing `modes` and `runProgram` from run_test.go and `CompileToC` for error samples.
- Wrote README and chapters 1 to 5. Ran samples through TestGuide; fixed outputs and error texts from the compiler's real messages.
- Compound assignment and C-style `for` (D76) are not in the compiler yet; the guide does not use them and lists them under Planned.
- Hand check: changed `hello, narra` to `hello, narrb` in chapter 1; TestGuide failed at `01-getting-started.md:16/debug` with a want/got diff. Reverted.
- 2026-11-05 (second worker): maintainer review of #87 asked for a guided tutorial, not a rule list. Rebased on origin/feat/language-guide as branch guide-basics-rewrite.
- Rewrote README and chapters 1 to 5: one thread per chapter (receipt, functions over the receipt, shop data types, a till), each section task, program, output, explanation; rules as `narra error` samples with fixes.
- Removed "Recursion needs no keyword" and the "Tail calls are loops" section. Recursion is taught with counting the ways to make change; a short note and sample follow on a call in last position running like a loop.
- Found the old chapter 2 claimed sized integers (`i32`, `u8`) and unused-local errors; neither is in the compiler. Removed; sized numbers added to Planned.
- TestGuide passes (71 sample subtests). No Go code touched.
- Review of b5f6121: fixed the Span fragment (`uses alloc` and a chapter 10 pointer, plus a main.nra fragment), built the chapter 1 and 4 fragments as real packages. Added value-block versus statement-block `return` sample to chapter 4, the Option-holds-in-place note, the correct shadowing rule, the helper-inside-a-function section after arrow functions, the escaping section from the want, `else if`, `break`/`continue`, `//` comments; dropped `narra explain` from the table; fixed the nits. TestGuide passes (76 sample subtests).
- Chapter 1 package commands: table note now says to run each command from inside the package folder with `.`. Ran run, build, build -o, build -release -o, check and test on the shop package from inside it (all work, 220), the same on a single hello.nra, and confirmed `narra build shop` from the parent fails at the linker.
