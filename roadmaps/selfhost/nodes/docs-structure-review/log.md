# Log: docs-structure-review

- 2026-11-19: branch `spike/docs-structure-review`, reset to `feat/batch-20` (`94107ca4`).
- Read the site sources: `website/src/pages.ts`, `routes/_layout.brm`, `routes/index.brm`, `components/DocPage.brm`, `plugins/markdown.ts`, `plugins/search.ts`, `scripts/vercel-build.sh`, `vercel.json`.
- Built `narrac` from `seed/narrac.c` in the session scratchpad (`cc -O1 -DNARRA_RELEASE -std=c11 -w ... -lm`) to check the leads:
  - `narrac query --package p interface p/main` works. `std/iterator`, `std/string` and `std/toml` print `error unknown module`.
  - `narrac builtins` prints the builtin registry (950 lines). 18 public std modules are builtin and have no source.
  - `narrac explain` takes one code. `narrac --version`, `narrac help` fail with "No such file or directory". Bare `narrac` prints the internal dump commands.
- Counted in the files: 25 `narra <command>` lines and 14 `narrac` lines in `SPEC.md`; 99 `D` numbers in `SPEC.md`; 479 distinct codes and 96 explanations in `selfhost/src/driver/diagnostics.nra`.
- Grepped the guide for comparisons with TypeScript and its runtime language outside chapter 13: three hits (02:66, 11:360, 11:432). No SPEC or `D` citations in the guide or `docs/install.md`.
- The guide index still lists workspaces as planned. Chapter 8 teaches them.
- No `LICENSE` file in the repository. Noted in the proposal as a maintainer decision.
- `pnpm --dir website build` not run: this node changes no site or docs file.
