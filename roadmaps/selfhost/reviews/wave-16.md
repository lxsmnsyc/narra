# Wave 16 report

## Merged

- #90 language-guide story: chapters 1 to 11 and `TestGuide`. Chapter 11 (guide-typescript) landed on the story branch as 56003e0 after a fresh review (two minors fixed: an error sample for `const this`, and a wrong `a && g(a)` row). The whole-guide review found one stale "Planned" line about generators, fixed in c574d57. The maintainer merged #90.
- #91 std-io-more, #92 closure-borrow, #94 unused-names: merged by the maintainer. #94 landed after #92 without the five renames listed below, so main failed `TestSourcesFormatted` (`emitexprs.nra:871: "local" is never read`); fixed in #95, which must merge before #93.

## Waiting for the maintainer

- #93 coroutines (D79): `Coroutine<In, Out>` and `Co.resume`. Known limit, in D79: passing an input to a plain generator reports the internal coroutine type. Rebased on main after #91, #92 and #94: the std module list, `resolve.nra`, and the borrow helpers that #92 moved from `stmts.nra` to `exprs.nra` (their `YieldStmt` pattern now has four fields).

## Notes on the merged pull requests

Each had a clean fresh review (or one whose findings were fixed and re-checked) and a passing full check. `plan.py can-merge` sent all four to the maintainer, because each changes files outside its node's paths.

- #91 std-io-more: stdin, directories, `std/path`. Review fixes: `Io` and `Path` in the add-import fix table, a plain C11 `fgetc` loop instead of POSIX `getline` (not yet run on Linux), out-of-memory aborts in `Path.normalize`.
- #92 closure-borrow (D77): non-escaping closures capture `let` and `mutable` by reference. The first review found a blocking use-after-free: builtin callbacks (`Results.map`, `mapErr`, `andThen`) skipped the borrow rule. Fixed and re-reviewed; the reviewer found no other path around the rule. Guide chapters 3, 9 and 11 rewritten for the two capture modes.
- #94 unused-names: unused locals and parameters are errors. A second review round found that the fix for a write-only shorthand pattern did not rename its assignments; fixed in 8661719, and the full check passes there.

## Decisions made for the maintainer (please confirm)

Taken while the maintainer was away, at their request to pick the best option. Each is recorded where it applies.

- unused-names: the `args` parameter of a binary's entry function is exempt (otherwise about 378 test files and every guide sample need `_args`). A `let` written but never read is unused; a `mutable` parameter that is only written counts as used (an output parameter); `x = x + 1` counts as a read. All in `compiler/README.md` "Differences from the spec". Unused imports and private functions (also in SPEC line 64) are a proposed follow-up node.
- coroutines: `Co.resume` charges `alloc` (it allocates the `Step`); non-test builds carry the panic-capture globals and one extra field per generator frame (D79 cost line).
- Refined outlines (commit 4671fa5), each with a "Decision to confirm" section and a gate: identity-cycles (D80: cycles leak, the leak check reports them, no collector), mutable-reference (D81: `mutable` stays second-class, adds `mutable` local bindings and `x = f(mutable x)`), sized-numbers (D82: no implicit conversion, `I32.from`, `tryFrom`, `wrap`), std-json (D83: `Value` in `std/shape` with separate `Int` and `Float`), shape-trait (D84: `Json.decode<User>(text)` instead of the spec's `Json.decode(text, User)`).
- coroutines brief: D79 records why coroutines are a separate type from generators (maintainer request).

## Plan changes

- The five outlines became 15 briefs: std-json, call-type-args, shape-trait, shape-derive; language-service, lsp-navigation, lsp-incremental; identity-cycles; mutable-reference, mutable-bindings; sized-numbers, sized-conversions, sized-f32, bytes-type; and std-tables-split, which splits the hot std table files once this wave lands.
- New node effects-indirect-calls (fix, should, S): calls through struct fields and call results do not charge their effects, in both checkers. From the inbox; wave 15 proposed it as must before closure-borrow and coroutines, but this wave started without it. Not started; it waits for the maintainer's go-ahead like any new node.

## Process lessons

- Every M node's first fresh review found something: a use-after-free (closure-borrow), a portability risk (std-io-more), leaking internal names in errors (coroutines), a too-broad exemption and an undocumented rule (unused-names). Re-reviewing a fix round caught a second bug in unused-names (the shorthand fix).
- Workers often fixed a last failing golden or guide sample after the full run and reran only the targeted tests. The repo has no CI, so the orchestrator ran the full check itself on #93 before the hand-off. Worker prompts should say: after any fix, rerun the full check, or say which tests were skipped.
- `can-merge` sends any change outside a node's paths to the maintainer. Goldens that list std modules and guide samples are touched by most language nodes, so nothing in this wave could auto-merge. Briefs for language changes should list `docs/guide/**` and the std-module golden as shared.
- The plain-writing hook limits commit subjects to 72 characters; titles from `pr_title` can exceed it.
- A merge-order note in this report did not stop a broken main: #92 and #94 had no text conflict, so GitHub merged both, and the five names #94's rule rejects in #92's files broke `TestSourcesFormatted`. When two ready pull requests interact without a text conflict, the orchestrator should rebase the second onto the first and fix it before the hand-off, or state the order in both pull request bodies.
