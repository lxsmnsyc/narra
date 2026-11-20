
## Review fixes

- Dropped `jq`: the page is made by `scripts/errors/generate.nra` (Process.run on `explain --list --json`, std/json decode), run by the generated suite with the compiler under test. `scripts/error-index.sh` stays as a wrapper.
- Explanation text is escaped for markdown (the site renders with `html: true`), code spans are kept.
- Code headings are `####`: the markdown plugin gives every heading a GitHub-style id, so `#e0101` still works; the ToC (levels 2 and 3) and search (levels 1 to 3) now have 17 entries. No other file links to an `errors#` anchor.
- An unknown hundred fails the generator.
- `writeCodeListJson` uses for-of, with one helper for the entry.
