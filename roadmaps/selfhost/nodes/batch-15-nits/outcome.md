# Outcome: batch-15-nits

## Built

- fmt keeps comments inside a `uses` list where written, for a declaration's list (`uses /* f */ io, /* g */ alloc`) and a type's list (`Generator<int> uses (io, /* c */ alloc)`, `uses /* d */ io`, `/* e */ uses io`, function types). The parser notes a trivia node for the `uses` keyword (key `u<line:col>` of the function keyword or the type) and one per effect (`u<line:col>.<index>`); `printUses` in `selfhost/src/fmtprinter.nra` prints them, and `fmt.nra` and `fmtexprs.nra` call it.
- `Simd.load` on a FixedArray says `expected Array<i32>, found FixedArray<i32, 4>; \`Simd.fromFixed\` makes a vector from a FixedArray` (E0119, the existing mismatch code).
- Guide chapter 12 line 90 says the C-style `for` loop works.
- Inbox: removed the entries for the uses-list comments, `Simd.load`, the C-style `for` text and the diagnostics guard.
- Goldens: `testdata/fmt/uses_list_comments.{nra,fmt}`, `testdata/errors/simd_load_fixed.{nra,err}`.

## Proof

- base: c4884fe1, `narrac fmt` on `uses_list_comments.nra` moves all five comments; `narrac check simd_load_fixed.nra` prints the message without the hint. Details in `log.md`.
- branch: the same commands give an unchanged file and the message with the hint; both goldens pass.

## Deviations

- Edited `selfhost/src/parser.nra`, `selfhost/src/fmtexprs.nra` and `selfhost/src/fmtprinter.nra`, outside the brief's shared paths: a comment can only stay in place with trivia nodes from the parser, and the printer helper lives with the other comment helpers.
- The diagnostics guard entry was in the inbox already fixed in batch-13; removed with the others.

## Gotchas

- `parseUsesList` and `parseTypeUses` take the owner's position as a second argument. A new caller must pass the function keyword position or the type's position.
- `Simd.load` reuses E0119; no code was reserved.

## Verification

- `narrac emit` of the base source with the base compiler and the branch compiler: byte-identical C (the parser only notes trivia nodes when the collector is on).
- `sh scripts/test.sh`: see the line in the final reply.

## Size

- Estimated S, actual S.
