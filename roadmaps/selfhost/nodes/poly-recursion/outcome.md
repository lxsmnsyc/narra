# Outcome: poly-recursion

## Built

- `checkPolyRecursion` in `compiler/internal/narra/check_types.go`, called at the end of `buildGeneric`. It walks the generic declarations reachable from a generic type's fields and variants. If a path returns to the type with an argument that holds a type parameter without being one (`G<Array<T>>` inside `G<T>`), it reports an error at the declaration: "generic type G contains G<Array<T>>, so each instance needs a larger one without end; ... polymorphic recursion is not supported". It then drops the fields so no later instance grows.
- Works through other generic types (mutual case) and for `derives Show`.
- Tests: `compiler/testdata/errors/poly_recursion.nra`, `poly_recursion_derives.nra`, `poly_recursion_mutual.nra`.

## Deviations

- None. Only `check_types.go` (shared, listed in the brief) and owned test files changed.

## Stubs left

- None.

## Gotchas

- Recursion with the same arguments (`List<T>` holding `Array<List<T>>`) is allowed and must stay so.

## Proposed learnings

- Check generic type declarations for growing self-reference at declaration time, not instantiation | instantiation loops forever otherwise | compiler/internal/narra | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass.
- On main's `check_types.go`, the three new error tests hang (test timeout at 15 s); on the branch they pass in under a second.

## Size

- Estimated S, actual S.
