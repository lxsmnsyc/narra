# Outcome: guide-proofs

## Built

Chapter 17 (`docs/guide/17-proofs-and-compile-time.md`) teaches the proofs, with a running example of a week of working hours. Each section has a runnable sample and at least one `narra error` sample:

- "Conditions between two parameters": `requires` relations (`start <= end`), how a caller proves them, bounds passed one step, no arithmetic in a relation (row 172).
- "One condition or another": `||` clauses, parentheses inside an `&&` list, the caller proves one side, a check of the whole `||` keeps only the hull (row 172).
- "Promise something about the result": `ensures` lines with comparison patterns, two lines for a range, what a returned value needs to be proven, a call is never proven, where a caller learns the facts (row 173).
- "Arrays with a fixed length": `FixedArray<T, N>`, the three ways to prove an index, `FixedArrays.get`, `FixedArrays.length`, the 1 MiB limit (row 83).
- "A type with a rule": refinement types, `T.from` proven at compile time, `T.check` at run time, the fact from unwrapping, `string` rules, `derives Shape` and `opaque` (row 95).
- "A length as a parameter": `const N: int` on functions and structs, one copy per value, the zero-length divisor error, no computing in type arguments, `N` does not prove `requires` (row 104).

The chapter title and the `Next:` line are unchanged. The last section is about proofs, so guide-compile-time can add its section at the end.

## Deviations

- The relation section does not repeat the divisor proof from chapter 2. It links to it and to the `requires` section of chapter 3.
- `docs/guide/README.md` still lists "SIMD vectors and fixed-size arrays" and "`@const` functions, `ensures` lines and refinement types" under "built but not taught". The file is outside this node's owned paths. Fixed-size arrays, `ensures` and refinement types are now taught, so whoever owns the README (or guide-compile-time, for `@const`) should update that list.

## Spec against compiler

No claim checked in `SPEC.md` sections 4 and 13 differs from the compiler. Checked with the prebuilt narrac:

- A relation is not proven by bounds that leave the order open (E0985), as SPEC 13 says.
- A caller's `if (a || b)` check does not prove `requires a || b`, also when the sides are `step < 0 || step > 0` and the caller checks `step != 0`. SPEC 13 says the caller knows only the hull, so this matches.
- An instance with `N` equal to 0 gives the divisor error naming `averageHours<0>` and its call site.
- `Vector<N + 1>` is an error, and passing `N` to a function with `requires` is E0985 even in a function with no instance.

Notes for spec-corrections, not mismatches:

- A struct literal does not infer a constant parameter from its fields: `Schedule { hours: [8, 8, ...] }` with `struct Schedule<const N: int> { hours: FixedArray<int, N> }` is "cannot tell constant parameter N of Schedule here". `Schedule<14> { ... }` works. SPEC 4 says only that a call infers it, so a sentence on struct literals would help.
- The E0985 hint for a `||` clause says "pass a constant, or a name checked first with an `if`". A check of the whole `||` does not prove it, so the hint can mislead. A hint that says to check one side would be clearer.

## Stubs left

- None. Compile-time constants (`@const` functions) are guide-compile-time's section.

## Gotchas

- An `is` pattern that binds names cannot sit inside `&&`. The `ensures` sample nests two `if` statements for that reason.
- The error text of a `narra error` sample is a substring match, so the divisor sample for `averageHours<0>` leaves out the `called at <file>:<line>:<col>` tail, which holds the scratch file name.

## Proposed learnings

- None.

## Verification

- `NARRAC=<prebuilt narrac> sh scripts/test.sh --suite guide`: PASS, 498 cases.
- `pnpm --dir website build`: passes, including `check:grammar`.
- Every sample checked with `narrac check` and `narrac run` in a scratch folder outside the repo.
- `check_prose.py` on the chapter: `prose: ok`.

## Size

- Estimated M, actual M.
