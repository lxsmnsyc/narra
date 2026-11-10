# Outcome: copy-builtin

## Built

- **`copy(x)`, a prelude function.** Row in `selfhost/src/stdprelude.nra` (`name: "copy", module: "", key: "copy", check: "checkCopy"`). The check, the emitter and the helpers are in the new module `selfhost/src/stdcopy.nra`:
  - `checkCopyCall` is called from `checkNamedCall` in `selfhost/src/calls.nra`, inside the prelude branch (`if (name == "copy")`). It does not go through `checkStdCall`: the fix needs the call's node id (`facts.callEnds`). It sets `Ref.Builtin(<copy row>)` on the call node and returns the argument's type. The argument is checked with the call's expected type.
  - E0975 reports a wrong argument count (`copy takes one argument, found N`), `copy(mutable x)`, and an argument whose type is never changed in place (`copy of <type> copies nothing: ...; remove the call`), with the fix "remove the copy call", offered when the call is on one line and written `copy(`. A primary argument (name, literal, field, index, call, array) loses `copy(` and `)`; any other argument keeps its parentheses (`copy(n + 1) * 2` becomes `(n + 1) * 2`). The untracked-type error is skipped while an instance of a generic function is checked (`checker.environment` binds a type parameter to a concrete type), because the generic body was already checked with `T`, where `copy` is allowed. The three formats are at the end of the table in `selfhost/src/diagnostics.nra`.
  - `isCopyTracked(checker, type)` decides which types `copy` accepts: every struct (all-scalar ones too), `Array`, `Box`, `Map`, `StringBuilder`, generators and coroutines, type parameters, a tuple with a tracked element, and a distinct type over a tracked base. Strings, bytes, scalars, enums (with `Option` and `Result`) and function values are rejected. `FixedArray` does not exist in the compiler yet. When it does, add it here.
  - **How a later pass sees a `copy` result as unshared:** `isCopyCall(checker, expression)` in `stdcopy.nra` returns true for a call whose node resolved to the prelude `copy` row (`getRef(checker, nodeId) is Ref.Builtin(i)` with `module == ""` and `key == "copy"`). It reads the side tables of the current body, so call it while that body is being checked (shared-write-check). Nothing else marks the result. No new field was added to `Checker`, and nothing calls `isCopyCall` yet: the brief's "the checker treats the result as unshared" is met only once shared-write-check uses it.
  - `emitCopyCall` (called from `emitPreludeCall` in `selfhost/src/emitexprs.nra`) emits `NARRA_COPY_GRANT(true);`, an owned temporary holding the value, its retain, and `NARRA_COPY_GRANT(false);`. A type without counted parts (`needsRc` false, an all-scalar struct) emits the value itself.
- **The debug copy check (runtime).** All of it is in `selfhost/runtime/runtime.h`, after `narra_free`, inside `#ifndef NARRA_RELEASE`:
  - Credit set: `narra_copy_slots`, an open-addressing table `{pointer, credits}` in plain `calloc` memory (not counted by `narra_live`). `narra_copy_credit`, `narra_copy_forget` (backward-shift delete), `narra_copy_find`.
  - Switch: `narra_copy_on`, set by `narra_copy_start()` from `NARRA_CHECK_COPIES`. Without the variable the hooks do no table work: one flag test per retain and one counter test per decrement.
  - Hooks (macros; under `NARRA_RELEASE` they expand to `((void)0)`):
    - `NARRA_COPY_SHARED(p)` after each retain. While granting is on, it gives `p` one credit.
    - `NARRA_COPY_DROP(p)` before each decrement of `p->rc`. When the count is about to fall to 1 or 0, it removes `p`'s credits. This also covers frees, so a reused address never inherits credits.
    - `NARRA_COPY_WRITE(p, kind)` at each copy on write. It uses one credit of `p` and turns granting on, so the children the copy retains each get one. Without a credit it counts an unexplained copy for `kind` (0 array, 1 box, 2 map, 3 string builder, 4 generator). `NARRA_COPY_WRITTEN()` turns granting off after the copy.
    - `NARRA_COPY_GRANT(on)` around the retain that `copy(x)` emits.
    - `NARRA_COPY_START()` and `NARRA_COPY_REPORT()`, emitted in `main` (`selfhost/src/emit.nra`, `emitMain`) and in the test child (`selfhost/runtime/test_runner.c`). The report comes after `narra_consts_release` and before the leak check.
  - Hook sites: array retain, release and the out-of-line `_copy` (the inline `_unique` only tests `rc == 1`), and box retain, release and `_unique`, in `selfhost/src/emitdecls.nra` (`arrayFunctions`, `boxFunctions`); map retain, release and `_unique` in `selfhost/runtime/runtime_map.h`; string builder retain, release and the copy in `narra_sb_push` in `selfhost/runtime/std_string.c`; generator retain, release, `narra_gen_next` and `narra_co_next` in `runtime.h`. Strings, closure environments and enum payloads have no hooks. They are not tracked.
  - Report at exit, on stderr, with exit code 103: `copy check: copies on write that no copy(...) explains: array 2, map 1`.
  - The generated `selfhost/src/runtimecore.nra`, `runtimemap.nra`, `runtimestdstring.nra` and `runtimetestrunner.nra` were regenerated.
