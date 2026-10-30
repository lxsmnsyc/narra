# Outcome: std-option-result

## Built

- `std/option` (`compiler/internal/narra/std_option.go`): `Option.map`, `andThen`, `withDefault`, `orElse`, `isSome`, `isNone`, `okOr`.
- `std/result` (`std_result.go`): `Result.map`, `mapErr`, `andThen`, `withDefault`, `ok`, `err`.
- Both register through `registerNamespaceModule`, so `import * as Option from "std/option"` works with the alias equal to the type name, and a call without the import gives E0908 with the add-import fix. The new modules appear in `stdModules` as ordinary modules; selfhost-check-imports should list `std/option` and `std/result` with these 13 functions.
- Callbacks: `Builtin.Callbacks []int` (new, `builtins.go`) lists argument indexes that are function values. The effect pass (`check_effects.go`, `effectWalker.call`) adds the effects of those arguments to the call, like a bound effect variable. `Effects` is set on every builtin: `alloc` for `Option.map`, `okOr`, `Result.map`, `mapErr`, `andThen`, `ok`, `err` (they build a payload); empty for the rest. `io` in a callback reaches the caller (test `errors/std_result_effect.nra`).
- `checkCallback` (std_option.go) mirrors `Array.map`'s check: arrow functions get the parameter types as a hint and `Stack = true`, a local function-typed parameter is allowed as non-escaping. Emit evaluates the arguments once into temporaries, calls `f.fn(f.env, payload)` in the taken branch, and builds the result with `e.enumAlloc`; values copied through are retained.
- Diagnostics E0909 to E0913 (`diagnostic_codes.go`): not an Option, not a Result, argument is not a function, callback returns the wrong enum, callback returns void.
- Tests: `testdata/std_option.nra/.out`, `std_result.nra/.out` (Some/None, Ok/Err, capturing and non-capturing callbacks, an `io` callback, a string and a struct payload); `testdata/errors/std_option_{wrong_callback,missing_import,andthen_type,not_option}.nra`, `std_result_{wrong_callback,andthen_error_type,effect}.nra`.
- Docs: SPEC.md (Optional values, section 5, standard modules list), DECISIONS.md D69, compiler/README.md.

Closure cost per call (debug build, `narra emit` of `std_option.nra`): a capturing arrow function passed to a helper compiles to a stack environment (`K_... k = {.h = {-1, NULL}, .c_v_k = ...}`), so the closure costs 0 heap allocations per call and no reference count traffic. The only allocation of a call is the result payload: 1 (`Some`/`Ok`/`Err`) for `map`, `mapErr`, `okOr`, `ok`, `err`, and for `andThen` on `Result` when the value passes an `Err` through; 0 for `withDefault`, `isSome`, `isNone`, the `Some` branch of `orElse`. A callback stored first (`const f = (x) => x + k`, then `Option.map(o, f)`) is a heap closure: 1 allocation once at creation and one retain/release pair per use, not per call of the helper. Non-capturing callbacks have `env = NULL`: 0.

Proposed SPEC.md text (gate; the full text is in the diff):

> **`std/option` helpers.** `import * as Option from "std/option"` adds `Option.map(x, f)`, `Option.andThen(x, f)` (`f` returns an `Option`), `Option.withDefault(x, fallback)`, `Option.orElse(x, f)` (`f` takes no argument and returns an `Option`), `Option.isSome(x)`, `Option.isNone(x)` and `Option.okOr(x, error)` (a `Result`). `withDefault` returns `fallback` when `x` is `None`. There is no `unwrap` or `expect`: read the value after an `is` test, or give a fallback.
>
> **Callbacks.** The function arguments of these helpers are non-escaping and take an effect variable, like `Array.map` in the effects section: a call uses the effects of its callback, so an `io` callback needs `uses io` on the caller. A capturing arrow function passed to a helper is allocated on the stack.
>
> **Helpers never return early.** They are expressions. The main style for a path that stops stays `is` and `match`.
>
> **`std/result` helpers** (section 5). `Result.map(r, f)`, `Result.mapErr(r, f)`, `Result.andThen(r, f)` (`f` returns a `Result` with the same error type), `Result.withDefault(r, fallback)`, and `Result.ok(r)` and `Result.err(r)`, which return the `Ok` or `Err` value as an `Option`. They never return from the caller, so they do not replace the visible `if` or `match` on the error path.

## Deviations

- Files edited outside the brief's owned and shared paths: `compiler/internal/narra/check_effects.go` (5 lines in `effectWalker.call` to add callback effects). No other way to flow a callback's effects into a builtin call.
- `builtins.go` gained the `Callbacks` field (shared file).
- Five diagnostic codes were taken (E0909 to E0913), the next free numbers in the E09xx range. If another branch appended the same numbers, renumber one side at merge.
- The implicit callback result for `Option.orElse` and an `andThen` result are typed from the callback; `Result.andThen` with `Err(...)` in the callback needs the arrow's return type written (`(x): Result<int, string> => ...`), because a constructor cannot infer the other type argument.

## Stubs left

- None.

## Gotchas

- Newlines inside call parentheses are ignored by the lexer, so an arrow with a block body passed inline cannot have several statements; define it as a `const` first.
- A generic function value passed as a callback needs its type arguments from context; the hint passed has no result type, so `Option.map(o, genericFn)` may report "cannot tell type parameter".
- `run` is the entry point name; sample files use `export function run(args: Array<string>): int`.
- Prelude enum detection is by name and module path of the `Option`/`Result` generic (`preludeEnumArg`); a user enum named `Option` is not accepted.

## Proposed learnings

- A std builtin that calls a function argument lists its index in `Builtin.Callbacks`, or its effects are lost | the effect pass only reads `Effects` otherwise | compiler/internal/narra | no
- Pass a builtin's arrow-function argument through `checkCallback` so it gets a stack environment | a heap closure costs an allocation per call | compiler/internal/narra | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (all samples under ASan, UBSan and the leak check in debug, and release).
- `narra run` output of both samples checked by hand against the expected values before the `.out` files were written.

## Size

- Estimated M, actual M.
