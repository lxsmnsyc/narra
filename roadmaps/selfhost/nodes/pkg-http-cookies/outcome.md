# Outcome: pkg-http-cookies

## Built

New public module `http/cookie` in `packages/http/src/cookie.nra`, listed in `[exports]` of `packages/http/narra.toml`. Checks are in `packages/http/src/checkcookie.nra`, called from `checks.nra`. The README has a "Cookies" section.

Names:

- `Cookie { name, value }`, `SameSite` (`Strict`, `Lax`, `None`).
- `SetCookie { name, value, path: Option<string>, domain: Option<string>, maxAge: Option<Duration>, expires: Option<int> (unix seconds), secure, httpOnly, sameSite: Option<SameSite> }`.
- `read(request): Array<Cookie>` and `get(request, name): Option<string>` (first match) over every `Cookie` header. Malformed pairs are skipped.
- `create(name, value): SetCookie`, `validate(cookie): Option<HttpError>`, `format(cookie): Result<string, HttpError>`, `formatDate(seconds): Option<string>`.
- `set(mutable response, cookie): Option<HttpError>` adds one `Set-Cookie` header per cookie, or adds nothing and gives the error. `remove(mutable response, name, path): Option<HttpError>` adds an expired cookie (`Expires` 1970, `Max-Age=0`).

Refused: a name that is not a token, a value outside the RFC 6265 cookie-octet set (optionally in one pair of quotes), a path not starting with `/` or holding control characters or `;`, a domain that is not letters, digits, `-`, `.`, an `expires` before 1970, `SameSite=None` without `Secure`, and the `__Host-` and `__Secure-` prefix rules.

## Deviations

- `set`, `remove` and `validate` give `Option<HttpError>` (None is success), because Narra has no `Result<void, E>`.
- `maxAge` below zero is written as `Max-Age=0`.
- Added beyond the brief: `validate`, `format`, `formatDate`, the `__Host-` and `__Secure-` prefix rules, and a path that must start with `/`.
- Shared paths touched: `narra.toml` (export), `checks.nra` (one call), `README.md`.

## Stubs left

- None.

## Gotchas

- `SameSite.None` is a variant of the enum; the field is `Option<SameSite>`, so "no attribute" is `None` and the attribute is `Some(SameSite.None)`.
- `remove` must use the path the cookie was set with, or browsers keep the cookie.
- Reading does not decode percent escapes and does not merge duplicate names; `get` gives the first.
- The curl check keeps its jar in `$TMPDIR` and removes it afterwards.

## Proposed learnings

- None.

## Verification

- `out/narrac fmt packages/http`: clean.
- `out/narrac run -sanitize packages/http`: `251 checks, 0 failed` (curl ran: set, read back with `-c`/`-b`, removed).
- `sh scripts/test.sh`: `test.sh: FAIL after 459s`. One case failed, `programs/async_api_all_cancel.nra/debug` (sleep-order timing; the run shared the machine with two other workers). It passes alone (`NARRAC=out/narrac sh scripts/test.sh --filter async_api_all_cancel`: `PASS, 5 cases`). Every other shard passed and the package checks printed `251 checks, 0 failed`.

## Size

- Estimated S, actual S to M: the date formatter and the check file made it larger than the module.
- Risk estimated low, measured low.
