# Outcome: error-index

## Built

- `narrac explain --list --json` (`writeCodeListJson` in `selfhost/src/driver/diagnostics.nra`, wired in `runExplain` of `selfhost/src/driver/driver.nra`): a JSON array, one object per line, one per code in table order: `{"code":"E0001","messages":["..."],"explanation":null}`. `messages` holds every format of the code (85 codes have more than one); `explanation` is the long text or `null`. 479 codes, 96 with an explanation.
- `scripts/errors/generate.nra`: a Narra program that runs `<narrac> explain --list --json` with `Process.run`, decodes it with `std/json` and prints `docs/reference/errors.md`. It needs only the compiler under test. Groups by hundred (`##`); each code is a `####` heading (id `#e0101`, so the page ToC and search list the 17 groups, not 479 codes). The messages stay in a `text` fence. Explanation text is escaped (`\<`, `\|`, `\*`, `\_`, a line-leading `#`, `-`, `+` or `1.`) outside balanced code spans, so generics never become HTML tags. A code in a hundred with no name in `AREA_NAMES` stops the generator with an error. `scripts/error-index.sh` is a wrapper that needs `sh` and `narrac`.
- `docs/reference/errors.md` committed; linked from `docs/reference/README.md` in place of the "Coming" line.
- Case `generated/errorindex` in `selfhost/tests/src/generated.nra`: runs the generator with the compiler under test and fails when the committed page differs. `sh scripts/test.sh --suite generated --update` rewrites the page.

## Deviations

- The brief says "checked like the std reference"; no such check exists yet, so the case joins the generated suite, which already checks other committed generated files.

## Stubs left

- None.

## Gotchas

- Adding or editing a code or explanation in `diagnostics.nra` makes the case fail until `--update` is run. The compiler under test must contain the change, so rebuild first.
- The 479 codes are `####` headings, so the page outline and the search index have 17 entries. The search text of a group is cut at 3000 characters, so the later codes of a group are found by their group, not by their own code.

## Verification

- `sh scripts/test.sh --suite generated`: pass; with a line appended to the page, `generated/errorindex` fails as stale.
- `sh scripts/test.sh`: `PASS, 10497 cases`. The generated suite also passes with `jq` off the PATH (nothing calls it).
- `pnpm --dir website build`: passes; `/reference/errors` is prerendered with 17 ToC links and 17 search entries.

## Size

- Estimated M, actual S.
