# Outcome: bytes-type

## Built

- Type `bytes`: `Type.TBytes` (`types.nra`), scope name `bytes` (`newScope` in `check.nra`), `typeString`, `typeKey`, `typeId` `bytes`, C type `lstr*` (`cnames.nra`), `stdType` name `bytes` (`stdtypes.nra`).
- Reference counting works as for `string` (`needsRc` and `refCountCode` in `emitstate.nra`). Traits: `bytes` is a primitive, so `Equatable`, `Hashable`, `Ordered` and `Show` come from the primitive rule in `traits.nra`. `emittraits.nra` maps eq, hash and cmp to `lstr_eq`, `lstr_hash` and `lstr_cmp`, and show to `narra_show_bytes` in `runtime.h` (`bytes[68 69 0a]`, or `bytes[]` when empty). `==`, `!=`, `<`, `<=`, `>` and `>=` work on two `bytes` values (`checkBinary` in `exprs.nra`, and `emitexprs.nra`). Map keys work without a `mapKeyProblem` case.
- `std/bytes` (`selfhost/src/stdbytes.nra`, runtime `selfhost/runtime/std_bytes.c`, generated `runtimestdbytes.nra`):
  - `fromString` shares the buffer.
  - `toString` returns `Result<string, string>`. Ok shares the buffer; Err reads `invalid UTF-8 at byte N`.
  - `length`.
  - `at` returns a `u8` and panics with `Bytes.at index out of range`.
  - `slice` copies and panics with `Bytes.slice range out of bounds`.
  - `concat`, `fromArray(Array<u8>)` and `toArray`.
- `Fs.readBytes(path): Result<bytes, string>` and `Fs.writeBytes(path, data): Result<int, string>` (`stdfs.nra`; in `std_fs.c`, `narra_fs_read_all` takes a UTF-8 flag).
- Tests:
  - `testdata/bytes.nra` and `.out`. The sample covers every function, invalid UTF-8 at two offsets, a binary file written and read back in `$TMPDIR`, `readText` refusing that file, a missing file, `bytes` as a map key, a derived struct with a `bytes` field, and `Show`, `Ordered` and `Hashable`. It is clean under ASan and the leak check (the programs suite runs it in debug and release; a run by hand with `-fsanitize=address,undefined` and `NARRA_CHECK_LEAKS=1` exits 0).
  - `testdata/errors/bytes_index`: `Bytes.at` with a `string` index and with a `u8` index (E0119).
  - `testdata/errors/bytes_string`: a `string` where `bytes` is expected, and the reverse (E0119).
- Docs: the SPEC "Standard modules" list has `std/bytes`. SPEC section 4 has a new "Bytes" subsection. `compiler/README.md` "Supported" has a `bytes` line.

## Acceptance

- Sample `bytes.nra` and `.out`: done, clean under ASan and the leak check.
- Error tests: done, E0119 for both. No new format was needed, so E0942 is not used.
- The SPEC "Standard modules" list has `std/bytes`: done.
- `@unique let buffer: Array<u8> = Arrays.withCapacity(65536)` in SPEC section 6: not done. `Arrays.withCapacity` has not landed (std-spec-functions), so the line is unchanged. Once that function exists the line needs nothing else: `Array<u8>` and `@unique` already work.
- README updated: done. "Go and selfhost emit the same C" no longer applies after the Go retirement.
- Suites: programs, errors, dumps, codes, sources, generated, std-tests, emit, query and bootstrap all PASS.

## New codes

- None. E0942 (reserved) is still free.

## Gate decisions

1. `bytes` is a primitive variant `Type.TBytes` with the representation of `string`. Alternatives: an opaque distinct type over `string` in a source module, or a struct that wraps `Array<u8>`. The primitive choice gets the std traits and map keys from the existing primitive rules, and it shares the buffer with `string` at no cost, as the brief asks.
2. `Fs.writeBytes` returns `Result<int, string>` (the number of bytes written), as `Fs.writeText` does. The alternative was the brief table's `Result<bool, string>`. The brief also says "as readText and writeText", and one result shape for both writers is simpler.
3. `Ordered` on `bytes` also allows the operators `<`, `<=`, `>` and `>=`, as on `string`. The alternative was `Ordered.compare` only. The two types share a representation, so they follow the same rule.
4. `narra_show_bytes` is in `runtime.h`, not in `std_bytes.c`, because a program can hold `bytes` from `Fs.readBytes` without importing std/bytes.
5. No DECISIONS entry was written (no number is reserved).

## For dependents

- `Shape` (shape-trait) does not cover `bytes`: `shapeImplements` in `traits.nra` returns false for it. A later node can add it, for example as base64 in `Value.Text` or as a list of ints.
- `Io.readBytes` (std/io) still returns `Result<string, string>`. A later node may change it to return `bytes`.
- `bytes` cannot be a literal pattern (E0709) or the type of a `mutable` parameter (it is not a contents type).
- `Bytes.fromString` and `Bytes.toString` share the buffer: both retain the same `lstr`. String literals stay static (negative count).
- The seed is not regenerated. `selfhost/src` must not use `bytes` until it is.
- `STD_ALIAS_PATHS` (`fixes.nra`) has no `Bytes` alias, so there is no add-import fix for it yet. The sized modules are not listed there either.

## Stubs left

- None.

## Size

- Estimated M, actual M.
