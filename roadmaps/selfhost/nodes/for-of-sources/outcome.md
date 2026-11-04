# Outcome: for-of-sources

## Built

- `for ... of` accepts any `Range` value, any `Array<T>` value and `Str.chars(s)` (an `Array<char>`), in the Go compiler and in `selfhost/`. The source is evaluated once; an array source is owned by the loop until the statement ends.
- `Range` is now a type name and a value: C struct `narra_range { low, high }`, defined in the `std/range` module runtime. `Range.upTo`/`Range.between` outside a loop head build it. A direct call in the head keeps its old C.
- `Arrays.values` / `Arrays.entries` outside a loop head: error E0920 (`std_array.go`, `selfhost/src/stdcalls.nra`), position of the call's `(`.
- E0232 text lists the accepted sources.
- Samples: `compiler/testdata/for_of_sources.nra` and `.out`. Errors: `compiler/testdata/errors/for_of_sources_values.nra`, `for_of_sources_int.nra`.
- `compiler/README.md` updated.

## Deviations

- Outside the brief's paths: `compiler/internal/narra/check.go` (`forSource` field), `check_types.go` (`Range` type name), `cnames.go`, `types.go` (`TRange` singleton, C name), `selfhost/src/check.nra`, `resolve.nra`, `cnames.nra`, `emitexprs.nra` (same changes), `selfhost/src/runtimetext.nra` (regenerated), `compiler/README.md`.
- `compiler/testdata/typesdump/stdcalls/arrays_maps.nra` and `.dump`: two `Arrays.length(Arrays.values([x]))` calls became `Arrays.length([x])`, because they are now E0920.

## Stubs left

- `Maps.keys` / `Maps.entries` outside a loop head still hit the emitter (unchanged; they are arrays in the checker).
- `Range` has no `==`, `Show` or hashing.

## Gotchas

- A program that names `Range` in a signature needs `std/range` imported somewhere in the program, because the C type comes with that module's runtime.
- Selfhost field defaults must be constants, so the loop source is kept as `forSourceLine`/`forSourceColumn`.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (full suite 258 s, includes `TestSelfhostBootstrap`).

## Size

- Estimated M, actual M.
