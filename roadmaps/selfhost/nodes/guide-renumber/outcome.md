# Outcome: guide-renumber

## Built

The guide has 20 chapter files in `docs/guide/`, in the order approved on 2026-11-20:

1. `01-getting-started.md`
2. `02-values-and-types.md`
3. `03-functions-and-closures.md`
4. `04-structs-and-enums.md`
5. `05-pattern-matching.md`
6. `06-option-and-result.md`
7. `07-traits.md`
8. `08-modules-and-packages.md`
9. `09-memory-and-mutable.md`
10. `10-effects-and-testing.md`
11. `11-files-and-the-system.md` (new stub)
12. `12-generators.md` (new stub)
13. `13-standard-library.md` (renamed from 11)
14. `14-async.md` (new stub)
15. `15-threads-and-shared-state.md` (new stub)
16. `16-unsafe-code-and-c.md` (new stub)
17. `17-proofs-and-compile-time.md` (new stub)
18. `18-tools.md` (renamed from 12)
19. `19-building-for-other-systems.md` (new stub)
20. `20-narra-for-typescript-programmers.md` (renamed from 13)

Each stub has a numbered title, one introductory paragraph and a `Next:` link. Chapter headings, `Next:` links, the chapter list in `docs/guide/README.md` and the two in-text chapter numbers (in `01-getting-started.md` and `20-narra-for-typescript-programmers.md`) use the new numbers.

## Deviations

- Stubs end with a `Next:` line, so the chain from chapter 10 to chapter 20 has no break. The brief said "nothing else".
- No change was needed in `docs/install.md`, `website/`, `skills/narra/`, `README.md` or `selfhost/tests/src/guide.nra`. None links to a renamed chapter. The site sorts chapters by file name and the search index reads every file in `docs/guide`.

## Stubs left

- Chapters 11, 12, 14, 15, 16, 17 and 19 hold only an introduction. The writing nodes of batch-21 replace them. A writer keeps the `Next:` line.

## Gotchas

- Each `Next:` link joins two files. A writer who reorders chapters must fix both ends.

## Verification

- `sh selfhost/bootstrap.sh <scratch>`: stage 1 and stage 2 emit identical C.
- `sh scripts/test.sh --suite guide`: PASS, 334 cases.
- `sh scripts/test.sh --suite sources`: PASS, 1 case.
- `pnpm --dir website build`: passes.
- A search for the old file names finds no link outside `roadmaps/`.

## Size

- Estimated S, actual S.
