# Outcome: strings-std

## Built

`std/string` (`compiler/internal/narra/std_string.go`) gained these functions. Positions are byte offsets of type `int`.

- `Str.slice(s, from, to)`, `Str.sliceFrom(s, from)`: copy of the bytes. Panic when a position is out of range, `from > to`, or a position is inside a character.
- `Str.byteAt(s, i)`: `int` 0..255; panics out of range.
- `Str.charAt(s, i)`: `char` at a character start; panics out of range, at the end, or inside a character.
- `Str.nextChar(s, i)`: position after the character at `i`; same panics.
- `Str.indexOf(s, needle)`: `Option<int>`; an empty needle gives `Some(0)`.
- `Str.charCount`, `Str.startsWith`, `Str.trim` (ASCII whitespace).
- `Str.split(s, sep)`: `Array<string>`; panics on an empty `sep`.
- `Str.chars(s)`: `Array<char>`.
- `StringBuilder`: `Str.builder()`, `Str.append(mutable sb, s)`, `Str.appendChar(mutable sb, c)`, `Str.build(sb)`. The type name `StringBuilder` resolves in any module. It is a new prim type `TStringBuilder` (C type `narra_sb*`, reference counted, copy on write when shared), so `const copy = sb` then appending to `sb` leaves `copy` unchanged.

Runtime C: `compiler/internal/narra/runtime_string.h`, embedded and set as `Module.Runtime` of `std/string` (so every program importing `std/string` now carries it).

Tests: `compiler/testdata/strings2.nra` (+ `.out`) tokenizes a string into words and numbers and exercises every function; `compiler/testdata/panics/strings2_inside.nra` and `strings2_range.nra` run by `TestStringPanics` in `compiler/internal/narra/std_string_test.go`. Debug builds run under ASan/UBSan and the leak check.

## Deviations

- `ByteIndex` is `int`, not a distinct type (the brief allows this until `opaque-distinct`). Recorded in `compiler/README.md` under "Differences from the spec".
- Slices are copies, not views of the original buffer (spec says views). Same README entry. A view needs a parent pointer in `lstr`, which touches `runtime.h` and every release path.
- Files outside owned paths: `compiler/README.md` (docs), `types.go` (`TStringBuilder`, `needsRC`), `emit_types.go` (retain/release), `check_traits.go` (builder implements no standard trait), `modules_check.go` (type name map), `check_call.go` (see Gotchas). Each is a few lines.
- Panic programs live in `testdata/panics/` and run from a new `std_string_test.go`, so `run_test.go` is untouched.

## Stubs left

- `ByteIndex` as a distinct opaque type: replaced by `opaque-distinct`.
- `Str.find`, `Str.copy`, `Str.charCount` iterators and unicode functions of the spec are not built.

## Gotchas

- `check_call.go` `builtinCall` checked `mutable` arguments before checking the argument expression, so `checkMutableArg` saw an unresolved `Local` and always failed. It now checks the expression first. Any builtin with `Mutable` was unusable before.
- `StringBuilder` is a `PrimType`, so code that switches on `*PrimType` treats it as a primitive. Equality, `Show` and pattern literals on it are not supported; `implements` returns false for it.
- `Str.split` and `Str.chars` return arrays built in the emitter from runtime-allocated pieces (like `Fs.listDir`).

## Proposed learnings

- A builtin with a `Mutable` parameter is checked through the generic path only after the argument expression is checked | otherwise the root local is unresolved | compiler/internal/narra | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass.
- `go run ./cmd/narra run testdata/strings2.nra`: output matches `strings2.out`.

## Size

- Estimated M, actual M.
