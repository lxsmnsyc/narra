# Outcome: build-targets

## Built

- Syntax: `@target(os = Os.X, arch = Arch.Y)` on its own line before a top-level function (either part, or both, each once). The lexer has `@` as punctuation (`SYMBOL_AT = 49`). The parser reads the attribute in `parseDecl` (`parseTargetAttribute`, `addTargetAttribute`) and stores it in the new `FuncDecl.target: Option<TargetAttribute> = None` (`ast.nra`: `TargetAttribute { pos, arguments }`, `TargetArgument { pos, name, typeName, typePos, variant }`; `typeName` keeps a qualified form such as `Build.Os`). The declaration's trivia record starts at `@`; the attribute has its own record (key `@<line>:<column>`) that runs up to the declaration keyword, so a comment after it on its line, or on lines of its own before `function`, is a dangling comment of the attribute, and `printTargetAttribute` prints it there (sample `testdata/fmt/comment_target_attribute.nra`). astdump prints a `.target` slot (`attribute`, `arg name= type= variant=`); fmt prints the attribute on its own line (`printTargetAttribute`).
- Checking (`buildconditions.nra`, `buildtarget.nra`):
  - `resolveFunctionTarget` (called by `declareFunc`) resolves each argument with `resolveType` to std/build's `Os` or `Arch` and checks the variant, then stores a `TargetRequirement { os, arch }` in the new `Checker.functionTargets` (by function index). Errors are E0990.
  - `recordBuildCondition(checker, condition, then, orElse)` also pushes the source range of the dropped branch (then block, or the whole else chain) to the new `Checker.droppedRanges: Array<DroppedRange { scope, start, end }>`.
  - `checkTargetUse(checker, pos, funcIndex)`, called at the top of `checkFunctionCall` (calls.nra) and `checkFunctionValue` (funcvalues.nra), reports E0991 when the target is known, the function has a target the current one does not meet, the use is not inside a dropped range of the same module, and the enclosing function does not have a `@target` that this target lacks.
  - `buildtarget.nra`: `TargetRequirement`, `DroppedRange`, `listTargets(debug)`, `isTargetMet`, `describeRequirement`.
- What counts as a guaranteeing branch (SPEC section 12): the check runs per target, and a target does not compile a branch its build conditions drop or the body of a `@target` function it does not have. Over all targets that is "inside a branch whose build condition holds only where the function exists, or inside a `@target` function with the same or narrower target". Ordinary `if`, `match`, an early `return` and `Build.debug` alone count for nothing.
- `narrac check --all-targets` (`driver.nra`: `checkForTarget`, `checkEveryTarget`): loads once, checks each `TARGETS` row as a debug and as a release build when the package imports std/build (else once), prints each diagnostic once, sorted by file, line and column. An E0991 found on several targets is one diagnostic whose message lists them (`MacOS-Arm64, MacOS-X86_64`; `(debug)` or `(release)` after a target that fails in one build only). Works with `--json`.
- Codes: E0990 (8 message rows: placement or repeat, unknown attribute, value not dotted, empty, unknown name, repeated name, unknown variant, wrong enum) and E0991 (with a `narrac explain` text).
- Docs: SPEC section 12 rewritten for `@target` (a new checked sample; the `extern` sample stays and is marked as waiting for C interop), D95 paragraph, compiler/README.md Supported bullet and Difference.
- Samples: `testdata/build_target.nra` (+ `.out`, uses in guaranteeing branches, a function value, a narrower `@target` calling a wider one, `Build.Os` qualified form). `testdata/build-targets/outside_branch.nra` (use outside: E0991 on both macOS targets) and `other_target.nra` (passes `check` with `NARRA_HOST_TARGET=Linux-X86_64`, fails `check --all-targets` only on Linux-Arm64), run by two new emit-suite cases. `testdata/errors/target_attribute_{arguments,const,twice,unknown}.nra`. `testdata/astdump/target_attribute.nra`.

## Deviations

- Files outside `shared:`, kept minimal:
  - `selfhost/src/calls.nra`, `selfhost/src/funcvalues.nra`: one import and one `checkTargetUse` line each. `setRef` (check.nra) has no source position, so the use check must sit where the position is known.
  - `selfhost/src/stmts.nra`: the `recordBuildCondition` call passes `then` and `orElse` (one line), so the dropped branch's source range is known.
  - `selfhost/tests/src/emitchecks.nra`: two cases (`build-targets/outside-branch`, `build-targets/other-target`) and `checkAllTargets`. E0991 names the target, so its output differs per host under the errors suite; `check --all-targets` gives the same output everywhere.
  - `testdata/lex/unexpected.*` and `bad_after_multibyte.*`: they used `@` as the unexpected character, which now lexes; they use `$`.
- `selfhost/tests/spec-skip.txt`: the SPEC `extern` sample keeps its skip line; its reason now names c-interop only (orchestrator's request).
- Rebased onto feat/batch-7 after contracts landed: `Checker` fields, the diagnostics table and the D95 text keep both features.

## Stubs left

- `@target` on `extern` functions, and linking C functions only for their target: c-interop.
- A `@target` function is still type-checked and emitted as C on every target (dead C: nothing the target compiles calls it). tree-shake can drop it.
- Only functions take `@target`; consts, types and impl functions do not.

## Gotchas

- Hover, references and definition in the language server do not resolve the names inside `@target(...)` arguments (`Os`, `Build.Os`, the variant): they are not expressions with node ids.
- `Checker.droppedRanges` is filled while bodies are checked, before the uses inside the branch are checked, so `checkTargetUse` must stay inside the body walk (it does). A pass that runs later and wants the same answer can read the ranges, or `foldedConditions`.
- Dumps, `narrac query` and the language server's unknown-target paths skip the check (target os is ""). The language server uses the host target, so it reports E0991 for the host only.
- A stage 1 compiler built without `-DNARRA_RELEASE` trips `NARRA_CHECK_COPIES` in `loadFile` (the test runner sets it); build test compilers with `-DNARRA_RELEASE`, as bootstrap.sh does.

## Proposed learnings

- A check that needs a source position goes where the checker has the position (the call or value check), not in `setRef` | setRef only has node ids | selfhost/src | no

## Verification

- Review fixes (after cbf5ea92): comments around `@target` stay in place; `--all-targets` checks debug and release (`testdata/build-targets/release_only.nra`, emit case `build-targets/release-only`); one E0991 per use; sorted output. The `_` fallbacks in `findStatementStart`/`findStatementEnd` stay (the match must be exhaustive); the unreachable return after `while (true)` is gone. Targeted suites fmt, fmt-stable, dumps, errors, emit, lsp, sources, codes, spec, trivia pass; `sh scripts/test.sh` PASS in 208 s.
- `NARRAC=<stage 2, -DNARRA_RELEASE> sh scripts/test.sh --suite <s>` for dumps, errors, codes, emit, lsp, sources, spec, programs, fmt, fmt-stable, trivia, generated, node-ids, query, habits: pass.
- Probes: `narrac run -sanitize testdata/build_target.nra`; `check` and `check --all-targets` (text and `--json`) on the build-targets samples; fmt of leading, trailing and between comments around the attribute (kept in place, stable).
- `sh scripts/test.sh`: PASS in 217 s before the rebase, and PASS in 478 s after the rebase onto contracts (stage 2 and 3 identical; seed older than source, as expected mid-batch). Targeted suites after the rebase: dumps, fmt, errors, emit, spec, codes, sources, trivia, programs pass.

## Size

- Estimated M, actual M.
