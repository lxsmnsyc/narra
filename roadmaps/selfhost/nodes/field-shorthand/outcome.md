# Outcome: field-shorthand

## Built

- `Point { x, y }` means `Point { x: x, y: y }` (D119) in struct literals, generic struct literals with or without type arguments, struct literals after a spread, and as an argument of an enum constructor. Mixed forms work. A plain `{ x }` without a type name or spread stays a block expression.
- `FieldInit.isShorthand: bool = false` in `selfhost/src/ast.nra`. The parser (`parseStructLit`) builds an `Expr.Ident` at the field's position for a shorthand field. No other pass changed.
- `selfhost/src/fieldshorthand.nra`: `countErrors` and `addShorthandHint`. An unknown name at a shorthand field reports E0203 with the hint "the field shorthand `x` reads a variable named x". Used in `checkStructLiteral` and `inferGenericStruct` (`exprs.nra`).
- fmt keeps the written form (`x` stays `x`, `x: x` stays `x: x`); astdump prints ` shorthand` after the `init` name.
- E0037 (the old "field shorthand is not allowed") is a retired row in `diagnostics.nra`. No E1280-E1284 code was used.
- Tests: `testdata/field_shorthand.nra/.out`, `testdata/errors/field_shorthand_{unknown,duplicate,unknown_generic}`, `testdata/astdump/field_shorthand`, `testdata/fmt/field_shorthand`. SPEC "Building a struct", guide chapters 4 and 12 updated.

## Proof

- base: 6c616a46, `narrac run testdata/field_shorthand.nra` -> `35:18: write `x: x`; field shorthand is not allowed`, exit 1.
- branch: same command prints the 11 lines of `field_shorthand.out`, exit 0.

## Deviations

- Removed `testdata/errors/patterns_shorthand.*` and `testdata/parse/shorthand.*` (outside owned paths): they tested the old E0037 error.
- Edited `docs/guide/04-structs-and-enums.md`, `docs/guide/12-narra-for-typescript-programmers.md` (the error sample became a working one) and `SPEC.md`; `compiler/README.md` needed no change.
- Hover, go-to-definition and rename: the language server has only diagnostics and `narrac query` has no struct-field sites, so there is nothing to extend. The value is a name at the field's position, so it resolves to the local. A future rename must treat a field with `isShorthand` as `x: newName`.

## Stubs left

- None. field-shorthand-sweep may use the form after the batch seed update.

## Gotchas

- Shorthand needs a type name or a spread; `{ x, y }` alone is a block.
- Keywords such as `from` cannot be field names, so they cannot be shorthand.

## Proposed learnings

- Sugar that later passes need not see is cheapest as a parser rewrite into the long form plus a flag for fmt and astdump | later passes stay unchanged | selfhost/src | no

## Verification

- `narrac emit selfhost` of the base tree from the base compiler and from the branch compiler: identical (247313 lines), so code that does not use the feature emits the same C.
- Suites errors, astdump, fmt and guide pass; full `sh scripts/test.sh` result in the report.

## Size

- Estimated M, actual S.
