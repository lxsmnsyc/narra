# Outcome: std-foundation

## Built

None of the three steps existed on main (f7fa23e).

- Generic mismatch errors (E0325), both compilers. `genericCall` (`compiler/internal/narra/check_call.go`) and `checkGenericCall` (`selfhost/src/calls.nra`) call `shapeMismatch` after unification. An argument whose type cannot fit a parameter that still holds an open type parameter gets `expected Option<T>, found int; the argument does not fit the generic parameter` at the argument. Only positions that hold an open parameter are compared; the rest stays with `expectType`. After a mismatch the call reports no "cannot tell type parameter" and returns its result type when that is known, so `Options.isSome(3)` reports one error. "cannot tell type parameter" stays when nothing pins T. No Option-specific case.
- E0325 has its own format (with the suffix) because E0119 owns `expected %s, found %s` and a code is found from the format string.
- `std/internal/`. `isInternalModule` (`builtins.go`); the import rule in `declareModule` (`check.go`) and `importModules` (`scopes.nra`): a module whose path does not start with `std/` importing `std/internal/<name>` gets E0426 `std/internal/<name> is internal to the standard library`. `moduleList` (both compilers) does not name them. One test-only primitive, `std/internal/probe` `code` (same code as `Char.code`), in `std_internal.go`, `stdlib.nra` and `emitstd.nra`.
- Stage 2 tests: `compiler/internal/narra/stage2_test.go` has `TestStage2Programs` (emit with the cached stage 2 binary, compile with the sanitizer flags, run with the leak check, compare `.out`), `TestStage2Errors`, `TestStage2OutputComparison` (the comparison fails on a difference) and `TestStdInternalFromStdModule` (a std-path module imports the primitive and checks). Samples: `testdata/stage2/generic_option.nra`, `testdata/stage2/errors/{std_internal_import,generic_mismatch}.nra`.
- Error tests: `std_option_andthen_type` and `std_option_not_option` name the mismatch again; new `generic_mismatch_array`, `generic_mismatch_unpinned`, `std_internal_user_import`.
- From the std-source review: removed `std/option` from the `emitOptionCall` condition in `emitexprs.nra`; `Options.isSome(3)` reports only the mismatch.
- SPEC.md (std section) and `compiler/README.md` (Tests) describe `std/internal/` and the selfhost-only modules with their test path.

## Deviations

- Files outside owned and shared paths: `compiler/internal/narra/builtins.go`, `check.go` and `std_internal.go` (new), `selfhost/src/builtins.nra`, `emitstd.nra`, `emitexprs.nra`, `stdlib.nra` (the import rule, the probe primitive, the moduleList filter, the dead Option dispatch), `compiler/internal/narra/selfhost_imports_test.go` (`TestStdModuleOrder` skips internal modules, which `moduleList` does not name), and the regenerated typesdump goldens if the full run changed any.
- Not done: sorting std source modules with `orderModules` in single-file packages (a note from the std-source review). No second source module imports another yet; do it with the first one that does.
- "A std source module importing it compiles" is covered by a Go test with a synthetic `std/probeuser` module, because no shipped std source module needs a primitive yet.

## Stubs left

- `std/internal/probe` is a test-only primitive. The first real primitive module can replace it.

## Gotchas

- A new error format needs its own code even when its text starts like another one: the code is found from the format string.
- `shapeMismatch` ignores positions without open type parameters, so a callback with a wrong parameter type still reports through `expectType` ("expected (int) => string, found (string) => string").

## Proposed learnings

- A new `std/internal/` module is listed in `stdlib.nra` after `std/fs` order rules like any module (Go registration order is file name order) | `TestSelfhostBuiltins` compares the order | compiler/internal/narra, selfhost/src | yes

## Verification

- `cd compiler && go vet ./... && go test -count=1 -timeout 40m ./...`: pass (508 s under load), including TestSelfhostBootstrap and TestStage2*. `typesdump/stdcalls/err_options.dump` was regenerated (new mismatch texts); a function argument of the wrong arity keeps its old "cannot tell" error because the arrow already reports the arity.
- Targeted: `go test -run 'TestCompileErrors|TestStage2|TestStdInternal|TestStdModuleOrder|TestEveryFormatHasCode'` passes.

## Size

- Estimated M, actual M.
