# Outcome: adopt-new-syntax

## Built

- `selfhost/src` (without the generated `runtime*.nra` and `stdtext.nra`) and `selfhost/tests/src` use `x++`, `x--`, `x op= y`, C-style `for` and or-pattern arms. Six commits, one per group: front end and dumps, checker, emitter, std module checkers and emitters, fmt/lsp/query, test runner.
- `std/src` is unchanged (see Deviations).

| Count | Before | After |
| --- | --- | --- |
| `x = x + 1` / `x = x - 1` lines | 410 (src 331, tests 51, std 28) | 28 (std only) |
| other `x = x op y` lines | 62 | 15 (10 would regroup, 5 in std) |
| `while` loops | 236 | 185 |
| C-style `for` loops | 0 | 51 (src 48, tests 3) |
| neighbouring `match` arm pairs with identical bodies | 320 | 26 |
| arm pairs merged into or-patterns | 0 | 294 |

`x = x - a + b` style lines (10) stay, because `x -= a + b` would group the value differently.

`while` loops left on purpose, by the first rule each fails:

| Reason | Count |
| --- | --- |
| no `let` of the counter right before the loop | 59 |
| the last statement is not a step of that counter | 76 |
| the counter is used after the loop | 37 |
| a `continue` targets the loop | 5 |
| the step names a local declared in the body (would not compile as a step) | 4 |
| the step spans lines | 1 |
| in `std/src`, which is left unchanged | 3 |
| total | 185 |

Arm pairs left (26): the alternatives bind one name with different types (`Decl.Func(inner)`, `Decl.Const(inner)`, ...), which is E0968. They are in `getDeclPos` (`selfhost/src/trivia.nra`), `getDeclarationName` and `isExported` (`selfhost/src/scopes.nra`) and one match in `selfhost/src/lspnavigation.nra`.

## Proof

Not a fix node.

## Result

Not an experiment.

## Deviations

- `std/src` is not edited. Its 28 `++` and 5 `op=` candidates compile and pass every suite, but the emitted C of 8 samples changes (`derive_shape`, `json`, `json_errors`, `json_floats`, `json_rpc`, `shape`, `std_math`, `std_random`): compound assignment emits a pointer temporary for the place, so later temporaries are renumbered, and the division check in `std/src/math.nra:53` reports column 5 instead of 27. That is more than the line-number differences allowed, so the edit was reverted. The orchestrator decides whether to accept that C change; the rewrite is mechanical (one script pass and `--suite generated --update`).
- The brief's measured counts (404, 457, 177) differ from the ones above: 457 seems to include the `+ 1` lines, and 177 counted fewer arm pairs (pairs binding the same names, like `Expr.A(pos, _) => pos`, were merged too).

## Stubs left

- None.

## Gotchas

- A C-style `for` step cannot name a `let` or `const` declared in the loop body; the step is outside the body's scope.
- `narrac fmt` lays out a short `return match` on one line with commas: `return match (type) { Type.TU8 | Type.TU16 => true, _ => false }`.
- Compound assignment emits different C than `x = x op y` (a pointer temporary for the place), so rewriting code whose C is checked (std, samples) changes goldens.

## Proposed learnings

- Merge neighbouring arms only when every alternative binds each name with one type | E0968 rejects the rest, so a sweep must compile after merging | selfhost/src | yes (the compiler)
- Leave std/src syntax sweeps for a batch that may change sample C | std code is inlined into every sample that imports it | std/src | yes (sample C diff)

## Verification

- After each group: `narrac fmt` on its files, `narrac emit selfhost` with the base compiler, a two-stage self-build with a fixed-point check, and the suites the group touches (`sources`, `errors`, `dumps`, `emit`, `programs`, `panics`, `leaks`, `fmt`, `fmt-stable`, `fmt-doc`, `trivia`, `query`, `lsp`, `std-tests`, `runner`, `runner-options`, `codes`, `guide`, `spec`, `generated`, `habits`, `edge`, `node-ids`, `packages`, `manifest`): all pass.
- After group 1, a full `scripts/test.sh` with the branch compiler: PASS (8 shards, 0 failed).
- Sample C: `narrac emit` of the 137 samples in `testdata/`, `testdata/panics`, `testdata/leaks` and `testdata/std-tests` with the base compiler and the final branch compiler: byte-identical.
- `sh selfhost/bootstrap.sh`: stage 1 and stage 2 emit identical C (183198 lines); the seed is noted as older than the source, as expected.
- `sh scripts/test.sh` (from the seed): PASS (8 shards of 722 or 723 cases and the bootstrap job, 0 failed, 83 s).
- `git diff origin/feat/batch-3 -- '*.nra'` read with a filter. Every removed `x = x op y` line has a matching `++`/`--`/`op=` line or is a converted loop's step (46). The other changed lines are `let`/`while` headers, merged arms and fmt layout.

## Size

- Estimated L, actual L.
