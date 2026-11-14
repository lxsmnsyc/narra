# Outcome: feature-review

## Built

- `roadmaps/selfhost/debate/feature-review/debate.md`: the proposals for the maintainer, with evidence, cost, value and a recommendation each.
- `roadmaps/selfhost/debate/feature-review/ledger.json` (with `sealed.json`): 16 judge facts, 25 proposals, 40 opponent claims, a ruling on each.
- `roadmaps/selfhost/debate/feature-review/evidence/`: `count-code.sh` and its output `code-counts.txt` (code only: comments dropped, string and template literals blanked), and `count.sh` with `usage-counts.txt` (raw matches, which also count diagnostic texts and generated files). The rulings do not use it.

Surviving proposals (7 of the 25 considered, plus one opponent alternative):

| # | Kind | Proposal | Recommendation | Cost |
| --- | --- | --- | --- | --- |
| 1 | remove | `Coroutine<In, Out>` and the two-argument `Co.resume` | remove | about -100 compiler lines, E0197 and 3 yield-input errors, one runtime field, two seed updates |
| 2 | remove | `lazy` top-level values and `std/lazy` | remove | about -260 lines, E1032-E1035 |
| 3 | remove | bare block expressions and `if` as a value | remove | about -40 parser lines, E0933 changes |
| 4 | adopt | field shorthand `Point { x, y }` | adopt (reverses a D26 item) | about +50 lines; usable in the compiler after a seed update |
| 5 | adopt | integer values in template holes | adopt if the maintainer rules it is not an implicit conversion | about +40 lines |
| 6 | simplify | SPEC 11-12: monomorphize only, drop layout tables, `@specialize`, `compact`, `.nrai`, TCC | adopt; drop node layout-tables | SPEC text only |
| 7 | remove | `@inline`, after a benchmark | spike first | about -830 lines if dropped |
| 8 | adopt | a descending range without `Option` in `std/range` | adopt (std) | about +20 std lines |

Rejected (18 proposals, 2 opponent alternatives): listed one line each in `debate.md`.

## Proof

Not a fix node.

## Result

Not an experiment node. What was measured, at 6c9c740e (`feat/batch-13`):

- Corpus: `selfhost/src` 171 files, 81,580 lines (73,311 code lines); `std/src` 21 files, 1,772 lines. Testdata 1,747 `.nra` files.
- Features with no code use in `selfhost/src` or `std/src`: `lazy`, `typeof`, type aliases, labeled loops, `if` as a value, bare block expressions, `requires`/`ensures`, `Dynamic`, `Coroutine`, `@specialize`, match guards. `@inline` has 3 uses (`std/option`), `@const` 1 (`Template.raw`).
- Repeated forms in `selfhost/src`: 512 `name: name` fields, 359 `${Int.toString(...)}` of 4,009 holes, 405 `x++`/`x--` against 3 `+= 1`, 416 `Ranges.upTo/between` against 45 C-style loops, 1,314 `Arrays.length`, 96 `is None) {` and 116 `is Err(` early exits.
- Compiler lines per feature: `typeof` 211 (resolve.nra), `lazy` about 260, `Dynamic` 514, `@inline` 829, `ensures`-only about 121.
- `Co.resume` is chosen by argument count (`calls.nra:715`), the one overload in the language. Every generator frame has a coroutine input field (`runtime.h:416`).
- `compiler/README.md:224,250,253`: no layout tables, no `.nrai` files. Every build monomorphizes.

## Deviations

- Short debate (one round). The judge wrote the facts and proposals. A fresh opponent agent wrote the attacks from the ledger and the repo only. The brief's caps held: 25 proposals, 7 survive, plus one opponent alternative.
- D119 is not written: the maintainer decides first, as the brief's step 3 says.
- No `log.md`: the ledger holds the record.

## Stubs left

- None.

## Gotchas

- Raw greps over `selfhost/src` count diagnostic texts and the generated `runtime*.nra` and `stdtext.nra`. Four opponent claims used such counts (`@inline` 29, `distinct type` 47, `lazy` 1, guards 2) and were rejected. Use `evidence/count-code.sh` or the compiler's own parse for feature counts.
- Proposal 4's context-typed `{ x }` works only together with proposal 3; `Point { x }` works without it.
- Proposal 1 edits `selfhost/runtime/`, so its batch needs two seed updates.
- Two questions go to the maintainer: whether integer holes count as an implicit conversion (O11), and whether `narrac fmt` rewrites `x: x` to `x` (oxfmt does not).

## Proposed learnings

- Count feature uses with string literals blanked and generated files excluded | the compiler's diagnostics and generated std text name every feature | roadmaps/selfhost/debate | no

## Verification

- `python3 <goal-debate>/scripts/debate.py roadmaps/selfhost/debate/feature-review check`: ok. `status`: settle (rulings: 54 upheld, 26 rejected, 1 open).
- No build or test ran (review node, load average 30 to 41).

## Size

- Estimated M, actual M.
