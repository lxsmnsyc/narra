# Outcome: std-hex-base64

## Built

Hex and base64 text for `std/bytes`, as Narra source in `std/src/ext__bytes.nra` (the same way `ext__char` extends the builtin `std/char`). No compiler or runtime change. Regenerated `selfhost/src/std/stdtext.nra` and `docs/reference/std/bytes.md`. Sample: `testdata/bytes_encoding.nra` with its `.out` (RFC 4648 section 10 vectors both ways, all 256 byte values, bad input).

## Public functions

```narra fragment
export function toHex(data: bytes): string uses alloc
export function fromHex(text: string): Option<bytes> uses alloc
export function toBase64(data: bytes): string uses alloc
export function toBase64Url(data: bytes, padded: bool): string uses alloc
export function fromBase64(text: string): Option<bytes> uses alloc
export function fromBase64Url(text: string): Option<bytes> uses alloc
```

- `toHex` writes lowercase, two digits per byte. `fromHex` accepts upper and lower case; odd length or a non-hex character gives `None`; empty text gives empty bytes.
- `toBase64` uses the standard alphabet with `=` padding. `toBase64Url` uses `-` and `_`; `padded` chooses `=` padding.
- `fromBase64` requires exact padding. `fromBase64Url` accepts padded or unpadded text. Both reject other characters, misplaced padding, a lone trailing character, and nonzero leftover bits.

## Deviations

- The brief says "an alphabet flag"; the URL-safe alphabet is a separate function pair, with a `padded` flag on the encoder only.
- Decoders reject nonzero leftover bits (strict RFC 4648 section 3.5).

## Gotchas

- Narra rejects a variable shift count unless a local is checked against 0 and 63 first; the code copies the count into a local and checks it.
- After editing `std/src/ext__*.nra`, regenerate stdtext (`sh scripts/test.sh --suite generated --update`) before `selfhost/bootstrap.sh`, since the compiler embeds the text.
- The 16 hand-written hex copies in the compiler are not replaced; that waits for a seed that knows these functions.
