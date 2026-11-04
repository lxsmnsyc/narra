# Outcome: for-destructure

## Built

- Approach: parser desugaring, as the brief prefers. `for (const <pattern> of xs) body` becomes `for (const __for_<line>_<column> of xs) { const <pattern> = __for_<line>_<column>; body }`. The hidden name uses the line and column of the pattern, so it is unique per loop and identical in both compilers. A `let` head (pattern or plain name) becomes a `let` declaration the same way. No `ForStmt`, checker or emitter change was needed for destructuring.
- `compiler/internal/narra/parser.go` and `selfhost/src/parser.nra`: the for head accepts `const` or `let`, then a name or the same patterns a declaration accepts.
- `Arrays.entries(items)` yields `#[int, T]` from 0 in a for-of: `std_array.go`, `check_stmt.go`, `check.go` (`forArrayEntries`), `emit_stmt.go`; selfhost `stdlib.nra`, `stdcalls.nra`, `stmts.nra`, `emitstmts.nra`.
- Samples: `compiler/testdata/for_destructure.nra` and `.out`. Errors: `compiler/testdata/errors/for_destructure_arity.nra`, `for_destructure_field.nra`.
- `compiler/README.md` updated. SPEC.md already names `Arrays.entries`, so it is unchanged.

## Deviations

- None. `diagnostic_codes.go` text for E0232 still lists only `Array.values`; it is not in the brief's paths.

## Stubs left

- `Arrays.entries` outside a for-of has no emitter (same as `Arrays.values`); node for-of-sources covers it.

## Gotchas

- The loop variable shows up in C as `v___for_L_C_N`. Error messages never name it, because the pattern errors come from the declaration.
- Two heads on one line and column cannot exist, so names cannot clash.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: see the PR.

## Size

- Estimated M, actual S.
