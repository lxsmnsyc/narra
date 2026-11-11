# Outcome: build-conditions

## Built

- `std/build` (`std/src/build.nra`, Narra source): `export enum Os { MacOS, Linux, Windows }`, `export enum Arch { Arm64, X86_64 }` (both `derives Show, Equatable`), and `export const os`, `arch`, `debug` with stand-in values (`Os.Linux`, `Arch.X86_64`, `true`). Registered like std/time: `selfhost/src/stdbuild.nra` (`BUILD_FUNCS`), one `registerAll` line in `stdlib.nra`, the regenerated `stdtext.nra`, and `"Build": "std/build"` in `STD_ALIAS_PATHS` (`fixes.nra`).
- The target table, `selfhost/src/buildtarget.nra` (no compiler imports, so `check.nra` can import it):
  - `struct BuildTarget { os: string, arch: string, debug: bool, unsupportedHost: string, hostError: string = "" }`. `os` and `arch` are variant names of std/build's `Os` and `Arch`. `os == ""` means no target: the package does not import std/build, the caller has no io (dumps, `narrac query`), the host is unsupported (`unsupportedHost` then holds the `uname -s -m` text or the override), or the host could not be read (`hostError` holds the reason).
  - `TARGETS: Array<TargetRow>` with `TargetRow { os, arch, system, machine }`: MacOS/Arm64 (Darwin arm64), MacOS/X86_64, Linux/X86_64, Linux/Arm64 (Linux aarch64). This is the one table cross-targets extends (add columns such as a C triple; add rows).
  - `createUnknownTarget(debug)`, `detectHostTarget(debug)` (reads `NARRA_HOST_TARGET=<os>-<arch>` when set, else runs `/usr/bin/uname -s -m` and matches `system machine`), `describeTargets()`.
- Folding and plumbing, `selfhost/src/buildconditions.nra`:
  - `chooseTarget(package, release): BuildTarget uses io` runs `detectHostTarget(!release)` only when a module of the package is `std/build`; otherwise `createUnknownTarget(!release)`.
  - `applyBuildTarget(module, target): SourceModule` replaces the three stand-in values in the std/build AST (same node ids, only the variant name or bool changes). `checkModules` calls it on each module before `declareSourceModule`, so the checker, effects and emitter see the target's values with no special case (`Build.os` emits as the real variant).
  - `reportUnsupportedHost(checker, package)`: E0989 once per module, at its first `import ... from "std/build"`: "std/build has no target for this host (...)" or "could not determine the host: <reason>".
  - `recordBuildCondition(checker, condition)`, called in `checkIf` (`stmts.nra`) right after the condition is checked. A condition built only from `Build.os`/`Build.arch`/`Build.debug` (by `Ref.Const` to a std/build const, through an alias or an `import { os }` name), `Os.X`/`Arch.X` constructors (`Ref.Constructor` of a std/build enum), `==`, `!=`, `!`, `&&`, `||` and `is` with a fieldless constructor pattern, and reading at least one constant, is recorded. Anything else (a runtime value anywhere, `match`) is an ordinary `if`.
- Where things live on the Checker (`check.nra`):
  - `Checker.buildTarget: BuildTarget`: set by `checkModules(package, skipped, target)` from its new third parameter. `newChecker` starts it as `createUnknownTarget(true)`.
  - `Checker.foldedConditions: Map<int, bool>`: key is the condition's expression id (`getExpressionId(condition)` of the `Stmt.IfStmt`), value true when the then branch is kept, false when the else branch (or nothing) is kept. Ids are package-global, so the map is not per body. An `else if` is its own IfStmt with its own entry. An `if` expression is a BlockExpr holding an IfStmt, so it folds the same way.
