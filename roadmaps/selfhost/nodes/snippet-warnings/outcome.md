# Outcome: snippet-warnings

## Built

- `selfhost/tests/src/guide.nra`: a `narra error` fence lists its diagnostics in leading `// error: LINE:COLUMN: MESSAGE` lines, and the guide suite fails on any printed diagnostic that no line names (`: note:` lines belong to the one before).
- `website/plugins/twoslash.ts`: `findUnexpected` applies the same rule to the query answer; packages of `narra file` fences are named after the fence and get the `main.run` binary, so their imports resolve.
- `website/plugins/markdown.ts`: `checkDiagnostics` fails the build when a sample (any page except `SPEC.md`) shows a diagnostic its fence does not announce.
- `docs/guide/**`: 15 unrelated diagnostics removed by fixing 14 samples; 12 related extras kept and marked. Chapter 2 has the three sentences (padding width in bytes with `00café`, ASCII-only `std/char`, `Float.parse` underflow reads as `0.0`).

## Proof

- base: the guide suite read only the first `// error:` line, so 27 extra diagnostics on error fences passed it and the site showed them; the site build checked nothing.
- branch: the same stray variable fails the guide suite (`FAIL guide/01-getting-started.md:23/debug`) and `pnpm --dir website build` (exit 1, `a sample shows a diagnostic its fence does not announce`). Chapter restored. Inventory in `log.md`.

Inventory (diagnostics shown on the site): before 172 (145 taught, 27 extra), after 157 (145 taught, 12 extra, all marked expected). Hidden on 4 package fences: 23 before, 0 after. Reference, install, README, home: 0.

## Deviations

- Narra has no warning level: unused names and the like are errors. So "an expected warning" is an expected diagnostic, marked by one more `// error:` line, not a new fence kind.
- The extra expected lines follow the same form as the first. This extends contract `guide-samples` (first line only). See `contracts/guide-samples.change-request.md`.
- `SPEC.md` is left out of the site check: its `narra` fences are fragments that the spec suite runs in its own way, and the site hides their diagnostics.
- The brief says Unicode classes come later in `std/unicode`; the chapter says `std/unicode` has no letter classes, which is true now, so that sentence is unchanged.

## Stubs left

- None.

## Gotchas

- Positions in `// error:` lines count the comment lines. Add a line and every later position moves by one.
- A sample that teaches an error must not use the value it broke (`print(book.title)` after a failed call) or the compiler adds a cascade error; use a statement, or an `export`ed constant for unused names.

## Proposed learnings

- A `narra error` sample lists every diagnostic it prints, one `// error:` line each | the guide suite and the site build both fail on others | docs/guide | yes

## Verification

- `NARRAC=<built> sh scripts/test.sh --suite guide` 622 cases pass; `--suite spec` 65 pass; `--suite sources` 1 pass.
- `pnpm --dir website build` and `pnpm --dir website typecheck` pass.
- Full `sh scripts/test.sh`: `PASS, 10918 cases in 291s`.

## Size

- Estimated M, actual M.
