# Outcome: typesdump-targs-spacing

## Built

- `compiler/internal/narra/typesdump.go`: the `targs=` header joins `dumpTypeString` of each argument with `,`.
- `selfhost/src/typesdump.nra`: `Str.join(parts, ",")`.
- New case `compiler/testdata/typesdump/bodies/targs_types.nra` and `.dump` (`targs=#[int, string]` and `targs=#[int, string],int`).

## Deviations

- No existing golden line changed: none had a space inside a type argument. A new case was needed for the tuple golden, so every parity test that walks the bodies inputs compares one more input (compared +1, skipped unchanged).

## Stubs left

- None.

## Gotchas

- Tuple indexing `x[0]` does not compile; destructure instead.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass.

## Size

- Estimated S, actual S.
