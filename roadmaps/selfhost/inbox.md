# Inbox: selfhost

Ideas and gaps found during work, not yet nodes. Triage with the user.

- 2026-10-30, from the selfhost-check-imports review (#34): `narrac decls` is quadratic in the number of declarations in one module (2000: 0.3 s, 4000: 1.15 s, 8000: 4.9 s); main shows the same, so it predates #34. Not located; likely the declaration passes or the dump. Fix before the selfhost compiler checks large generated files.
- 2026-10-30, same review: `TestSelfhostBuiltins` checks fixed-signature builtins with a one-way compile probe, so a Go parameter widened beyond the Narra table, or a result that is a subtype, passes. A Go-side dump of `checkFixed` signatures would make it exact.
- 2026-11-01, from the std-iterator discussion: async generators (`async function*`, "not in the first version" in SPEC section 8) would let `std/iterator` operators work on async streams, like IxJS AsyncIterable. Add time operators (`debounce`, `throttle`, `merge`) to `std/async` with them. Needs async and generators first.
- 2026-11-01, user direction: make closures more powerful and closer to TypeScript semantics after m11-closures. Candidates to weigh: inferring `escaping` where the cost is still visible in tooling, capturing `let` from escaping closures through an automatic `Cell`, and async closures. Each needs a decision record; run a goal debate before planning.
- 2026-11-02: compound assignment (`+=`, `-=` and the rest), increment and decrement (`++`, `--`), user request. Added as outline node compound-assign.
