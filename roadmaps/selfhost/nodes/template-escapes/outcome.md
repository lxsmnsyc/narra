# Outcome: template-escapes

## Result

Already fulfilled. No code changed.

## Evidence

- Commit fb246d8 ("parser: decode every escape in template strings") is on feat/batch-4. `parseTemplate` calls `decodeEscape`, which wraps `lexEscape`, so templates share the plain-string decoding and its errors.
- `testdata/template_escapes.nra` covers n, t, r, 0, backslash, double quote, single quote, backtick, dollar, `\u{...}` (1, 2 and 4 byte), escapes mixed with interpolation, and checks `` `\r\0\u{41}` `` equals the plain string.
- Invalid escapes are tested in `testdata/errors/template_unknown_escape.nra`, `testdata/errors/template_escape_multiline.nra` and `testdata/parse/template_unknown_escape.ast`.
- `sh scripts/test.sh` on this branch (same tree as feat/batch-4): PASS, 8 shards, 720 passed each.

## Open

- The inbox entry (2026-11-09, from language-service) is still there. The orchestrator removes it.
- The `CRLF` constant in `selfhost/src/lsp.nra` and `selfhost/tests/src/lsp.nra` stays, as instructed. It can go after the next seed regeneration.

## Deviations

None.
