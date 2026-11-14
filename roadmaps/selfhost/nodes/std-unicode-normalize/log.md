# Log: std-unicode-normalize

## Proof (base)

- Base: feat/batch-14 at 87bba23c. Compiler: `sh selfhost/bootstrap.sh <scratch>/base` (release).
- `base/narrac run testdata/unicode_normalize_basic.nra`: exit 1, first line `unicode_normalize_basic.nra:4:1: module "std/unicode" has no export "NormalForm"`, then `no export "normalize"` and `no export "isNormalized"` at every call.

## Inputs

- UCD 17.0.0 from https://www.unicode.org/Public/17.0.0/ucd/: the seven std-unicode inputs plus CompositionExclusions.txt, DerivedNormalizationProps.txt and NormalizationTest.txt (17.0.0, dated 2025-06-30).

## Table format

- Counts at 17.0.0 (one-step decompositions of UnicodeData): 2,081 canonical, 3,833 compatibility (2,685 to one code point in 1,249 runs, 1,148 longer with 2,928 code points); 924 ranges of combining class plus quick check.
- Chosen: one-step decompositions, decomposed recursively at run time (depth at most 4), so the composition table is a list of 3-digit indexes into CANONICAL (2.9 KB) rather than a third copy of the pairs. Fully expanded decompositions were estimated at about 133 KB with 6-digit fields; this layout is 92 KB.
- The generator derives Full_Composition_Exclusion from CompositionExclusions.txt plus singletons and non-starter decompositions, and panics when it differs from DerivedNormalizationProps.

## Conformance

- `branch/narrac run -release testdata/unicode_normalize.nra <ucd>/NormalizationTest.txt`: `20034 of 20034 lines pass` (all five parts, four forms, normalize and isNormalized of every column), 1.2 s.
- Every code point not in part 1 through the four forms (scratch `part2.nra`): `1094978 checked, 0 changed`.

## Review fix: canonical ordering

- Insertion on every push was quadratic in a run of marks. `testdata/unicode_normalize_linear.nra` (80,000 alternating U+0301/U+0316): 605ab75a compiler `over 1 s` in both forms (25.2 s in all); after `orderMarks` (insertion up to 8, merge sort above) `under 1 s` (0.9 s in all, compile included). Full NormalizationTest.txt still 20034 of 20034.
