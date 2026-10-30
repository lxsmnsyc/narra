# Inbox: selfhost

Ideas and gaps found during work, not yet nodes. Triage with the user.

- 2026-10-30, from the prelude-imports review (#29): two imports with the same alias are accepted silently (`import * as Int from "std/int"` twice). `import * as Option from "std/array"` is accepted too, so an alias can take a prelude type's name for an unrelated module. Proposed: a small check node that rejects duplicate aliases, and decides whether an alias may name a prelude type other than its own.
- 2026-10-30, from the prelude-imports review: `SPEC.md` documents `Range.stepped`, but no std module registers it. Either implement it in `std/range` or drop it from the spec. This was already true on main.
- 2026-10-30, from the selfhost-check-imports review (#34): `narrac decls` is quadratic in the number of declarations in one module (2000: 0.3 s, 4000: 1.15 s, 8000: 4.9 s); main shows the same, so it predates #34. Not located; likely the declaration passes or the dump. Fix before the selfhost compiler checks large generated files.
- 2026-10-30, same review: `TestSelfhostBuiltins` checks fixed-signature builtins with a one-way compile probe, so a Go parameter widened beyond the Narra table, or a result that is a subtype, passes. A Go-side dump of `checkFixed` signatures would make it exact.
