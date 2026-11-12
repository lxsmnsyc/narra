# Outcome: checker-emitter-copies

## Built

- D104 in `DECISIONS.md`. Gap 1 is closed: a `const` or `let` (destructuring ones too) holds its value until its block ends. A write that would copy an array, map, `Box`, `StringBuilder` or generator the local still shares is now E0978, also after the local's last use. The message says `which X still holds until its block ends at line N` and offers the usual `copy(...)` fix. Writes that copy nothing are still accepted: a struct of numbers, or a value shared only through an enum. Code: `selfhost/src/sharedwrites.nra` (`scopeEnds`, `heldWrites`, `LiveReason.ScopeEnd`, `canWriteCopy`, `findPlaceType`, `holdsHeapValue`, `isHeapObject`, `collectScopedLocals`). No new diagnostic code.
- After the fresh review: an enum counts as holding what its payloads can hold (`const held = Some(bag.items)` keeps the array, so a push to `bag.items` is E0978). The bindings of an `is` or `match` on a value that is not a place, and a `for (const ...)` binding over one, hold that value until their block or arm ends. A pattern binding's subject is now the origin of its sharing, so destructuring errors offer the `copy(...)` fix. A local initialized with `copy(...)` holds its value like any other (second review: the credit ends with the original's part, so an exception hid copies); `copy_enum.nra`, `copy_enum_credit_kept.nra` and the SPEC sample now change a part of a live enum copy through `copy(...)` of the part, and `copy_enum.nra` adds a part taken out of a temporary inline-enum copy, which the copy's credit explains. Error tests `shared_write_scope_end_copy_result` and `_copy_replaced`. Found and documented: a part taken out of a temporary copy of a heap enum (`takeLeftRow(copy(tree))`) compiles and the copy check reports it (D104 item 4, SPEC known gaps). This added 52 sites in `selfhost/src`: 36 take the `copy(...)` fix, 10 wrap the subject of an `is` in `copy(...)`, and the rest copy one pushed value or push a literal directly. Error tests: `shared_write_scope_end_enum`, `_is`, `_match`.
- Gap 2 (shared elements from `Arrays.filled(row, n)` and `[row, row]`) stays documented in SPEC section 6. D104 says why: the sharing belongs to the array value and moves across functions, and nothing is inferred across functions. The one compiler site, `filledArray([], n)` in `query.nra`, now gives each node its own list.
- Found a third gap and documented it in SPEC section 6 and D104: storing a value that shares one parameter into a `mutable` parameter (`into.pairs = other.pairs` in `mergeInto`). It caused 1785 of the 1927 hidden copies. The site now uses `copy(...)`. No rule checks this pattern yet.
- Compiler fixes for the patterns the rule rejects (119 sites in `selfhost/src`, 2 in `selfhost/tests/src`):
  - `parse` moves the lexer's tokens into the parser inside `readTokens`.
  - `orderModules` builds its tables inside the walk.
  - The layout tables in `check.nra` are built in `decideHeapEnums` and `decideLayouts`.
  - `pushBlockContext` is new in `stmts.nra`.
  - Literals are pushed straight into their owner in `declare.nra`, `check.nra`, `traits.nra` and `instances.nra`.
  - About 75 sites take the compiler's `copy(...)` fix.
- `exitWithoutRelease` (`emitter.nra`) now returns normally when `NARRA_CHECK_COPIES` is set. Before, `emit` ended through `Process.exit`, which skipped the copy report. So `NARRA_CHECK_COPIES=1 narrac emit` exited 0 even with hidden copies.
- New `copies` suite with one case, `selfhost-no-hidden-copies` (`selfhost/tests/src/copies.nra`, registered in `registry.nra`). It builds a debug compiler from `selfhost/`, runs `emit selfhost` with `NARRA_CHECK_COPIES=1`, and fails on a nonzero exit. It takes about 7 s.
- Tests:
  - `testdata/errors/shared_write_scope_end{,_call,_payload}.nra`: a write after the last use, a `mutable` call, and a narrowed `Result` payload.
  - `testdata/shared_write_scope_end_ok.nra`: the writes that copy nothing stay accepted.
  - `testdata/copy_dead_name.*` is removed. It kept a hidden copy that is now a compile error, and `shared_write_scope_end.nra` replaces it.
