# Outcome: pkg-http-fetch-parity

## Built

- `http/body`: `readAll(body)`, `readText(body)`, `readJson(body)` lose `maxBytes`. New: `limit(body, maxBytes): Body`, `fromQuery(params: Array<QueryParam>): Body uses alloc`, `fromForm(form: FormData): Body uses io, alloc`. `chunks` gives `Err(TooLarge("body"))` over the limit.
- `http/form`: `readFields(request)` loses `maxBytes`. New: opaque `FormData`, `create(): FormData uses alloc`, `append(mutable form, name, value) uses alloc`, `appendFile(mutable form, name, filename, contentType, body: Body) uses alloc`. `parts(request, options)` and `FormOptions` are unchanged.
- `http/client`: `ClientOptions.maxBodyBytes` (default 64 MiB), a response body carries it. `http/server`: a request body carries `ServerOptions.maxBodyBytes`.
- Content type on a body: `fromText` gives `text/plain; charset=utf-8`, `fromQuery` gives `application/x-www-form-urlencoded`, `fromForm` gives `multipart/form-data; boundary=...`; `empty`, `fromBytes`, `fromStream` give none. Server and client write it when the message has no `Content-Type`; a caller's type wins.
- Private: `packages/http/src/bodystate.nra` (limit and content type in `Body`, `limit`, `withContentType`, `findContentType`, `fromIncoming(pieces, length, maxBytes)`), `packages/http/src/formdata.nra` (`FormData`, multipart writer). `Wire.NO_LIMIT` is removed.
- Multipart writer: random boundary (`narra-form-` plus 30 hex digits), re-picked when it occurs in a name, filename, content type or held text. A form of held parts is held bytes with a length and replayable; a form with a stream part is a stream, chunked, sent once.
- Checks: `packages/http/src/checksend.nra` (limits, content types, exact wire bytes, round trips through the server, hostile names, client and server limits). Callers in `checkcookie`, `checknet`, `checkform` updated. `packages/websocket` and `examples/` used no reader and needed no change.
- README: limits, content types, sending forms.

## Deviations

- The brief says "a body the program built has no limit"; `Bodies.limit` on a held body does limit it (the body then fails over the limit). This is what `Bodies.limit` is for.
- The multipart reader now decodes `%22`, `%0D`, `%0A` in names and filenames (coordinator decision, after the first commit).
- The old multipart `readFields` counted field names and values against `maxBytes`. The body limit now counts every byte of the request body (boundaries and heads too), so a limit is a little larger than the sum of the fields.
- `Forms.append` and `Forms.appendFile` are wrapper functions, not aliases of the private ones, because a `mutable` parameter function cannot be a value.

## Stubs left

- None.

## Gotchas

- The multipart reader decodes only `%22`, `%0D` and `%0A` in names and filenames (as the Fetch standard does), so a literal `%22` typed in a name also reads back as a quote.
- A `Content-Type` that the caller sets on a request with `fromForm` must carry the same boundary; the body does not tell the caller its boundary.
- A form with a stream part is sent chunked and cannot be replayed on a redirect; the client returns the redirect as it is.
- `Bodies.limit` shares the stream with the original body, so reading one uses up the other.

## Proposed learnings

- A function with a `mutable` parameter cannot be re-exported with `export const`; write a wrapper | the checker refuses it as a value | packages/ | no

## Verification

- `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1 ASAN_OPTIONS=detect_leaks=0 out/narrac run -sanitize -bin http-tests packages/http`: 411 checks, 0 failed.
- `out/narrac run -sanitize -bin websocket-tests packages/websocket`: 210 checks, 0 failed.
- `out/narrac check .` and `out/narrac fmt --check .`: clean.
- `sh scripts/test.sh`: `PASS, 11201 cases in 240s`.

## Size

- Estimated M, actual M. A follow-up commit made the reader decode the three escapes.
- Risk estimated medium, measured medium. The body struct change touches every sender and reader; the checks cover both.
