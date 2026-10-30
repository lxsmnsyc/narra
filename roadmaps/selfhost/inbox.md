# Inbox: selfhost

Ideas and gaps found during work, not yet nodes. Triage with the user.

- 2026-10-30, from the prelude-imports review (#29): two imports with the same alias are accepted silently (`import * as Int from "std/int"` twice). `import * as Option from "std/array"` is accepted too, so an alias can take a prelude type's name for an unrelated module. Proposed: a small check node that rejects duplicate aliases, and decides whether an alias may name a prelude type other than its own.
- 2026-10-30, from the prelude-imports review: `SPEC.md` documents `Range.stepped`, but no std module registers it. Either implement it in `std/range` or drop it from the spec. This was already true on main.