- Docs: SPEC section 6 "When it ends" and the known gaps, the SPEC struct sample (`let current = copy(config)`), the `selfhost/CLAUDE.md` line about `let x = holder.field`, and the E0978 entry in `compiler/README.md`.

## Proof

- base (`feat/batch-10`, ececc85a), debug compiler built with `cc -O2 -std=c11 -w -o narrac-debug seed/narrac.c -lm`:
  - `NARRA_CHECK_COPIES=1 narrac-debug check testdata/arrays.nra`: exit 103, `array 3, map 1`.
  - `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1 narrac-debug emit selfhost`: exit 103, `array 956, map 971`. With `NARRA_CHECK_COPIES=1` alone it exits 0, because the report is skipped.
- branch: a stage 2 debug compiler built from the branch source (`narrac emit selfhost > c; cc -O1 -std=c11 -w c -lm`):
  - `NARRA_CHECK_COPIES=1 narrac-dbg emit selfhost`: exit 0, also with `NARRA_CHECK_LEAKS=1`. The emitted C is the same with and without the leak check.
  - `NARRA_CHECK_COPIES=1 narrac-dbg check testdata/arrays.nra`: exit 0.
- `narrac emit selfhost` instructions, release build (`-O2 -DNARRA_RELEASE`, `/usr/bin/time -l`, two runs each, load average about 20): before 2.463 G and 2.430 G, after 2.416 G and 2.425 G. That is within the noise. Peak memory went from 158 MB to 157 MB. Each compiler emits its own source.

## Deviations

- Files outside the owned and shared paths:
  - `selfhost/tests/src/registry.nra`: one import and one `Suite` entry for `copies`.
  - `selfhost/tests/src/files.nra` and `guide.nra`: two `copy(...)` fixes the new rule requires.
- The "reported once" dedupe of E0978 now skips a write when another error falls between the start and the end of the written place (`WriteSite.end`, `findPlaceEnd`), so `mutable_contents_destructure.err` keeps its single line. Two goldens (`shared_write_pattern.err`, `shared_write_scalar_pattern.err`) now offer the `copy(...)` fix where they said none is offered.

## Stubs left

- None.

## Gotchas

- The seed still has the old rule. `selfhost/src` compiles under both rules, but the stage 1 compiler is the first one to enforce D104. Regenerating the seed at the end of the batch makes the seed enforce it too.
- The rule is conservative across calls. A deep `mutable checker` call beside a local read out of `checker` is an error even when the callee never writes that part. Most of the new `copy(...)` calls are retains that never copy.
- Taking a value out of a narrowed `Result` and then changing it (`let names = listed.value; sortNames(mutable names)`) always copied. It is now an error, so code must write `copy(...)` or move the value inside a helper.
- If copy-check-holders lands first, the orchestrator should rerun the debug `emit selfhost` on the merged branch, as the batch brief says.

## Proposed learnings

- Build a value into its owner directly (push the literal, or fill it in a helper) instead of keeping a `const` that holds it while a `mutable` call follows | a local holds its value until its block ends, so every later write through the owner is E0978 or a copy | selfhost/src | yes (E0978)
- Do not end a debug run of a compiler command through `Process.exit` | it skips the leak and copy reports | selfhost/src | yes (copies suite)

## Verification

- `NARRAC=<branch narrac> sh scripts/test.sh` (all suites): one failure (a SPEC sample), fixed. Then `--suite spec`: PASS, 62 cases.
- `NARRAC=<branch narrac> sh scripts/test.sh --suite errors --update`: PASS, 606 cases.
- `NARRAC=<branch narrac> sh scripts/test.sh --suite copies`: PASS, 1 case, 7 s.
- `sh scripts/test.sh` from the seed, after the review fixes on 423966e4: PASS, 7172 cases in 175s.

## Size

- Estimated L, actual L. The first cut of the rule reported 173 sites. Narrowing it to writes that copy a heap object brought that to 119, and those took most of the time.