- **Turning the report on in test suites.** `selfhost/tests/src/copies.nra` (module `testrunner/copies`) holds `isCopySample(testCase)` (true for `copy_*.nra`), `COPY_CHECK_ENVIRONMENT` (`NARRA_CHECK_COPIES=1`) and `checkCopyRun`. `programs.nra` runs copy samples through `Common.runNarracWith(context, "", [Copies.COPY_CHECK_ENVIRONMENT], arguments)`. To turn the check on for every program (shared-write-error), make `isCopySample` return true, or add `COPY_CHECK_ENVIRONMENT` to `Shell.leakCheckEnvironment()` to cover every suite. Expect reports from existing programs until the shared-write rule and sweep land. A sample `<name>.copies` holds the expected report line: the debug run must then exit with code 103 and print that line. For each copy sample, the release variant also builds with `-emit-c` and checks that the C has no `narra_copy` text after preprocessing with `-DNARRA_RELEASE`, and that it does have it without the flag.
- **Samples.** `testdata/copy_nested.nra` covers struct with nested arrays and a map, arrays of arrays, an all-scalar struct, a `Box`, a `StringBuilder`, a generator, and a parameter named `copy`. The check is on and reports nothing. `testdata/copy_report.nra` with `.copies` has two hidden copies: one plain shared write, and one write after the `copy(...)` that gave the credit was dropped. `testdata/errors/copy_argument_count.nra`, `copy_untracked.nra`, `copy_mutable_argument.nra`. `testdata/copy_generic.nra` (a generic `copy` used with `int` and `Array`). `testdata/diagnostics/copy_untracked_fix.nra` and `copy_binary_fix.nra` (the habits suite applies the fix and the result checks clean). `testdata/fmt/copy_call.nra` and `.fmt`. `testdata/astdump/copy_call.nra` and `.ast`.
- **Hints.** The 13 hints that said `` copy it first with `const copy = X` `` (brief said ten) now say `` `copy(X)` ``. The two E0290 hints now say `` copy it into a local first with `copy(X)` ``. Report sites are in `exprs.nra` and `funcvalues.nra`, formats in `diagnostics.nra`. The `.err` and typesdump goldens, the `// error:` first lines and the guide's error lines (chapters 9 and 11) are updated.
- **A local named `copy`:** allowed, as for `panic`. `checkNamedCall` looks up locals and constants before the prelude, so a local or parameter `copy` hides the builtin inside its scope. A top-level function named `copy` is never called, because the prelude is checked before module functions, the same as for `panic` today. Recorded in SPEC.md section 6, "Copies on shared values".
- **SPEC.md:** "Copies on shared values" now describes `copy(x)`, E0975, the name rule and the debug check, with a checked sample. `@unique`, `--report-copies` and `NARRA_TRACE_COPIES` are removed, and so is the bindings line about `@unique`. The prelude list adds `copy`. The two `const copy = ...` mentions now use `copy(...)` or another name. The `@unique` line in `selfhost/tests/spec-skip.txt` is removed. The guide never mentioned the removed flags.
- **DECISIONS.md D94:** its status line says what is built.

