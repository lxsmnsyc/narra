# Log: grammar-fixes-two

- Base feat/batch-21. Sample appended to a scratch copy of docs/install.md (reverted), `pnpm --dir website check:grammar`: 3 mismatches.
  - `struct Bundle<T: Label> implements Label {`: TextMate identifier, site keyword, on `implements`. Reproduced.
  - `typeof PRICES` (twice): TextMate class_name, site constant. Reproduced.
- Cause: the `:` in `<T: Label>` and in `x: typeof PRICES` opens `type-annotation`, whose `type-expression` has no keyword rule and colors every capitalized name as a type, with no constant rule.
- Fix: `type-expression` gets a keyword rule (`implements derives is in typeof`) and `#constant-name` before the type-name rule.
- After the fix the same sample: 0 problems.
- Docs: chapter 7 `Bundle` header on one line; chapter 4 uses `SHOP_PRICES`, `STANDARD_DISCOUNT`. `narrac fmt --check` passes on both samples.
