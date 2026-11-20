# Outcome: guide-basics-gaps

## Built

Chapter 2 (`docs/guide/02-values-and-types.md`) has new sections, each with a runnable example:

- "Dividing by a name", "Division when the divisor may be zero" and "Powers", in "Whole numbers and decimals". They teach divisor proofs and what a check proves (rows 115, 117), `Int.divmod`, `Int.floorDiv`, `Int.compare` (46) and the missing power operator (45).
- String `+=`, `Str.concat` and the string builder (`Str.builder`, `Str.append`, `Str.appendChar`, `Str.build`), in "Text" (10, 120).
- `? :`, and "Long lines" for line continuation and the parenthesised multi-line `? :` (5, 36), in "Make a decision".
- "Count with a `for` header" and "Leave an outer loop", for the C-style `for`, labels, `break label` and `continue label` (33, 34), after "Stop early or skip an element".
- "More array functions" (`Arrays.map`, `Arrays.slice`, `Arrays.withCapacity`, `Arrays.values`) and "Count with a map" (`Maps.update`, `Maps.entries`, the `counts[key] += 1` error) (127, 79).

Chapter 6 keeps one sentence on division and links to the chapter 2 section.

## Deviations

- `website/scripts/check-grammar.ts`, outside the owned paths: the TextMate grammar reads the loop after a label (`outer: for ...`) as a type annotation, so the website build failed on the label sample. `explainDifference` now lists that case as a known difference. A grammar fix in `editors/textmate` would remove the need.
- `Int.pow` and `Float.pow` do not exist; the audit row is stale. The guide teaches `Math.powInt` (returns an `Option`) and `Math.pow`.
- `Arrays` and `Maps` have few functions: only `entries`, `filled`, `length`, `push`, `values`, `withCapacity`, `map`, `slice` for arrays and `entries`, `keys`, `remove`, `size`, `update` for maps. All are now taught in chapter 2 (`filled`, `length`, `push`, `entries`, `keys`, `remove` and `size` already were).

## Stubs left

- None.

## Gotchas

- An `// error:` line in a `narra error` sample counts the comment line, so the position is one more than in a bare file.

## Verification

- `NARRAC=<prebuilt narrac> sh scripts/test.sh --suite guide`: PASS, 363 cases.
- `pnpm --dir website build`: passes.
- `check_prose.py` on chapters 2 and 6: two long-sentence warnings left, both in text this node did not write; one semicolon warning in chapter 6 in text this node did not write.

## Size

- Estimated M, actual M.