- Emitter: `emitIf` (`emitstmts.nra`) looks up `emitter.checker.foldedConditions[getExpressionId(condition)]` first; on a hit `emitKeptBranch` writes the kept block inside `{ }` (or nothing when false with no else) and never evaluates the condition. To tell which branch was dropped, any pass reads the same map: `Some(true)` drops `orElse`, `Some(false)` drops `then`, `None` keeps both.
- Entry points: `Body.checkPackageFor(package, target)` (new), `Body.checkPackage(package)` (unchanged, unknown target), `checkModules(package, skipped, target)`. The driver's `loadAndCheck(path, release)` and `compileToC(..., release)` pass `-release` through (`build`, `run`, `test`); `check` and `apply-fix` use debug. `narrac emit` (`emitter.nra`) and the language server use a debug host target. The server reads the host once per process (`CheckCache.hostTarget`, `chooseServerTarget` in `lspcache.nra`); each CLI command reads it at most once.
- Effects of every branch count (the effects pass already walks both branches); recorded in SPEC section 12 and D95.
- SPEC section 12 rewritten for what is built (enums, `-release`, host as target, the operators that fold, mixed conditions, effects), its sample fixed to compile (`import { Os }`, `Env.get` returns an Option) and removed from `spec-skip.txt`; `std/build` added to the standard module list. D95 says what is built. compiler/README.md: one Supported bullet, one Difference (host only; `@target` and `--all-targets` are build-targets).
- E0989 `std/build has no target for this host (%s); the supported targets are %s`.
- Samples: `testdata/build_os.nra` (golden per host: `build_os.darwin.out`, `build_os.linux.out`), `testdata/build_debug.nra` (`build_debug.out` debug, `build_debug.release.out` release), `testdata/errors/build_conditions_other_branch.nra` (type error in the Windows branch). Emit suite: `build-conditions/one-branch` (the C of build_os.nra holds only the host's branch string), `build-conditions/unsupported-host` (`NARRA_HOST_TARGET=Plan9-mips` gives E0989 exactly once for a module with two std/build imports).

## Deviations

- Files outside `shared:`, kept minimal:
  - `std/src/build.nra` (new std source module; the brief asks for a std module and std sources live there).
  - `selfhost/src/body.nra`: `checkModules` takes the target, sets `checker.buildTarget`, calls `reportUnsupportedHost` and `applyBuildTarget`; new `checkPackageFor`. The checker is created there, so the target cannot be set elsewhere without an io effect in the checker.
  - `selfhost/src/stdlib.nra`: one import and one `registerAll` line (the std module rule in selfhost/CLAUDE.md).
  - `selfhost/src/buildconditions.nra`, `buildtarget.nra`, `stdbuild.nra`: new modules of this node.
  - `selfhost/tests/src/programs.nra`: `<name>.release.out` and `<name>.<darwin|linux>.out` goldens (needed for the two-run debug/release sample and the per-host golden). `selfhost/tests/src/emitchecks.nra`: the two emit cases. `selfhost/tests/spec-skip.txt`: the section 12 sample now passes, so its skip line is removed (the spec suite fails otherwise).
  - `testdata/errors/unknown_module.err` and `testdata/declsdump/imports/err_unknown_module.dump`: the list of std modules in that message now includes std/build.
- Host detection spawns `/usr/bin/uname` (no io-free way exists that the seed knows; a runtime function would need two seed updates). It runs only when the package imports std/build, once per process. `NARRA_HOST_TARGET` overrides it (SPEC section 12).
- `is` tests of a fieldless variant fold too: SPEC says an `if` that tests only the constants folds, and `Build.os is Os.MacOS` tests only them.
- No guide chapter: guide samples are run and compared, and build output is host- or mode-dependent. SPEC/D95/README carry the docs.

## Stubs left

- `@target` declarations and `narra check --all-targets`: build-targets.
- No `--target` flag; the host is the only target (cross-targets / build-targets).

## Gotchas

- Dumps (`typesdump`, `declsdump`), `narrac query` and lspnavigation still call `Body.checkPackage` with no target: std/build keeps its stand-in values there and nothing folds. Give them `chooseTarget` if a dump must show folding.
- The dropped branch is still walked by every checker pass and by emitter pre-scans (closure collection, hoist scans, tail scans). Closures and generic instances that only a dropped branch uses are still emitted as C (valid but dead); tree-shake should skip `foldedConditions` branches when it marks what is reachable.
- `emitKeptBranch` does not evaluate the condition; a build condition has no side effects by construction (only constants and constructors).
- The "could not determine the host" wording has no test: it needs `/usr/bin/uname` missing or failing.
- `narrac emit` writes debug C; release C comes from `narrac build -release -emit-c` (SPEC section 12).

## Proposed learnings

- Replace a std source module's stand-in constants in its AST before declaring it, keeping node ids | the checker and emitter then need no special case for target values | selfhost/src | no

## Verification

- `NARRAC=<stage1> sh scripts/test.sh --suite <s>` for emit, errors, codes, spec, generated, sources, programs (--filter build): pass.
- Probes under `-sanitize`: closures in both branches, a generator and a generic function with folded ifs; `Env.get` in a dropped Windows branch of an exported function without `uses io` is an effect error.
- `sh scripts/test.sh`: PASS (386 s, and 486 s after the review fixes; bootstrap fixed point holds; seed older than source, as expected mid-batch).

## Size

- Estimated M, actual M.

Landed on feat/batch-7 as b6e88093.
