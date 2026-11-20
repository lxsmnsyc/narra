# Outcome: guide-types-more

## Built

Audit rows 87, 103, 107, 53, 91, 92, 109, 111, 108 and 94 are taught, each with a runnable example.

`docs/guide/04-structs-and-enums.md`, 4 new sections and 1 extended:

- **Structs and enums for any type** (rows 87, 103). `Stock<T>` and `Lookup<T>`, type arguments inferred from the fields, the error when a variant does not fix `T`, and a pointer to chapter 7 for constraints.
- **Give a type a second name** (row 91). Aliases of an array, an enum (variants through the alias), a function type, generic aliases `Pair<T>` and `Shelf<V>`, the missing-arguments error, and the rule against recursive aliases.
- **Name the type of a value** (row 92). `typeof` of a typed top-level constant and of a function, the names-only rule, and the error for `typeof` of an untyped constant in a signature.
- **Hide the fields of a struct**, extended (row 94). `export opaque distinct type ReceiptNumber` in a `receipts` package, with the error text for `from` outside the module.
- **Export a name from another module** (row 53). A `shop` package where `src/till.nra` passes on `Payments.Payment` with `export type` and `Payments.isVoucher` with `export const`.
- The distinct-type section has one new paragraph: a distinct type is not an alias, with a link to chapter 7.

`docs/guide/07-traits.md`, 5 new sections and a longer intro:

- **Traits for a generic type** (rows 107, 103). `struct Bundle<T: Label> implements Label`, `implement Label for Bundle<T>`, the error when the struct lacks the constraint, the error when a use does not meet it, no specialization, and generic enums.
- **Map keys and `Hashable`** (row 111). The float-key error, `Equatable` and `Hashable` written together to key books by title, `Hash.of`, and `<T: Hashable>`.
- **A trait function that makes a value** (row 108). A `Placeholder` trait with `Self` only in the result, `Placeholder.placeholder<Book>(...)`, the expected-type form, a generic caller, and the error without a type. It does not use `Shape`.
- **Values of different types in one array** (row 109). `Array<Dynamic<Label>>`, no import needed, the expected-type rule for `Dynamic.of`, the no-fields error, no `==` or map key, no constraint match, and which traits qualify.
- **Traits on a distinct type** (row 94). `distinct type CustomerId = int implements Label derives Ordered, Show`, derives use the base type, and the error when the base lacks the trait. This is the last section, so guide-serialization can add its section after it.

## Deviations

- Row 53 (re-export) is in chapter 4, not chapter 8, because chapter 8 is outside this node's paths. Chapter 8 has no re-export text, so nothing is said twice.
- `Bundle` in chapter 7 puts `implements Label` on its own line. On one line, `struct Bundle<T: Label> implements Label {` fails `pnpm --dir website build`: the TextMate grammar reads `: Label>` as a type annotation and colors `implements` as an identifier, while the site colors it as a keyword. `narrac fmt` joins the two lines. Node grammar-fixes should fix the grammar and join the line again.
- The `typeof` samples use lowercase constant names (`shopPrices`, `standardDiscount`). `typeof PRICES` fails the same build check: TextMate says class name and the site says constant on a SCREAMING_CASE name after `typeof`. This is a second case for grammar-fixes.
- The re-exported function in the `shop` sample has no effects. See Gotchas for why.
- `docs/guide/README.md` still lists "`typeof` in types and `Dynamic` trait values" under "built but not taught". The file is outside this node's paths. The orchestrator should remove that line.

## Stubs left

- None.

## Gotchas

- Compiler bug: a top-level `const` without a written type cannot hold a function that has effects. `const show = label` with `function label(cents: int): string uses alloc` gives `expected (int) => string, found (int) => string uses alloc`. So `export const describe = Payments.describe` fails when `describe` uses `alloc`. Writing the type works: `export const describe: (payment: Payment) => string uses alloc = Payments.describe`. The guide avoids the case. A compiler node should fix it, and then the guide can re-export a function that allocates.
- A generic function can use `Map<T, V>` with an unconstrained `T`. The compiler checks the key rules per instance and reports a float key at the generic body's lines. The guide therefore does not claim that map keys need `<T: Hashable>`, only that `Hash.of` does.
- `narra error` positions count the `// error:` line as line 1.

## Proposed learnings

- Check every new guide sample with `pnpm --dir website build` as well as the `guide` suite | the TextMate comparison fails on forms the compiler accepts | docs/guide | yes, the build already does.

## Verification

- Every sample was checked with `narrac check` and `narrac run` from the prebuilt compiler in a scratch folder outside the repo, before it went into the chapter.
- `NARRAC=<prebuilt narrac> sh scripts/test.sh --suite guide`: PASS, 487 cases.
- `pnpm --dir website build`: passes, with 0 grammar problems.
- `check_prose.py` on both chapters: no errors. One long-sentence warning remains at `07-traits.md:62`, in text this node did not write.

## Size

- Estimated M, actual M.
