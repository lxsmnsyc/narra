# Log: identity-cycles

- Branch `spike/identity-cycles` from `origin/feat/batch-1`. Built `out/narrac` with `sh selfhost/bootstrap.sh out`.
- closure-borrow is done; SPEC section 6 "Cycles" already carries its wording ("only a closure that cannot outlive the call captures a variable by reference"). The brief's bullet "A `let` variable or `mutable` parameter cannot be captured or stored" no longer exists, so nothing to align.
- Probes in `probes/`, run with `NARRA_CHECK_LEAKS=1 out/narrac run -sanitize`:
  - `probe_values.nra`: `Arrays.push(mutable tree.children, alias)` with `alias = tree`; `link.value = Link { next: Some(link) }` on a `Box`; a closure capturing a snapshot of the struct stored in the struct's field. Output `1 0 2 3`, exit 0, no leak report. The writes copied, as section 6 says.
  - `probe_self_closure.nra`: `const loop = (): int => loop()` is "unknown function loop": a `const` cannot name itself.
  - `probe_let_escape.nra`: an escaping closure stored in a `let` struct's field, capturing it, is rejected (E0612 text).
  - Note: D81 "const fixes the name, not the contents" is not built yet (`tree` as `const` cannot be passed as `mutable`); probes use `let`.
- No positive control is possible: no identity type exists, and `Box` is a copy-on-write value, not an identity type.
- Found while reading the runtime: `narra_live` is a plain `int64_t`; the leak check reads it at exit. `narra_co_next` resets it after a panicking coroutine in test builds (D79), which hides a cycle built in that coroutine.
- Wrote D80, SPEC section 6 "Cycles", section 8 "Channels" and "Shared state", section 16 "Deferred".
