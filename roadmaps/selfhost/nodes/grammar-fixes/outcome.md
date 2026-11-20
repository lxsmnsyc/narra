# Outcome: grammar-fixes

## Built

- `editors/textmate/narra.tmLanguage.json`: `label` and `named-mutable` rules, `=>` inside type arguments, one constant/class rule for names.
- `website/src/highlight/narra.ts`: `promoteCapitalNames` replaces the library upper-snake, PascalCase and text-set rules.
- `website/scripts/check-grammar.ts`: the label known difference is gone.
- Guide: chapter 9 writes `Cell<Option<() => int>>`, chapter 16 uses `CStrings`, `I32.toString` and `U64.toString`, chapter 5 has a runnable named `mutable` argument.

## Proof

- base: feat/batch-21, a sample with the cases in docs/install.md, `pnpm --dir website check:grammar`: 5 mismatches (see log.md); the label case showed once its known-difference entry was removed.
- branch: the same sample: 0 problems.

## Deviations

- Names like `I32`, `F32`, `U8` and single capitals are classes, not constants (the brief did not say which side was right). `A1B` and `MAX` stay constants.

## Stubs left

- None.

## Gotchas

- The grammar reads every `:` in an expression as a type annotation; a new construct that puts a word after `:` needs its own rule before `type-annotation`.

## Proposed learnings

- A name-color rule lives in two places, the TextMate regexes and `promoteCapitalNames` | keep them equal | website/src/highlight | yes (check-grammar)

## Verification

- `pnpm --dir website build`, `pnpm --dir website typecheck`: pass.
- `NARRAC=<prebuilt> sh scripts/test.sh --suite guide`: PASS, 460 cases.

## Size

- Estimated S, actual S.
