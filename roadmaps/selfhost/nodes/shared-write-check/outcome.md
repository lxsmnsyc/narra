# Outcome: shared-write-check

## Built

- **Warnings exist now.** Until this node the compiler only had errors.
  - `Checker.warnings: Array<CheckError>` and `reportWarningFix(mutable checker, pos, code, fixes, message)` in `selfhost/src/check.nra`. Warnings never touch `checker.errors`, so nothing that tests "no errors" changes.
  - `narrac check` (`runCheck`, `selfhost/src/driver.nra`) prints warnings only when the package has no errors: text `file:line:col: warning: message` on stderr, or with `--json` one object per warning with `"severity":"warning"` (`writeDiagnosticJson` in `diagnostics.nra` takes a `warning` flag). Exit code 0 when it found only warnings, 1 with errors, as before.
  - `narrac apply-fix` takes the warnings when there are no errors, so their fix is applied the usual way.
  - `build`, `run`, `test` and `emit` do not run the analysis and print nothing new.
- **The analysis: `selfhost/src/sharedwrites.nra`, warning E0978.** `reportSharedWrites(mutable checker)` runs once after the whole package is checked (called from `findWarnings` in `driver.nra`). It walks every function, implementation function and test body of a source module, with that body's side tables current; generic instances are skipped. One forward walk per body:
  - A *place* is a local or a top-level constant followed by fields and `[]` for any element (`L12.items[]`). A *pair* says two places hold the same value. Pairs come from `const`/`let` declarations, assignments, pattern bindings (`is`, `match`, destructuring), `for (const x of xs)` (x shares `xs[]`), struct, array and map literals (the part's path is kept: `{ items: a.items }` gives `n.items` with `a.items`), spreads (each copied field), enum payloads, calls whose callee has `from` (`findSharedArgument`), std `mutable` arguments (`Arrays.push(mutable list, item)` makes `item` an element of `list`), and arrow functions (a value made from an arrow shares its captures). `copy(...)` (`isCopyCall`) gives nothing. Chains are closed when a pair is added (`const b = a; const c = b` pairs all three; `const items = a.items` pairs with whatever shares `a`). Only a part whose type `canShareType` accepts can share, so strings, scalars and enums without tracked payloads never pair.
  - A write: an assignment or compound assignment hits a pair side that strictly contains the target; a `mutable` argument to a std builtin hits sides that contain or equal the container; any other `mutable` argument (and an arrow function's `refCaptures`, at the place the arrow is written) also hits sides inside the target when the shared value is `isCopyTracked`. A hit waits for the next use of the pair's other side (any read of an overlapping place, a write inside it, or a capture by an arrow function) and then warns once per write. The other side counts as live at once when its root is a parameter, a constant, a captured local seen from inside an arrow function, or a `Cell`. Every pair the write reaches ends.
  - Assigning a place whole (or declaring a local again) ends its pairs and the writes waiting on it. Assigning one element (`grid[0] = x`) is not a whole assignment: indexes are not compared, so the pairs and pending writes of the other elements (`grid[]`) stay. Branches start from the same state and merge; a loop body is walked again until no new pair or pending write appears (bounded by the number of locals in the body plus one, which a chain through every local needs). Each write has one site, keyed by its position and text, that every round reuses, so a round that adds no pair adds no pending write either and the walk stops; `break`, `continue` and `return` end a path; `return` inside a block expression gives its value.
  - Loops: `for (const x of xs)` over an array or a map, directly or through a std iterator (`Arrays.values(xs)`), pairs `x` with `xs[]`. `for (mutable item of xs)` makes `item` stand for `xs[]`, so a write through it writes the element of `xs`. A loop over `copy(xs)` shares and holds nothing. While the body of any `for ... of` loop other than `for (mutable ...)` runs, the iterator holds the iterated place: a write that copies it warns with `..., which the loop over rows still holds; change it after the loop, or loop over a copy(...)`.
  - Messages (six formats in `diagnostics.nra`), for example `startFunction(mutable emitter, ...) may change emitter.names, which info still uses at line 17; take a copy where it is shared, with copy(...)`, or `..., which items still holds for the caller; ...`, `which captured items still holds`, `which the cell counter still holds`, `which constant PRIMES still holds`, `which the loop over rows still holds`. When no fix is offered, the advice reads `no fix is offered here: copy the value by hand where the sharing starts, or end the sharing before the write`.
  - The fix (`buildCopyFix`) wraps the value of the statement that started the sharing in `copy(...)`, when the parser noted where it ends and `isCopyTracked` accepts its type. The parser now records `SourceFacts.valueEnds` for `const`/`let` values and assignment values too (`parser.nra`, two lines).
- **Small API changes.** `findExpressionStart` in `sharedresults.nra` is exported. The codes suite knows `reportWarningFix` as a reporter.
- **Test runner.** The errors suite reads the first line of a sample: `// warning:` means exit code 0 with the golden on stderr (an empty golden means the sample checks clean). The habits suite's clean check now also rejects warnings left in stderr, and its codes check accepts warning-only samples.
- **Samples.** `testdata/errors/shared_write_*.nra` with `.err`, one per bullet of the rule: `field` (write to a place while a name shares it), `reverse` (write through `items`, later use of `bag.items`), `chain`, `mutable_whole` (`mutable R`), `mutable_field` (`mutable R.f`, a sibling field is not reported), `closure` (non-escaping arrow that changes a captured `const`), `parameter`, `pattern` (destructuring and `for (const ...)`), `captured` (arrow value and a capture inside an arrow), `cell`, `call_result` (`from`), `for_const` and `for_mutable` (loops over an array written directly), `loop_held` (writing the iterated array inside a direct loop and an `Arrays.values` loop), `nested_loops` (three nested loops over 50 locals: a guard against a loop walk that never settles; it takes about 21 M instructions). `shared_write_clean` also keeps a loop over `copy(items)` that writes `items` clean. `shared_write_clean.nra` (empty golden) keeps clean: whole-place assignment, a `copy(...)` source, scalars and strings, a write after the sharing name's last use. `testdata/diagnostics/shared_write_fix.nra`: the debate's sample; the habits suite applies the fix and the result checks clean.
- **Docs.** DECISIONS.md D94 amendment (built as a warning, the three decisions below, the counts). `compiler/README.md` line.

### Warning counts (for sizing shared-write-sweep)

| Where | Warnings |
|---|---|
| `selfhost/src` (`narrac check selfhost`) | 198 (on feat/batch-6 cb2e90ee, with type-of and proven-operands) |
| `std/src` | 0 |
| `selfhost/tests/src` (`narrac check selfhost/tests`) | 3 (all `newCommand(...)` results, which are `from arguments`) |
| `testdata` (every program and package outside `errors/`) | 36, plus the new fix sample; they are the copy samples that keep a hidden copy on purpose (`copy_report`, `loop_unique_hoist`, `copy_hoist_*`, `mutable_for*`, `box_assign`, `arrays`, `maps`, ...); 2 of them are loop-held writes |

- Save/restore sites with a `mutable` call between save and reset (O9), where the later use is the restore `place = saved`: **4** (`resolve.nra:510`, `closures.nra:162`, `stmts.nra:1125` and `:1132`).
- `selfhost/src` by write: 177 `mutable` calls to compiler functions, 14 std `mutable` calls (`Arrays.push`, `Str.append`), 7 assignments. By liveness: 112 wait for a later use, 52 are live because the other side is a parameter (mostly a value taken out of a `mutable` parameter, changed, and put back), 34 are writes inside a `for ... of` loop over a place the write may copy (mostly `for (const x of Arrays.values(checker.<list>)) { reportError(mutable checker, ...) }`; these have no automatic fix).
- `selfhost/src` by area: checker modules 140, emitter 32, fmt/query/lsp/trivia 17, parser/modules/driver 9.
- About 8 come only from all-scalar structs (`Pos`), measured before the review fixes with a build that does not track them.
- The debate estimated 120 to 220 edits; 198 is inside that range. Most fixes are `copy(...)` around one `const x = checker.<path>` before a `mutable checker` call, which `narrac apply-fix` writes; the 34 loop-held sites need a hand edit (loop over a `copy(...)` or move the write).

### Decisions (brief's "Decide and record")

- **Std builtin rows need no `shared` column.** Arrays and maps have no reading builtins: `items[i]` and `table[key]` are places, which the walk follows. The builtins that hand out elements are the iterators, and `for ... of` over a std builtin call shares the elements of its arguments. `Arrays.filled`, `Maps.keys` and the rest count as fresh. A std builtin's `mutable` argument changes only the container (WRITE_CONTAINER), which removed about 30 false reports on elements of pushed-to arrays.
- **Reassigning a local whole ends its sharing.** The walk is flow-sensitive: an assignment to a place (or a new declaration of the local in a later loop round) drops every pair with a side at or under it, and every pending write that waited on it. `let list = items; list = []; list[0] = ...` is clean.
- **The analysis has its own walk; it does not reuse `sharedresults.nra`'s.** That walk is flow-insensitive and only follows values to returns. This one reuses `findSharedArgument` (call results), `canShareType` and `isCopyCall`, and the place model of `LoopBorrow` (`[]` for an element).

## Deviations

- **The analysis runs only in `narrac check` and `apply-fix`, after checking.** The brief says the warning "is on for every package, including selfhost (it does not fail the build)". It is on for every package, but builds do not compute or print it: computing it inside every body check made `narrac emit selfhost` cost 18% more instructions (1.78 G to 2.11 G). As a separate pass, emit stays at 1.75 G (base 1.78 G) and only `check selfhost` pays (1.26 G to 1.59 G, +27%).
- **A package with errors shows no warnings.** Error goldens stay unchanged, and the walk needs complete side tables.
- **The language server does not show E0978 yet** (`lspcache.nra` publishes `checkErrors` only). A later node can call `reportSharedWrites` there and publish `checker.warnings` with severity 2.
- **Std modules** imported by a package are walked when the module is a source module of the checker (that is how `std/src` counted 0). No std warning appears today.
- Files outside the owned paths, all in shared paths: `selfhost/src/{check,diagnostics,driver,parser,sharedresults}.nra`, `selfhost/tests/src/{codes,errors,habits}.nra`, `DECISIONS.md`, `compiler/README.md`, `testdata/diagnostics/shared_write_fix.nra` (the brief owns `testdata/diagnostics/shared_write_*`).

## Stubs left

- None. shared-write-sweep fixes the sites; shared-write-error turns E0978 into an error.

## Gotchas

- The rule is conservative where it cannot see: a `mutable` argument to a compiler function writes everything under it, so `const scope = checker.scopes[i]; reportError(mutable checker, ...); use(scope)` warns although `reportError` only pushes to `checker.errors`. That is D94 as decided (no inference across functions); it is most of the 153.
- Indexes are not compared: `items[i]` and `items[j]` are the same place `items[]`. An element assignment keeps the sharing of `items[]`. **Known limit:** reassigning the same index still warns: `let row = grid[0]; grid[0] = [9]; row[0] = 5; use(grid[0])` reports that the write to `row` may change `grid[0]`, although `grid[0]` no longer shares `row`. shared-write-error decides whether this false positive stays.
- A labeled `break` to an outer loop is treated as a break of the innermost loop (the state reaches a bit more code; only more warnings can result).
- Two values that share inside one name (`a.first = a.second`) are not tracked; the rule is about two names.
- A `mutable` argument to a user function does not make it keep the other arguments (only std builtins do). `f(mutable checker, scope)` followed by a write to `scope` is not reported.
- Map reads (`table[key]`) have an `Option` type, so a deep write to the map does not hit a name that holds `table[key]`'s value (the shared value is an enum). A write to `table[key].field` does.
- The loop bound is the number of locals plus one per loop, and an inner loop is walked again on each round of the outer one. Real bodies settle in two or three rounds.
- `Checker.warnings` is filled only by `reportSharedWrites`. shared-write-error should report through `reportErrorFix` instead and keep the samples (their first lines become `// error:`, exit 1).

## Proposed learnings

- New warning-only samples go in `testdata/errors/` with a `// warning:` first line | the errors suite then wants exit 0 and an `.err` golden of the warnings | testdata/errors | yes (the errors suite)
- A pass that only reports (no effect on emitted code) runs after checking, from the driver command that shows it | it keeps build cost flat | selfhost/src | no

## Verification

- `NARRAC=<built> sh scripts/test.sh --suite errors --suite habits --suite codes --suite sources --suite dumps --suite fmt --suite guide --suite spec --suite query`: PASS.
- `narrac check` on every sample: the goldens in `testdata/errors/shared_write_*.err`; `narrac apply-fix testdata/diagnostics/shared_write_fix.nra` gives `const info = copy(emitter.names)`, which checks clean.
- Counts: `narrac check selfhost` (173), `narrac check selfhost/tests` (3), every testdata program and package (33 lines, one is the fix sample); std/src with a throwaway build that walks every module (0).
- Instructions retired (`/usr/bin/time -l`, release builds from the seed): `narrac emit selfhost` base 1,776,654,940, branch 1,752,195,022; `narrac check selfhost` base 1,257,623,471, first commit 1,591,890,926, after the review fixes 1,612,936,091, and 1,672,709,514 after the rebase on feat/batch-6 cb2e90ee, which adds compiler code (the second review found 34.7 G while each loop round added new sites; fixed with stable sites).
- First commit: `sh scripts/test.sh` (full, built from the seed): PASS, 6082 cases, 0 failed.
- After the rebase on feat/batch-6 cb2e90ee: `sh scripts/test.sh` PASS, 6231 cases, 0 failed, with the bootstrap job.
- Review fixes (direct-array loops, element assignment, unbounded loop rounds, loop-held places, advice text): the reviewer's probes p1, t1, u2, u2b, r1, r2, q7 now warn and s5 says that no fix is offered. Targeted suites errors, habits, codes, sources: PASS. Counts above are after the fixes. `sh scripts/test.sh`: PASS, 6097 cases, 0 failed. `sh selfhost/bootstrap.sh`: stage 1 and stage 2 emit identical C; the seed is older than the source (expected).

## Size

- Estimated L, actual L.
