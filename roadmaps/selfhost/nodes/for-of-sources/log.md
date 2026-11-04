# Log: for-of-sources

- Nothing existed yet. `Str.chars` returns `Array<char>` (README, SPEC calls it a generator), so an Array source covers it.
- A `Range` had no runtime value (C type `void`) and no type name. It is now the type `Range`: the C struct `narra_range { low, high }`, defined in `runtime.h` (always emitted, so a program may name `Range` without importing `std/range`), with a `TRange` singleton in Go. `Range.upTo` and `Range.between` got emitters that build the struct.
- A direct `Range.upTo(n)` / `Range.between(a, b)` loop head keeps its old C, so earlier output is unchanged.
- Array source: any `Array<T>` expression that is not an `Array.values`/`Array.entries`/`Map.*` call goes through the existing array loop (the loop owns the array until the statement ends).
- `Arrays.values` and `Arrays.entries` outside a loop head: error E0920. The checker holds the loop iterable (`Checker.forSource` in Go; line and column in `forSourceLine`/`forSourceColumn` in selfhost, since a field default must be a constant). The error position is the call's `(`.
- E0232 text now reads "for ... of needs a Range, an Array, Str.chars, Array.entries, Map.keys or Map.entries, found %s".
- `runtimetext.nra` regenerated with `NARRA_UPDATE=1 go test -run TestSelfhostRuntimeText`.
- Review fixes: the loop source is saved and restored around nested loops; a Range map key is a compile error (SPEC line 379 describes a user opaque struct, so Range keeps no == or hash); outcome corrected.
