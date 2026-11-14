# Outcome: parameter-defaults-design

## Built

- D115 in `DECISIONS.md`, "Proposed, waiting for the maintainer": constant defaults on trailing parameters of declared functions, filled in at the call.
- SPEC section 2, Functions: the bullet "Parameter defaults (D115, not built yet: node parameter-defaults)"; "A call passes every argument" now adds "that has no default"; the "no optional parameters" line and its `OpenOptions` example become "No varargs".
- `selfhost/src/diagnostics.nra` header: `E1210-E1219  parameter defaults`.
- `nodes/parameter-defaults/brief.md` refined into a full M brief (owns, shared, codes E1210-E1216, samples, error tests, guide and README changes). Not split.

## Answers

1. Allowed: yes. Trailing only, compile-time constants by the struct field default rule. A middle default is skipped by naming a later parameter.
2. No `(a: int, b?: int) => int`. A function value's type lists every parameter; calls through it pass every argument.
3. Named arguments: defaults are filled after matching. Generics: checked with type parameters open, no inference from left-out arguments. `mutable`: no default (E1211). Traits: none for now (E1214). extern: none (E1215). Arrows and function types: none (E1213). `name?: T`: E1216 with a fix.
4. Effects: a default does none except `alloc` for an array literal, charged at the call that leaves it out.
5. SPEC as above; guide chapters 3 and 11 change in the build node; std has no option-struct APIs today, so no std change.

## Deviations

- No goal debate run: the answer follows from D26, D87 and the measured counts.
- No SPEC sample added, since the spec suite checks every ts block and its skip list is outside this node; the build node can add one.

## Stubs left

- None.

## Gotchas

- `parseFuncHeader` is shared by trait and extern functions.
- The effect pass may not charge `alloc` for an omitted non-empty array struct field default (`Expr.StructLit` in `effects.nra`). Unverified.
- The compiler cannot use defaults until the batch's seed update; an adoption sweep (`checkExpression(..., None)` and others) would be a later node.

## Proposed learnings

- none

## Verification

- No code changed; only the header comment of `diagnostics.nra`. Suites not run, since no SPEC sample changed.

## Size

- Estimated S, actual S.
