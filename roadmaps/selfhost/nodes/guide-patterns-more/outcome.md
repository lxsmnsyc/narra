# Outcome: guide-patterns-more

## Built

Three additions to `docs/guide/05-pattern-matching.md`, each with a runnable example:

- Unreachable arms: an error sample after the `int` exhaustiveness error (audit row 101).
- `## Match an array`: array patterns with `[]`, literals, `..._` and `...rest`, the length check in exhaustiveness, `is` and nesting, and the cost of a named rest (rows 100, 131).
- In `## Take a value apart`: why `const [a, b] = items` fails on an Array, the lone `[...rest]` copy, and `FixedArray` destructuring (row 16).
- `## Calls: names, std functions, order`: named arguments in full, std functions as values with the arrow-function fix, and left-to-right argument evaluation (rows 21, 20, 23).

## Deviations

- Calls (rows 21, 20, 23) sit in a section before "The till", not near patterns, since the chapter has no earlier call topic.
- Not covered: the `E0715`, `E0723`, `E0724`, `E0726` codes by number (the guide cites no codes), the trait-function case of named arguments, and the effects a std function value carries (chapter 10 teaches effects).

## Stubs left

- None.

## Gotchas

- Error samples count the `// error:` comment as line 1.

## Verification

- `NARRAC=<prebuilt> sh scripts/test.sh --suite guide`: PASS, 408 cases.
- Every sample run with `narrac check` and `run` in a scratch folder.
- plain-writing checker: no new warnings (one old semicolon warning at line 56 remains).
- `pnpm --dir website build`: passes. The site highlighter flags `name: mutable x` in a call as a keyword mismatch, so the chapter shows that form in prose only.

## Size

- Estimated S, actual S.
