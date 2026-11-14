# Outcome: uses-list-design

## Built

- D117 in `DECISIONS.md`, "Proposed, waiting for the maintainer": in a type, `uses` takes one effect or a parenthesized list (`() => int uses (io, alloc)`); a declaration's own list stays bare (`function f(): int uses io, alloc`), and parentheses are not valid there.
- SPEC: the section 4 function types line, a section 7 rule for `uses` lists in types (not built yet), and the `relay` and `lines` examples in sections 7 and 8.
- `selfhost/src/diagnostics.nra` header: `E1240-E1249  effect lists inside types (uses-list-syntax)`.
- `nodes/uses-list-syntax/brief.md` refined in full (size S to M, not split). Codes: E1240 (parser, a comma after a type's one effect where the list cannot go on), E1241 (checker, an effect name used as a type); E1242-E1249 free.

## Proof

- base: `feat/batch-13` (b23bbd4f), stage 2 narrac, `narrac parse` on the four shapes in `log.md`: wrong AST or parse errors as listed there.
- branch: not built (spike).

## Deviations

- E1240 has no automatic fix, per the roadmap rule against fixes for old forms.

## Stubs left

- None.

## Gotchas

- `usesSuffix` serves both type display and declaration lists; the build needs a separate helper for the type form.
- fmt's `endsWithUses` trailing-comma exception must go with the parser change, or broken parameter lists keep losing their comma.
- Existing fmt bug: a comment after `uses` moves into the type arguments. Not part of the build.

## Proposed learnings

- A list inside a type must be closed by its own delimiter when the type can sit in another comma list | `uses io, int` inside `Map<...>` was ambiguous | selfhost/src/parser.nra | yes (parse dumps)

## Verification

- `NARRAC=<stage 2 narrac> sh scripts/test.sh --suite spec`: PASS, 62 cases.
- `--suite sources`: PASS, 1 case.

## Size

- Estimated S, actual S.
