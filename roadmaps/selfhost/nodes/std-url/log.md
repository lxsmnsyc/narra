# Log: std-url

- std/url is plain Narra source with one registry row (stdurl.nra), the same shape as std/set. The std text is embedded in the compiler, so every edit to url.nra needs the generated suite update and a bootstrap before a sample sees it.
- Parts are kept as written and percent-encoded, with `Option` for absent parts, so `a` and `a?` and `//h` and `//h/` stay different. parse accepts relative references because the RFC 5.4 examples are relative; resolve needs a base with a scheme.
- `Query.parse` and `Query.toString` from the brief cannot be two namespaces in one module, so they are `parseQuery` and `queryToString`.
- Dead end: `for (const [index, segment] of ...)` does not compile for entries; the pattern is `#[index, segment]`.
- Review fix: the port is `Option<string>` (the digits as written) so `http://h:` and `http://h:0080/` round-trip. An unbracketed host with a second colon is an Err.
- Doubt: no check of host characters beyond spaces and controls; WHATWG host rules are out of scope.
