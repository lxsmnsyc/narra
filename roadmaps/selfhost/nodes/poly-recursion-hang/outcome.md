# Outcome: poly-recursion-hang

## Built

- `typeDepth` and `maxDepth` in `selfhost/src/generics.nra`: the nesting depth of type arguments.
- `addFuncInstance` in `selfhost/src/instances.nra` reports E0310 ("too many instances of generic functions; F may recurse with ever larger types") when an instance's type arguments are deeper than `MAX_TYPE_DEPTH` (64), and queues no body. Function and trait implementation instances both go through it.
- Tests: `testdata/errors/poly_recursion_function.nra` and `poly_recursion_trait.nra`, each with its `.err`.

## Proof

- base: af35fdcb, `narrac check` on the function case and on the trait case: killed at the 20 s cap, no output.
- branch: the same commands exit 1 in under a second with E0310 at the function (6:1) and at the `show` member (7:3).

## Deviations

- No new error code: E0310 already says "may recurse with ever larger types", so the depth limit reuses it.
- The Go-era check for generic types (`checkPolyRecursion`) is declaration based and does not apply to functions, whose growth shows only at the call; a depth limit was used instead.

## Stubs left

- None.

## Gotchas

- Type arguments nested deeper than 64 levels are rejected, also without recursion.
- `seed/narrac.c` is older than the source after this change; the orchestrator regenerates it per batch.

## Proposed learnings

- Bound anything that builds a type per instance by depth, not only by count | the count limit is reached only after minutes when each instance is larger | selfhost/src | yes

## Verification

- `sh selfhost/bootstrap.sh`: fixed point holds.
- `sh scripts/test.sh`: see the final line in the report.

## Size

- Estimated S, actual S.
