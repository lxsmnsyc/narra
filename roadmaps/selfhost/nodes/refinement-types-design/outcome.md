# Outcome: refinement-types-design

## Built

- D116 in `DECISIONS.md`, "Proposed, waiting for the maintainer": `distinct type Port = int requires >= 1 && <= 65535`, `distinct type Method = string requires "GET" | "POST"`.
  1. Predicate: `ensures` patterns (comparisons, literals, or-patterns) joined by `&&` at the top of the list; bases are integer types and `string` (literals only).
  2. `Port.from(value): Port` is proven at compile time like a `requires` call; `Port.check(value): Option<Port>` tests at run time. Opaque hides both.
  3. Unwrapping a refined integer (`Port.toInt(e)`, any `e`, fields included) gives the hull of the patterns through `findOperandFact`; `const n = Port.toInt(p)` is seeded.
  4. `derives Shape` calls `check`; failure is `Err(".port: 70000 does not satisfy Port (>= 1 && <= 65535)")`. The error stays one string.
  5. No generic refinements (distinct types take no type parameters); derived traits other than Shape unchanged; E1220 parse, E1221 base, E1222 pattern, E1223 unproven `from`, each naming the patterns as written.
- SPEC section 4, "Refinements" under "Distinct types", marked not built yet.
- `selfhost/src/diagnostics.nra` header: `E1220-E1239  refinement types on distinct types`.
- `nodes/refinement-types/brief.md` refined to a full L brief (not split).

## Proof

- base: with the batch-13 compiler, `slot(Port.toInt(port))` against `requires index >= 1` is E0985, and `requires` after a distinct base is not parsed.
- branch: not built (spike).

## Deviations

- The inbox draft had `Port.from(raw): Option<Port>`. D116 keeps `from` total and proven, and adds `check` for the run-time test; the rejected list says why.
- No code for a refined `from` used as a value: no distinct function is a value today (probed).

## Stubs left

- None.

## Gotchas

- SPEC samples are inline code, not `ts` blocks, until the build lands (an unbuilt block fails the spec suite).
- Other batch-13 spikes add header range lines to `diagnostics.nra` too; keep every line, in number order.

## Proposed learnings

- A new value-level fact should go through `findOperandFact` | divisors, shift counts, `requires` arguments and `ensures` returns all read it | selfhost/src | no

## Verification

- `NARRAC=/private/tmp/claude-501/refinement-types-design/out/narrac sh scripts/test.sh --suite spec`: PASS, 62 cases.
- `... --suite sources`: PASS, 1 case.

## Size

- Estimated S, actual S.

## Questions for the maintainer

1. `from` proven at compile time plus `check` at run time, or `from` returning `Option` as in the inbox draft? Recommendation: proven `from` plus `check`.
2. The name of the run-time constructor: `check`, `tryFrom` or another? Recommendation: `check` (`if (Port.check(raw) is Some(port))`).
3. Bases in the first version: integer types and `string` only? Recommendation: yes; struct, enum and call predicates later.
