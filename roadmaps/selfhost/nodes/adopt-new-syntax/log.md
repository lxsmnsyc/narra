# Log: adopt-new-syntax

## 2026-11-09 measure (base origin/feat/batch-3 at 55140d6)

Scope: `selfhost/src` without the generated files (`runtime*.nra` except `runtimetext.nra`, `stdtext.nra`), `selfhost/tests/src`, `std/src`. Counted with a line script (`sweep.py measure`, kept outside the repo).

| Count | src | tests | std | total |
| --- | --- | --- | --- | --- |
| `x = x + 1` / `x = x - 1` lines | 331 | 51 | 28 | 410 |
| `x = x op y` lines (other) | 46 | 9 | 7 | 62 |
| of those, kept: `x op= (y)` would regroup (`x = x - a + b`) | 7 | 2 | 1 | 10 |
| `while` loops | 189 | 36 | 11 | 236 |
| neighbouring `match` arm pairs with identical bodies | 320 | 0 | 0 | 320 |
| C-style `for` / or-pattern arms already | 0 | 0 | 0 | 0 |

The brief's 404 and 457 were measured earlier; 457 seems to include the `+ 1` lines (410 + 52 = 462). The arm count is higher than 177 because pairs that bind the same names (`Expr.A(pos, _) => pos`, `Expr.B(_, pos) => pos`) are counted too.

`while` loops by first failing rule (before rewriting):

| Reason | Count |
| --- | --- |
| can be rewritten | 55 |
| no `let` of a counter right before the loop | 59 |
| last statement is not a step of that counter | 76 |
| counter used after the loop | 37 |
| `continue` targeting the loop | 8 |
| step spans lines | 1 |

## 2026-11-09 to 11-10 rewrite

- Passes per group, in order: compound (`x = x op y` to `op=`, `+ 1`/`- 1` to `++`/`--`), loops (`let` + `while` to C-style `for`), arms (neighbouring identical bodies to or-patterns), then `narrac fmt` on the group's files and `narrac emit selfhost` with the base compiler.
- Arms whose alternatives bind the same name with different types fail with E0968; those runs were left as they were: `getDeclPos` (trivia.nra), `getDeclarationName` and `isExported` (scopes.nra), one match in lspnavigation.nra. 26 pairs.
- A loop step that names a local declared in the body (`current = trimmed` in fmtstable.nra) does not compile as a `for` step; such loops stay `while` (4).
- Converted loops with `continue` in the body: one (unused.nra), where the `continue` is in a nested `while`.
- std/src: the compound pass compiles and every test passes, but the emitted C of 8 samples changes: compound assignment emits a pointer temporary for the place, the temporaries after it are renumbered, and the position of the division check in `std/src/math.nra:53` moves from column 27 to 5. Reverted; std/src is unchanged.
- Full `scripts/test.sh` after group 1 took 2635 s on this machine (load from elsewhere); later groups ran targeted suites.
