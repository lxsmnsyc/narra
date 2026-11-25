# Outcome: std-url

## Built

`std/url`, a Narra source module in `std/src/url.nra`. It parses URI references (RFC 3986), writes them back, resolves a reference against a base, percent-encodes and decodes text, and reads and writes form-encoded query strings. The sample is `testdata/url_std.nra` with its `.out`. It covers the RFC 3986 section 5.4 normal and abnormal examples, round trips, bad escapes and repeated keys.

## Public API

- `struct Url { scheme: Option<string>, userInfo: Option<string>, host: Option<string>, port: Option<string>, path: string, query: Option<string>, fragment: Option<string> }`. Parts stay as written and percent-encoded. A missing part is `None`, so `a` and `a?` differ. The port is the text after `:` (validated as 0 to 65535, leading zeros kept, `Some("")` for a bare `:`), not an `int`, so `http://h:` and `http://h:0080/` round-trip; read it with `Int.parse(port, 10)`. An unbracketed host with a second colon is an Err.
- `struct QueryParam { name: string, value: string }`, decoded.
- `parse(text): Result<Url, string>`, `toString(url): string`, `resolve(base, reference): Result<Url, string>` (Err when the base has no scheme).
- `encode(text): string`, `decode(text): Result<string, string>` (`+` stays `+`; bad escapes and non-UTF-8 give an Err).
- `parseQuery(text): Result<Array<QueryParam>, string>`, `queryToString(params): string` (`+` is a space; repeated keys keep their order).

## Deviations

- The brief names `Query.parse` and `Query.toString`. One module has one namespace, so they are `parseQuery` and `queryToString`.
- Registry and generated files outside the owned paths: `selfhost/src/std/stdurl.nra` (new row file), one import and one `registerAll` line in `selfhost/src/std/stdlib.nra`, and an entry in `scripts/std/generate.nra`. These are the same steps std/set took.
- Seven `testdata/errors/*.err`, `testdata/declsdump/imports/err_unknown_module.dump` and `testdata/stage2/errors/std_internal_import.err` list the std modules; each now ends `std/regex, std/unicode and std/url`. Other std nodes in batch-24 change the same lines, so regenerate them on landing with `sh scripts/test.sh --update`.

## Gotchas

- `selfhost/src/std/stdtext.nra` embeds `url.nra`. After any edit to `std/src/url.nra`, run `sh scripts/test.sh --suite generated --update` and rebuild the compiler before running the sample.
- Host rules of the WHATWG standard (case folding, IDNA, default ports) are not applied.
