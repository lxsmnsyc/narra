# Log: grammar-fixes

Reproduction on feat/batch-21 (sample appended to docs/install.md, `pnpm --dir website check:grammar`):

- `outer: for (const x of xs) {}`: TextMate reads `:` as the start of a type annotation, so `for`, `const`, `of` are identifiers. Hidden by an `explainDifference` entry.
- `Cell<Option<() => int>>`: the `>` of `=>` closed the inner type arguments, so the real `>>` ended wrong ("TextMate says punctuation, the site says operator" on `>`).
- `I32`, `U8`, `U64`: TextMate class_name (type-name `[A-Z][a-z0-9]...` came first), site constant (library upper-snake rule).
- `CStrings`: TextMate identifier (no rule matches two capitals then lowercase), site class_name.
- `fill(buffer: mutable slots, count: 3)`: TextMate reads `: mutable slots` as a type annotation; `mutable` is an identifier, the site says keyword.

Cause of the name colors: two different rules. Now one rule on both sides: a capitalized name is a constant when it has two or more capitals or underscores and no lowercase (`MAX`, `HTTP_OK`), else a class (`T`, `I32`, `F32`, `CStrings`).
