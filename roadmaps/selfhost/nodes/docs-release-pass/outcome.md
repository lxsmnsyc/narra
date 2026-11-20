# Outcome: docs-release-pass

## Built

- Chapter 1 installs with the one-line command, links to `docs/install.md` and checks with `narrac --version`; the guide index says "install the compiler".
- "Workspaces of several packages" removed from the guide's Planned list.
- Three TypeScript and JavaScript comparisons removed from chapters 2 and 11.
- Bootstrap stages, exit codes and `linux-check.sh` moved from `docs/install.md` to `seed/README.md`.
- `SPEC.md`: command is `narrac` (30 lines); section 15 titled "Packages".
- `README.md`: install first, release status, "Working on the compiler", License.
- `skills/narra/evals/results.md` (totals of the 40-run eval); the home page links the 19 of 20 claim to it.
- `LICENSE`: MIT, Copyright (c) 2026 Alexis H. Munsayac.

## Proof

Not a bug fix.

## Deviations

- The eval results file has only the totals from the narra-skills outcome; the per-run files were gone.
- The links to `seed/README.md` (install page) and `results.md` (home page) are GitHub URLs, because the site does not publish those files.
- README states the 19 of 20 result too, with a link.

## Stubs left

- Change 8 of the proposal (prerelease paragraph of the install page) waits for the first tag.

## Gotchas

- The README and install page assume the GitHub repository path `lxsmnsyc/narra`.

## Proposed learnings

None.

## Verification

- `pnpm --dir website build`: passes (18 pages).
- `NARRAC=<seed build> sh scripts/test.sh --suite guide`: PASS, 334 cases. `--suite sources`: PASS, 1 case.
- `check_prose.py` on the changed Markdown: only pre-existing warnings remain.

## Size

- Estimated M, actual M.