## Deviations

- The brief asks for "a fix that drops the call" for an untracked type. The same E0975 also covers `copy(mutable x)`, which has no fix. The brief's only paths are E0975 and the hint rewrite, so this stays in scope.
- Thirteen hints were rewritten, not ten. The constant hint (E0290) depends on the write: a write to a part of a constant whose type `isCopyTracked` accepts suggests `copy(X)`; a scalar constant, a constant whose type is not known yet, or a whole reassign suggests a `let` local (`let limit = LIMIT`, `let startingKit = STARTING_KIT`), built by `buildConstantHint` in `stdcopy.nra`. E0290 has a second format row for it.
- Files edited outside the brief's owned paths, all within its shared paths: `selfhost/runtime/{runtime.h,runtime_map.h,std_string.c,test_runner.c}`, `selfhost/src/{calls,diagnostics,emit,emitdecls,emitexprs,exprs,funcvalues,stdprelude}.nra`, the new `selfhost/src/stdcopy.nra`, the 4 generated runtime `.nra` files, `selfhost/tests/src/{programs,copies}.nra`, `selfhost/tests/spec-skip.txt`, `testdata/errors/*.err` and some `// error:` lines, `testdata/typesdump/*`, `testdata/diagnostics/copy_untracked_fix.nra`, `testdata/astdump/copy_call.*`, `docs/guide/09-*.md`, `docs/guide/11-*.md`, `SPEC.md`, `DECISIONS.md`.

## Stubs left

- None. The report is off by default in the test suites until shared-write-error, as the brief asks.

## Gotchas

- The seed is not regenerated. `bootstrap.sh` prints "seed/narrac.c is older than the source". The compiler must not use `copy(...)` until the orchestrator regenerates the seed. Runtime text changed too, so this branch needs the two seed updates the GOAL describes.
- `runtimecore.nra` is one generated line. Any other branch that edits `runtime.h` conflicts there. Resolve by merging `runtime.h` and rerunning `sh scripts/test.sh --suite generated --update`.
- Rebased on array-write-cold-paths: the array hooks sit in the `NARRA_NOINLINE` `_copy` path, so the inline `rc == 1` fast path is unchanged. Keep the four hook lines (`NARRA_COPY_WRITE` before the copy, `NARRA_COPY_WRITTEN(); NARRA_COPY_DROP(a);` before `a->rc--`, `NARRA_COPY_SHARED` in retain, `NARRA_COPY_DROP` before the release decrement) wherever that code moves. Any new copy-on-write or decrement site for a tracked type needs the same hooks.
- O18 shows in practice: a child given a credit by a credited copy on write keeps it until its count falls to 1. A later hidden copy of that child goes unreported, at most once per credit.
- Sharing that never goes through a retain is invisible to the check: borrowed values (Perceus borrow optimizations) and inline all-scalar structs. The compile-time rule (shared-write-check) covers those.
- In `narrac test` builds the report runs inside each test child, so a test with an unexplained copy exits with code 103 instead of 0.

## Proposed learnings

- Every reference-count decrement and copy on write of a tracked type carries a `NARRA_COPY_*` hook | otherwise the debug copy check misses copies or keeps stale credits | selfhost/runtime, selfhost/src/emitdecls.nra | yes (grep for `rc--` without a preceding `NARRA_COPY_DROP`)

## Verification

- `NARRAC=out/new/narrac sh scripts/test.sh --suite errors --suite dumps --suite habits --suite codes --suite guide --suite spec --suite sources --suite generated`: PASS.
- `NARRAC=out/new/narrac sh scripts/test.sh --suite programs --filter copy_`: 4 cases pass, debug and release.
- `sh scripts/test.sh` (full, built from the seed): PASS in 234 s, 0 failed; again after the rebase on feat/batch-5 e13504a5: PASS in 243 s, 0 failed. After the review fixes, rebased on feat/batch-5 61199a64: PASS in 221 s, 0 failed.
- `sh selfhost/bootstrap.sh out/new`: stage 2 and stage 3 emit identical C. The note says the seed is older than the source (expected).

## Size

- Estimated L, actual L.
