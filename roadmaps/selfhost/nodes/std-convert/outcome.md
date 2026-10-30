# Outcome: std-convert

## Built

- New std module `std/char` (`compiler/internal/narra/std_char.go`): `Char.code(c): int`, `Char.fromCode(n): Option<char>` (None for surrogates, negatives, above 0x10FFFF). Chosen as a module, like `Int` and `Float`: `import * as Char from "std/char"`.
- `std/int`: `Int.parse(s, base): Result<int, string>` (bases 2-36, optional `-`, no prefix; errors: "empty string", "no digits after '-'", "invalid digit for the base", "number does not fit in 64 bits", "base must be between 2 and 36"), `Int.toHex(n)`. Runtime C is `runtimeInt` in `std_int.go`.
- `std/string`: `Str.endsWith`, `Str.contains`, `Str.replace` (empty `from` returns a copy of `s`), `Str.join`. Runtime C in `runtime_string.h`.
- Bug fix: `lstr_from_char` (`runtime.h`) builds from the encoded length; `Str.fromChar('\0')` has length 1 and `appendChar` keeps NUL.
- Test: `compiler/testdata/convert.nra` (+ `.out`), run in all modes (ASan/UBSan, leak check).
- README: supported-list line.

## Deviations

- None to the brief. Files touched outside owns/shared: none (`std_string.go`, `std_int.go`, `runtime.h`, `runtime_string.h`, `README.md` are shared).
- The NUL test is in `convert.nra`, not `std_string_test.go`.

## Stubs left

- None.

## Gotchas

- `print` takes only `string`; tests use `if (...) { print(...) }` for bools.
- `Int.parse` of the `-9223372036854775808` literal text works; the literal itself in source is written as `-9223372036854775807 - 1`.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass.

## Size

- Estimated M, actual M.
