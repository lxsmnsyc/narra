# Outcome: guide-final-pass

## Built

Read `docs/guide/README.md` and chapters 1 to 20 in order. Changes by file:

- `README.md`: one line per chapter that lists what each chapter now teaches. The "built but not taught" list is gone, since chapters 4, 7, 14 to 17 and 19 now teach every item on it. "Planned" keeps C-callable functions and `@layout`, with a link to chapter 16 "Not built yet".
- Chapter 1: "dividing by zero ... stops the program" removed. Chapters 2 and 6 say integer division by zero cannot compile, and float division gives `inf` or `nan`.
- Chapter 2: the `+=` sample's `run` no longer writes `uses io, alloc` (effects come in chapter 10, and `run` never needs the list). The first arrow function, in `Arrays.map`, links to chapter 3. Sized `parse` says what `Ok` and `Err` hold and links to chapter 6. "A generator" in the random section now reads "a random number generator", so it is not confused with chapter 12.
- Chapter 3: the `requires` rules say they are the simple form and link to chapter 17 for relations and `||`.
- Chapter 4: `let moreTea = tea` is described as rejected, as chapters 2, 9 and 20 say and as the compiler does (it was "an error when changed while `tea` is still used"). The re-export sample's `run` drops `uses alloc`, and the sentence says an exported caller of `describe` declares it. "Chapter 8" is a link.
- Chapter 5: the `FixedArray` destructuring links forward to chapter 17.
- Chapter 7: the four `Shape` samples drop `uses io, alloc` from `run` and the private `report`, since effects come in chapter 10.
- Chapter 8: "What a module can hold" names type aliases and links to chapter 4 "Export a name from another module".
- Chapter 9: generators link to chapter 12, the one-thread `Cell` note links to chapter 15 "Share a value with a lock", and the identity-type list links to chapters 14 and 15.
- Chapter 11: the intro said every function that calls an `io` function "has to declare" `uses io, alloc`, which contradicts chapter 10. It now says exported functions must, and the samples write the list on private functions and `run` by choice.
- Chapter 14: "Chapter 15" is a link. The `Cell` named in the capture error links to chapter 9.
- Chapter 15: the divisor check linked to chapter 6. It now links to chapter 2 "Dividing by a name". `Cell` links to chapter 9. The non-`Sendable` list adds `Dynamic` trait values with a link to chapter 7 (checked: `Dynamic<Label> is not Sendable`).
- Chapter 16: the paragraph that taught the sized integers again is one sentence linking to chapter 2 "Numbers of other sizes".
- Chapter 17: `derives Shape` pointed at chapter 13. It now links to chapter 7 "Saving and loading values".
- Chapter 18: `-target` became `--target`, as chapter 19 writes it (the compiler takes both).
- Chapter 19: the `std/simd` lane index links to chapter 17's fixed-length arrays.
- `skills/narra/SKILL.md`: the chapter list in "More" had the old numbers (standard library 11, tools 12). It now lists the 20 chapters. The `const` rule adds that a `const` array or struct can still change through an index, a field or `mutable`.

Every "Next:" line points at the next chapter, and every `[Chapter N](...)` link names the file of chapter N with an anchor that exists (checked with a script over all chapters).

## Deviations

- Chapters 12, 13 and 17 still write `uses io, alloc` on `run`. They come after chapter 10, so the list is explained there and is harmless; I left them.

## Found but not fixed

- `docs/guide/08-modules-and-packages.md`: the cache folder is explained twice (in the download steps and in the trust paragraph). Both are short, and merging them means rewriting the section.
- `Ranges.between` (chapter 11) and string escapes such as `\u{1b}` (chapter 13) are used without being taught. Both read clearly in context; a sentence in chapter 2 would cover them.
- `skills/narra/SKILL.md` does not mention the second-name rule (`const b = a` on an array or struct is an error without `copy`), a rule agents likely get wrong. The brief asks only to fix what differs from the guide.
- The chapter 18 command table and `narrac --help` disagree (guide-tools-gaps noted `--help` omits several commands). Compiler work.

## Stubs left

- None.

## Gotchas

- A guide link's anchor follows the heading text with punctuation and backticks dropped (`#state-that-closures-share-cell`). Renaming a heading breaks links in other chapters; the guide suite and the website build do not check anchors.

## Proposed learnings

- A guide sample before chapter 10 does not write `uses` on `run` or on private functions | effects are taught in chapter 10, and neither needs the list | docs/guide | no

## Verification

- `NARRAC=<branch narrac> sh scripts/test.sh --suite guide`: PASS, 610 cases.
- `NARRAC=<branch narrac> sh scripts/test.sh --suite sources`: PASS, 1 case.
- `pnpm --dir website build`: passes, 29 pages prerendered.
- A link script over `docs/guide`: 0 missing files, anchors, wrong chapter labels or bad `Next:` lines.
- Claims checked with the branch narrac: `let backup = hero` is rejected without a write; `+=` sample runs without `uses`; a thread capturing a `Dynamic<Label>` is rejected as not `Sendable`; `check -target` and `check --target` both work.
- `check_prose.py` on the added lines: no errors. Long-sentence warnings remain on the README index lines and on sentences this node only touched.

## Size

- Estimated M, actual M.
