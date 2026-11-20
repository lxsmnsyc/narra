# Outcome: error-index

## Built

- `narrac explain --list --json` (`writeCodeListJson` in `selfhost/src/driver/diagnostics.nra`, wired in `runExplain` of `selfhost/src/driver/driver.nra`): a JSON array, one object per line, one per code in table order: `{"code":"E0001","messages":["..."],"explanation":null}`. `messages` holds every format of the code (85 codes have more than one); `explanation` is the long text or `null`. 479 codes, 96 with an explanation.
- `scripts/error-index.sh`: prints `docs/reference/errors.md` from that JSON with `jq`; groups by hundred (E0000 to E0099 and so on), one `### E0101` heading per code (anchor `#e0101`), the messages in a `text` fence, the explanation one line per paragraph.
- `docs/reference/errors.md` committed; linked from `docs/reference/README.md` in place of the "Coming" line.
- Case `generated/errorindex` in `selfhost/tests/src/generated.nra`: runs the script with the compiler under test and fails when the committed page differs. `sh scripts/test.sh --suite generated --update` rewrites the page.

## Deviations

- The brief says "checked like the std reference"; no such check exists yet, so the case joins the generated suite, which already checks other committed generated files.
- `jq` is required to run the script (and so the case); the script exits 2 with a message when it is missing.

## Stubs left

- None.

## Gotchas

- Adding or editing a code or explanation in `diagnostics.nra` makes the case fail until `--update` is run. The compiler under test must contain the change, so rebuild first.
- The page has 479 `###` headings, so the site search outline for `/reference/errors` is long.

## Verification

- `sh scripts/test.sh --suite generated`: pass; with a line appended to the page, `generated/errorindex` fails as stale.
- `sh scripts/test.sh`: see the parent report.
- `pnpm --dir website build`: passes; `/reference/errors` is prerendered.

## Size

- Estimated M, actual S.
