# Outcome: std-digest

## Built

`std/digest` in `std/src/digest.nra`: SHA-256, SHA-1, HMAC and PBKDF2-HMAC-SHA-256 in Narra with `u32` wrapping arithmetic, no C library. Import it as `import * as Digest from "std/digest"`.

## Public API

```narra fragment
export enum Algorithm { Sha256, Sha1 }
export struct Hasher { algorithm: Algorithm, state: Array<u32>, pending: Array<u8>, schedule: Array<u32>, length: int }
export function create(algorithm: Algorithm): Hasher uses alloc
export function update(mutable hasher: Hasher, data: bytes) uses alloc
export function finish(hasher: Hasher): bytes uses alloc
export function sha256(data: bytes): bytes uses alloc
export function sha1(data: bytes): bytes uses alloc
export function hmac(algorithm: Algorithm, key: bytes, message: bytes): bytes uses alloc
export function pbkdf2Sha256(password: bytes, salt: bytes, iterations: int, length: int): Option<bytes> uses alloc
export function equal(first: bytes, second: bytes): bool
```

- `finish` does not change the hasher; it can take more input afterwards. `copy(hasher)` forks it.
- `pbkdf2Sha256` gives `None` for `iterations < 1`, a negative `length`, or `length` above 32 * (2^32 - 1).
- `equal` visits every byte; different lengths are not equal.
- Digests are `bytes`; print them with `Bytes.toHex`.

## Files

- `std/src/digest.nra`, `testdata/digest.nra` and `.out`
- `selfhost/src/std/stddigest.nra` (registry row), one import and one `registerAll` line in `stdlib.nra`
- `scripts/std/generate.nra` (ModuleInfo); generated `selfhost/src/std/stdtext.nra`, `docs/reference/std/digest.md`, `docs/reference/std/README.md`
- Goldens listing std modules: `testdata/errors/{async_io_internal,exports_std_ext,std_internal_user_import,std_spec_ext_import,terminal_internal_import,unicode_internal_import,unknown_module}.err`, `testdata/stage2/errors/std_internal_import.err`, `testdata/declsdump/imports/err_unknown_module.dump`

## Tests

`testdata/digest.nra`: NIST vectors (empty, `abc`, 448-bit, 896-bit, one million `a`) for both hashes, one call and in uneven parts (1, 3, 63, 64, 65, 7, 129, 2, 100 bytes); RFC 4231 cases 1 to 4, 6, 7; RFC 2202 case 2 (SHA-1); RFC 7914 section 11 PBKDF2 (c=1 and c=80000); bad arguments; `equal`.

## Benchmark

64 MiB of random data, `narrac build -release`, hashing with `Digest.sha256` after `Fs.readBytes`: 0.43 s user, 0.74 s wall (includes reading the file), load average about 5.7 (the machine was shared). `shasum -a 256` on the same file: 0.26 s wall. Both print `9b21b7e6872038ee032ce74a4f6d029154ae35a3b04eaef14b0a2d03c24ff952`.

`Hasher` is an opaque struct. `pbkdf2Sha256` allocates on every iteration, so it is slow for large iteration counts (80000 iterations take about a second in the test).

## Deviations

- Touched, as every new std module does: `stddigest.nra`, `stdlib.nra`, `scripts/std/generate.nra`, the module-list goldens above.
- The brief mentions `Digest.hex`; it is not added, `Bytes.toHex` does that.
- The brief says to start from `sha256` in `scripts/unicode/generate.nra`; that one works on `int` with masks and a whole string, so this module is new code on `u32`. The seed is not touched and that script keeps its own copy.

## Gotchas

- Std text is embedded in the compiler: regenerate stdtext and rebuild with `selfhost/bootstrap.sh` before a sample sees a std change.
- `--update` does not create a missing `.out`; write it with `narrac run`.
- An unused loop variable in std source is an error; name it `_index`.
- A template part must be a `string`; a `bool` needs a conversion.
