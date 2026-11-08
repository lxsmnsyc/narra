# Log: goldens-tools

- Reset to origin/feat/c-seed, branch feat/goldens-tools.
- Built narrac with `sh selfhost/bootstrap.sh <dir>` to develop against.
- Wrote suites dumps, fmt, fmt-stable, fmt-doc, trivia, query, sources, node-ids, plus helpers scratch, fmtrun, and findSources/listCorpus in files.nra.
- Golden rule for dumps: exit 0 means stdout is the golden; exit 1 means stderr is the golden and stdout is empty.
- Bug met: byte slicing of dump lines panics inside a multibyte character; compare bytes and step with Str.nextChar.
- `find` also lists directories named `dir.nra` (pkgdump/nested_order); use `-type f`.
- test.sh: NARRAC unset runs selfhost/bootstrap.sh into the work dir.
- Break tests (a line appended to a golden): dumps (ast, decls, types, pkg), fmt, fmt-doc, trivia, query each fail with a diff and pass after restoring.
