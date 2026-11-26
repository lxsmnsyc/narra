# Outcome: pkg-http-forms

## Built

Body readers, response helpers and forms in `packages/http`. Nothing of this existed before.

- `http/body` (edited): async `readText(body, maxBytes): Result<string, HttpError>` (`Malformed` for bytes that are not UTF-8, `TooLarge` over the limit) and async `readJson(body, maxBytes): Result<Value, HttpError>` (through `std/json`; `Malformed` for bad JSON), `Value` from `std/shape`.
- `http/response` (edited): `json(status, value: Value): Response`, `redirect(url, status): Result<Response, HttpError>` (301, 302, 303, 307, 308 with `Location`; any other status gives `Malformed`), `isOk(response): bool` (200 to 299).
- `http/form` (new, in `[exports]`, `packages/http/src/form.nra`): `FormField { name, value }`, `FormPart { name, filename: Option<string>, contentType: Option<string>, body: Body }`, `FormOptions { maxParts, maxPartHeaderBytes, maxPartBytes }`, `defaultOptions()` (1000, 16 KiB, 16 MiB), async `readFields(request, maxBytes): Result<Array<FormField>, HttpError>` and async generator `parts(request, options): AsyncGenerator<Result<FormPart, HttpError>> uses (io, alloc)`.
- `packages/http/src/multipart.nra` (private): the boundary, the part head and the part body stream.
- `packages/http/src/checkform.nra`: 113 new checks, registered in `checks.nra` (`runFormChecks`). The README has the new sections "Reading bodies and building responses" and "Forms".

`readFields` picks the format from `Content-Type`: urlencoded through `Urls.parseQuery` (repeated names kept), multipart through `parts`. In multipart every part must be text without a filename, and the fields together stay under `maxBytes`.

## Deviations

- `readFields` gives `Malformed` for a multipart file part (a part with a filename) and for any `Content-Type` that is not a form, instead of skipping them. The brief said "text parts only" and did not name the error.
- `parts` yields `Result<FormPart, HttpError>` so an error ends the stream with a value. A part not read before the stream goes on is drained and skipped.
- Shared files edited: `body.nra`, `response.nra`, `narra.toml` (one export), `checks.nra` (one call), `README.md`. No file outside the brief's paths.

## Stubs left

- None.

## Gotchas

- The parts share one request body, so a part must be read before the next one is asked for; a skipped part cannot be read later.
- `parts` has no limit on the whole body; a server's `maxBodyBytes` bounds it, a client caller must bound it.
- `redirect` refuses an empty url and a url with CR, LF or NUL. The client's request writer already refuses a header value with a line break (`Malformed`, through `findBadHeader`), as the server answers 500; a check covers it.
- `Content-Disposition` values: a quoted value ends at the next quote and keeps its backslashes (browsers send paths raw); `filename*` and `Content-Transfer-Encoding` are ignored. Browsers percent-encode a quote as `%22`, which stays as written.
- A delimiter in a part is the bytes CR LF `--boundary`. Text after it on the same line that is not spaces, tabs or the line break gives `Malformed`.
- The boundary search reads bytes one at a time with `Bytes.at`; the buffer never holds more than one piece plus the delimiter length, so it is linear in the body.
- `let request = copy(created.value)` is needed before `Fields.add(mutable request.headers, ...)` on a request from `Requests.create` (E0978).

## Proposed learnings

- None.

## Verification

- `out/narrac check packages/http`, `out/narrac fmt --check .`: clean.
- `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1 ASAN_OPTIONS=detect_leaks=0 out/narrac run -sanitize packages/http`: `241 checks, 0 failed`, curl included (`curl -F` with a binary file, `curl -d` with repeated names).
- Checks: a body split into two pieces at every byte position, one byte per chunk, chunks of 2 to 39 bytes, every prefix of a body, 150 bodies with changed bytes, 150 random bodies and 150 random bodies behind a real opening boundary give a result and no panic; the limits (parts, head, part body) and each malformed case give the stated error; `readText`, `readJson`, `json`, `redirect` and `isOk` run over loopback.
- `sh scripts/test.sh`: `test.sh: PASS, 11018 cases in 401s` (the package step ran its 241 checks inside it).

## Size

- Estimated M, actual M.
- Risk estimated low, measured low: new modules only, plus small additions to two existing ones.
