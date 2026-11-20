# Outcome: guide-cell-and-state

## Built

- `docs/guide/09-memory-and-mutable.md` has three new sections: "State that closures share: `Cell`", "Enums and tuples are replaced", and, inside "Freeing memory", "Cycles through a `Cell`" and "Copies the compiler does not show".
- Concepts taught, each with a runnable example (audit rows 29, 137, 138, 132): `Cell` with `create`, `get`, `set` and `update` (a score shared by escaping closures), a struct inside an `Option` is replaced and not changed, identity-type cycles with the `NARRA_CHECK_LEAKS` report and the two ways out (`Cells.set(cell, None)`, `std/arena` with a link to chapter 13), and `NARRA_CHECK_COPIES`.
- `docs/guide/20-narra-for-typescript-programmers.md`: the counter error line now carries the full E0612 message, and a `Cell` counter is the idiomatic answer before the `mutable` struct version.

## Deviations

- The leaking program and the program with a hidden copy are `narra fragment` blocks. The guide suite runs every `narra` block with the leak and copy checks on, so a block that leaks or hides a copy would fail. A passing `narra` block follows each.
- A type alias `Handler` stands for `() => int` in the cycle samples, because the site highlighter and TextMate disagree on `>>` after `=> int` and fail the website build.
- Tuples are covered by one sentence, not an example: the same rule as enums, and no tuple field in the running game.

## Stubs left

- None.

## Gotchas

- Chapter 9 now links to chapter 15 (`Shared`, threads) and to chapter 13 (`#graphs-stdarena`). Chapter 15 is still a stub when this lands.

## Proposed learnings

- A guide sample cannot show a leak or a hidden copy as a `narra` block | the guide suite runs with both checks on | docs/guide | yes

## Verification

- `NARRAC=<prebuilt> sh scripts/test.sh --suite guide`: PASS, 348 cases.
- `pnpm --dir website build`: passes.
- `check_prose.py` on both chapters: only long-sentence warnings, none on new text except one sentence kept from the old "Freeing memory" paragraph.
- Run by hand in a scratch folder: leak check exit 102 with `leak check: 2 allocations still live`, copy check exit 103 with `array 1`.

## Size

- Estimated M, actual S.
