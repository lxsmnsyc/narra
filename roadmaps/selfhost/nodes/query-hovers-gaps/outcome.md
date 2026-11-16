# Outcome: query-hovers-gaps

## Built

- Alias uses (`Int.toString`) get their hovers from the FieldExpr sites instead of lexer tokens, so they work inside template parts, arguments, array literals and match arms.
- A builtin called by name shows its signature with effects: `print` gives `(string) => void uses io`. The lookup covers the prelude and names imported from a std module.
- A struct literal shows `struct Player` at the type name and `name: string` at each written field (not shorthand fields, which stay the local's hover). Spread forms work.
- Cases added to `testdata/query/hovers_names`: templates, call arguments, array literal, match arms, `print`, struct literals with a spread.
- The `print` signature fix is in `describeSite`, so the LSP hover and `type-of` show it too (goldens `lsp/nav_stale`, `query/basic`, `query/hovers` changed). The alias and struct passes are in `collectHovers` only, so the LSP does not get them.

## Proof

Before, on the brief's sample: `print` showed `void`, `toString` inside the template showed `string`, `Int` inside the template had no hover. After: `(string) => void uses io`, `(int) => string uses alloc`, `module std/int`. See `log.md`.

## Result

Not an experiment.

## Deviations

- Template alias hovers and struct literal hovers are in the query command only; the LSP hover still lacks them.
- `Shapes.Point { ... }` and labelled variant arguments are not valid syntax, so they have no case.

## Stubs left

- None.

## Gotchas

- A template string is one token. Code that scans `view.tokens` misses everything inside `${...}`; use the sites.

## Proposed learnings

- None.

## Verification

- `sh scripts/seed-update.sh` until "seed unchanged".
- `sh scripts/test.sh`.

## Size

- Estimated S, actual S.
