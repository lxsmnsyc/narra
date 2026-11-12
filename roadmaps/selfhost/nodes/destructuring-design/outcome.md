# Outcome: destructuring-design

## Built

- D103 in `DECISIONS.md`, revised after the maintainer's answers (2026-11-12). Array patterns are built, including `[first, ...rest]`, where `rest` is a copied Array with its cost stated. Destructured parameters are rejected: they have no name, so D87 named calls cannot pass them. Destructuring assignment is rejected for now, with the rules kept for reopening.
- `roadmaps/selfhost/nodes/destructuring-design/count.py`: the counts behind D103 (`-v` lists each site).
- Refined brief, `outline: false`, `status: todo`: array-patterns, size L, hard dependency on destructuring-design, reserves E0722, E0723, E0724 and E0726.
- Dropped by D103: param-destructuring and destructuring-assign. Each brief keeps its text with a "Dropped by D103" line at the top of its Goal, and reserves no codes (E0721 and E0725 released). The orchestrator sets their status fields.

## Proof

- base: feat/batch-10 head b25c2b23. The first run was on ececc85a; the one commit since then only edits seed-switch-cleanup's brief. `out/narrac run` on samples in /private/tmp/claude-501/destructuring-design/s/: `function area({ width, height }: Rect)` gives expected a name, found "{". `(#[x, y]: #[int, int]) => ...` gives found "#[". `match (items) { [] => ... }` gives found "[". `#[x, y] = #[y, x]` gives only variables, fields and array elements can be assigned. Nested struct destructuring and `for (const { c: count } of items)` work.
- branch: not built (spike, no compiler code).

## Deviations

- None after the gate. The first draft deferred the named rest and proposed building destructured parameters; the maintainer reversed both.

## Stubs left

- None.

## Gotchas

- `isIrrefutable(checker, pattern)` in `patterns.nra` takes no type. FixedArray patterns need the scrutinee type at every nesting level.
- `const [a, b] = items` on an Array needs no new code: E0715 already reports refutable destructuring (`checkDestructure` in `match.nra`).
- A named rest copies the tail on each match that binds it. A recursive `[first, ...rest] => f(rest)` is quadratic, so SPEC states the cost next to the slice rule.
- Codes E0722, E0723, E0724 and E0726 are proposed in the array-patterns brief. The orchestrator confirms that no other open node takes them (E0720 was the last code used in the patterns range).

## Proposed learnings

- Count corpus uses with a script kept in the node folder, and test each loop spelling the corpus uses | the first count.py run checked only C-style loops and missed 15 `Ranges.between(1, ...)` tail loops | roadmaps/selfhost/nodes | no

## Verification

- `sh selfhost/bootstrap.sh /private/tmp/claude-501/destructuring-design/out`: fixed point, seed current.
- `python3 roadmaps/selfhost/nodes/destructuring-design/count.py` from the repo root: the counts in D103.
- No `sh scripts/test.sh`: the node changes no compiler code, samples or std text.

## Size

- Estimated M, actual M.
