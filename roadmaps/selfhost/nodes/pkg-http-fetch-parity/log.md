# Log: pkg-http-fetch-parity

- Checked first: no limit on the body, no content type on the body, no form builder existed on `feat/batch-27`. `websocket` and `examples/` use no body reader, so nothing there changed.
- The limit lives in `Body` (`limit: Option<int>`), set by `fromIncoming` from `ServerOptions.maxBodyBytes` and the new `ClientOptions.maxBodyBytes`. `chunks` enforces it, and `readAll` is built on `chunks`, so every reader gives `TooLarge("body")`.
- `Wire.NO_LIMIT` was only used by the client and is removed; the client passes `maxBodyBytes` to `streamChunked`.
- The content type is a second field of `Body`. `fromText` sets `text/plain; charset=utf-8`; `fromQuery` and `fromForm` set theirs through `State.withContentType`. Server and client add `Content-Type` only when the headers have none.
- Multipart writer: new private module `formdata.nra` holds the opaque `FormData` and `toBody`. A form whose parts are all held bytes is joined into one held body (length known, replayable). Otherwise the body is a stream that yields each part head and then the chunks of the part body.
- A function with a `mutable` parameter cannot be used as a value, so `Forms.append` and `Forms.appendFile` are wrapper functions in `form.nra`, not `const` aliases.
- Old multipart `readFields` counted names and values against `maxBytes`. Now the body limit covers the whole request body, so the three tests that counted names and values were rewritten to limit the body.
- Escaping follows the WHATWG form encoding (`%22`, `%0D`, `%0A`). The reader first did not decode them; a follow-up makes `multipart.nra` decode exactly those three in names and filenames, so names round-trip.
