# Outcome: finish-facts-lines

## Built

- `findHeaderLine` scans bytes of the source and allocates nothing (no `Str.split`).
- Template `${}` parts are no longer lexed a second time. `parseTemplate` already has their tokens; it now adds their referenced names to `facts.referenced` and merges the names of nested templates. `collectTemplateNames` is removed, and `finishFacts` starts from the names collected during parsing.
- Only `selfhost/src/parser.nra` changed. No language change, no new diagnostics.

## Result

- Instructions retired for `narrac emit selfhost`, 3 runs each: base 9.339e9, 9.326e9, 9.341e9; branch 9.058e9, 9.068e9, 9.071e9. About 2.9% fewer; the base spread is 0.16%, so the gain is real.
- Peak memory is unchanged within noise (about 131 MB).
- Emitted C is identical between base and branch compilers; the bootstrap fixed point holds.

## Deviations

- None.
