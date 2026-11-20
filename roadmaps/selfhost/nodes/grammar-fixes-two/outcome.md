# Outcome: grammar-fixes-two

## Built

- `editors/textmate/narra.tmLanguage.json`: `type-expression` has a keyword rule and the constant rule, so `implements` after `<T: Label>` and `typeof PRICES` match the site.
- `docs/guide/07-traits.md`: `struct Bundle<T: Label> implements Label {` on one line.
- `docs/guide/04-structs-and-enums.md`: `SHOP_PRICES` and `STANDARD_DISCOUNT` in the `typeof` samples.

## Proof

- base: feat/batch-21, a sample with both cases in docs/install.md, `pnpm --dir website check:grammar`: 3 mismatches (log.md).
- branch: the same sample: 0 problems.

## Deviations

- None. The site highlighter needed no change.

## Stubs left

- None.

## Gotchas

- `type-expression` is entered by every `:`, including a type parameter constraint, so it runs up to the `{` of a struct and sees `implements`.

## Proposed learnings

- A rule added to the expression patterns of the grammar may need a twin in `type-expression` | check both | editors/textmate | yes (check-grammar)

## Verification

- `pnpm --dir website build`, `pnpm --dir website typecheck`: pass.
- `NARRAC=<prebuilt> sh scripts/test.sh --suite guide`: PASS, 601 cases.

## Size

- Estimated S, actual S.
