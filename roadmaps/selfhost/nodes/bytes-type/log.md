# Log: bytes-type

- 2026-11-09: branch feat/bytes-type from origin/feat/batch-1. E0942 is free in `diagnostics.nra` (E0940 and E0941 are used; the next used code is E0948).
- `Arrays.withCapacity` is not in `selfhost/src` or `std/src`, so std-spec-functions has not landed it. The SPEC section 6 `@unique let buffer` line is left alone.
- Built `bytes` as a primitive `Type.TBytes` with the C type `lstr*`, following sized-f32's list of primitive sites. Eq, hash and cmp reuse `lstr_eq`, `lstr_hash`, `lstr_cmp`. `narra_show_bytes` is new in `runtime.h` (core), since `Fs.readBytes` gives a `bytes` value without an import of std/bytes.
- `std/bytes` is in `stdbytes.nra`. Every row has a fixed signature, so `checkStdCall` needs no case. Runtime `std_bytes.c`, generated `runtimestdbytes.nra`, one case in `runtimetext.nra`.
- `narra_fs_read` became `narra_fs_read_all(path, utf8, ...)` with two wrappers. `Fs.writeBytes` reuses `narra_fs_write`.
- No new diagnostic format was needed: wrong argument types are E0119 from the fixed-signature check. E0942 stays unused.
- Golden updates: the unknown-module list now ends with std/bytes (`errors/unknown_module.err`, `declsdump/imports/err_unknown_module.dump`).
- Suites run with out/narrac: programs, errors, dumps, codes, sources, generated, std-tests, emit, query, bootstrap. All PASS.
