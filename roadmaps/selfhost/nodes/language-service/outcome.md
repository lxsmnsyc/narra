# Outcome: language-service

## Built

- `narrac lsp`: an LSP server on stdin and stdout. Dispatch: one `if` in `run` (`selfhost/src/main.nra`).
  - `selfhost/src/lsp.nra`: Content-Length framing, lifecycle, request and notification dispatch, error codes.
  - `selfhost/src/lspserver.nra`: `Server` state (open documents, URIs with a non-empty list), `checkDocumentPackage`, `closeDocument`, `formatDocument`, URI and path conversion, UTF-16 and UTF-8 column conversion.
  - `selfhost/src/lspserverjson.nra`: `getField`, `getText`, `getList`, `getPath` over `Value`, and `makeRecord`, `makeResponse`, `makeError`, `makeNotification`. Messages are parsed and written with `std/json`.
- `selfhost/src/modules.nra`: `loadWith(path, overrides: Map<string, string>)`. A source path that is a key of `overrides` is read from the map. `load(path)` is `loadWith(path, Map {})`, so `narrac check` is unchanged.
- Runner suite `lsp` (`selfhost/tests/src/lsp.nra`, registered in `registry.nra`). It runs `testdata/lsp/<case>/session.txt` and compares the framed output with `expected.txt`. Cases: lifecycle, earlyexit, eof, malformed, change, twomodules, format, utf16, utf8.
- Docs: SPEC.md section 14 "Language server", compiler/README.md "Language server".

## Behavior to know (for lsp-navigation and lsp-incremental)

- Keys of `Server.documents` are URIs. `convertUriToPath` maps them to the paths that `loadWith` overrides (`DIR/src/x.nra`).
- Each open or change reloads and rechecks the whole package (`Body.checkPackage`, then `Effects.checkEffects` when the bodies have no errors). lsp-incremental can replace this inside `checkDocumentPackage`.
- Publish order: package modules in load order, then files that only problems mention, then empty lists for URIs in `server.reported` that this check did not mention.
- Columns: the compiler's character columns are converted with `countUnits` per line of the document text (the override, else the file on disk).
- `lsp.nra` `handleRequest` is where new requests (hover, definition) go. An unknown request gets -32601.
- Template strings in the compiler decode only `\n` and `\t` (see "Found" below). Use a plain-string constant (`CRLF` in `lsp.nra`) for `\r`.

## Acceptance

- [x] Framing: Content-Length read, other headers ignored (lifecycle case sends Content-Type), body read with `Io.readBytes`, replies framed and flushed. End of input without `exit` exits 1 (eof case).
- [x] Lifecycle: capabilities with full sync, `documentFormattingProvider` and `positionEncoding` (utf-8 case, utf-16 default). `exit` gives 0 after `shutdown` (lifecycle) and 1 before (earlyexit). -32002 before `initialize`, -32601 unknown request, unknown notification ignored (`$/setTrace`), -32700 for bad JSON and for a header block with no Content-Length, after which the server keeps answering (malformed).
- [x] Documents: open, change and close override disk. The package of the nearest `narra.toml` is checked and every module gets a list (twomodules: `text.nra` gets `[]`). Load errors (syntax error in the change case) go to their file and position.
- [x] Formatting: one whole-document edit, `[]` when formatted, `null` on a syntax error, with the `[fmt]` table of the case's narra.toml (format case).
- [x] Runner suite `lsp` (the Go `TestSelfhostLsp` is retired). `--update` rewrites `expected.txt`.
- [x] SPEC section 14 and compiler/README.md.
- [x] `bootstrap` suite and full `scripts/test.sh`: see the report.

## Gate decisions

- JSON comes from `std/json` and `std/shape`. Alternative: the in-tree JSON parser and writer the first worker wrote while the seed did not know `std/json`. Removed once the batch seed (f25cce2) knew it: one JSON implementation, and the server tests it.
- `didClose` rechecks the package from disk, sends the other modules' lists, and sends the closed document one empty list. Alternatives: publish the disk diagnostics of the closed file (most servers do not, and the editor no longer shows the file), or only forget the text with no recheck (would leave other modules' lists stale when the closed text differed from disk).
- Documents outside `src/` of a package (for example a file under `tests/`) are not modules of the package, so they get no diagnostics of their own. Opening one still rechecks the package. A load error with no file goes to the top of the changed document. Alternative: check such a file alone (wrong, since it is not a module and imports would not resolve).
- One `Problem` severity: every diagnostic is an error (1); the compiler has no warnings.

## Found (outside this node)

- `parseTemplate` (`selfhost/src/parser.nra`, about line 2741) decodes only `\n` and `\t` inside template strings. `\r`, `\0` and `\u{...}` become the bare letter, so `` `a\r\n` `` is `ar` plus a newline. Plain strings decode them in `lexEscape`. A fix should decode with the same table as `lexEscape`; it changes the compiler's own output only where templates use those escapes. Worked around here; not fixed, because `parser.nra` is not in this node's paths.

## New codes

- None.

## Stubs left

- None. Hover, definition, references and symbols belong to lsp-navigation; incremental recheck to lsp-incremental.

## Size

- Estimated M, actual M.
