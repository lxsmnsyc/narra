# Log: pkg-http-cookies

- Checked first: no `cookie` module existed under `packages/http` or `std`.
- `Result<void, E>` does not exist, so `set`, `remove` and `validate` give `Option<HttpError>` (None is success).
- Wrote `cookie.nra`, then `checkcookie.nra` (pure checks, 4000-round fuzz, loopback server, curl with a cookie jar). First run caught my own wrong weekday for 2099-12-31 (a Thursday) and a fuzz that never produced a valid cookie.
- Mutable-argument rule: `randomText(mutable random)` twice in one expression is an error, so the fuzz uses locals.
- `curl -c jar -b jar` with four URLs on one connection: set, read, clear, read. The jar goes in `$TMPDIR` and is removed after the thread scope.
