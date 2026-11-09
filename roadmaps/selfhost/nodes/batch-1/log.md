# batch-1 log

## Review fixes for PR #102 (branch fix/batch-1-review)

- Named arguments: `holdReorderedArguments` (emitexprs.nra) now holds every non-mutable argument up to the last one that a later-written argument is passed before. Before, `three(note("1"), c: note("3"), b: note("2"))` ran 3, 1, 2. The bootstrap fixed point still holds, so calls without reordering emit the same C.
- `Json.parse`: `readWord` compares bytes instead of slicing, so `"tru€"` and `"\ud83d€"` return `Err` with a position. The other `Str.slice` calls in json.nra end on ASCII bytes.
- `narrac lsp`: `Server.reported` is a map from package (directory, or the file checked alone) to URIs. The check cache holds one package and already does a whole-package check when the package changes, so it needed no change. New case `testdata/lsp/twopackages`.
- Derived `Shape`: an `Option` field whose default is not written `None` encodes `None` as `Null`. Decoding already turned `Null` into `None` through `Option`'s Shape. SPEC.md updated. A default that is a constant holding `None` counts as "not None" and writes `null`; that is still a correct round trip.
- LSP header names match in any case. No session test: the runner always appends its own `Content-Length`, and a bad header leaves the body unread, which breaks the rest of the session. Checked by hand: `CONTENT-LENGTH: 41` is read as the length.
- The seed was not regenerated; bootstrap prints the "seed is older" note and runs stage 3.
