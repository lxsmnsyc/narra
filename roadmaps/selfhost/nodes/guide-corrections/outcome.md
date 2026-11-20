# Outcome: guide-corrections

## Built

- `docs/guide/02-values-and-types.md`: tuples can be indexed. The example now prints `cheapest[0]` and `cheapest[1]`, and a `narra error` sample shows `pair[0] = "milk"` rejected. A `const` name can be passed as `mutable`.
- `docs/guide/01-getting-started.md`, `docs/guide/08-modules-and-packages.md`, `skills/narra/SKILL.md`: `copy` is listed with `print`, `assert` and `panic`.
- `docs/guide/08-modules-and-packages.md`: all 55 std modules, in five groups with one line each.
- `SPEC.md`: the title reads "Narra specification".

## Proof

- Compiler (`narrac check`) accepted `pair[0]`, `add(mutable xs)` with `const xs`, and `copy(xs)` without an import.
- `narrac builtins` lists `copy` with `module=""`.
- An import of `std/nope` prints the supported list: 55 modules, one more than the audit's 54 (it counts `std/async/time`, `std/async/net` and `std/async/file` as three).

## Deviations

- The audit says 54 modules. The compiler lists 55, and the guide says 55.
- Prelude types `Dynamic` and `Async` are not added to the "no import" list. The brief named only `copy`.

## Stubs left

- None. The std list in chapter 8 stays until the std reference exists.

## Gotchas

- The list in chapter 8 is by hand. A new std module needs a line there.

## Proposed learnings

- None.

## Verification

- `NARRAC=<prebuilt> sh scripts/test.sh --suite guide` and `pnpm --dir website build`: guide suite PASS, 335 cases; website build passes.
- `check_prose.py` on the changed guide files.

## Size

- Estimated S, actual S.
